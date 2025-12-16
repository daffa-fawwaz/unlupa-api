package handlers

import (
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/usecases"
	"hifzhun-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authUC usecases.AuthUsecase
}

func NewAuthHandler(authUC usecases.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUC}
}

// ================= REGISTER =================
// POST /register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
		FullName string `json:"full_name"`
		School   string `json:"school"`
		Domicile string `json:"domicile"`
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

	user := &entities.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
		FullName: req.FullName,
		School:   req.School,
		Domicile: req.Domicile,
	}

	if err := h.authUC.Register(user); err != nil {
		return utils.Error(
			c,
			fiber.StatusBadRequest,
			err.Error(),
			"REGISTER_FAILED",
			nil,
		)
	}

	message := "registration success"
	if user.Role == "teacher" {
		message = "registration success, waiting admin approval"
	}

	return utils.Success(
		c,
		fiber.StatusCreated,
		message,
		fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
		nil,
	)
}

// ================= LOGIN =================
// POST /login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	user, token, err := h.authUC.Login(req.Email, req.Password)
	if err != nil {
		return utils.Error(
			c,
			fiber.StatusUnauthorized,
			err.Error(),
			"LOGIN_FAILED",
			nil,
		)
	}

	return utils.Success(
		c,
		fiber.StatusOK,
		"login success",
		fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
			"token": token,
		},
		nil,
	)
}

// ================= ADMIN APPROVE TEACHER =================
// PUT /admin/approve/:id
func (h *AuthHandler) ApproveTeacher(c *fiber.Ctx) error {
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

	if err := h.authUC.ApproveTeacher(id); err != nil {
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
		"teacher approved successfully",
		nil,
		nil,
	)
}
