package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/khiemnd777/andy_api/shared/utils"
)

type LocalStorage struct {
	BasePath          string
	PublicBaseURL     string
	PublicStripPrefix string
}

func NewLocalStorage(basePath string, publicBaseURL string, publicStripPrefix string) *LocalStorage {
	return &LocalStorage{
		BasePath:          utils.ExpandHomeDir(basePath),
		PublicBaseURL:     strings.TrimRight(strings.TrimSpace(publicBaseURL), "/"),
		PublicStripPrefix: strings.Trim(strings.TrimSpace(publicStripPrefix), "/"),
	}
}

func (s *LocalStorage) Upload(ctx context.Context, relPath string, file io.Reader) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	cleanRelPath := path.Clean(strings.TrimLeft(relPath, "/"))
	fullPath := filepath.Join(s.BasePath, filepath.FromSlash(cleanRelPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", fmt.Errorf("create storage directory: %w", err)
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create storage file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("write storage file: %w", err)
	}

	publicRelPath := cleanRelPath
	if s.PublicStripPrefix != "" {
		publicRelPath = strings.TrimPrefix(cleanRelPath, s.PublicStripPrefix+"/")
	}
	if s.PublicBaseURL == "" {
		return "/" + publicRelPath, nil
	}

	return s.PublicBaseURL + "/" + path.Clean(strings.TrimLeft(publicRelPath, "/")), nil
}
