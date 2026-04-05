package handlers

import (
	"time"

	"hifzhun-api/pkg/cache"
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/services"
	"hifzhun-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type BookHandler struct {
	bookSvc services.BookService
	cache   *cache.Cache
}

func NewBookHandler(bookSvc services.BookService, c *cache.Cache) *BookHandler {
	return &BookHandler{bookSvc: bookSvc, cache: c}
}

// ==================== BOOK ENDPOINTS ====================

// CreateBook godoc
// @Summary Create a new book
// @Description Create a new book (draft status)
// @Tags Book
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateBookRequest true "Create book request"
// @Success 201 {object} utils.SuccessResponse{data=entities.Book}
// @Failure 400 {object} utils.ErrorResponse
// @Router /books [post]
func (h *BookHandler) CreateBook(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		CoverImage  string `json:"cover_image"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", "BAD_REQUEST", nil)
	}

	book, err := h.bookSvc.CreateBook(userID, req.Title, req.Description, req.CoverImage)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "CREATE_BOOK_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusCreated, "book created successfully", book, nil)
}

// GetMyBooks godoc
// @Summary Get my books
// @Description Get all books owned by the user
// @Tags Book
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=[]entities.Book}
// @Failure 500 {object} utils.ErrorResponse
// @Router /books [get]
func (h *BookHandler) GetMyBooks(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	books, err := h.bookSvc.GetMyBooks(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_BOOKS_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "books fetched successfully", books, nil)
}

// GetPublishedBooks godoc
// @Summary Get published books
// @Description Get all published/approved books
// @Tags Book
// @Accept json
// @Produce json
// @Success 200 {object} utils.SuccessResponse{data=[]entities.Book}
// @Failure 500 {object} utils.ErrorResponse
// @Router /books/published [get]
func (h *BookHandler) GetPublishedBooks(c *fiber.Ctx) error {
	// Try cache first
	cacheKey := "books:published"
	var cached []entities.Book
	if h.cache.Get(c.Context(), cacheKey, &cached) {
		return utils.Success(c, fiber.StatusOK, "published books fetched successfully", cached, nil)
	}

	books, err := h.bookSvc.GetPublishedBooks()
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_BOOKS_FAILED", nil)
	}

	h.cache.Set(c.Context(), cacheKey, books, 30*time.Minute)

	return utils.Success(c, fiber.StatusOK, "published books fetched successfully", books, nil)
}

// GetBookDetail godoc
// @Summary Get book detail
// @Description Get book detail with modules and items
// @Tags Book
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {object} utils.SuccessResponse{data=services.BookDetailWithStability}
// @Failure 404 {object} utils.ErrorResponse
// @Router /books/{id} [get]
func (h *BookHandler) GetBookDetail(c *fiber.Ctx) error {
	bookID := c.Params("id")
	userIDInterface := c.Locals("user_id")
	userRole := c.Locals("role")

	var userID *uuid.UUID
	if userIDInterface != nil {
		id := userIDInterface.(uuid.UUID)
		userID = &id
	}

	var role string
	if userRole != nil {
		role = userRole.(string)
	}

	book, err := h.bookSvc.GetBookDetailWithStability(bookID, userID, role)
	if err != nil {
		return utils.Error(c, fiber.StatusNotFound, err.Error(), "BOOK_NOT_FOUND", nil)
	}

	return utils.Success(c, fiber.StatusOK, "book fetched successfully", book, nil)
}

// GetPublishedBookDetail godoc
// @Summary Get published book detail
// @Description Get book detail for a book that must be in published status
// @Tags Book
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {object} utils.SuccessResponse{data=entities.Book}
// @Failure 404 {object} utils.ErrorResponse
// @Router /books/published/{id} [get]
func (h *BookHandler) GetPublishedBookDetail(c *fiber.Ctx) error {
	bookID := c.Params("id")

	book, err := h.bookSvc.GetPublishedBookDetail(bookID)
	if err != nil {
		return utils.Error(c, fiber.StatusNotFound, err.Error(), "BOOK_NOT_FOUND", nil)
	}

	return utils.Success(c, fiber.StatusOK, "published book fetched successfully", book, nil)
}

// UpdateBook godoc
// @Summary Update book
// @Description Update book title, description, or cover image
// @Tags Book
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Param request body UpdateBookRequest true "Update book request"
// @Success 200 {object} utils.SuccessResponse{data=entities.Book}
// @Failure 400 {object} utils.ErrorResponse
// @Router /books/{id} [put]
func (h *BookHandler) UpdateBook(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	bookID := c.Params("id")

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		CoverImage  string `json:"cover_image"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", "BAD_REQUEST", nil)
	}

	book, err := h.bookSvc.UpdateBook(bookID, userID, req.Title, req.Description, req.CoverImage)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "UPDATE_BOOK_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "book updated successfully", book, nil)
}

