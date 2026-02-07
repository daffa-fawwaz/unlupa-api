package handlers

import (
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ClassHandler struct {
	classSvc services.ClassService
}

func NewClassHandler(classSvc services.ClassService) *ClassHandler {
	return &ClassHandler{classSvc}
}

// ==================== TEACHER ENDPOINTS ====================

// CreateClass godoc
// @Summary Create a new class
// @Description Teacher creates a new class (quran or book type)
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateClassRequest true "Create class request"
// @Success 201 {object} utils.SuccessResponse{data=entities.Class}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /classes [post]
func (h *ClassHandler) CreateClass(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"` // quran | book
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", "BAD_REQUEST", nil)
	}

	class, err := h.classSvc.CreateClass(userID, req.Name, req.Description, req.Type)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "CREATE_CLASS_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusCreated, "class created successfully", class, nil)
}

// GetMyClasses godoc
// @Summary Get teacher's classes
// @Description Get all classes owned by the teacher
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=[]entities.Class}
// @Failure 403 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /classes [get]
func (h *ClassHandler) GetMyClasses(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	classes, err := h.classSvc.GetMyClasses(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_CLASSES_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "classes fetched successfully", classes, nil)
}

// GetClassDetail godoc
// @Summary Get class detail
// @Description Get class detail with members and books
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class ID"
// @Success 200 {object} utils.SuccessResponse{data=entities.Class}
// @Failure 404 {object} utils.ErrorResponse
// @Router /classes/{id} [get]
func (h *ClassHandler) GetClassDetail(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	classID := c.Params("id")

	class, err := h.classSvc.GetClassDetail(classID, userID)
	if err != nil {
		return utils.Error(c, fiber.StatusNotFound, err.Error(), "CLASS_NOT_FOUND", nil)
	}

	return utils.Success(c, fiber.StatusOK, "class fetched successfully", class, nil)
}

// UpdateClass godoc
// @Summary Update class
// @Description Update class name, description, or active status
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class ID"
// @Param request body UpdateClassRequest true "Update class request"
// @Success 200 {object} utils.SuccessResponse{data=entities.Class}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /classes/{id} [put]
func (h *ClassHandler) UpdateClass(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	classID := c.Params("id")

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsActive    *bool  `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", "BAD_REQUEST", nil)
	}

	class, err := h.classSvc.UpdateClass(classID, userID, req.Name, req.Description, req.IsActive)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "UPDATE_CLASS_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "class updated successfully", class, nil)
}

// DeleteClass godoc
// @Summary Delete class
// @Description Delete a class and all its members/books
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /classes/{id} [delete]
func (h *ClassHandler) DeleteClass(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	classID := c.Params("id")

	if err := h.classSvc.DeleteClass(classID, userID); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "DELETE_CLASS_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "class deleted successfully", nil, nil)
}

// AddBookToClass godoc
// @Summary Add book to class
// @Description Add a book to book-type class (teacher only)
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class ID"
// @Param request body AddBookRequest true "Add book request"
// @Success 201 {object} utils.SuccessResponse{data=entities.ClassBook}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /classes/{id}/books [post]
func (h *ClassHandler) AddBookToClass(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	classID := c.Params("id")

	var req struct {
		BookID string `json:"book_id"`
		Order  int    `json:"order"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", "BAD_REQUEST", nil)
	}

	classBook, err := h.classSvc.AddBookToClass(classID, userID, req.BookID, req.Order)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "ADD_BOOK_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusCreated, "book added to class successfully", classBook, nil)
}

// RemoveBookFromClass godoc
// @Summary Remove book from class
// @Description Remove a book from class (teacher only)
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class ID"
// @Param book_id path string true "Book ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /classes/{id}/books/{book_id} [delete]
func (h *ClassHandler) RemoveBookFromClass(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	classID := c.Params("id")
	bookID := c.Params("book_id")

	if err := h.classSvc.RemoveBookFromClass(classID, userID, bookID); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "REMOVE_BOOK_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "book removed from class successfully", nil, nil)
}

