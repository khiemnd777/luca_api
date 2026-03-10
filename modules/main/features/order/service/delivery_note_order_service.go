package service

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	promotionengine "github.com/khiemnd777/andy_api/modules/main/features/promotion/engine"
	promotionrepo "github.com/khiemnd777/andy_api/modules/main/features/promotion/repository"
	promotionservice "github.com/khiemnd777/andy_api/modules/main/features/promotion/service"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/clinic"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/department"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/material"
	"github.com/khiemnd777/andy_api/shared/utils"
)

type DeliveryNotePrintRequest struct {
	OrderID            int64                           `json:"order_id"`
	Company            *DeliveryNoteCompany            `json:"company,omitempty"`
	Attachments        *DeliveryNoteAttachments        `json:"attachments,omitempty"`
	ImplantAccessories *DeliveryNoteImplantAccessories `json:"implant_accessories,omitempty"`
	PaymentMethod      *DeliveryNotePaymentMethod      `json:"payment_method,omitempty"`
}

func (s *orderService) GenerateDeliveryNoteByOrderID(ctx context.Context, req DeliveryNotePrintRequest) ([]byte, string, error) {
	if req.OrderID <= 0 {
		return nil, "", fmt.Errorf("invalid order_id")
	}

	orderDTO, err := s.repo.GetByID(ctx, req.OrderID)
	if err != nil {
		return nil, "", err
	}
	if orderDTO.LatestOrderItem == nil || orderDTO.LatestOrderItem.ID <= 0 {
		return nil, "", fmt.Errorf("latest order item not found")
	}

	products, err := s.repo.GetAllOrderProductsByOrderItemID(ctx, orderDTO.LatestOrderItem.ID)
	if err != nil {
		return nil, "", err
	}

	materials, err := s.repo.GetAllOrderMaterialsByOrderItemID(ctx, orderDTO.LatestOrderItem.ID)
	if err != nil {
		return nil, "", err
	}

	company := s.resolveDeliveryNoteCompany(ctx, orderDTO)
	shippingAddress := s.resolveDeliveryNoteShippingAddress(ctx, orderDTO)
	attachments := s.resolveDeliveryNoteAttachments(ctx, orderDTO, materials)
	implantAccessories := s.resolveDeliveryNoteImplantAccessories(ctx, orderDTO, materials)
	note, err := buildDeliveryNoteFromOrder(orderDTO, products, materials, company, shippingAddress, attachments, implantAccessories, req)
	if err != nil {
		return nil, "", err
	}

	discountAmount, finalAmount, err := s.calculateDeliveryNotePricingByPromotion(ctx, orderDTO, products)
	if err != nil {
		return nil, "", err
	}
	note.DiscountAmount = discountAmount
	note.FinalAmount = finalAmount

	deliveryQRSvc := NewOrderDeliveryQRService(s.deps.Ent.(*generated.Client), s.deps)
	rawToken, err := deliveryQRSvc.GenerateDeliveryQRToken(ctx, int(req.OrderID))
	if err != nil {
		if !errors.Is(err, model.ErrOrderAlreadyDelivered) {
			return nil, "", err
		}
	} else {
		note.QRCode = BuildDeliveryQRStartURL(s.deps.Config.DeliveryQR.ClientBaseURL, rawToken)
		note.QRCodeImageURL = BuildQRCodeImageURL(note.QRCode, 160)
	}

	if strings.TrimSpace(note.Company.LogoPath) != "" {
		logoData, err := ConvertImageToBase64(note.Company.LogoPath)
		if err != nil {
			return nil, "", fmt.Errorf("convert company logo to base64: %w", err)
		}
		note.Company.LogoData = template.URL(logoData)
	}

	pdf, err := GenerateDeliveryNotePDF(note)
	if err != nil {
		return nil, "", err
	}

	fileName := fmt.Sprintf("hoa-don-%s.pdf", strings.ReplaceAll(note.Order.Number, "/", "-"))
	return pdf, fileName, nil
}