// DeleteBook godoc
// @Summary Delete book
// @Description Delete a book and all its modules/items
// @Tags Book
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /books/{id} [delete]
func (h *BookHandler) DeleteBook(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	bookID := c.Params("id")

	if err := h.bookSvc.DeleteBook(bookID, userID); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "DELETE_BOOK_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "book deleted successfully", nil, nil)
}

// RequestPublish godoc
// @Summary Request book publish
// @Description Submit book for admin approval
// @Tags Book
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /books/{id}/request-publish [post]
func (h *BookHandler) RequestPublish(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	bookID := c.Params("id")

	if err := h.bookSvc.RequestPublish(bookID, userID); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "REQUEST_PUBLISH_FAILED", nil)
	}

	// Invalidate published books cache
	h.cache.Delete(c.Context(), "books:published")

	return utils.Success(c, fiber.StatusOK, "publish request submitted successfully", nil, nil)
}

// AddPublishedBookToMyBook godoc
// @Summary Add a published book into my book items
// @Description Create Item rows for each BookItem in the published book (preparing user memorization items)
// @Tags Book
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {object} utils.SuccessResponse{data=services.AddPublishedBookToMyBookResult}
// @Failure 400 {object} utils.ErrorResponse
// @Router /books/published/{id}/add-to-my-books [post]
func (h *BookHandler) AddPublishedBookToMyBook(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	bookID := c.Params("id")

	result, err := h.bookSvc.AddPublishedBookToMyBook(userID, bookID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "ADD_PUBLISHED_BOOK_FAILED", nil)
	}

	// Invalidate my-items cache so the newly added items show up.
	h.cache.Delete(c.Context(), "myitems:"+userID.String()+":book")
	h.cache.Delete(c.Context(), "myitems:"+userID.String()+":all")

	return utils.Success(c, fiber.StatusOK, "published book added to my book successfully", result, nil)
}

// CopyPublishedBookToDraft godoc
// @Summary Copy published book to my draft
// @Description Copy published book structure (modules & items) into a new draft owned by the user
// @Tags Book
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Published Book ID"
// @Param request body CopyPublishedBookToDraftRequest true "Copy request"
// @Success 200 {object} utils.SuccessResponse{data=entities.Book}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /books/published/{id}/copy-to-draft [post]
func (h *BookHandler) CopyPublishedBookToDraft(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	publishedBookID := c.Params("id")

	var req CopyPublishedBookToDraftRequest
	if err := c.BodyParser(&req); err != nil {
		// Body can be empty; BodyParser fails for empty body in some cases.
		// Treat empty body as default values (use source book values).
		req = CopyPublishedBookToDraftRequest{}
	}

	book, err := h.bookSvc.CopyPublishedBookToDraft(userID, publishedBookID, req.Title, req.Description, req.CoverImage)
	if err != nil {
		// If the service says "book not found"/invalid status, treat as 404/400.
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "COPY_PUBLISHED_BOOK_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "published book copied to draft successfully", book, nil)
}

// ==================== ADMIN ENDPOINTS ====================

// GetPendingBooks godoc
// @Summary Get pending books (Admin)
// @Description Get all books pending for approval
// @Tags Book Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=[]entities.Book}
// @Failure 403 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /admin/books/pending [get]
func (h *BookHandler) GetPendingBooks(c *fiber.Ctx) error {
	books, err := h.bookSvc.GetPendingBooks()
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_BOOKS_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "pending books fetched successfully", books, nil)
}

// ApproveBook godoc
// @Summary Approve book (Admin)
// @Description Approve a pending book for publishing
// @Tags Book Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /admin/books/{id}/approve [post]
func (h *BookHandler) ApproveBook(c *fiber.Ctx) error {
	bookID := c.Params("id")

	if err := h.bookSvc.ApproveBook(bookID); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "APPROVE_BOOK_FAILED", nil)
	}

	// Invalidate published books cache
	h.cache.Delete(c.Context(), "books:published")

	return utils.Success(c, fiber.StatusOK, "book approved successfully", nil, nil)
}

