package handlers

import (
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type AIHandler struct {
	aiService services.AIService
}

func NewAIHandler(aiService services.AIService) *AIHandler {
	return &AIHandler{
		aiService: aiService,
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
// @Description Creates a full book outline with chapters and cards from a topic or raw text notes
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body GenerateAIRequest true "AI Generation Request"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /ai/generate-book [post]
func (h *AIHandler) GenerateBook(c *fiber.Ctx) error {
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

	return utils.Success(c, fiber.StatusOK, "Buku berhasil digenerate oleh AI", fiber.Map{
		"book": book,
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
