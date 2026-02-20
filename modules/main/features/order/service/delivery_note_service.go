package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

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
	Name     string       `json:"name"`
	LogoPath string       `json:"logo_path"`
	LogoData template.URL `json:"logo_data"`
	Address  string       `json:"address"`
	Phone    string       `json:"phone"`
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
	if err := os.WriteFile(htmlPath, htmlBytes, 0o600); err != nil {
		return nil, fmt.Errorf("write html temp file: %w", err)
	}

	// Prefer Chrome/Chromium via chromedp (best CSS + full control).
	if browserBin, ok := lookupHeadlessBrowser(); ok {
		pdf, err := renderPDFWithChromedp(browserBin, htmlPath, printOptionsA4())
		if err == nil {
			return pdf, nil
		}
		// If chromedp fails, fallback to wkhtmltopdf (more stable in some environments).
		// NOTE: Keep original error for troubleshooting.
		if wkhtmlBin, ok2 := lookupWkhtmltopdf(); ok2 {
			pdf2, err2 := renderPDFWithWkhtmltopdf(wkhtmlBin, htmlPath, tmpDir)
			if err2 == nil {
				return pdf2, nil
			}
			return nil, fmt.Errorf("chromedp failed: %v; wkhtmltopdf failed: %v", err, err2)
		}
		return nil, fmt.Errorf("chromedp failed: %w", err)
	}

	// Fallback to wkhtmltopdf if no Chrome found.
	if wkhtmlBin, ok := lookupWkhtmltopdf(); ok {
		return renderPDFWithWkhtmltopdf(wkhtmlBin, htmlPath, tmpDir)
	}

	return nil, errors.New("no supported PDF engine found; install chromium/chrome or wkhtmltopdf, or set DELIVERY_NOTE_BROWSER_BIN")
}

type pdfPrintOptions struct {
	// A4 in inches (CDP uses inches). A4: 8.27 x 11.69 in
	PaperWidthIn  float64
	PaperHeightIn float64

	// Margins in inches
	MarginTopIn    float64
	MarginRightIn  float64
	MarginBottomIn float64
	MarginLeftIn   float64

	PrintBackground      bool
	DisplayHeaderFooter  bool
	PreferCSSPageSize    bool
	Scale                float64
	WaitForNetworkIdleMs int
}

func printOptionsA4() pdfPrintOptions {
	// mm -> inches: mm / 25.4
	mm := func(v float64) float64 { return v / 25.4 }

	return pdfPrintOptions{
		PaperWidthIn:  8.27,
		PaperHeightIn: 11.69,

		// match your wkhtml margins: top/bottom 14mm, left/right 10mm
		MarginTopIn:    mm(14),
		MarginRightIn:  mm(10),
		MarginBottomIn: mm(14),
		MarginLeftIn:   mm(10),

		PrintBackground:     true,
		DisplayHeaderFooter: false, // IMPORTANT: prevents Chrome from injecting date/title/url
		PreferCSSPageSize:   true,
		Scale:               1.0,

		// If HTML loads images/fonts, waiting reduces flakiness.
		WaitForNetworkIdleMs: 300,
	}
}

func renderPDFWithChromedp(browserBin, htmlPath string, opt pdfPrintOptions) ([]byte, error) {
	// Use a dedicated context with timeout to avoid hangs.
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// Build Chrome allocator options.
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(browserBin),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("no-sandbox", true), // needed in many docker envs
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("allow-file-access-from-files", true),
		chromedp.Flag("disable-web-security", true), // helps local file assets; can be removed if not needed
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	// Navigate to local HTML file.
	targetURL := "file://" + filepath.ToSlash(htmlPath)

	var pdfBuf []byte

	actions := []chromedp.Action{
		chromedp.Navigate(targetURL),

		// Wait for DOM ready.
		chromedp.WaitReady("body", chromedp.ByQuery),

		// Optional: wait a little for fonts/images.
		chromedp.Sleep(time.Duration(opt.WaitForNetworkIdleMs) * time.Millisecond),

		// Print to PDF with full control.
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(opt.PrintBackground).
				WithDisplayHeaderFooter(opt.DisplayHeaderFooter).
				WithPreferCSSPageSize(opt.PreferCSSPageSize).
				WithPaperWidth(opt.PaperWidthIn).
				WithPaperHeight(opt.PaperHeightIn).
				WithMarginTop(opt.MarginTopIn).
				WithMarginBottom(opt.MarginBottomIn).
				WithMarginLeft(opt.MarginLeftIn).
				WithMarginRight(opt.MarginRightIn).
				WithScale(opt.Scale).
				Do(ctx)
			if err != nil {
				return err
			}
			pdfBuf = buf
			return nil
		}),
	}

	if err := chromedp.Run(taskCtx, actions...); err != nil {
		return nil, fmt.Errorf("chromedp render pdf failed: %w", err)
	}
	if len(pdfBuf) == 0 {
		return nil, errors.New("chromedp produced empty pdf")
	}

	// Copy to prevent unexpected reuse.
	return bytes.Clone(pdfBuf), nil
}

