package utils

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
)

const (
	MaxImageSize = 3 * 1024 * 1024 // 3 MB
	UploadDir    = "uploads/covers"
)

var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

// SaveCoverImage validates and saves an uploaded cover image.
// Returns a public URL when Supabase Storage is configured, otherwise a local
// relative file path.
func SaveCoverImage(file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExts[ext] {
		return "", errors.New("only png, jpg, jpeg, and webp files are allowed")
	}

	if file.Size > MaxImageSize {
		return "", errors.New("image size must be 3MB or less")
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	data, err := io.ReadAll(io.LimitReader(src, MaxImageSize+1))
	if err != nil {
		return "", fmt.Errorf("failed to read uploaded file: %w", err)
	}
	if int64(len(data)) > MaxImageSize {
		return "", errors.New("image size must be 3MB or less")
	}

	// Decode image
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}
	if !validImageFormatForExt(ext, format) {
		return "", errors.New("uploaded file content does not match the allowed image format")
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	objectPath := "covers/" + filename

	var buf bytes.Buffer
	// Encode based on format
	switch format {
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return "", fmt.Errorf("failed to encode png: %w", err)
		}
	case "webp":
		buf.Write(data)
	default:
		// jpeg / jpg
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
			return "", fmt.Errorf("failed to encode jpeg: %w", err)
		}
	}

	contentType := "image/jpeg"
	if format == "png" {
		contentType = "image/png"
	} else if format == "webp" {
		contentType = "image/webp"
	}

	if supabaseStorageConfigured() {
		return uploadCoverImageToSupabase(objectPath, buf.Bytes(), contentType)
	}

	return saveCoverImageLocally(filename, buf.Bytes())
}

func saveCoverImageLocally(filename string, data []byte) (string, error) {
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	destPath := filepath.Join(UploadDir, filename)
	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write output file: %w", err)
	}

	return "/" + destPath, nil
}

func supabaseStorageConfigured() bool {
	return supabaseBaseURL() != "" &&
		supabaseAPIKey() != "" &&
		os.Getenv("SUPABASE_BUCKET") != ""
}

func uploadCoverImageToSupabase(objectPath string, data []byte, contentType string) (string, error) {
	baseURL := supabaseBaseURL()
	apiKey := supabaseAPIKey()
	bucket := os.Getenv("SUPABASE_BUCKET")

	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", baseURL, url.PathEscape(bucket), escapeObjectPath(objectPath))
	req, err := http.NewRequest(http.MethodPost, uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create supabase upload request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Cache-Control", "3600")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload image to supabase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("supabase upload failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", baseURL, url.PathEscape(bucket), escapeObjectPath(objectPath)), nil
}

func supabaseBaseURL() string {
	baseURL := strings.TrimRight(os.Getenv("SUPABASE_URL"), "/")
	baseURL = strings.TrimSuffix(baseURL, "/rest/v1")
	baseURL = strings.TrimSuffix(baseURL, "/storage/v1")
	return strings.TrimRight(baseURL, "/")
}

func supabaseAPIKey() string {
	if key := os.Getenv("SUPABASE_SERVICE_ROLE_KEY"); key != "" {
		return key
	}
	return os.Getenv("SUPABASE_ANON_KEY")
}

func validImageFormatForExt(ext string, format string) bool {
	switch ext {
	case ".jpg", ".jpeg":
		return format == "jpeg"
	case ".png":
		return format == "png"
	case ".webp":
		return format == "webp"
	default:
		return false
	}
}

func escapeObjectPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

// DeleteFromSupabase removes an object from Supabase Storage given its public URL.
// It is a no-op when Supabase is not configured or the URL is empty / not a Supabase URL.
func DeleteFromSupabase(publicURL string) error {
	if publicURL == "" {
		return nil
	}
	if !supabaseStorageConfigured() {
		return nil
	}

	baseURL := supabaseBaseURL()
	bucket := os.Getenv("SUPABASE_BUCKET")

	// Extract object path from public URL.
	// Expected format: {baseURL}/storage/v1/object/public/{bucket}/{objectPath}
	prefix := fmt.Sprintf("%s/storage/v1/object/public/%s/", baseURL, url.PathEscape(bucket))
	if !strings.HasPrefix(publicURL, prefix) {
		// Not a Supabase URL we manage — skip silently
		return nil
	}

	objectPath, err := url.PathUnescape(publicURL[len(prefix):])
	if err != nil {
		objectPath = publicURL[len(prefix):]
	}

	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s", baseURL, url.PathEscape(bucket))
	body := fmt.Sprintf(`{"prefixes":["%s"]}`, strings.ReplaceAll(objectPath, `"`, `\"`))

	req, err := http.NewRequest(http.MethodDelete, deleteURL, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create supabase delete request: %w", err)
	}

	apiKey := supabaseAPIKey()
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("apikey", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete image from supabase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("supabase delete failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}
