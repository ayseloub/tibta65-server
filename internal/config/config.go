package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string
	AppPort       string
	DatabaseURL   string
	JWTSecret     string
	UploadDir     string
	UploadURL     string
	StorageDriver string
	S3Endpoint    string
	S3AccessKey   string
	S3SecretKey   string
	S3BucketName  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		AppPort:       getEnv("APP_PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		StorageDriver: getEnv("STORAGE_DRIVER", "local"),
		S3Endpoint:    getEnv("S3_ENDPOINT", ""),
		S3AccessKey:   getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:   getEnv("S3_SECRET_KEY", ""),
		S3BucketName:  getEnv("S3_BUCKET_NAME", ""),
		UploadURL:     getEnv("UPLOAD_URL", "/uploads"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
