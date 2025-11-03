package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func ExpandHomeDir(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path // fallback, không thay đổi
		}
		if path == "~" {
			return home
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func EnsureDirExists(path string) error {
	fullPath := ExpandHomeDir(path)
	return os.MkdirAll(fullPath, os.ModePerm)
}
