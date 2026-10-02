package utils

import (
	"errors"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var allowedExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
}

var allowedMIME = map[string]bool{
	"image/jpeg": true, "image/png": true, "image/webp": true,
}

func SaveUpload(file *multipart.FileHeader, destDir string, maxMB int64) (string, error) {
	if file.Size > maxMB*1024*1024 {
		return "", errors.New("ukuran file maksimal 5 MB")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExt[ext] {
		return "", errors.New("format file harus JPG, JPEG, PNG, atau WEBP")
	}
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	buf := make([]byte, 512)
	n, _ := src.Read(buf)
	mime := http.DetectContentType(buf[:n])
	if !allowedMIME[mime] {
		return "", errors.New("tipe file tidak valid (harus gambar)")
	}
	name := uuid.New().String() + ext
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	dst := filepath.Join(destDir, name)
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()
	// re-read full file
	src.Seek(0, 0)
	if _, err := out.ReadFrom(src); err != nil {
		// fallback copy
		src.Seek(0, 0)
		b := make([]byte, 32*1024)
		for {
			nr, er := src.Read(b)
			if nr > 0 {
				out.Write(b[:nr])
			}
			if er != nil {
				break
			}
		}
	}
	return "/uploads/" + name, nil
}
