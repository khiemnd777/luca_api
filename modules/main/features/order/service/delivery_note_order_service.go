package service

import (
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"

	model "github.com/khiemnd777/andy_api/modules/main/features/__model"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/clinic"
	"github.com/khiemnd777/andy_api/shared/db/ent/generated/department"
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

	products, err := s.repo.GetAllOrderProducts(ctx, req.OrderID)
	if err != nil {
		return nil, "", err
	}

	materials, err := s.repo.GetAllOrderMaterials(ctx, req.OrderID)
	if err != nil {
		return nil, "", err
	}

	company := s.resolveDeliveryNoteCompany(ctx, orderDTO)
	shippingAddress := s.resolveDeliveryNoteShippingAddress(ctx, orderDTO)
	note, err := buildDeliveryNoteFromOrder(orderDTO, products, materials, company, shippingAddress, req)
	if err != nil {
		return nil, "", err
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

	fileName := fmt.Sprintf("delivery-note-%s.pdf", strings.ReplaceAll(note.Order.Number, "/", "-"))
	return pdf, fileName, nil
}

func buildDeliveryNoteFromOrder(
	orderDTO *model.OrderDTO,
	products []*model.OrderItemProductDTO,
	materials []*model.OrderItemMaterialDTO,
	company DeliveryNoteCompany,
	shippingAddress string,
	req DeliveryNotePrintRequest,
) (DeliveryNote, error) {
	if orderDTO == nil {
		return DeliveryNote{}, fmt.Errorf("order not found")
	}

	note := DeliveryNote{
		Company: company,
		Order: DeliveryNoteOrder{
			Number:          firstNonEmpty(utils.DerefString(orderDTO.CodeLatest), utils.DerefString(orderDTO.Code)),
			BS:              utils.DerefString(orderDTO.DentistName),
			BN:              utils.DerefString(orderDTO.PatientName),
			ClinicName:      utils.DerefString(orderDTO.ClinicName),
			ShippingAddress: shippingAddress,
			Date:            pickOrderDate(orderDTO),
		},
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

	latestCF := map[string]any{}
	if orderDTO.LatestOrderItem != nil {
		latestCF = orderDTO.LatestOrderItem.CustomFields
	}
	mergedCF := mergeMap(orderDTO.CustomFields, latestCF)

	note.Attachments = DeliveryNoteAttachments{
		KhayLayDau: cfBool(mergedCF, "attachment_khay_lay_dau", "khay_lay_dau"),
		HamDoi:     cfBool(mergedCF, "attachment_ham_doi", "ham_doi"),
		SapCan:     cfBool(mergedCF, "attachment_sap_can", "sap_can"),
		GiaKhop:    cfBool(mergedCF, "attachment_gia_khop", "gia_khop"),
		MauRang:    cfBool(mergedCF, "attachment_mau_rang", "mau_rang"),
	}
	if req.Attachments != nil {
		note.Attachments = *req.Attachments
	}

	note.ImplantAccessories = DeliveryNoteImplantAccessories{
		CayLayDau:     cfBool(mergedCF, "implant_cay_lay_dau", "cay_lay_dau"),
		Analog:        cfBool(mergedCF, "implant_analog", "analog"),
		OcLabo:        cfBool(mergedCF, "implant_oc_labo", "oc_labo"),
		CayVanTinhLuc: cfBool(mergedCF, "implant_cay_van_tinh_luc", "cay_van_tinh_luc"),
		VitNgan:       cfBool(mergedCF, "implant_vit_ngan", "vit_ngan"),
		OcLamSang:     cfBool(mergedCF, "implant_oc_lam_sang", "oc_lam_sang"),
		NuouNhua:      cfBool(mergedCF, "implant_nuou_nhua", "nuou_nhua"),
		KhoaChuyen:    cfBool(mergedCF, "implant_khoa_chuyen", "khoa_chuyen"),
		Khac:          cfBool(mergedCF, "implant_khac", "khac"),
		KhacNote:      firstNonEmpty(utils.SafeGetString(mergedCF, "implant_khac_note"), utils.SafeGetString(mergedCF, "khac_note")),
	}
	if req.ImplantAccessories != nil {
		note.ImplantAccessories = *req.ImplantAccessories
	}

	note.PaymentMethod = DeliveryNotePaymentMethod{
		TienMat: cfBool(mergedCF, "payment_tien_mat", "tien_mat"),
		CongNo:  cfBool(mergedCF, "payment_cong_no", "cong_no"),
	}
	if req.PaymentMethod != nil {
		note.PaymentMethod = *req.PaymentMethod
	}

	items := make([]DeliveryNoteItem, 0, len(products)+len(materials))
	for _, p := range products {
		if p == nil {
			continue
		}
		items = append(items, DeliveryNoteItem{
			Description: firstNonEmpty(utils.DerefString(p.ProductName), utils.DerefString(p.ProductCode)),
			Quantity:    float64(p.Quantity),
			UnitPrice:   derefFloat64(p.RetailPrice),
		})
	}
	for _, m := range materials {
		if m == nil {
			continue
		}
		items = append(items, DeliveryNoteItem{
			Description: firstNonEmpty(utils.DerefString(m.MaterialName), utils.DerefString(m.MaterialCode)),
			Quantity:    float64(m.Quantity),
			UnitPrice:   derefFloat64(m.RetailPrice),
		})
	}
	note.Items = items

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

func cfBool(m map[string]any, keys ...string) bool {
	for _, key := range keys {
		if _, ok := m[key]; ok {
			return utils.SafeGetBool(m, key)
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
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
