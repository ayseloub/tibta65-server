package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/oklog/ulid/v2"
)

type S3Storage struct {
	client     *minio.Client
	bucketName string
	publicURL  string
}

func NewS3Storage(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*S3Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	scheme := "http"
	if useSSL {
		scheme = "https"
	}

	return &S3Storage{
		client:     client,
		bucketName: bucketName,
		publicURL:  fmt.Sprintf("%s://%s/%s", scheme, endpoint, bucketName),
	}, nil
}

func (s *S3Storage) Upload(ctx context.Context, fileHeader *multipart.FileHeader, folder string) (string, error) {
	if err := validateFile(fileHeader); err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	objectName := fmt.Sprintf("%s/%s%s", folder, ulid.Make().String(), ext)

	_, err = s.client.PutObject(ctx, s.bucketName, objectName, src, fileHeader.Size, minio.PutObjectOptions{
		ContentType: fileHeader.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", err
	}

	publicURL := fmt.Sprintf("%s/%s", s.publicURL, objectName)
	return publicURL, nil
}

func (s *S3Storage) Delete(ctx context.Context, path string) error {
	objectName := strings.TrimPrefix(path, s.publicURL+"/")
	return s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
}