func buildDeliveryNoteFromOrder(
	orderDTO *model.OrderDTO,
	products []*model.OrderItemProductDTO,
	materials []*model.OrderItemMaterialDTO,
	company DeliveryNoteCompany,
	shippingAddress string,
	attachments DeliveryNoteAttachments,
	implantAccessories DeliveryNoteImplantAccessories,
	req DeliveryNotePrintRequest,
) (DeliveryNote, error) {
	if orderDTO == nil {
		return DeliveryNote{}, fmt.Errorf("order not found")
	}

	note := DeliveryNote{
		Company: company,
		Order: DeliveryNoteOrder{
			Number:          firstNonEmpty(utils.DerefString(orderDTO.Code), utils.DerefString(orderDTO.Code)),
			BS:              utils.DerefString(orderDTO.DentistName),
			BN:              utils.DerefString(orderDTO.PatientName),
			ClinicName:      utils.DerefString(orderDTO.ClinicName),
			ShippingAddress: shippingAddress,
			Date:            pickOrderDate(orderDTO),
		},
		Attachments:        attachments,
		ImplantAccessories: implantAccessories,
	}

	if req.Company != nil {
		// Keep department as source of truth, fallback to request for missing values only.
		if strings.TrimSpace(note.Company.Name) == "" {
			note.Company.Name = req.Company.Name
		}
		if strings.TrimSpace(note.Company.LogoPath) == "" {
			note.Company.LogoPath = req.Company.LogoPath
		}
		if strings.TrimSpace(string(note.Company.LogoData)) == "" {
			note.Company.LogoData = req.Company.LogoData
		}
		if strings.TrimSpace(note.Company.Address) == "" {
			note.Company.Address = req.Company.Address
		}
		if strings.TrimSpace(note.Company.Phone) == "" {
			note.Company.Phone = req.Company.Phone
		}
	}

	note.PaymentMethod = DeliveryNotePaymentMethod{
		TienMat: false,
		CongNo:  false,
	}
	if orderDTO.LatestOrderItem != nil {
		// Source of truth:
		// - Tiền mặt  <- order_items.is_cash
		// - Công nợ   <- order_items.is_credit
		note.PaymentMethod.TienMat = orderDTO.LatestOrderItem.IsCash
		note.PaymentMethod.CongNo = orderDTO.LatestOrderItem.IsCredit
	}

	items := make([]DeliveryNoteItem, 0, len(products)+len(materials))
	for _, p := range products {
		if p == nil {
			continue
		}
		noteText := utils.DerefString(p.Note)
		items = append(items, DeliveryNoteItem{
			Description: composeDescriptionWithNote(
				firstNonEmpty(utils.DerefString(p.ProductName), utils.DerefString(p.ProductCode)),
				noteText,
			),
			Note:      noteText,
			Quantity:  float64(p.Quantity),
			UnitPrice: resolveDeliveryNoteUnitPrice(p.IsCloneable, p.RetailPrice),
		})
	}
	for _, m := range materials {
		if m == nil {
			continue
		}
		noteText := utils.DerefString(m.Note)
		items = append(items, DeliveryNoteItem{
			Description: composeDescriptionWithNote(
				firstNonEmpty(utils.DerefString(m.MaterialName), utils.DerefString(m.MaterialCode)),
				noteText,
			),
			Note:      noteText,
			Quantity:  float64(m.Quantity),
			UnitPrice: resolveDeliveryNoteUnitPrice(m.IsCloneable, m.RetailPrice),
		})
	}
	note.Items = items
	note.PromotionCode = utils.DerefString(orderDTO.PromotionCode)

	if strings.TrimSpace(note.Order.Number) == "" {
		return DeliveryNote{}, fmt.Errorf("order code is empty")
	}

	return note, nil
}

