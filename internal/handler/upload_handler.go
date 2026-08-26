package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type UploadHandler struct {
	uploadDir string
}

func NewUploadHandler(uploadDir string) *UploadHandler {
	return &UploadHandler{uploadDir: uploadDir}
}

func (h *UploadHandler) UploadImage(c echo.Context) error {
	c.Request().Body = http.MaxBytesReader(c.Response().Writer, c.Request().Body, 6*1024*1024)
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "no image file provided"})
	}
	if fileHeader.Size > 5*1024*1024 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "image must be 5 MB or smaller"})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to read image"})
	}
	defer src.Close()

	contentType, err := detectImageType(src)
	if err != nil || !strings.HasPrefix(contentType, "image/") {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "only image files are allowed"})
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to read image"})
	}

	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to prepare image storage"})
	}

	ext := safeExtension(contentType)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	destPath := filepath.Join(h.uploadDir, filename)
	dst, err := os.Create(destPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save image"})
	}
	defer dst.Close()

	if _, err := copyFile(dst, src); err != nil {
		_ = os.Remove(destPath)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save image"})
	}

	return c.JSON(http.StatusOK, map[string]string{"url": "/uploads/" + filename})
}

func detectImageType(src multipart.File) (string, error) {
	buf := make([]byte, 512)
	n, err := src.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	return http.DetectContentType(buf[:n]), nil
}

func safeExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".img"
	}
}

func copyFile(dst *os.File, src multipart.File) (int64, error) {
	return io.Copy(dst, src)
}
