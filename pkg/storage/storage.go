package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
)

type Storage interface {
	Upload(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)

	Delete(ctx context.Context, path string) error
}

const MaxUploadSizeBytes = 2 * 1024 * 1024

func validateFile(fileHeader *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedImageExt[ext] {
		return fmt.Errorf("tipe file tidak didukung: %s", ext)
	}
	if fileHeader.Size > MaxUploadSizeBytes {
		return fmt.Errorf("ukuran file maksimal 1MB")
	}
	return nil
}
