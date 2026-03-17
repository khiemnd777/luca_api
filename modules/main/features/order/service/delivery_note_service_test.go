package service

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetDeliveryNoteTemplate_Cached(t *testing.T) {
	tpl1, err := getDeliveryNoteTemplate()
	if err != nil {
		t.Fatalf("getDeliveryNoteTemplate first call: %v", err)
	}
	tpl2, err := getDeliveryNoteTemplate()
	if err != nil {
		t.Fatalf("getDeliveryNoteTemplate second call: %v", err)
	}
	if tpl1 == nil || tpl2 == nil {
		t.Fatal("expected non-nil templates")
	}
	if tpl1 != tpl2 {
		t.Fatal("expected cached template instance to be reused")
	}
}

func TestConvertImageToBase64_WithFileSchemePNG(t *testing.T) {
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "logo.png")

	// 1x1 transparent PNG
	pngB64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+X8xkAAAAASUVORK5CYII="
	pngBytes, err := base64.StdEncoding.DecodeString(pngB64)
	if err != nil {
		t.Fatalf("decode test png: %v", err)
	}
	if err := os.WriteFile(imgPath, pngBytes, 0o600); err != nil {
		t.Fatalf("write test png: %v", err)
	}

	dataURI, err := ConvertImageToBase64("file://" + imgPath)
	if err != nil {
		t.Fatalf("ConvertImageToBase64 returned error: %v", err)
	}
	if !strings.HasPrefix(dataURI, "data:image/png;base64,") {
		t.Fatalf("unexpected data URI prefix: %s", dataURI[:min(32, len(dataURI))])
	}
}

func TestConvertImageToBase64_NotFound(t *testing.T) {
	_, err := ConvertImageToBase64("/no/such/file.png")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBuildQRCodeImageURL_EscapesPayload(t *testing.T) {
	out := BuildQRCodeImageURL("https://example.com/delivery/qr/a b", 160)
	if !strings.Contains(out, "data=https%3A%2F%2Fexample.com%2Fdelivery%2Fqr%2Fa+b") {
		t.Fatalf("unexpected QR image URL: %s", out)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
