package handlers

import (
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TeacherRequestHandler struct {
	teacherReqSvc services.TeacherRequestService
}

func NewTeacherRequestHandler(teacherReqSvc services.TeacherRequestService) *TeacherRequestHandler {
	return &TeacherRequestHandler{teacherReqSvc}
}

// RequestTeacher godoc
// @Summary Request to become teacher
// @Description Submit a request to become a teacher
// @Tags Teacher Request
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body TeacherReqBody true "Teacher request body"
// @Success 201 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /teacher-request [post]
func (h *TeacherRequestHandler) RequestTeacher(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req struct {
		Message string `json:"message"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(
			c,
			fiber.StatusBadRequest,
			"invalid request body",
			"BAD_REQUEST",
			nil,
		)
	}

	if err := h.teacherReqSvc.RequestTeacher(userID, req.Message); err != nil {
		return utils.Error(
			c,
			fiber.StatusBadRequest,
			err.Error(),
			"REQUEST_FAILED",
			nil,
		)
	}

	return utils.Success(
		c,
		fiber.StatusCreated,
		"teacher request submitted successfully",
		nil,
		nil,
	)
}

// GetMyRequest godoc
// @Summary Get my teacher request
// @Description Get the current user's teacher request status
// @Tags Teacher Request
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=entities.TeacherRequest}
// @Failure 404 {object} utils.ErrorResponse
// @Router /teacher-request/my [get]
func (h *TeacherRequestHandler) GetMyRequest(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	req, err := h.teacherReqSvc.GetMyRequest(userID)
	if err != nil {
		return utils.Error(
			c,
			fiber.StatusNotFound,
			"no teacher request found",
			"NOT_FOUND",
			nil,
		)
	}

	return utils.Success(
		c,
		fiber.StatusOK,
		"teacher request fetched successfully",
		req,
		nil,
	)
}

// ================= ADMIN =================

// GetPendingRequests godoc
// @Summary Get pending teacher requests (Admin)
// @Description Get all pending teacher requests
// @Tags Teacher Request Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=[]entities.TeacherRequest}
// @Failure 403 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /admin/teacher-requests [get]
func (h *TeacherRequestHandler) GetPendingRequests(c *fiber.Ctx) error {
	requests, err := h.teacherReqSvc.GetPendingRequests()
	if err != nil {
		return utils.Error(
			c,
			fiber.StatusInternalServerError,
			"failed to fetch teacher requests",
			"FETCH_FAILED",
			nil,
		)
	}

	return utils.Success(
		c,
		fiber.StatusOK,
		"teacher requests fetched successfully",
		requests,
		nil,
	)
}

// ApproveRequest godoc
// @Summary Approve teacher request (Admin)
// @Description Approve a pending teacher request
// @Tags Teacher Request Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Request ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /admin/teacher-requests/{id}/approve [post]
func (h *TeacherRequestHandler) ApproveRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	if id == "" {
		return utils.Error(
			c,
			fiber.StatusBadRequest,
			"id is required",
			"BAD_REQUEST",
			nil,
		)
	}

	if err := h.teacherReqSvc.ApproveRequest(id); err != nil {
		return utils.Error(
			c,
			fiber.StatusBadRequest,
			err.Error(),
			"APPROVE_FAILED",
			nil,
		)
	}

	return utils.Success(
		c,
		fiber.StatusOK,
		"teacher request approved successfully",
		nil,
		nil,
	)
}

// RejectRequest godoc
// @Summary Reject teacher request (Admin)
// @Description Reject a pending teacher request
// @Tags Teacher Request Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Request ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /admin/teacher-requests/{id}/reject [post]
func (h *TeacherRequestHandler) RejectRequest(c *fiber.Ctx) error {
	id := c.Params("id")

	if id == "" {
		return utils.Error(
			c,
			fiber.StatusBadRequest,
			"id is required",
			"BAD_REQUEST",
			nil,
		)
	}

	if err := h.teacherReqSvc.RejectRequest(id); err != nil {
		return utils.Error(
			c,
			fiber.StatusBadRequest,
			err.Error(),
			"REJECT_FAILED",
			nil,
		)
	}

	return utils.Success(
		c,
		fiber.StatusOK,
		"teacher request rejected successfully",
		nil,
		nil,
	)
}

// GetStats godoc
// @Summary Get teacher request statistics
// @Description Admin gets count of pending, approved, and rejected teacher requests
// @Tags Teacher Request Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=repositories.TeacherRequestStats}
// @Failure 403 {object} utils.ErrorResponse
// @Router /admin/teacher-requests/stats [get]
func (h *TeacherRequestHandler) GetStats(c *fiber.Ctx) error {
	stats, err := h.teacherReqSvc.GetStats()
	if err != nil {
		return utils.Error(
			c,
			fiber.StatusInternalServerError,
			err.Error(),
			"GET_STATS_FAILED",
			nil,
		)
	}

	return utils.Success(
		c,
		fiber.StatusOK,
		"teacher request stats fetched successfully",
		stats,
		nil,
	)
}

// ==================== REQUEST MODELS ====================

// TeacherReqBody represents teacher request body
type TeacherReqBody struct {
	Message string `json:"message" example:"Saya ingin menjadi guru di platform ini"`
}
