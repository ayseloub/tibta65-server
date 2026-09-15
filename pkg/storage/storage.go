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

	UploadFromURL(ctx context.Context, sourceURL, folder string) (string, error)

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

func extFromContentType(ct string) string {
	switch {
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "webp"):
		return ".webp"
	default:
		return ".jpg"
	}
}
