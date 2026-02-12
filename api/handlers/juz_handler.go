package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"
)

type JuzHandler struct {
	service     *services.HafalanService
	juzRepo     *repositories.JuzRepository
	juzItemRepo *repositories.JuzItemRepository
}

func NewJuzHandler(s *services.HafalanService, juzRepo *repositories.JuzRepository, juzItemRepo *repositories.JuzItemRepository) *JuzHandler {
	return &JuzHandler{service: s, juzRepo: juzRepo, juzItemRepo: juzItemRepo}
}

// Create godoc
// @Summary Create juz for hafalan
// @Description Create a new juz entry for Quran memorization
// @Tags Juz
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param index path int true "Juz index (1-30)"
// @Success 201 {object} utils.SuccessResponse{data=entities.Juz}
// @Failure 400 {object} utils.ErrorResponse
// @Router /juz/{index} [post]
func (h *JuzHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	juzIndex, err := strconv.Atoi(c.Params("index"))
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "Invalid juz index parameter", "INVALID_PARAMETER", nil)
	}

	juz, err := h.service.CreateJuz(userID, juzIndex)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "CREATE_JUZ_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusCreated, "Juz created successfully", juz, nil)
}

// JuzWithStatusCount response struct
type JuzWithStatusCount struct {
	JuzID      uuid.UUID `json:"juz_id"`
	JuzIndex   int       `json:"juz_index"`
	TotalItems int       `json:"total_items"`
	Menghafal  int       `json:"menghafal"`
	Interval   int       `json:"interval"`
	FSRSActive int       `json:"fsrs_active"`
	Graduate   int       `json:"graduate"`
}

// GetMyJuz godoc
// @Summary Get my juz list with item status counts
// @Description Get all juz entries with counts per status (menghafal, interval, fsrs_active, graduate)
// @Tags Juz
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /juz [get]
func (h *JuzHandler) GetMyJuz(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	// Fetch all juz for this user
	juzs, err := h.juzRepo.FindByUser(userID.String())
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_JUZ_FAILED", nil)
	}

	if len(juzs) == 0 {
		return utils.Success(c, fiber.StatusOK, "juz fetched successfully", []JuzWithStatusCount{}, nil)
	}

	// Collect juz IDs
	juzIDs := make([]string, len(juzs))
	for i, j := range juzs {
		juzIDs[i] = j.ID.String()
	}

	// Batch fetch item status counts
	statusCounts, err := h.juzItemRepo.CountItemStatusByJuzIDs(juzIDs)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_JUZ_FAILED", nil)
	}

	// Build per-juz status map
	type statusMap struct {
		Menghafal  int
		Interval   int
		FSRSActive int
		Graduate   int
		Total      int
	}
	juzStatusMap := make(map[string]*statusMap)
	for _, sc := range statusCounts {
		if _, exists := juzStatusMap[sc.JuzID]; !exists {
			juzStatusMap[sc.JuzID] = &statusMap{}
		}
		sm := juzStatusMap[sc.JuzID]
		sm.Total += sc.Count
		switch sc.Status {
		case "menghafal":
			sm.Menghafal = sc.Count
		case "interval":
			sm.Interval = sc.Count
		case "fsrs_active":
			sm.FSRSActive = sc.Count
		case "graduate":
			sm.Graduate = sc.Count
		}
	}

	// Build response
	resp := make([]JuzWithStatusCount, 0, len(juzs))
	for _, j := range juzs {
		entry := JuzWithStatusCount{
			JuzID:    j.ID,
			JuzIndex: j.Index,
		}
		if sm, ok := juzStatusMap[j.ID.String()]; ok {
			entry.TotalItems = sm.Total
			entry.Menghafal = sm.Menghafal
			entry.Interval = sm.Interval
			entry.FSRSActive = sm.FSRSActive
			entry.Graduate = sm.Graduate
		}
		resp = append(resp, entry)
	}

	return utils.Success(c, fiber.StatusOK, "juz fetched successfully", resp, nil)
}

