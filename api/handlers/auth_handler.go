package handlers

import (
	"errors"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
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

// Register godoc
// @Summary Register new user
// @Description Register a new user account as student
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register request"
// @Success 201 {object} utils.SuccessResponse{data=RegisterResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
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

	if req.Email == "" || req.Password == "" || req.FullName == "" {
		return utils.Error(
			c,
			fiber.StatusBadRequest,
			"validation error",
			"VALIDATION_ERROR",
			[]utils.FieldError{
				{Field: "email", Messages: []string{"email is required"}},
				{Field: "password", Messages: []string{"password is required"}},
				{Field: "full_name", Messages: []string{"full name is required"}},
			},
		)
	}

	user := &entities.User{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Role:     "student",
		IsActive: true,
	}

	err := h.authUC.Register(user)
	if err != nil {

		if errors.Is(err, repositories.ErrEmailAlreadyExists) {
			return utils.Error(
				c,
				fiber.StatusBadRequest,
				"email already registered",
				"EMAIL_ALREADY_EXISTS",
				[]utils.FieldError{
					{
						Field:    "email",
						Messages: []string{"email already exists"},
					},
				},
			)
		}

		return utils.Error(
			c,
			fiber.StatusInternalServerError,
			"failed to register user",
			"REGISTER_FAILED",
			nil,
		)
	}

	return utils.Success(
		c,
		fiber.StatusCreated,
		"registration success",
		fiber.Map{
			"id":        user.ID,
			"email":     user.Email,
			"full_name": user.FullName,
			"role":      user.Role,
		},
		nil,
	)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} utils.SuccessResponse{data=LoginResponse}
// @Failure 401 {object} utils.ErrorResponse
// @Router /auth/login [post]
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
			"name":  user.FullName,
			"role":  user.Role,
			"token": token,
		},
		nil,
	)
}

// ==================== REQUEST/RESPONSE MODELS ====================

// RegisterRequest represents register request body
type RegisterRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
	FullName string `json:"full_name" example:"John Doe"`
}

// RegisterResponse represents register response data
type RegisterResponse struct {
	ID       string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email    string `json:"email" example:"user@example.com"`
	FullName string `json:"full_name" example:"John Doe"`
	Role     string `json:"role" example:"student"`
}

// LoginRequest represents login request body
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
}

// LoginResponse represents login response data
type LoginResponse struct {
	ID    string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email string `json:"email" example:"user@example.com"`
	Name  string `json:"name" example:"John Doe"`
	Role  string `json:"role" example:"student"`
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