func pickOrderDate(dto *model.OrderDTO) time.Time {
	if dto == nil {
		return time.Time{}
	}
	if dto.DeliveryDate != nil && !dto.DeliveryDate.IsZero() {
		return *dto.DeliveryDate
	}
	if !dto.UpdatedAt.IsZero() {
		return dto.UpdatedAt
	}
	return dto.CreatedAt
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func composeDescriptionWithNote(base, note string) string {
	base = strings.TrimSpace(base)
	note = strings.TrimSpace(note)
	if note == "" {
		return base
	}
	if base == "" {
		return note
	}
	return fmt.Sprintf("%s (%s)", base, note)
}

func mergeMap(a, b map[string]any) map[string]any {
	out := make(map[string]any, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func derefFloat64(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func resolveDeliveryNoteUnitPrice(isCloneable *bool, retailPrice *float64) float64 {
	if isCloneable != nil && *isCloneable {
		return 0
	}
	return derefFloat64(retailPrice)
}

func calculateOrderProductsTotalPrice(products []*model.OrderItemProductDTO) float64 {
	if len(products) == 0 {
		return 0
	}

	var total float64
	for _, p := range products {
		if p == nil {
			continue
		}
		unitPrice := resolveDeliveryNoteUnitPrice(p.IsCloneable, p.RetailPrice)
		total += unitPrice * float64(p.Quantity)
	}
	return total
}

func buildDeliveryNotePromotionOrder(orderDTO *model.OrderDTO, products []*model.OrderItemProductDTO) *model.OrderDTO {
	if orderDTO == nil {
		return nil
	}

	orderCopy := *orderDTO
	if orderDTO.LatestOrderItem == nil {
		return &orderCopy
	}

	latestCopy := *orderDTO.LatestOrderItem
	latestCopy.Products = make([]*model.OrderItemProductDTO, 0, len(products))
	for _, p := range products {
		if p == nil {
			continue
		}
		productCopy := *p
		unitPrice := resolveDeliveryNoteUnitPrice(p.IsCloneable, p.RetailPrice)
		productCopy.RetailPrice = &unitPrice
		latestCopy.Products = append(latestCopy.Products, &productCopy)
	}

	orderCopy.LatestOrderItem = &latestCopy
	return &orderCopy
}

func (s *orderService) calculateDeliveryNotePricingByPromotion(
	ctx context.Context,
	orderDTO *model.OrderDTO,
	products []*model.OrderItemProductDTO,
) (float64, float64, error) {
	baseTotal := calculateOrderProductsTotalPrice(products)
	if orderDTO == nil {
		return 0, baseTotal, nil
	}

	promoCode := strings.TrimSpace(utils.DerefString(orderDTO.PromotionCode))
	if promoCode == "" {
		return 0, baseTotal, nil
	}

	entClient, ok := s.deps.Ent.(*generated.Client)
	if !ok || entClient == nil {
		return 0, baseTotal, nil
	}

	// Reuse the same apply flow as promotion_handler.CalculateTotalPrice.
	repo := promotionrepo.NewPromotionRepository(entClient, s.deps.DB)
	promoSvc := promotionservice.NewPromotionService(repo, s.deps)
	promotionOrder := buildDeliveryNotePromotionOrder(orderDTO, products)
	result, err := promoSvc.ApplyPromotion(ctx, nil, promotionOrder, promoCode)
	if err != nil {
		// Keep behavior consistent with promotion handler for invalid promo scenarios.
		if _, isPromoErr := promotionengine.IsPromotionApplyError(err); isPromoErr {
			return 0, baseTotal, nil
		}
		return 0, 0, err
	}

	finalAmount := baseTotal - result.DiscountAmount
	if finalAmount < 0 {
		finalAmount = 0
	}

	return result.DiscountAmount, finalAmount, nil
}

func (s *orderService) resolveDeliveryNoteCompany(ctx context.Context, orderDTO *model.OrderDTO) DeliveryNoteCompany {
	if orderDTO == nil || orderDTO.DepartmentID == nil || *orderDTO.DepartmentID <= 0 {
		return DeliveryNoteCompany{}
	}

	entClient, ok := s.deps.Ent.(*generated.Client)
	if !ok || entClient == nil {
		return DeliveryNoteCompany{}
	}

	deptEntity, err := entClient.Department.Query().
		Where(
			department.ID(*orderDTO.DepartmentID),
			department.Deleted(false),
		).
		Only(ctx)
	if err != nil {
		return DeliveryNoteCompany{}
	}

	return DeliveryNoteCompany{
		Name:     deptEntity.Name,
		LogoPath: s.resolveDepartmentLogoPath(utils.DerefString(deptEntity.Logo)),
		Address:  utils.DerefString(deptEntity.Address),
		Phone:    utils.DerefString(deptEntity.PhoneNumber),
	}
}

func (s *orderService) resolveDepartmentLogoPath(logo string) string {
	logo = strings.TrimSpace(logo)
	if logo == "" {
		return ""
	}

	// keep remote/data/file URLs as-is
	if strings.Contains(logo, "://") || strings.HasPrefix(logo, "data:") {
		return logo
	}

	// already absolute path
	if filepath.IsAbs(logo) {
		return logo
	}

	basePath := strings.TrimSpace(s.deps.Config.Storage.PhotoPath)
	if basePath == "" {
		basePath = "./storage/photo"
	}
	basePath = utils.ExpandHomeDir(basePath)

	// align with photo module path convention: <photo_path>/<size>/<filename>
	fullPath := filepath.Join(basePath, "original", logo)
	if absPath, err := filepath.Abs(fullPath); err == nil {
		fullPath = absPath
	}

	return fullPath
}

func (s *orderService) resolveDeliveryNoteShippingAddress(ctx context.Context, orderDTO *model.OrderDTO) string {
	if orderDTO == nil || orderDTO.ClinicID == nil || *orderDTO.ClinicID <= 0 {
		return ""
	}

	entClient, ok := s.deps.Ent.(*generated.Client)
	if !ok || entClient == nil {
		return ""
	}

	clinicEntity, err := entClient.Clinic.Query().
		Where(
			clinic.ID(*orderDTO.ClinicID),
			clinic.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return ""
	}

	return utils.DerefString(clinicEntity.Address)
}

func (s *orderService) resolveDeliveryNoteAttachments(
	ctx context.Context,
	orderDTO *model.OrderDTO,
	orderMaterials []*model.OrderItemMaterialDTO,
) DeliveryNoteAttachments {
	if orderDTO == nil || orderDTO.DepartmentID == nil || *orderDTO.DepartmentID <= 0 {
		return DeliveryNoteAttachments{}
	}

	entClient, ok := s.deps.Ent.(*generated.Client)
	if !ok || entClient == nil {
		return DeliveryNoteAttachments{}
	}

	orderLoanerIDs := map[int]struct{}{}
	for _, om := range orderMaterials {
		if om == nil || strings.ToLower(utils.DerefString(om.Type)) != "loaner" {
			continue
		}
		orderLoanerIDs[om.MaterialID] = struct{}{}
	}

	loanerMaterials, err := entClient.Material.Query().
		Where(
			material.DepartmentIDEQ(*orderDTO.DepartmentID),
			material.TypeEQ("loaner"),
			material.DeletedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		return DeliveryNoteAttachments{}
	}

	items := make([]DeliveryNoteAttachmentItem, 0, len(loanerMaterials))
	for _, m := range loanerMaterials {
		_, checked := orderLoanerIDs[m.ID]
		items = append(items, DeliveryNoteAttachmentItem{
			ID:      m.ID,
			Name:    utils.DerefString(m.Name),
			Checked: checked,
		})
	}

	return DeliveryNoteAttachments{Items: items}
}

func (s *orderService) resolveDeliveryNoteImplantAccessories(
	ctx context.Context,
	orderDTO *model.OrderDTO,
	orderMaterials []*model.OrderItemMaterialDTO,
) DeliveryNoteImplantAccessories {
	if orderDTO == nil || orderDTO.DepartmentID == nil || *orderDTO.DepartmentID <= 0 {
		return DeliveryNoteImplantAccessories{}
	}

	entClient, ok := s.deps.Ent.(*generated.Client)
	if !ok || entClient == nil {
		return DeliveryNoteImplantAccessories{}
	}

	orderLoanerIDs := map[int]struct{}{}
	for _, om := range orderMaterials {
		if om == nil || strings.ToLower(utils.DerefString(om.Type)) != "loaner" {
			continue
		}
		orderLoanerIDs[om.MaterialID] = struct{}{}
	}

	implantMaterials, err := entClient.Material.Query().
		Where(
			material.DepartmentIDEQ(*orderDTO.DepartmentID),
			material.TypeEQ("loaner"),
			material.IsImplantEQ(true),
			material.DeletedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		return DeliveryNoteImplantAccessories{}
	}

	items := make([]DeliveryNoteImplantAccessoryItem, 0, len(implantMaterials))
	for _, m := range implantMaterials {
		_, checked := orderLoanerIDs[m.ID]
		items = append(items, DeliveryNoteImplantAccessoryItem{
			ID:      m.ID,
			Name:    utils.DerefString(m.Name),
			Checked: checked,
		})
	}

	return DeliveryNoteImplantAccessories{Items: items}
}
