package utils

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func GetGoModuleName() string {
	root := GetProjectRootDir()
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

func GetProjectRootDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "." // fallback
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "." // fallback
}

func GetFullPath(paths ...string) string {
	root := GetProjectRootDir()
	return filepath.Join(append([]string{root}, paths...)...)
}

func GetModulePath(moduleName string, subpaths ...string) string {
	all := append([]string{"modules", moduleName}, subpaths...)
	return GetFullPath(all...)
}

func GetModuleConfigPath(moduleName string) string {
	return GetModulePath(moduleName, "config.yaml")
}

func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func CopyDir(src, dst string) error {
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func Coalesce[T comparable](v, def T) T {
	var zero T
	if v == zero {
		return def
	}
	return v
}

func PtrCoalesce[T comparable](ptr *T, def T) T {
	if ptr == nil {
		return def
	}
	return Coalesce(*ptr, def)
}

func Ptr[T any](v T) *T {
	return &v
}

func Deref(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}
