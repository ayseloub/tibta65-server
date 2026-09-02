package storage

import (
	"context"
	"mime/multipart"
)

// Storage adalah kontrak abstrak buat penyimpanan file (gambar, dokumen, dll).
// Implementasi konkretnya bisa macam-macam (local disk, S3, R2, dst),
// tapi kode yang MANGGIL interface ini gak perlu tau/peduli implementasi mana yang dipakai.
type Storage interface {
	// Upload nyimpen file ke folder tertentu, return URL publik buat akses file-nya.
	Upload(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)

	// Delete ngapus file berdasarkan path/URL yang tersimpan di database.
	Delete(ctx context.Context, path string) error
}