// RejectBook godoc
// @Summary Reject book (Admin)
// @Description Reject a pending book
// @Tags Book Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /admin/books/{id}/reject [post]
func (h *BookHandler) RejectBook(c *fiber.Ctx) error {
	bookID := c.Params("id")

	if err := h.bookSvc.RejectBook(bookID); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "REJECT_BOOK_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "book rejected successfully", nil, nil)
}

// DeletePublishedBook godoc
// @Summary Delete published book (Admin)
// @Description Delete a published book and all its modules/items
// @Tags Book Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Router /admin/books/{id} [delete]
func (h *BookHandler) DeletePublishedBook(c *fiber.Ctx) error {
	bookID := c.Params("id")

	if err := h.bookSvc.DeletePublishedBook(bookID); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "DELETE_BOOK_FAILED", nil)
	}

	// Invalidate published books cache
	h.cache.Delete(c.Context(), "books:published")

	return utils.Success(c, fiber.StatusOK, "book deleted successfully", nil, nil)
}

// GetBookDetailForAdmin godoc
// @Summary Get book detail (Admin)
// @Description Get any user's book detail with modules and items (Admin only)
// @Tags Book Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {object} utils.SuccessResponse{data=entities.Book}
// @Failure 404 {object} utils.ErrorResponse
// @Router /admin/books/{id} [get]
func (h *BookHandler) GetBookDetailForAdmin(c *fiber.Ctx) error {
	bookID := c.Params("id")

	book, err := h.bookSvc.GetBookDetailForAdmin(bookID)
	if err != nil {
		return utils.Error(c, fiber.StatusNotFound, err.Error(), "BOOK_NOT_FOUND", nil)
	}

	return utils.Success(c, fiber.StatusOK, "book fetched successfully", book, nil)
}

// ==================== MODULE ENDPOINTS ====================

// AddModule godoc
// @Summary Add module to book
// @Description Add a new module to a book
// @Tags Book Module
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param book_id path string true "Book ID"
// @Param request body AddModuleRequest true "Add module request"
// @Success 201 {object} utils.SuccessResponse{data=entities.BookModule}
// @Failure 400 {object} utils.ErrorResponse
// @Router /books/{book_id}/modules [post]
func (h *BookHandler) AddModule(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	bookID := c.Params("id")

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Order       int    `json:"order"`
		ParentID    string `json:"parent_id"` // optional UUID string
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", "BAD_REQUEST", nil)
	}

	var parentPtr *uuid.UUID
	if req.ParentID != "" {
		pid, err := uuid.Parse(req.ParentID)
		if err != nil {
			return utils.Error(c, fiber.StatusBadRequest, "invalid parent_id", "BAD_REQUEST", nil)
		}
		parentPtr = &pid
	}

	module, err := h.bookSvc.AddModule(bookID, userID, req.Title, req.Description, req.Order, parentPtr)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "ADD_MODULE_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusCreated, "module added successfully", module, nil)
}

// GetBookTree godoc
// @Summary Get book module tree
// @Description Get hierarchical modules (with items) for a book
// @Tags Book
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /books/{id}/tree [get]
func (h *BookHandler) GetBookTree(c *fiber.Ctx) error {
	bookID := c.Params("id")
	userIDInterface := c.Locals("user_id")
	userRole := c.Locals("role")

	var userID *uuid.UUID
	if userIDInterface != nil {
		id := userIDInterface.(uuid.UUID)
		userID = &id
	}

	var role string
	if userRole != nil {
		role = userRole.(string)
	}

	tree, err := h.bookSvc.GetBookTree(bookID, userID, role)
	if err != nil {
		return utils.Error(c, fiber.StatusNotFound, err.Error(), "BOOK_NOT_FOUND", nil)
	}

	return utils.Success(c, fiber.StatusOK, "book tree fetched successfully", tree, nil)
}

