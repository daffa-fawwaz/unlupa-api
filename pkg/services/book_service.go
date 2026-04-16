package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"

	"github.com/google/uuid"
)

type BookService interface {
	// Book CRUD
	CreateBook(ownerID uuid.UUID, title, description, coverImage string) (*entities.Book, error)
	GetMyBooks(ownerID uuid.UUID) ([]entities.Book, error)
	GetPublishedBooks() ([]entities.Book, error)
	GetPublishedBookDetail(bookID string) (*entities.Book, error)
	GetBookDetail(bookID string, userID *uuid.UUID, role string) (*entities.Book, error)
	GetBookDetailWithStability(bookID string, userID *uuid.UUID, role string) (*BookDetailWithStability, error)
	GetBookDetailForAdmin(bookID string) (*entities.Book, error)
	GetBookTree(bookID string, userID *uuid.UUID, role string) (*BookTreeResponse, error)
	UpdateBook(bookID string, ownerID uuid.UUID, title, description, coverImage string) (*entities.Book, error)
	DeleteBook(bookID string, ownerID uuid.UUID) error

	// Publish workflow
	RequestPublish(bookID string, ownerID uuid.UUID) error
	GetPendingBooks() ([]entities.Book, error)
	ApproveBook(bookID string) error
	RejectBook(bookID string) error
	DeletePublishedBook(bookID string) error

	// Book update requests (for published books)
	RequestBookUpdate(bookID string, ownerID uuid.UUID, title, description, coverImage string) (*entities.BookUpdateRequest, error)
	GetBookUpdateRequests(bookID string) ([]entities.BookUpdateRequest, error)
	ApproveBookUpdate(requestID string, adminID uuid.UUID) error
	RejectBookUpdate(requestID string, adminID uuid.UUID, reason string) error
	GetPendingBookUpdates() ([]entities.BookUpdateRequest, error)

	// Module CRUD
	AddModule(bookID string, ownerID uuid.UUID, title, description string, order int, parentID *uuid.UUID) (*entities.BookModule, error)
	UpdateModule(moduleID string, ownerID uuid.UUID, title, description string, order int) (*entities.BookModule, error)
	DeleteModule(moduleID string, ownerID uuid.UUID) error

	// Item CRUD
	AddItem(bookID string, moduleID *uuid.UUID, ownerID uuid.UUID, title, content, answer string, order int, estimateVal int, estimateUnit string) (*entities.BookItem, error)
	UpdateItem(itemID string, ownerID uuid.UUID, title, content, answer string, order int) (*entities.BookItem, error)
	DeleteItem(itemID string, ownerID uuid.UUID) error

	// Memorization
	StartItemMemorization(userID uuid.UUID, bookID, bookItemID string) (*StartMemorizationResult, error)

	// Add published book into user's "my book items" (creates Item rows for each BookItem)
	AddPublishedBookToMyBook(userID uuid.UUID, bookID string) (*AddPublishedBookToMyBookResult, error)

	// Copy published book structure into a new draft owned by the user
	CopyPublishedBookToDraft(userID uuid.UUID, publishedBookID string, title, description, coverImage string) (*entities.Book, error)

	// My Book Collection
	GetMyBookCollection(userID uuid.UUID) ([]BookCollectionItem, error)
	RemoveFromMyBookCollection(userID uuid.UUID, bookID string) error
}

// BookItemWithStability represents a BookItem with stability information
type BookItemWithStability struct {
	entities.BookItem
	Stability string `json:"stability"` // "item belum masuk ujian" or days until next review
}

// BookDetailWithStability represents book detail with stability information for items
type BookDetailWithStability struct {
	entities.Book
	Items   []BookItemWithStability `json:"items"`
	Modules []ModuleWithStability   `json:"modules"`
}

type ModuleWithStability struct {
	entities.BookModule
	Items    []BookItemWithStability `json:"items"`
	Children []ModuleWithStability   `json:"children"`
}

// calculateStability calculates stability based on Item status and review dates
// Returns "item belum masuk ujian" if status is start or no review data, otherwise calculates days until next review
func calculateStability(item *entities.Item) string {
	if item == nil {
		return "item belum masuk ujian"
	}

	// If status is start, item hasn't entered FSRS phase yet
	if item.Status == entities.ItemStatusStart {
		return "item belum masuk ujian"
	}

	// For items in FSRS or other phases with NextReviewAt, calculate days until next review
	if item.NextReviewAt != nil && item.LastReviewAt != nil {
		now := time.Now().In(config.AppLocation)
		// Calculate days from last review to next review (total interval)
		totalInterval := item.NextReviewAt.Sub(*item.LastReviewAt).Hours() / 24
		// Calculate days elapsed since last review
		elapsed := now.Sub(*item.LastReviewAt).Hours() / 24
		// Calculate remaining days until next review
		remaining := totalInterval - elapsed
		
		if remaining < 0 {
			remaining = 0
		}
		
		// Return as string representation of integer days
		return fmt.Sprintf("%.0f", remaining)
	}

	return "item belum masuk ujian"
}

