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

// POST /api/v1/teacher-request
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

// GET /api/v1/teacher-request/my
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

// GET /api/v1/admin/teacher-requests
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

// POST /api/v1/admin/teacher-requests/:id/approve
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

// POST /api/v1/admin/teacher-requests/:id/reject
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