// UpdateModule godoc
// @Summary Update module
// @Description Update a book module
// @Tags Book Module
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Module ID"
// @Param request body UpdateModuleRequest true "Update module request"
// @Success 200 {object} utils.SuccessResponse{data=entities.BookModule}
// @Failure 400 {object} utils.ErrorResponse
// @Router /books/modules/{id} [put]
func (h *BookHandler) UpdateModule(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	moduleID := c.Params("id")

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Order       int    `json:"order"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", "BAD_REQUEST", nil)
	}

	module, err := h.bookSvc.UpdateModule(moduleID, userID, req.Title, req.Description, req.Order)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "UPDATE_MODULE_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "module updated successfully", module, nil)
}

// DeleteModule godoc
// @Summary Delete module
// @Description Delete a book module
// @Tags Book Module
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Module ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /books/modules/{id} [delete]
func (h *BookHandler) DeleteModule(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	moduleID := c.Params("id")

	if err := h.bookSvc.DeleteModule(moduleID, userID); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "DELETE_MODULE_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "module deleted successfully", nil, nil)
}

// ==================== ITEM ENDPOINTS ====================

// AddItemToBook godoc
// @Summary Add item to book
// @Description Add a new item directly to a book
// @Tags Book Item
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param book_id path string true "Book ID"
// @Param request body AddBookItemRequest true "Add item request"
// @Success 201 {object} utils.SuccessResponse{data=entities.BookItem}
// @Failure 400 {object} utils.ErrorResponse
// @Router /books/{book_id}/items [post]
func (h *BookHandler) AddItemToBook(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	bookID := c.Params("id")

	var req struct {
		Title         string `json:"title"`
		Content       string `json:"content"`
		Answer        string `json:"answer"`
		Order         int    `json:"order"`
		EstimateValue int    `json:"estimate_value"`
		EstimateUnit  string `json:"estimate_unit"` // "seconds" | "minutes"
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", "BAD_REQUEST", nil)
	}

	item, err := h.bookSvc.AddItem(bookID, nil, userID, req.Title, req.Content, req.Answer, req.Order, req.EstimateValue, req.EstimateUnit)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "ADD_ITEM_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusCreated, "item added successfully", item, nil)
}

// AddItemToModule godoc
// @Summary Add item to module
// @Description Add a new item to a book module
// @Tags Book Item
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param module_id path string true "Module ID"
// @Param request body AddModuleItemRequest true "Add item request"
// @Success 201 {object} utils.SuccessResponse{data=entities.BookItem}
// @Failure 400 {object} utils.ErrorResponse
// @Router /books/modules/{module_id}/items [post]
func (h *BookHandler) AddItemToModule(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	moduleIDStr := c.Params("module_id")

	moduleID, err := uuid.Parse(moduleIDStr)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid module id", "BAD_REQUEST", nil)
	}

	var req struct {
		BookID        string `json:"book_id"`
		Title         string `json:"title"`
		Content       string `json:"content"`
		Answer        string `json:"answer"`
		Order         int    `json:"order"`
		EstimateValue int    `json:"estimate_value"`
		EstimateUnit  string `json:"estimate_unit"` // "seconds" | "minutes"
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", "BAD_REQUEST", nil)
	}

	item, err := h.bookSvc.AddItem(req.BookID, &moduleID, userID, req.Title, req.Content, req.Answer, req.Order, req.EstimateValue, req.EstimateUnit)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "ADD_ITEM_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusCreated, "item added successfully", item, nil)
}

// UpdateItem godoc
// @Summary Update book item
// @Description Update a book item
// @Tags Book Item
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Item ID"
// @Param request body UpdateBookItemRequest true "Update item request"
// @Success 200 {object} utils.SuccessResponse{data=entities.BookItem}
// @Failure 400 {object} utils.ErrorResponse
// @Router /books/items/{id} [put]
func (h *BookHandler) UpdateItem(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	itemID := c.Params("id")

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Answer  string `json:"answer"`
		Order   int    `json:"order"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "invalid request body", "BAD_REQUEST", nil)
	}

	item, err := h.bookSvc.UpdateItem(itemID, userID, req.Title, req.Content, req.Answer, req.Order)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "UPDATE_ITEM_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "item updated successfully", item, nil)
}

// DeleteItem godoc
// @Summary Delete book item
// @Description Delete a book item
// @Tags Book Item
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Item ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /books/items/{id} [delete]
func (h *BookHandler) DeleteItem(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	itemID := c.Params("id")

	if err := h.bookSvc.DeleteItem(itemID, userID); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "DELETE_ITEM_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "item deleted successfully", nil, nil)
}

// ==================== REQUEST MODELS ====================

// CreateBookRequest represents create book request
type CreateBookRequest struct {
	Title       string `json:"title" example:"Belajar Tajwid"`
	Description string `json:"description" example:"Panduan belajar tajwid"`
	CoverImage  string `json:"cover_image" example:"https://example.com/cover.jpg"`
}

// UpdateBookRequest represents update book request
type UpdateBookRequest struct {
	Title       string `json:"title" example:"Belajar Tajwid Lengkap"`
	Description string `json:"description" example:"Panduan lengkap"`
	CoverImage  string `json:"cover_image" example:"https://example.com/cover2.jpg"`
}

