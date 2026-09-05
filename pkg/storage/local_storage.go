package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/oklog/ulid/v2"
)

var allowedImageExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

type LocalStorage struct {
	baseDir string
	baseURL string
}

func NewLocalStorage(baseDir, baseURL string) *LocalStorage {
	return &LocalStorage{baseDir: baseDir, baseURL: baseURL}
}

func (s *LocalStorage) Upload(ctx context.Context, fileHeader *multipart.FileHeader, folder string) (string, error) {
	if err := validateFile(fileHeader); err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	targetDir := filepath.Join(s.baseDir, folder)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}

	filename := ulid.Make().String() + ext
	targetPath := filepath.Join(targetDir, filename)

	dst, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	publicURL := fmt.Sprintf("%s/%s/%s", s.baseURL, folder, filename)
	return publicURL, nil
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	relativePath := strings.TrimPrefix(path, s.baseURL+"/")
	fullPath := filepath.Join(s.baseDir, relativePath)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return nil
}
