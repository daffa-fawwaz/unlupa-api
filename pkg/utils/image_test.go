package utils

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"os"
	"strings"
	"testing"
)

func createTestFileHeader(filename string, data []byte) *multipart.FileHeader {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("cover", filename)
	part.Write(data)
	writer.Close()

	reader := multipart.NewReader(&body, writer.Boundary())
	form, _ := reader.ReadForm(10 * 1024 * 1024)
	return form.File["cover"][0]
}

func TestSupabaseStorageConfigured(t *testing.T) {
	if os.Getenv("SUPABASE_URL") != "" && os.Getenv("SUPABASE_BUCKET") != "" &&
		(os.Getenv("SUPABASE_SERVICE_ROLE_KEY") != "" || os.Getenv("SUPABASE_ANON_KEY") != "") {
		if !supabaseStorageConfigured() {
			t.Errorf("expected supabaseStorageConfigured to be true, got false")
		}
	}
}

func TestSaveCoverImage_ValidPNG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			img.Set(x, y, color.RGBA{R: 0, G: 128, B: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode error: %v", err)
	}

	fh := createTestFileHeader("sample.png", buf.Bytes())
	resultURL, err := SaveCoverImage(fh)
	if err != nil {
		t.Fatalf("SaveCoverImage failed: %v", err)
	}

	if resultURL == "" {
		t.Fatalf("expected non-empty result URL")
	}

	if supabaseStorageConfigured() {
		if !strings.HasPrefix(resultURL, "https://") || !strings.Contains(resultURL, "supabase.co") {
			t.Errorf("expected Supabase public URL, got: %s", resultURL)
		}
		// Clean up from supabase
		_ = DeleteFromSupabase(resultURL)
	} else {
		if !strings.HasPrefix(resultURL, "/uploads/covers/") {
			t.Errorf("expected local upload path, got: %s", resultURL)
		}
		// Clean up local
		_ = os.Remove(strings.TrimPrefix(resultURL, "/"))
	}
}

func TestSaveCoverImage_InvalidExtension(t *testing.T) {
	fh := createTestFileHeader("malicious.exe", []byte("MZ binary header"))
	_, err := SaveCoverImage(fh)
	if err == nil {
		t.Errorf("expected error for invalid extension, got nil")
	}
}

func TestSaveCoverImage_OversizedFile(t *testing.T) {
	oversized := make([]byte, MaxImageSize+1024)
	fh := createTestFileHeader("large.jpg", oversized)
	_, err := SaveCoverImage(fh)
	if err == nil {
		t.Errorf("expected error for oversized image, got nil")
	}
}

func TestDeleteFromSupabase_NonSupabaseURL(t *testing.T) {
	if err := DeleteFromSupabase(""); err != nil {
		t.Errorf("expected nil error for empty url, got: %v", err)
	}
	if err := DeleteFromSupabase("/uploads/covers/test.jpg"); err != nil {
		t.Errorf("expected nil error for local url, got: %v", err)
	}
}