// AddModuleRequest represents add module request
type AddModuleRequest struct {
	Title       string `json:"title" example:"Bab 1: Pengenalan"`
	Description string `json:"description" example:"Pengenalan tajwid"`
	Order       int    `json:"order" example:"1"`
	ParentID    string `json:"parent_id,omitempty" example:""`
}

type CopyPublishedBookToDraftRequest struct {
	Title       string `json:"title,omitempty" example:"Belajar Tajwid (Copy)"`
	Description string `json:"description,omitempty" example:"Panduan belajar tajwid (versi draft)"`
	CoverImage  string `json:"cover_image,omitempty" example:"https://example.com/cover.jpg"`
}

// UpdateModuleRequest represents update module request
type UpdateModuleRequest struct {
	Title       string `json:"title" example:"Bab 1: Intro"`
	Description string `json:"description" example:"Updated desc"`
	Order       int    `json:"order" example:"1"`
}

// AddBookItemRequest represents add item to book request
type AddBookItemRequest struct {
	Title         string `json:"title" example:"Hukum Nun Mati"`
	Content       string `json:"content" example:"Penjelasan hukum nun mati..."`
	Answer        string `json:"answer" example:"Jawaban dari pertanyaan..."`
	Order         int    `json:"order" example:"1"`
	EstimateValue int    `json:"estimate_value" example:"2"`
	EstimateUnit  string `json:"estimate_unit" example:"minutes"`
}

// AddModuleItemRequest represents add item to module request
type AddModuleItemRequest struct {
	BookID        string `json:"book_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title         string `json:"title" example:"Sub Item"`
	Content       string `json:"content" example:"Konten item..."`
	Answer        string `json:"answer" example:"Jawaban..."`
	Order         int    `json:"order" example:"1"`
	EstimateValue int    `json:"estimate_value" example:"90"`
	EstimateUnit  string `json:"estimate_unit" example:"seconds"`
}

// UpdateBookItemRequest represents update item request
type UpdateBookItemRequest struct {
	Title   string `json:"title" example:"Updated Title"`
	Content string `json:"content" example:"Updated content..."`
	Answer  string `json:"answer" example:"Updated answer..."`
	Order   int    `json:"order" example:"2"`
}

// ==================== MEMORIZATION ====================

// StartMemorization godoc
// @Summary Start memorizing a book item
// @Description User starts memorizing a specific item from a book (published or owned)
// @Tags Book
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Param item_id path string true "Book Item ID"
// @Success 200 {object} utils.SuccessResponse{data=services.StartMemorizationResult}
// @Failure 400 {object} utils.ErrorResponse
// @Router /books/{id}/items/{item_id}/start [post]
func (h *BookHandler) StartMemorization(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	bookID := c.Params("id")
	bookItemID := c.Params("item_id")

	result, err := h.bookSvc.StartItemMemorization(userID, bookID, bookItemID)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "START_MEMORIZATION_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "Item memorization started", result, nil)
}

// ==================== MY BOOK COLLECTION ====================

// GetMyBookCollection godoc
// @Summary Get my book collection
// @Description Get all published books added to user's collection (read-only books from other users)
// @Tags My Book Collection
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=[]services.BookCollectionItem}
// @Failure 500 {object} utils.ErrorResponse
// @Router /books/my-collection [get]
func (h *BookHandler) GetMyBookCollection(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	books, err := h.bookSvc.GetMyBookCollection(userID)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, err.Error(), "GET_COLLECTION_FAILED", nil)
	}

	return utils.Success(c, fiber.StatusOK, "book collection fetched successfully", books, nil)
}

// RemoveFromMyBookCollection godoc
// @Summary Remove book from my collection
// @Description Remove a book from user's collection (deletes all memorization items from that book)
// @Tags My Book Collection
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /books/my-collection/{id} [delete]
func (h *BookHandler) RemoveFromMyBookCollection(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	bookID := c.Params("id")

	if err := h.bookSvc.RemoveFromMyBookCollection(userID, bookID); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, err.Error(), "REMOVE_FROM_COLLECTION_FAILED", nil)
	}

	// Invalidate cache
	h.cache.Delete(c.Context(), "myitems:"+userID.String()+":book")
	h.cache.Delete(c.Context(), "myitems:"+userID.String()+":all")

	return utils.Success(c, fiber.StatusOK, "book removed from collection successfully", nil, nil)
}
