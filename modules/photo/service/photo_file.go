package service

import (
	"fmt"
	"image"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"mime/multipart"

	"github.com/disintegration/imaging"
	"github.com/khiemnd777/andy_api/shared/utils"
	"github.com/rwcarlsen/goexif/exif"
)

func SaveAndResizeFile(fileHeader *multipart.FileHeader, filename string, basePath string) error {
	// Tạo các thư mục nếu chưa có
	sizes := []string{"original", "medium", "thumbnail"}
	for _, s := range sizes {
		if err := utils.EnsureDirExists(filepath.Join(basePath, s)); err != nil {
			return fmt.Errorf("failed to create folder: %w", err)
		}
	}

	// Decode file → image.Image
	srcFile, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Read EXIF orientation (nếu có)
	exifData, _ := exif.Decode(srcFile)
	orientation := 1
	if exifData != nil {
		if tag, err := exifData.Get(exif.Orientation); err == nil {
			orientation, _ = tag.Int(0)
		}
	}

	// Reset lại để decode ảnh
	_, err = srcFile.Seek(0, 0)
	if err != nil {
		return err
	}

	img, _, err := image.Decode(srcFile)
	if err != nil {
		return fmt.Errorf("decode error: %w", err)
	}

	// Xoay ảnh nếu cần thiết
	img = rotateAccordingToExif(img, orientation)

	// ✅ Convert và lưu lại toàn bộ ở định dạng JPG
	// 1. Original
	originalPath := filepath.Join(basePath, "original", filename)
	if err := saveAsJPG(img, originalPath); err != nil {
		return err
	}

	// 2. Medium (width 1024)
	medium := imaging.Resize(img, 1024, 0, imaging.Lanczos)
	if err := saveAsJPG(medium, filepath.Join(basePath, "medium", filename)); err != nil {
		return err
	}

	// 3. Thumbnail (width 256)
	thumb := imaging.Resize(img, 256, 0, imaging.Lanczos)
	if err := saveAsJPG(thumb, filepath.Join(basePath, "thumbnail", filename)); err != nil {
		return err
	}

	return nil
}

func rotateAccordingToExif(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.Rotate180(imaging.FlipH(img))
	case 5:
		return imaging.Rotate270(imaging.FlipH(img))
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.Rotate90(imaging.FlipH(img))
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}

func IsSupportedImageType(fileHeader *multipart.FileHeader) (bool, string) {
	f, err := fileHeader.Open()
	if err != nil {
		return false, ""
	}
	defer f.Close()

	// Read MIME type from buffer
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	if err != nil {
		return false, ""
	}

	mimeType := http.DetectContentType(buffer)
	mimeType = strings.ToLower(mimeType)

	switch mimeType {
	case "image/jpeg", "image/jpg":
		return true, ".jpg"
	case "image/png":
		return true, ".png"
	case "image/webp":
		return true, ".webp"
	case "image/gif":
		return true, ".gif"
	case "image/bmp":
		return true, ".bmp"
	case "image/tiff":
		return true, ".tiff"
	case "image/heic", "image/heif":
		return true, ".heic"
	case "image/avif":
		return true, ".avif"
	default:
		return false, mimeType
	}
}

func DetectMime(fileHeader *multipart.FileHeader) string {
	f, _ := fileHeader.Open()
	defer f.Close()

	buf := make([]byte, 512)
	f.Read(buf)
	return http.DetectContentType(buf)
}

func saveAsJPG(img image.Image, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	return imaging.Encode(out, img, imaging.JPEG)
}