// GetStudentProgress godoc
// @Summary Get student progress
// @Description Get progress of all students in quran-type class
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class ID"
// @Success 200 {object} utils.SuccessResponse{data=[]services.StudentProgress}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /classes/{id}/progress [get]
func (h *ClassHandler) GetStudentProgress(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	classID := c.Params("id")

	progress, err := h.classSvc.GetStudentProgress(classID, userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "GET_PROGRESS_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "student progress fetched successfully", progress, nil)
}

// GetClassMembers godoc
// @Summary Get class members
// @Description Get all members of a class (teacher only)
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class ID"
// @Success 200 {object} utils.SuccessResponse{data=[]services.MemberInfo}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /classes/{id}/members [get]
func (h *ClassHandler) GetClassMembers(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	classID := c.Params("id")

	members, err := h.classSvc.GetClassMembers(classID, userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "GET_MEMBERS_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "class members fetched successfully", members, nil)
}

// ==================== STUDENT ENDPOINTS ====================

// JoinClass godoc
// @Summary Join class with code
// @Description Student joins a class using class code
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body JoinClassRequest true "Join class request"
// @Success 200 {object} utils.SuccessResponse{data=entities.Class}
// @Failure 400 {object} utils.ErrorResponse
// @Router /classes/join [post]
func (h *ClassHandler) JoinClass(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req struct {
		Code string `json:"code"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", "BAD_REQUEST", nil)
	}

	class, err := h.classSvc.JoinClass(userID, req.Code)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "JOIN_CLASS_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "joined class successfully", class, nil)
}

// LeaveClass godoc
// @Summary Leave class
// @Description Student leaves a class
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /classes/{id}/leave [delete]
func (h *ClassHandler) LeaveClass(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	classID := c.Params("id")

	if err := h.classSvc.LeaveClass(userID, classID); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "LEAVE_CLASS_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "left class successfully", nil, nil)
}

// GetMyJoinedClasses godoc
// @Summary Get joined classes
// @Description Get all classes the student has joined
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=[]entities.Class}
// @Failure 500 {object} utils.ErrorResponse
// @Router /classes/joined [get]
func (h *ClassHandler) GetMyJoinedClasses(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	classes, err := h.classSvc.GetMyJoinedClasses(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_CLASSES_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "joined classes fetched successfully", classes, nil)
}

// GetClassBooks godoc
// @Summary Get class books
// @Description Get all books in a book-type class
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class ID"
// @Success 200 {object} utils.SuccessResponse{data=[]entities.ClassBook}
// @Failure 400 {object} utils.ErrorResponse
// @Router /classes/{id}/books [get]
func (h *ClassHandler) GetClassBooks(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	classID := c.Params("id")

	books, err := h.classSvc.GetClassBooks(classID, userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "GET_BOOKS_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "class books fetched successfully", books, nil)
}

// ==================== REQUEST/RESPONSE MODELS FOR SWAGGER ====================

// CreateClassRequest represents create class request body
type CreateClassRequest struct {
	Name        string `json:"name" example:"Kelas Tajwid A"`
	Description string `json:"description" example:"Belajar tajwid dari dasar"`
	Type        string `json:"type" example:"book" enums:"quran,book"`
}

// UpdateClassRequest represents update class request body
type UpdateClassRequest struct {
	Name        string `json:"name" example:"Kelas Tajwid B"`
	Description string `json:"description" example:"Updated description"`
	IsActive    *bool  `json:"is_active" example:"true"`
}

// AddBookRequest represents add book to class request body
type AddBookRequest struct {
	BookID string `json:"book_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Order  int    `json:"order" example:"1"`
}

// JoinClassRequest represents join class request body
type JoinClassRequest struct {
	Code string `json:"code" example:"AB3K7X"`
}

// ==================== GRADUATION APPROVAL ENDPOINTS ====================

// GetPendingGraduations godoc
// @Summary Get pending graduations
// @Description Teacher gets all items pending graduation approval in a class
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class ID"
// @Success 200 {object} utils.SuccessResponse{data=[]services.PendingGraduation}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /classes/{id}/graduations/pending [get]
func (h *ClassHandler) GetPendingGraduations(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	classID := c.Params("id")

	pending, err := h.classSvc.GetPendingGraduations(classID, userID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "GET_PENDING_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "pending graduations fetched successfully", pending, nil)
}

// ApproveGraduation godoc
// @Summary Approve graduation
// @Description Teacher approves an item for graduation
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class ID"
// @Param item_id path string true "Item ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /classes/{id}/graduations/{item_id}/approve [post]
func (h *ClassHandler) ApproveGraduation(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	classID := c.Params("id")
	itemID := c.Params("item_id")

	if err := h.classSvc.ApproveGraduation(classID, userID, itemID); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "APPROVE_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "graduation approved successfully", nil, nil)
}

// RejectGraduation godoc
// @Summary Reject graduation
// @Description Teacher rejects an item for graduation (returns to fsrs_active)
// @Tags Class
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class ID"
// @Param item_id path string true "Item ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /classes/{id}/graduations/{item_id}/reject [post]
func (h *ClassHandler) RejectGraduation(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	classID := c.Params("id")
	itemID := c.Params("item_id")

	if err := h.classSvc.RejectGraduation(classID, userID, itemID); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "REJECT_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "graduation rejected, item returned to fsrs_active", nil, nil)
}
