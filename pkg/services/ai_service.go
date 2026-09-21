package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type GeneratedCard struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type GeneratedChapter struct {
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Cards       []GeneratedCard `json:"cards"`
}

type GeneratedBook struct {
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Chapters    []GeneratedChapter `json:"chapters"`
}

type AIService interface {
	GenerateBook(ctx context.Context, topic, text, language string) (*GeneratedBook, error)
	GenerateCards(ctx context.Context, topic, text, language string) ([]GeneratedCard, error)
}

type aiService struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewAIService() AIService {
	apiKey := os.Getenv("GEMINI_API_KEY")
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.0-flash"
	}
	return &aiService{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Gemini REST API request / response structs
type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiGenerationConfig struct {
	ResponseMimeType string      `json:"responseMimeType,omitempty"`
	ResponseSchema   interface{} `json:"responseSchema,omitempty"`
}

type geminiRequest struct {
	Contents         []geminiContent         `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiCandidatePart struct {
	Text string `json:"text"`
}

type geminiCandidateContent struct {
	Parts []geminiCandidatePart `json:"parts"`
}

type geminiCandidate struct {
	Content geminiCandidateContent `json:"content"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func cleanBilingualText(t string) string {
	t = strings.TrimSpace(t)
	return t
}

func (s *aiService) callGemini(ctx context.Context, prompt string, responseSchema interface{}) (string, error) {
	if s.apiKey == "" {
		return "", errors.New("GEMINI_API_KEY belum dikonfigurasi di server")
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", s.model, s.apiKey)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			ResponseMimeType: "application/json",
			ResponseSchema:   responseSchema,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("error connecting to Gemini API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var gErr geminiResponse
		if json.Unmarshal(respBytes, &gErr) == nil && gErr.Error != nil {
			return "", fmt.Errorf("Gemini API error (%d %s): %s", gErr.Error.Code, gErr.Error.Status, gErr.Error.Message)
		}
		return "", fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var gResp geminiResponse
	if err := json.Unmarshal(respBytes, &gResp); err != nil {
		return "", fmt.Errorf("failed to decode Gemini response: %w", err)
	}

	if len(gResp.Candidates) == 0 || len(gResp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("tidak ada hasil yang dihasilkan oleh model AI")
	}

	return gResp.Candidates[0].Content.Parts[0].Text, nil
}

func (s *aiService) GenerateBook(ctx context.Context, topic, text, language string) (*GeneratedBook, error) {
	if topic == "" && text == "" {
		return nil, errors.New("topik atau teks catatan wajib diisi")
	}

	langLabel := "Bahasa Indonesia"
	if language == "en" {
		langLabel = "English"
	}

	prompt := fmt.Sprintf(`Tugas Anda adalah memformat dan menyusun teks/topik materi berikut menjadi struktur buku hafalan pintar berbab (maksimal 5 bab) untuk aplikasi Unlupa.
%s%s
Petunjuk:
1. Kelompokkan materi ke dalam bab-bab (chapters) yang logis dan runtut.
2. Setiap bab harus memiliki daftar kartu tanya-jawab (Q&A flashcards) yang berkualitas, tajam, dan mudah dihafal.
3. PANDUAN KHUSUS TEKS DWI-BAHASA (ARAB & INDONESIA):
   - Jika pertanyaan atau jawaban memuat teks Arab dan terjemahan/penjelasan bahasa Indonesia, Anda WAJIB memisahkannya dengan baris baru (\n). JANGAN PERNAH menggabungkan teks Arab dan teks Indonesia dalam satu baris bersambung tanpa enter.
   - Pertahankan teks Arab aslinya secara utuh lengkap dengan harakat/tanda baca.
   - Contoh pertanyaan: "سُوْرَةُ النَّبَإِ مِنْ أَيِّ السُّوَرِ؟\nSurah An-Naba' termasuk surah apa?"
   - Contoh jawaban: "مَكِّيَّةٌ، أَيْ نَزَلَتْ قَبْلَ هِجْرَةِ رَسُولِ اللَّهِ ﷺ إِلَى الْمَدِينَةِ.\nSurah An-Naba' termasuk surah Makkiyah, yaitu surah yang turun sebelum hijrahnya Rasulullah ﷺ ke Madinah."
4. Bahasa Output: %s.`,
		func() string {
			if topic != "" {
				return fmt.Sprintf("Topik Utama: %s\n", topic)
			}
			return ""
		}(),
		func() string {
			if text != "" {
				return fmt.Sprintf("Teks/Catatan/Q&A Sumber:\n%s\n", text)
			}
			return ""
		}(),
		langLabel,
	)

	schema := map[string]interface{}{
		"type": "OBJECT",
		"properties": map[string]interface{}{
			"title":       map[string]interface{}{"type": "STRING", "description": "Judul Buku"},
			"description": map[string]interface{}{"type": "STRING", "description": "Deskripsi singkat isi buku"},
			"chapters": map[string]interface{}{
				"type": "ARRAY",
				"items": map[string]interface{}{
					"type": "OBJECT",
					"properties": map[string]interface{}{
						"title":       map[string]interface{}{"type": "STRING", "description": "Judul Bab/Modul"},
						"description": map[string]interface{}{"type": "STRING", "description": "Deskripsi Bab"},
						"cards": map[string]interface{}{
							"type": "ARRAY",
							"items": map[string]interface{}{
								"type": "OBJECT",
								"properties": map[string]interface{}{
									"question": map[string]interface{}{"type": "STRING", "description": "Pertanyaan kartu hafalan"},
									"answer":   map[string]interface{}{"type": "STRING", "description": "Jawaban kartu hafalan"},
								},
								"required": []string{"question", "answer"},
							},
						},
					},
					"required": []string{"title", "cards"},
				},
			},
		},
		"required": []string{"title", "description", "chapters"},
	}

	jsonText, err := s.callGemini(ctx, prompt, schema)
	if err != nil {
		return nil, err
	}

	var book GeneratedBook
	if err := json.Unmarshal([]byte(jsonText), &book); err != nil {
		return nil, fmt.Errorf("gagal memproses output JSON dari AI: %w", err)
	}

	// Clean up strings
	book.Title = cleanBilingualText(book.Title)
	book.Description = cleanBilingualText(book.Description)
	for ci := range book.Chapters {
		book.Chapters[ci].Title = cleanBilingualText(book.Chapters[ci].Title)
		book.Chapters[ci].Description = cleanBilingualText(book.Chapters[ci].Description)
		for cdi := range book.Chapters[ci].Cards {
			book.Chapters[ci].Cards[cdi].Question = cleanBilingualText(book.Chapters[ci].Cards[cdi].Question)
			book.Chapters[ci].Cards[cdi].Answer = cleanBilingualText(book.Chapters[ci].Cards[cdi].Answer)
		}
	}

	return &book, nil
}

func (s *aiService) GenerateCards(ctx context.Context, topic, text, language string) ([]GeneratedCard, error) {
	if topic == "" && text == "" {
		return nil, errors.New("topik atau teks catatan wajib diisi")
	}

	langLabel := "Bahasa Indonesia"
	if language == "en" {
		langLabel = "English"
	}

	prompt := fmt.Sprintf(`Tugas Anda adalah membuat kumpulan kartu hafalan tanya-jawab (flashcards) dari topik atau teks catatan berikut untuk aplikasi Unlupa.
%s%s
Petunjuk:
1. Buat pasangan tanya-jawab (Q&A) yang akurat, jelas, dan mempermudah hafalan jangka panjang.
2. PANDUAN KHUSUS DWI-BAHASA (ARAB & INDONESIA):
   - Jika memuat teks Arab dan terjemahan/penjelasan Indonesia, pisahkan dengan baris baru (\n).
   - Jangan gabungkan teks Arab dan teks Indonesia dalam satu baris bersambung.
   - Pertahankan harakat teks Arab.
3. Bahasa Output: %s.`,
		func() string {
			if topic != "" {
				return fmt.Sprintf("Topik: %s\n", topic)
			}
			return ""
		}(),
		func() string {
			if text != "" {
				return fmt.Sprintf("Teks Catatan:\n%s\n", text)
			}
			return ""
		}(),
		langLabel,
	)

	schema := map[string]interface{}{
		"type": "OBJECT",
		"properties": map[string]interface{}{
			"cards": map[string]interface{}{
				"type": "ARRAY",
				"items": map[string]interface{}{
					"type": "OBJECT",
					"properties": map[string]interface{}{
						"question": map[string]interface{}{"type": "STRING", "description": "Pertanyaan"},
						"answer":   map[string]interface{}{"type": "STRING", "description": "Jawaban"},
					},
					"required": []string{"question", "answer"},
				},
			},
		},
		"required": []string{"cards"},
	}

	jsonText, err := s.callGemini(ctx, prompt, schema)
	if err != nil {
		return nil, err
	}

	var res struct {
		Cards []GeneratedCard `json:"cards"`
	}
	if err := json.Unmarshal([]byte(jsonText), &res); err != nil {
		return nil, fmt.Errorf("gagal memproses output JSON dari AI: %w", err)
	}

	for i := range res.Cards {
		res.Cards[i].Question = cleanBilingualText(res.Cards[i].Question)
		res.Cards[i].Answer = cleanBilingualText(res.Cards[i].Answer)
	}

	return res.Cards, nil
}