// BookTreeResponse represents hierarchical modules and items for a book
type BookTreeResponse struct {
	BookID  string                 `json:"book_id"`
	Title   string                 `json:"title"`
	Items   []BookItemWithStability `json:"items"` // items directly under book
	Modules []ModuleNodeWithItems  `json:"modules"`
}

type ModuleNodeWithItems struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Order       int                    `json:"order"`
	Items       []BookItemWithStability `json:"items"`
	Children    []ModuleNodeWithItems  `json:"children"`
}

// StartMemorizationResult represents the result of starting book item memorization
type StartMemorizationResult struct {
	ItemID     uuid.UUID `json:"item_id"`
	BookItemID uuid.UUID `json:"book_item_id"`
	BookTitle  string    `json:"book_title"`
	ItemTitle  string    `json:"item_title"`
	Status     string    `json:"status"`
}

type AddPublishedBookToMyBookResult struct {
	BookID           string   `json:"book_id"`
	AddedCount       int      `json:"added_count"`
	SkippedCount     int      `json:"skipped_count"`
	AddedContentRefs []string `json:"added_content_refs,omitempty"`
}

// BookCollectionItem represents a book in user's collection
type BookCollectionItem struct {
	BookID        string    `json:"book_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	CoverImage    string    `json:"cover_image,omitempty"`
	OwnerName     string    `json:"owner_name,omitempty"`
	ItemCount     int       `json:"item_count"`
	AddedAt       string    `json:"added_at"`
}

type bookService struct {
	bookRepo               repositories.BookRepository
	bookModuleRepo         repositories.BookModuleRepository
	bookItemRepo           repositories.BookItemRepository
	itemRepo               *repositories.ItemRepository
	userRepo               repositories.UserRepository
	updateRequestRepo      *repositories.BookUpdateRequestRepository
}

func NewBookService(
	bookRepo repositories.BookRepository,
	bookModuleRepo repositories.BookModuleRepository,
	bookItemRepo repositories.BookItemRepository,
	itemRepo *repositories.ItemRepository,
	userRepo repositories.UserRepository,
	updateRequestRepo *repositories.BookUpdateRequestRepository,
) BookService {
	return &bookService{
		bookRepo:               bookRepo,
		bookModuleRepo:         bookModuleRepo,
		bookItemRepo:           bookItemRepo,
		itemRepo:               itemRepo,
		userRepo:               userRepo,
		updateRequestRepo:      updateRequestRepo,
	}
}

// ==================== BOOK CRUD ====================

func (s *bookService) CreateBook(ownerID uuid.UUID, title, description, coverImage string) (*entities.Book, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}

	book := &entities.Book{
		OwnerID:     ownerID,
		Title:       title,
		Description: description,
		CoverImage:  coverImage,
		Status:      entities.BookStatusDraft,
	}

	if err := s.bookRepo.Create(book); err != nil {
		return nil, err
	}

	return book, nil
}

func (s *bookService) GetMyBooks(ownerID uuid.UUID) ([]entities.Book, error) {
	return s.bookRepo.FindByOwner(ownerID.String())
}

func (s *bookService) GetPublishedBooks() ([]entities.Book, error) {
	return s.bookRepo.FindPublished()
}

func (s *bookService) GetPublishedBookDetail(bookID string) (*entities.Book, error) {
	book, err := s.bookRepo.FindByIDWithRelations(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	if book.Status != entities.BookStatusPublished {
		return nil, errors.New("book is not published")
	}

	return book, nil
}

func (s *bookService) GetBookDetail(bookID string, userID *uuid.UUID, role string) (*entities.Book, error) {
	book, err := s.bookRepo.FindByIDWithRelations(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	// Admin can view any book regardless of status
	if role == "admin" {
		return book, nil
	}

	// If not published, only owner can view
	if book.Status != entities.BookStatusPublished {
		if userID == nil || book.OwnerID != *userID {
			return nil, errors.New("you don't have access to this book")
		}
	}

	return book, nil
}

// GetBookDetailWithStability returns book detail with stability information for each item
func (s *bookService) GetBookDetailWithStability(bookID string, userID *uuid.UUID, role string) (*BookDetailWithStability, error) {
	book, err := s.bookRepo.FindByIDWithRelations(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	// Admin can view any book regardless of status
	if role != "admin" {
		// If not published, only owner can view
		if book.Status != entities.BookStatusPublished {
			if userID == nil || book.OwnerID != *userID {
				return nil, errors.New("you don't have access to this book")
			}
		}
	}

	// Load all modules and items for this book (same as GetBookTree)
	modules, err := s.bookModuleRepo.FindByBookID(bookID)
	if err != nil {
		return nil, err
	}
	items, err := s.bookItemRepo.FindByBookID(bookID)
	if err != nil {
		return nil, err
	}

	// Build content_ref map to fetch Item entities for stability calculation
	contentRefs := make([]string, 0, len(items))
	for _, it := range items {
		contentRefs = append(contentRefs, "book:"+bookID+":item:"+it.ID.String())
	}

	// Fetch Item entities for stability calculation (if user is logged in)
	itemByContentRef := make(map[string]*entities.Item)
	if userID != nil {
		for _, ref := range contentRefs {
			existingItems, err := s.itemRepo.FindByOwnerAndContentRef(*userID, ref)
			if err == nil && len(existingItems) > 0 {
				itemByContentRef[ref] = &existingItems[0]
			}
		}
	}

	// Build module map and children links (same as GetBookTree)
	modMap := make(map[string]*entities.BookModule)
	childrenByParent := make(map[string][]string)
	for i := range modules {
		m := &modules[i]
		id := m.ID.String()
		modMap[id] = m
		parentKey := ""
		if m.ParentID != nil {
			parentKey = m.ParentID.String()
		}
		childrenByParent[parentKey] = append(childrenByParent[parentKey], id)
	}

	// Group items by module_id (nil goes to book-level)
	bookItems := make([]BookItemWithStability, 0)
	itemsByModule := make(map[string][]BookItemWithStability)
	for _, it := range items {
		contentRef := "book:" + bookID + ":item:" + it.ID.String()
		stability := calculateStability(itemByContentRef[contentRef])
		itemWithStability := BookItemWithStability{
			BookItem:  it,
			Stability: stability,
		}
		if it.ModuleID == nil {
			bookItems = append(bookItems, itemWithStability)
			continue
		}
		key := it.ModuleID.String()
		itemsByModule[key] = append(itemsByModule[key], itemWithStability)
	}

	// Recursive function to build modules with children
	var buildModules func(parentID string) []ModuleWithStability
	buildModules = func(parentID string) []ModuleWithStability {
		childIDs := childrenByParent[parentID]
		nodes := make([]ModuleWithStability, 0, len(childIDs))
		for _, cid := range childIDs {
			m := modMap[cid]
			node := ModuleWithStability{
				BookModule: *m,
				Items:      itemsByModule[cid],
				Children:   buildModules(cid),
			}
			nodes = append(nodes, node)
		}
		return nodes
	}

	return &BookDetailWithStability{
		Book:    *book,
		Items:   bookItems,
		Modules: buildModules(""),
	}, nil
}

func (s *bookService) GetBookDetailForAdmin(bookID string) (*entities.Book, error) {
	book, err := s.bookRepo.FindByIDWithRelations(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	// Admin can view any book regardless of status
	return book, nil
}

func (s *bookService) GetBookTree(bookID string, userID *uuid.UUID, role string) (*BookTreeResponse, error) {
	// Access control like GetBookDetail
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	// Admin can view any book regardless of status
	if role == "admin" {
		// Continue to load data
	} else if book.Status != entities.BookStatusPublished {
		if userID == nil || book.OwnerID != *userID {
			return nil, errors.New("you don't have access to this book")
		}
	}

	// Load all modules and items for this book
	modules, err := s.bookModuleRepo.FindByBookID(bookID)
	if err != nil {
		return nil, err
	}
	items, err := s.bookItemRepo.FindByBookID(bookID)
	if err != nil {
		return nil, err
	}

	// Build content_ref map to fetch Item entities for stability calculation
	contentRefs := make([]string, 0, len(items))
	for _, it := range items {
		contentRefs = append(contentRefs, "book:"+bookID+":item:"+it.ID.String())
	}

	// Fetch Item entities for stability calculation (if user is logged in)
	itemByContentRef := make(map[string]*entities.Item)
	if userID != nil {
		for _, ref := range contentRefs {
			existingItems, err := s.itemRepo.FindByOwnerAndContentRef(*userID, ref)
			if err == nil && len(existingItems) > 0 {
				itemByContentRef[ref] = &existingItems[0]
			}
		}
	}

	// Group items by module_id (nil goes to book-level) and calculate stability
	bookItems := make([]BookItemWithStability, 0)
	itemsByModule := make(map[string][]BookItemWithStability)
	for _, it := range items {
		contentRef := "book:" + bookID + ":item:" + it.ID.String()
		stability := calculateStability(itemByContentRef[contentRef])

		itemWithStability := BookItemWithStability{
			BookItem:  it,
			Stability: stability,
		}

		if it.ModuleID == nil {
			bookItems = append(bookItems, itemWithStability)
			continue
		}
		key := it.ModuleID.String()
		itemsByModule[key] = append(itemsByModule[key], itemWithStability)
	}

	// Build module map and children links
	type modWrap struct {
		mod      entities.BookModule
		children []string
	}
	modMap := make(map[string]*entities.BookModule)
	childrenByParent := make(map[string][]string)
	for i := range modules {
		m := &modules[i]
		id := m.ID.String()
		modMap[id] = m
		parentKey := ""
		if m.ParentID != nil {
			parentKey = m.ParentID.String()
		}
		childrenByParent[parentKey] = append(childrenByParent[parentKey], id)
	}

	var build func(parentID string) []ModuleNodeWithItems
	build = func(parentID string) []ModuleNodeWithItems {
		childIDs := childrenByParent[parentID]
		nodes := make([]ModuleNodeWithItems, 0, len(childIDs))
		for _, cid := range childIDs {
			m := modMap[cid]
			node := ModuleNodeWithItems{
				ID:          m.ID.String(),
				Title:       m.Title,
				Description: m.Description,
				Order:       m.Order,
				Items:       itemsByModule[cid],
				Children:    build(cid),
			}
			nodes = append(nodes, node)
		}
		// Preserve original order: modules slice was ordered by "order" ASC
		return nodes
	}

	tree := &BookTreeResponse{
		BookID:  book.ID.String(),
		Title:   book.Title,
		Items:   bookItems,
		Modules: build(""),
	}
	return tree, nil
}

func (s *bookService) AddPublishedBookToMyBook(userID uuid.UUID, bookID string) (*AddPublishedBookToMyBookResult, error) {
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	if book.Status != entities.BookStatusPublished {
		return nil, errors.New("book is not published")
	}

	bookItems, err := s.bookItemRepo.FindByBookID(bookID)
	if err != nil {
		return nil, err
	}

	result := &AddPublishedBookToMyBookResult{
		BookID:           bookID,
		AddedCount:       0,
		SkippedCount:     0,
		AddedContentRefs: nil,
	}

	// Create Item rows (source_type=book) for each BookItem in the published book.
	// We prevent duplicates by checking `content_ref` for the user.
	for _, bi := range bookItems {
		contentRef := "book:" + bookID + ":item:" + bi.ID.String()

		existingItems, err := s.itemRepo.FindByOwnerAndContentRef(userID, contentRef)
		if err != nil {
			return nil, err
		}
		if len(existingItems) > 0 {
			result.SkippedCount++
			continue
		}

		item := &entities.Item{
			OwnerID:                userID,
			SourceType:             "book",
			ContentRef:             contentRef,
			Status:                 entities.ItemStatusMenghafal,
			EstimatedReviewSeconds: bi.EstimatedReviewSeconds,
		}

		if err := s.itemRepo.Create(item); err != nil {
			return nil, err
		}

		result.AddedCount++
		// optional: return refs for UI debugging
		if result.AddedContentRefs == nil {
			result.AddedContentRefs = make([]string, 0, len(bookItems))
		}
		result.AddedContentRefs = append(result.AddedContentRefs, contentRef)
	}

	return result, nil
}

func (s *bookService) CopyPublishedBookToDraft(
	userID uuid.UUID,
	publishedBookID string,
	title, description, coverImage string,
) (*entities.Book, error) {
	srcBook, err := s.bookRepo.FindByIDWithRelations(publishedBookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	if srcBook.Status != entities.BookStatusPublished {
		return nil, errors.New("book is not published")
	}

	// Allow optional overrides; empty means "use source book value".
	finalTitle := srcBook.Title
	if title != "" {
		finalTitle = title
	}
	finalDesc := srcBook.Description
	if description != "" || srcBook.Description == "" {
		// If both are empty, it doesn't matter; but if src description isn't empty,
		// an empty override keeps using source to avoid accidental wipe.
		if description != "" {
			finalDesc = description
		}
	}
	finalCover := srcBook.CoverImage
	if coverImage != "" {
		finalCover = coverImage
	}

	draft := &entities.Book{
		OwnerID:     userID,
		Title:       finalTitle,
		Description: finalDesc,
		CoverImage:  finalCover,
		Status:      entities.BookStatusDraft,
		PublishedAt: nil,
	}
	if err := s.bookRepo.Create(draft); err != nil {
		return nil, err
	}

	// 1) Copy modules first (without parent pointers), so we can map IDs.
	newModulesByOldID := make(map[uuid.UUID]*entities.BookModule)
	for _, m := range srcBook.Modules {
		newMod := &entities.BookModule{
			BookID:      draft.ID,
			ParentID:    nil, // fix in second pass
			Title:       m.Title,
			Description: m.Description,
			Order:       m.Order,
		}
		if err := s.bookModuleRepo.Create(newMod); err != nil {
			return nil, err
		}
		newModulesByOldID[m.ID] = newMod
	}

	// 2) Restore nesting (parent-child) using the ID map.
	for _, m := range srcBook.Modules {
		if m.ParentID == nil {
			continue
		}
		newMod := newModulesByOldID[m.ID]
		if newMod == nil {
			return nil, errors.New("failed to copy module mapping")
		}
		parentNew := newModulesByOldID[*m.ParentID]
		if parentNew == nil {
			return nil, errors.New("failed to copy parent module mapping")
		}
		pid := parentNew.ID
		newMod.ParentID = &pid
		if err := s.bookModuleRepo.Update(newMod); err != nil {
			return nil, err
		}
	}

	// 3) Copy book-level items (module_id IS NULL)
	for _, it := range srcBook.Items {
		newItem := &entities.BookItem{
			BookID:                 draft.ID,
			ModuleID:               nil,
			Title:                  it.Title,
			Content:                it.Content,
			Answer:                 it.Answer,
			Order:                  it.Order,
			EstimatedReviewSeconds: it.EstimatedReviewSeconds,
		}
		if err := s.bookItemRepo.Create(newItem); err != nil {
			return nil, err
		}
	}

	// 4) Copy module items
	for _, m := range srcBook.Modules {
		newMod := newModulesByOldID[m.ID]
		if newMod == nil {
			return nil, errors.New("failed to copy module")
		}
		for _, it := range m.Items {
			modID := newMod.ID
			newItem := &entities.BookItem{
				BookID:                 draft.ID,
				ModuleID:               &modID,
				Title:                  it.Title,
				Content:                it.Content,
				Answer:                 it.Answer,
				Order:                  it.Order,
				EstimatedReviewSeconds: it.EstimatedReviewSeconds,
			}
			if err := s.bookItemRepo.Create(newItem); err != nil {
				return nil, err
			}
		}
	}

	return draft, nil
}

func (s *bookService) UpdateBook(bookID string, ownerID uuid.UUID, title, description, coverImage string) (*entities.Book, error) {
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	if book.OwnerID != ownerID {
		return nil, errors.New("you don't have permission to update this book")
	}

	// For published books, create an update request instead of direct update
	if book.Status == entities.BookStatusPublished {
		// Check if there's already a pending update request
		existingPending, err := s.updateRequestRepo.FindPendingByBookID(bookID)
		if err == nil && existingPending != nil {
			// Update the existing pending request
			if title != "" {
				existingPending.Title = title
			}
			if description != "" {
				existingPending.Description = description
			}
			if coverImage != "" {
				existingPending.CoverImage = coverImage
			}
			if err := s.updateRequestRepo.Update(existingPending); err != nil {
				return nil, err
			}
			return book, nil
		}

		// Create new update request
		updateReq := &entities.BookUpdateRequest{
			BookID:      uuid.MustParse(bookID),
			OwnerID:     ownerID,
			Title:       title,
			Description: description,
			CoverImage:  coverImage,
			Status:      entities.BookUpdateStatusPending,
		}
		if err := s.updateRequestRepo.Create(updateReq); err != nil {
			return nil, err
		}
		return book, nil
	}

	if title != "" {
		book.Title = title
	}
	if description != "" {
		book.Description = description
	}
	if coverImage != "" {
		book.CoverImage = coverImage
	}

	if err := s.bookRepo.Update(book); err != nil {
		return nil, err
	}

	return book, nil
}

func (s *bookService) DeleteBook(bookID string, ownerID uuid.UUID) error {
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return errors.New("book not found")
	}

	if book.OwnerID != ownerID {
		return errors.New("you don't have permission to delete this book")
	}

	if book.Status == entities.BookStatusPublished {
		return errors.New("cannot delete published book")
	}

	// Delete all related items and modules
	if err := s.bookItemRepo.DeleteByBookID(bookID); err != nil {
		return err
	}
	if err := s.bookModuleRepo.DeleteByBookID(bookID); err != nil {
		return err
	}

	return s.bookRepo.Delete(bookID)
}

// ==================== PUBLISH WORKFLOW ====================

func (s *bookService) RequestPublish(bookID string, ownerID uuid.UUID) error {
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return errors.New("book not found")
	}

	if book.OwnerID != ownerID {
		return errors.New("you don't have permission to publish this book")
	}

	if book.Status != entities.BookStatusDraft && book.Status != entities.BookStatusRejected {
		return errors.New("book must be in draft or rejected status to request publish")
	}

	return s.bookRepo.UpdateStatus(bookID, entities.BookStatusPending)
}

func (s *bookService) GetPendingBooks() ([]entities.Book, error) {
	return s.bookRepo.FindPendingPublish()
}

func (s *bookService) ApproveBook(bookID string) error {
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return errors.New("book not found")
	}

	if book.Status != entities.BookStatusPending {
		return errors.New("book is not pending for approval")
	}

	return s.bookRepo.UpdateStatus(bookID, entities.BookStatusPublished)
}

func (s *bookService) RejectBook(bookID string) error {
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return errors.New("book not found")
	}

	if book.Status != entities.BookStatusPending {
		return errors.New("book is not pending for approval")
	}

	return s.bookRepo.UpdateStatus(bookID, entities.BookStatusRejected)
}

// DeletePublishedBook deletes a published book (Admin only)
func (s *bookService) DeletePublishedBook(bookID string) error {
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return errors.New("book not found")
	}

	if book.Status != entities.BookStatusPublished {
		return errors.New("book is not published")
	}

	// Delete all related items and modules
	if err := s.bookItemRepo.DeleteByBookID(bookID); err != nil {
		return err
	}
	if err := s.bookModuleRepo.DeleteByBookID(bookID); err != nil {
		return err
	}

	return s.bookRepo.Delete(bookID)
}

// ==================== BOOK UPDATE REQUESTS ====================

// RequestBookUpdate creates an update request for a published book
func (s *bookService) RequestBookUpdate(bookID string, ownerID uuid.UUID, title, description, coverImage string) (*entities.BookUpdateRequest, error) {
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	if book.OwnerID != ownerID {
		return nil, errors.New("you don't have permission to update this book")
	}

	if book.Status != entities.BookStatusPublished {
		return nil, errors.New("book must be published to request update")
	}

	// Check if there's already a pending update request
	existingPending, err := s.updateRequestRepo.FindPendingByBookID(bookID)
	if err == nil && existingPending != nil {
		return nil, errors.New("there is already a pending update request for this book")
	}

	updateReq := &entities.BookUpdateRequest{
		BookID:      uuid.MustParse(bookID),
		OwnerID:     ownerID,
		Title:       title,
		Description: description,
		CoverImage:  coverImage,
		Status:      entities.BookUpdateStatusPending,
	}

	if err := s.updateRequestRepo.Create(updateReq); err != nil {
		return nil, err
	}

	return updateReq, nil
}

// GetBookUpdateRequests returns all update requests for a book
func (s *bookService) GetBookUpdateRequests(bookID string) ([]entities.BookUpdateRequest, error) {
	return s.updateRequestRepo.FindByBookID(bookID)
}

// ApproveBookUpdate approves an update request and applies changes to the book
func (s *bookService) ApproveBookUpdate(requestID string, adminID uuid.UUID) error {
	updateReq, err := s.updateRequestRepo.FindByID(requestID)
	if err != nil {
		return errors.New("update request not found")
	}

	if updateReq.Status != entities.BookUpdateStatusPending {
		return errors.New("update request is not pending")
	}

	// Get the book
	book, err := s.bookRepo.FindByID(updateReq.BookID.String())
	if err != nil {
		return errors.New("book not found")
	}

	// Apply changes to the book
	if updateReq.Title != "" {
		book.Title = updateReq.Title
	}
	if updateReq.Description != "" {
		book.Description = updateReq.Description
	}
	if updateReq.CoverImage != "" {
		book.CoverImage = updateReq.CoverImage
	}

	now := time.Now().In(config.AppLocation)
	updateReq.Status = entities.BookUpdateStatusApproved
	updateReq.ApprovedAt = &now
	updateReq.ApprovedBy = &adminID

	// Update both the book and the request
	if err := s.bookRepo.Update(book); err != nil {
		return err
	}

	return s.updateRequestRepo.Update(updateReq)
}

// RejectBookUpdate rejects an update request
func (s *bookService) RejectBookUpdate(requestID string, adminID uuid.UUID, reason string) error {
	updateReq, err := s.updateRequestRepo.FindByID(requestID)
	if err != nil {
		return errors.New("update request not found")
	}

	if updateReq.Status != entities.BookUpdateStatusPending {
		return errors.New("update request is not pending")
	}

	now := time.Now().In(config.AppLocation)
	updateReq.Status = entities.BookUpdateStatusRejected
	updateReq.ApprovedAt = &now
	updateReq.ApprovedBy = &adminID
	updateReq.RejectReason = reason

	return s.updateRequestRepo.Update(updateReq)
}

// GetPendingBookUpdates returns all pending update requests
func (s *bookService) GetPendingBookUpdates() ([]entities.BookUpdateRequest, error) {
	return s.updateRequestRepo.FindAllPending()
}

// ==================== MODULE CRUD ====================

func (s *bookService) AddModule(bookID string, ownerID uuid.UUID, title, description string, order int, parentID *uuid.UUID) (*entities.BookModule, error) {
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	if book.OwnerID != ownerID {
		return nil, errors.New("you don't have permission to add module to this book")
	}

	// Allow owner to add modules to published books
	// (Changes are visible immediately, but owner should request update for metadata)
	if title == "" {
		return nil, errors.New("module title is required")
	}

	module := &entities.BookModule{
		BookID:      uuid.MustParse(bookID),
		ParentID:    parentID,
		Title:       title,
		Description: description,
		Order:       order,
	}

	if err := s.bookModuleRepo.Create(module); err != nil {
		return nil, err
	}

	return module, nil
}

func (s *bookService) UpdateModule(moduleID string, ownerID uuid.UUID, title, description string, order int) (*entities.BookModule, error) {
	module, err := s.bookModuleRepo.FindByID(moduleID)
	if err != nil {
		return nil, errors.New("module not found")
	}

	book, err := s.bookRepo.FindByID(module.BookID.String())
	if err != nil {
		return nil, errors.New("book not found")
	}

	if book.OwnerID != ownerID {
		return nil, errors.New("you don't have permission to update this module")
	}

	if book.Status == entities.BookStatusPublished {
		return nil, errors.New("cannot update module in published book")
	}

	if title != "" {
		module.Title = title
	}
	if description != "" {
		module.Description = description
	}
	if order > 0 {
		module.Order = order
	}

	if err := s.bookModuleRepo.Update(module); err != nil {
		return nil, err
	}

	return module, nil
}

func (s *bookService) DeleteModule(moduleID string, ownerID uuid.UUID) error {
	module, err := s.bookModuleRepo.FindByID(moduleID)
	if err != nil {
		return errors.New("module not found")
	}

	book, err := s.bookRepo.FindByID(module.BookID.String())
	if err != nil {
		return errors.New("book not found")
	}

	if book.OwnerID != ownerID {
		return errors.New("you don't have permission to delete this module")
	}

	if book.Status == entities.BookStatusPublished {
		return errors.New("cannot delete module from published book")
	}

	// Get all BookItems in this module to find their Item entities
	bookItems, err := s.bookItemRepo.FindByModuleID(moduleID)
	if err == nil && len(bookItems) > 0 {
		// Delete Item entities for each BookItem
		for _, bookItem := range bookItems {
			contentRef := "book:" + bookItem.BookID.String() + ":item:" + bookItem.ID.String()
			existingItems, err := s.itemRepo.FindByContentRef(contentRef)
			if err == nil && len(existingItems) > 0 {
				for _, existingItem := range existingItems {
					s.itemRepo.DeleteByID(existingItem.ID)
				}
			}
		}
	}

	// Delete all BookItems in this module
	if err := s.bookItemRepo.DeleteByModuleID(moduleID); err != nil {
		return err
	}

	return s.bookModuleRepo.Delete(moduleID)
}

// ==================== ITEM CRUD ====================

func (s *bookService) AddItem(bookID string, moduleID *uuid.UUID, ownerID uuid.UUID, title, content, answer string, order int, estimateVal int, estimateUnit string) (*entities.BookItem, error) {
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	if book.OwnerID != ownerID {
		return nil, errors.New("you don't have permission to add item to this book")
	}

	// Allow owner to add items to published books
	// Title is optional, but content or answer must be provided
	if content == "" && answer == "" {
		return nil, errors.New("either content or answer must be provided")
	}

	// Validate module belongs to book if provided
	if moduleID != nil {
		module, err := s.bookModuleRepo.FindByID(moduleID.String())
		if err != nil {
			return nil, errors.New("module not found")
		}
		if module.BookID.String() != bookID {
			return nil, errors.New("module does not belong to this book")
		}
	}

	// Normalize estimation into seconds
	estSeconds := 0
	if estimateVal > 0 {
		switch estimateUnit {
		case "minutes", "minute", "min", "m":
			estSeconds = estimateVal * 60
		default:
			estSeconds = estimateVal
		}
		if estSeconds < 0 {
			estSeconds = 0
		}
	}

	item := &entities.BookItem{
		BookID:                 uuid.MustParse(bookID),
		ModuleID:               moduleID,
		Title:                  title,
		Content:                content,
		Answer:                 answer,
		Order:                  order,
		EstimatedReviewSeconds: estSeconds,
		CreatedAt:              time.Now().In(config.AppLocation),
		UpdatedAt:              time.Now().In(config.AppLocation),
	}

	if err := s.bookItemRepo.Create(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *bookService) UpdateItem(itemID string, ownerID uuid.UUID, title, content, answer string, order int) (*entities.BookItem, error) {
	item, err := s.bookItemRepo.FindByID(itemID)
	if err != nil {
		return nil, errors.New("item not found")
	}

	book, err := s.bookRepo.FindByID(item.BookID.String())
	if err != nil {
		return nil, errors.New("book not found")
	}

	if book.OwnerID != ownerID {
		return nil, errors.New("you don't have permission to update this item")
	}

	if book.Status == entities.BookStatusPublished {
		return nil, errors.New("cannot update item in published book")
	}

	if title != "" {
		item.Title = title
	}
	if content != "" {
		item.Content = content
	}
	if answer != "" {
		item.Answer = answer
	}
	if order > 0 {
		item.Order = order
	}
	item.UpdatedAt = time.Now().In(config.AppLocation)

	if err := s.bookItemRepo.Update(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *bookService) DeleteItem(itemID string, ownerID uuid.UUID) error {
	item, err := s.bookItemRepo.FindByID(itemID)
	if err != nil {
		return errors.New("item not found")
	}

	book, err := s.bookRepo.FindByID(item.BookID.String())
	if err != nil {
		return errors.New("book not found")
	}

	if book.OwnerID != ownerID {
		return errors.New("you don't have permission to delete this item")
	}

	if book.Status == entities.BookStatusPublished {
		return errors.New("cannot delete item from published book")
	}

	// Delete Item entity if it exists (user already started memorizing this item)
	contentRef := "book:" + item.BookID.String() + ":item:" + itemID
	existingItems, err := s.itemRepo.FindByContentRef(contentRef)
	if err == nil && len(existingItems) > 0 {
		for _, existingItem := range existingItems {
			if err := s.itemRepo.DeleteByID(existingItem.ID); err != nil {
				// Log error but continue with BookItem deletion
				// This ensures BookItem is deleted even if Item deletion fails
			}
		}
	}

	return s.bookItemRepo.Delete(itemID)
}

// ==================== MEMORIZATION ====================

// StartItemMemorization starts memorizing a book item
// If item already exists, returns the existing item instead of error
func (s *bookService) StartItemMemorization(userID uuid.UUID, bookID, bookItemID string) (*StartMemorizationResult, error) {
	// 1. Get book and validate access
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	// Check if book is accessible (published or owned by user)
	if book.Status != entities.BookStatusPublished && book.OwnerID != userID {
		return nil, errors.New("you don't have access to this book")
	}

	// 2. Get book item and validate it belongs to book
	bookItem, err := s.bookItemRepo.FindByID(bookItemID)
	if err != nil {
		return nil, errors.New("book item not found")
	}

	if bookItem.BookID.String() != bookID {
		return nil, errors.New("book item does not belong to this book")
	}

	// 3. Check if user already has this item
	contentRef := "book:" + bookID + ":item:" + bookItemID
	existingItems, err := s.itemRepo.FindByOwnerAndContentRef(userID, contentRef)
	if err == nil && len(existingItems) > 0 {
		existingItem := &existingItems[0]
		
		// If item exists but status is 'menghafal', update to 'start'
		// This handles items created from AddPublishedBookToMyBook
		if existingItem.Status == entities.ItemStatusMenghafal {
			existingItem.Status = entities.ItemStatusStart
			if err := s.itemRepo.Update(existingItem); err != nil {
				return nil, err
			}
			return &StartMemorizationResult{
				ItemID:     existingItem.ID,
				BookItemID: bookItem.ID,
				BookTitle:  book.Title,
				ItemTitle:  bookItem.Title,
				Status:     entities.ItemStatusStart,
			}, nil
		}
		
		// Item already exists with other status, return as-is
		return &StartMemorizationResult{
			ItemID:     existingItem.ID,
			BookItemID: bookItem.ID,
			BookTitle:  book.Title,
			ItemTitle:  bookItem.Title,
			Status:     existingItem.Status,
		}, nil
	}

	// 4. Create new Item with status "start" for book items
	// Book items flow: START → FSRS_ACTIVE → GRADUATE
	item := &entities.Item{
		OwnerID:    userID,
		SourceType: "book", // book items use "book" as source type
		ContentRef: contentRef,
		Status:     entities.ItemStatusStart, // Start phase for book items
	}
	// copy estimation from book item into Item for daily usage
	item.EstimatedReviewSeconds = bookItem.EstimatedReviewSeconds

	if err := s.itemRepo.Create(item); err != nil {
		return nil, err
	}

	return &StartMemorizationResult{
		ItemID:     item.ID,
		BookItemID: bookItem.ID,
		BookTitle:  book.Title,
		ItemTitle:  bookItem.Title,
		Status:     item.Status,
	}, nil
}

// ==================== MY BOOK COLLECTION ====================

func (s *bookService) GetMyBookCollection(userID uuid.UUID) ([]BookCollectionItem, error) {
	// Fetch all book items (source_type = "book")
	items, err := s.itemRepo.FindByOwnerAndSourceType(userID, "book")
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return []BookCollectionItem{}, nil
	}

	// Extract unique book IDs and track earliest added_at per book
	bookIDMap := make(map[string]*BookCollectionItem)
	bookOrder := make([]string, 0)

	for _, item := range items {
		// Parse content_ref: "book:BOOK_ID:item:BOOK_ITEM_ID"
		parts := strings.Split(item.ContentRef, ":")
		if len(parts) != 4 || parts[0] != "book" || parts[2] != "item" {
			continue
		}
		bookID := parts[1]

		if _, exists := bookIDMap[bookID]; !exists {
			// Fetch book details
			book, err := s.bookRepo.FindByID(bookID)
			if err != nil {
				continue // Skip if book not found
			}

			// Fetch owner name
			ownerName := ""
			owner, err := s.userRepo.FindByID(book.OwnerID.String())
			if err == nil && owner != nil {
				ownerName = owner.FullName
			}

			bookIDMap[bookID] = &BookCollectionItem{
				BookID:      bookID,
				Title:       book.Title,
				Description: book.Description,
				CoverImage:  book.CoverImage,
				OwnerName:   ownerName,
				ItemCount:   0,
				AddedAt:     item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
			bookOrder = append(bookOrder, bookID)
		}

		bookIDMap[bookID].ItemCount++
	}

	// Build result in order
	result := make([]BookCollectionItem, 0, len(bookOrder))
	for _, bookID := range bookOrder {
		result = append(result, *bookIDMap[bookID])
	}

	return result, nil
}

func (s *bookService) RemoveFromMyBookCollection(userID uuid.UUID, bookID string) error {
	// Verify book exists
	_, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return errors.New("book not found")
	}

	// Delete all items with content_ref starting with "book:BOOK_ID:"
	// We need to fetch items first to delete them
	items, err := s.itemRepo.FindByOwnerAndSourceType(userID, "book")
	if err != nil {
		return err
	}

	prefix := "book:" + bookID + ":item:"
	for _, item := range items {
		if strings.HasPrefix(item.ContentRef, prefix) {
			if err := s.itemRepo.DeleteByID(item.ID); err != nil {
				return err
			}
		}
	}

	return nil
}
