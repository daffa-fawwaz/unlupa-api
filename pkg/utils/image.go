package utils

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/image/draw"
)

const (
	MaxImageSize    = 3 * 1024 * 1024 // 3 MB
	UploadDir       = "uploads/covers"
	TargetMaxWidth  = 1200
	TargetMaxHeight = 1200
)

var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

// SaveCoverImage validates, optionally resizes, and saves an uploaded cover image.
// Returns the relative file path on success.
func SaveCoverImage(file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExts[ext] {
		return "", errors.New("only jpg, jpeg, and png files are allowed")
	}

	// Ensure upload directory exists
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Decode image
	img, format, err := image.Decode(src)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize if file exceeds 3MB
	if file.Size > MaxImageSize {
		img = resizeImage(img, TargetMaxWidth, TargetMaxHeight)
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	destPath := filepath.Join(UploadDir, filename)

	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	// Encode based on format
	switch format {
	case "png":
		if err := png.Encode(out, img); err != nil {
			return "", fmt.Errorf("failed to encode png: %w", err)
		}
	default:
		// jpeg / jpg
		if err := jpeg.Encode(out, img, &jpeg.Options{Quality: 85}); err != nil {
			return "", fmt.Errorf("failed to encode jpeg: %w", err)
		}
	}

	return "/" + destPath, nil
}

// resizeImage scales down img so it fits within maxW x maxH while preserving aspect ratio.
func resizeImage(src image.Image, maxW, maxH int) image.Image {
	bounds := src.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	if origW <= maxW && origH <= maxH {
		return src
	}

	ratio := float64(origW) / float64(origH)
	newW, newH := maxW, maxH

	if float64(maxW)/float64(maxH) > ratio {
		newW = int(float64(maxH) * ratio)
	} else {
		newH = int(float64(maxW) / ratio)
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}
