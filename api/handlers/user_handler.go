package handlers

import (
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userSvc services.UserService
}

func NewUserHandler(userSvc services.UserService) *UserHandler {
	return &UserHandler{userSvc}
}

// ================= ADMIN =================

// GET /api/v1/admin/users?role=teacher
func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	role := c.Query("role")

	users, err := h.userSvc.GetAllUsers(role)
	if err != nil {
		return utils.Error(
			c,
			fiber.StatusInternalServerError,
			"failed to fetch users",
			"FETCH_USERS_FAILED",
			nil,
		)
	}

	return utils.Success(
		c,
		fiber.StatusOK,
		"users fetched successfully",
		users,
		nil,
	)
}

// POST /api/v1/admin/users/:id/activate
func (h *UserHandler) ActivateUser(c *fiber.Ctx) error {
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

	if err := h.userSvc.ActivateUser(id); err != nil {
		return utils.Error(
			c,
			fiber.StatusBadRequest,
			err.Error(),
			"ACTIVATE_FAILED",
			nil,
		)
	}

	return utils.Success(
		c,
		fiber.StatusOK,
		"user activated successfully",
		nil,
		nil,
	)
}

// POST /api/v1/admin/users/:id/deactivate
func (h *UserHandler) DeactivateUser(c *fiber.Ctx) error {
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

	if err := h.userSvc.DeactivateUser(id); err != nil {
		return utils.Error(
			c,
			fiber.StatusBadRequest,
			err.Error(),
			"DEACTIVATE_FAILED",
			nil,
		)
	}

	return utils.Success(
		c,
		fiber.StatusOK,
		"user deactivated successfully",
		nil,
		nil,
	)
}


