package service

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	templateasset "github.com/khiemnd777/andy_api/modules/main/features/order/template"
)

const deliveryNoteDateFormat = "02/01/2006 15:04"

// DeliveryNote is the root payload for delivery note rendering.
type DeliveryNote struct {
	Company            DeliveryNoteCompany            `json:"company"`
	Order              DeliveryNoteOrder              `json:"order"`
	Items              []DeliveryNoteItem             `json:"items"`
	Attachments        DeliveryNoteAttachments        `json:"attachments"`
	ImplantAccessories DeliveryNoteImplantAccessories `json:"implant_accessories"`
	PaymentMethod      DeliveryNotePaymentMethod      `json:"payment_method"`
}

type DeliveryNoteCompany struct {
	LogoURL string `json:"logo_url"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

type DeliveryNoteOrder struct {
	Number          string    `json:"number"`
	BS              string    `json:"bs"`
	BN              string    `json:"bn"`
	Date            time.Time `json:"date"`
	ClinicName      string    `json:"clinic_name"`
	ShippingAddress string    `json:"shipping_address"`
}

type DeliveryNoteItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type DeliveryNoteAttachments struct {
	KhayLayDau bool `json:"khay_lay_dau"`
	HamDoi     bool `json:"ham_doi"`
	SapCan     bool `json:"sap_can"`
	GiaKhop    bool `json:"gia_khop"`
	MauRang    bool `json:"mau_rang"`
}

type DeliveryNoteImplantAccessories struct {
	CayLayDau     bool   `json:"cay_lay_dau"`
	Analog        bool   `json:"analog"`
	OcLabo        bool   `json:"oc_labo"`
	CayVanTinhLuc bool   `json:"cay_van_tinh_luc"`
	VitNgan       bool   `json:"vit_ngan"`
	OcLamSang     bool   `json:"oc_lam_sang"`
	NuouNhua      bool   `json:"nuou_nhua"`
	KhoaChuyen    bool   `json:"khoa_chuyen"`
	Khac          bool   `json:"khac"`
	KhacNote      string `json:"khac_note"`
}

type DeliveryNotePaymentMethod struct {
	TienMat bool `json:"tien_mat"`
	CongNo  bool `json:"cong_no"`
}

type deliveryNoteTemplateData struct {
	Company            DeliveryNoteCompany
	Order              deliveryNoteOrderView
	Items              []deliveryNoteItemView
	Attachments        DeliveryNoteAttachments
	ImplantAccessories DeliveryNoteImplantAccessories
	PaymentMethod      DeliveryNotePaymentMethod
	TotalQuantity      float64
	TotalAmount        float64
}

type deliveryNoteOrderView struct {
	Number          string
	BS              string
	BN              string
	DateDisplay     string
	ClinicName      string
	ShippingAddress string
}

type deliveryNoteItemView struct {
	Description string
	Quantity    float64
	UnitPrice   float64
	LineTotal   float64
}

// GenerateDeliveryNotePDF renders delivery note HTML and converts it to PDF bytes.
func GenerateDeliveryNotePDF(data DeliveryNote) ([]byte, error) {
	if strings.TrimSpace(data.Order.Number) == "" {
		return nil, errors.New("order number is required")
	}
	if len(data.Items) == 0 {
		return nil, errors.New("items must not be empty")
	}

	tpl, err := template.New("delivery_note").Funcs(template.FuncMap{
		"add1": func(i int) int {
			return i + 1
		},
		"number":   formatNumber,
		"currency": formatNumber,
		"checked": func(v bool) string {
			if v {
				return "X"
			}
			return ""
		},
	}).Parse(templateasset.DeliveryNoteHTML)
	if err != nil {
		return nil, fmt.Errorf("parse delivery note template: %w", err)
	}

	viewData := buildDeliveryNoteViewData(data)

	var htmlBuf bytes.Buffer
	if err := tpl.Execute(&htmlBuf, viewData); err != nil {
		return nil, fmt.Errorf("render delivery note html: %w", err)
	}

	pdfBytes, err := htmlToPDF(htmlBuf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("convert delivery note to pdf: %w", err)
	}

	return pdfBytes, nil
}

func buildDeliveryNoteViewData(data DeliveryNote) deliveryNoteTemplateData {
	items := make([]deliveryNoteItemView, 0, len(data.Items))
	totalQty := 0.0
	totalAmount := 0.0

	for _, it := range data.Items {
		lineTotal := it.Quantity * it.UnitPrice
		totalQty += it.Quantity
		totalAmount += lineTotal
		items = append(items, deliveryNoteItemView{
			Description: it.Description,
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice,
			LineTotal:   lineTotal,
		})
	}

	dateDisplay := ""
	if !data.Order.Date.IsZero() {
		dateDisplay = data.Order.Date.Format(deliveryNoteDateFormat)
	}

	return deliveryNoteTemplateData{
		Company: data.Company,
		Order: deliveryNoteOrderView{
			Number:          data.Order.Number,
			BS:              data.Order.BS,
			BN:              data.Order.BN,
			DateDisplay:     dateDisplay,
			ClinicName:      strings.ToUpper(strings.TrimSpace(data.Order.ClinicName)),
			ShippingAddress: data.Order.ShippingAddress,
		},
		Items:              items,
		Attachments:        data.Attachments,
		ImplantAccessories: data.ImplantAccessories,
		PaymentMethod:      data.PaymentMethod,
		TotalQuantity:      totalQty,
		TotalAmount:        totalAmount,
	}
}

func htmlToPDF(htmlBytes []byte) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "delivery-note-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	htmlPath := filepath.Join(tmpDir, "delivery_note.html")
	pdfPath := filepath.Join(tmpDir, "delivery_note.pdf")

	if err := os.WriteFile(htmlPath, htmlBytes, 0o600); err != nil {
		return nil, fmt.Errorf("write html temp file: %w", err)
	}

	browserBin, err := lookupHeadlessBrowser()
	if err != nil {
		return nil, err
	}

	args := []string{
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--allow-file-access-from-files",
		"--print-to-pdf-no-header",
		"--print-to-pdf=" + pdfPath,
		"file://" + htmlPath,
	}
	cmd := exec.Command(browserBin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("headless browser command failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}

	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("read generated pdf: %w", err)
	}

	if len(pdfBytes) == 0 {
		return nil, errors.New("generated pdf is empty")
	}

	return pdfBytes, nil
}

func lookupHeadlessBrowser() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("DELIVERY_NOTE_BROWSER_BIN")); configured != "" {
		if _, err := exec.LookPath(configured); err == nil {
			return configured, nil
		}
		return "", fmt.Errorf("DELIVERY_NOTE_BROWSER_BIN=%q is not executable", configured)
	}

	candidates := []string{
		"chromium",
		"chromium-browser",
		"google-chrome",
		"google-chrome-stable",
		"chrome",
		"microsoft-edge",
		"msedge",
	}
	for _, name := range candidates {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}

	return "", errors.New("no supported headless browser found; set DELIVERY_NOTE_BROWSER_BIN or install chromium/chrome")
}

func formatNumber(v float64) string {
	neg := v < 0
	if neg {
		v = -v
	}

	intPart := int64(v)
	frac := int(math.Round((v - float64(intPart)) * 100))
	if frac == 100 {
		intPart++
		frac = 0
	}

	intText := formatThousands(intPart)
	fracText := ""
	if frac > 0 {
		if frac%10 == 0 {
			fracText = fmt.Sprintf(".%d", frac/10)
		} else {
			fracText = fmt.Sprintf(".%02d", frac)
		}
	}

	if neg {
		return "-" + intText + fracText
	}
	return intText + fracText
}

func formatThousands(n int64) string {
	if n == 0 {
		return "0"
	}

	digits := fmt.Sprintf("%d", n)
	var b strings.Builder
	for i, r := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	return b.String()
}
