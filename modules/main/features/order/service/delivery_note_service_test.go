package service

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