func renderPDFWithWkhtmltopdf(wkhtmlBin, htmlPath, tmpDir string) ([]byte, error) {
	pdfPath := filepath.Join(tmpDir, "delivery_note.pdf")

	args := []string{
		"--enable-local-file-access",
		"--encoding", "utf-8",
		"--page-size", "A4",
		"--orientation", "Portrait",
		"--margin-top", "14mm",
		"--margin-right", "10mm",
		"--margin-bottom", "14mm",
		"--margin-left", "10mm",
		"--disable-smart-shrinking",
		htmlPath,
		pdfPath,
	}

	cmd := exec.Command(wkhtmlBin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("wkhtmltopdf command failed: %w (%s)", err, strings.TrimSpace(string(out)))
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

// ConvertImageToBase64 reads a local image file and returns a data URI string.
func ConvertImageToBase64(path string) (string, error) {
	p := strings.TrimSpace(path)
	if p == "" {
		return "", fmt.Errorf("image path is empty")
	}

	p = strings.TrimPrefix(p, "file://")
	if p == "" {
		return "", fmt.Errorf("invalid image path after removing file:// prefix")
	}

	info, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("image file does not exist: %s", p)
		}
		return "", fmt.Errorf("stat image file %q: %w", p, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("image path is a directory: %s", p)
	}

	b, err := os.ReadFile(p)
	if err != nil {
		return "", fmt.Errorf("read image file %q: %w", p, err)
	}
	if len(b) == 0 {
		return "", fmt.Errorf("image file is empty: %s", p)
	}

	mimeType := http.DetectContentType(b)
	switch mimeType {
	case "image/png", "image/jpeg":
	default:
		return "", fmt.Errorf("unsupported image type %q for file %s (only png, jpg, jpeg)", mimeType, p)
	}

	encoded := base64.StdEncoding.EncodeToString(b)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
}

func readGeneratedPDF(pdfPath string) ([]byte, error) {
	pdfBytes, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("read generated pdf: %w", err)
	}
	if len(pdfBytes) == 0 {
		return nil, errors.New("generated pdf is empty")
	}
	return pdfBytes, nil
}

func lookupHeadlessBrowser() (string, bool) {
	if configured := strings.TrimSpace(os.Getenv("DELIVERY_NOTE_BROWSER_BIN")); configured != "" {
		if p, err := exec.LookPath(configured); err == nil {
			return p, true
		}
		if st, err := os.Stat(configured); err == nil && !st.IsDir() {
			return configured, true
		}
		return "", false
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
			return p, true
		}
	}

	// Common absolute paths where browser binary is installed but not in PATH.
	for _, p := range absoluteBrowserCandidates() {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, true
		}
	}

	return "", false
}

func lookupWkhtmltopdf() (string, bool) {
	if p, err := exec.LookPath("wkhtmltopdf"); err == nil {
		return p, true
	}
	for _, p := range absoluteWkhtmltopdfCandidates() {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, true
		}
	}
	return "", false
}

func absoluteBrowserCandidates() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		}
	case "linux":
		return []string{
			"/usr/bin/google-chrome",
			"/usr/bin/google-chrome-stable",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
			"/usr/bin/microsoft-edge",
			"/usr/bin/msedge",
		}
	case "windows":
		return []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		}
	default:
		return nil
	}
}

func absoluteWkhtmltopdfCandidates() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/usr/local/bin/wkhtmltopdf",
			"/opt/homebrew/bin/wkhtmltopdf",
		}
	case "linux":
		return []string{
			"/usr/bin/wkhtmltopdf",
			"/usr/local/bin/wkhtmltopdf",
		}
	case "windows":
		return []string{
			`C:\Program Files\wkhtmltopdf\bin\wkhtmltopdf.exe`,
			`C:\Program Files (x86)\wkhtmltopdf\bin\wkhtmltopdf.exe`,
		}
	default:
		return nil
	}
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
