package handlers

import (
	"time"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const WeeklyAIBookLimit = 3

type AIHandler struct {
	aiService services.AIService
	aiLogRepo repositories.AIBookGenerationRepository
}

func NewAIHandler(aiService services.AIService, aiLogRepo repositories.AIBookGenerationRepository) *AIHandler {
	return &AIHandler{
		aiService: aiService,
		aiLogRepo: aiLogRepo,
	}
}

type GenerateAIRequest struct {
	Topic    string `json:"topic"`
	Text     string `json:"text"`
	Mode     string `json:"mode,omitempty"`
	Language string `json:"language,omitempty"`
}

// GenerateBook godoc
// @Summary Generate structured book using Gemini AI
// @Description Creates a full book outline with chapters and cards from a topic or raw text notes with weekly limit
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body GenerateAIRequest true "AI Generation Request"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 429 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /ai/generate-book [post]
func (h *AIHandler) GenerateBook(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.Error(c, fiber.StatusUnauthorized, "User ID not found", "UNAUTHORIZED", nil)
	}

	// Batasi setiap user hanya bisa menggunakan fitur buat buku dengan AI 3x / minggu
	since := time.Now().AddDate(0, 0, -7)
	count, err := h.aiLogRepo.CountWeeklyBookGenerations(userID, since)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal memeriksa kuota buat buku AI", "INTERNAL_SERVER_ERROR", nil)
	}

	if count >= WeeklyAIBookLimit {
		return utils.Error(
			c,
			fiber.StatusTooManyRequests,
			"Batas penggunaan buat buku dengan AI telah tercapai (maksimal 3x per minggu)",
			"AI_LIMIT_EXCEEDED",
			nil,
		)
	}

	var req GenerateAIRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_REQUEST", nil)
	}

	if req.Topic == "" && req.Text == "" {
		return utils.Error(c, fiber.StatusBadRequest, "Topik atau teks catatan wajib diisi", "BAD_REQUEST", nil)
	}

	book, err := h.aiService.GenerateBook(c.Context(), req.Topic, req.Text, req.Language)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "AI_GENERATE_FAILED", nil)
	}

	// Catat riwayat pembuatan buku dengan AI
	logEntry := &entities.AIBookGenerationLog{
		UserID: userID,
		Topic:  req.Topic,
	}
	_ = h.aiLogRepo.LogBookGeneration(logEntry)

	remaining := int64(WeeklyAIBookLimit) - (count + 1)
	if remaining < 0 {
		remaining = 0
	}

	return utils.Success(c, fiber.StatusOK, "Buku berhasil digenerate oleh AI", fiber.Map{
		"book":      book,
		"limit":     WeeklyAIBookLimit,
		"used":      count + 1,
		"remaining": remaining,
	}, nil)
}

// GetBookUsage godoc
// @Summary Get user's AI book generation quota usage
// @Description Returns weekly limit, usage, and remaining quota for creating books with AI
// @Tags AI
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Router /ai/book-usage [get]
func (h *AIHandler) GetBookUsage(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return utils.Error(c, fiber.StatusUnauthorized, "User ID not found", "UNAUTHORIZED", nil)
	}

	since := time.Now().AddDate(0, 0, -7)
	count, err := h.aiLogRepo.CountWeeklyBookGenerations(userID, since)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "Gagal mengambil data kuota AI", "INTERNAL_SERVER_ERROR", nil)
	}

	remaining := int64(WeeklyAIBookLimit) - count
	if remaining < 0 {
		remaining = 0
	}

	return utils.Success(c, fiber.StatusOK, "Data kuota buat buku dengan AI berhasil diambil", fiber.Map{
		"limit":     WeeklyAIBookLimit,
		"used":      count,
		"remaining": remaining,
		"period":    "weekly",
	}, nil)
}

// GenerateCards godoc
// @Summary Generate flashcards using Gemini AI
// @Description Creates flashcard pairs (question & answer) from a topic or text notes
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body GenerateAIRequest true "AI Generation Request"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /ai/generate-cards [post]
func (h *AIHandler) GenerateCards(c *fiber.Ctx) error {
	var req GenerateAIRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid request body", "INVALID_REQUEST", nil)
	}

	if req.Topic == "" && req.Text == "" {
		return utils.Error(c, fiber.StatusBadRequest, "Topik atau teks catatan wajib diisi", "BAD_REQUEST", nil)
	}

	cards, err := h.aiService.GenerateCards(c.Context(), req.Topic, req.Text, req.Language)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "AI_GENERATE_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "Kartu berhasil digenerate oleh AI", fiber.Map{
		"cards": cards,
	}, nil)
}
