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

// GetAllUsers godoc
// @Summary Get all users (Admin)
// @Description Get all users, optionally filtered by role
// @Tags User Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param role query string false "Filter by role (student, teacher, admin)"
// @Success 200 {object} utils.SuccessResponse{data=[]entities.User}
// @Failure 403 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /admin/users [get]
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

// ActivateUser godoc
// @Summary Activate user (Admin)
// @Description Activate a user account
// @Tags User Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /admin/users/{id}/activate [post]
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

// DeactivateUser godoc
// @Summary Deactivate user (Admin)
// @Description Deactivate a user account
// @Tags User Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /admin/users/{id}/deactivate [post]
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
