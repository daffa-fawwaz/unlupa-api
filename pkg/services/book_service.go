package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/utils"

	"github.com/google/uuid"
)

type BookService interface {
	// Book CRUD
	CreateBook(ownerID uuid.UUID, title, description, coverImage string) (*entities.Book, error)
	GetMyBooks(ownerID uuid.UUID) ([]entities.Book, error)
	GetPublishedBooks() ([]PublishedBookWithStats, error)
	GetPublishedBookDetail(bookID string) (*entities.Book, error)
	GetBookDetail(bookID string, userID *uuid.UUID, role string) (*entities.Book, error)
	GetBookDetailWithStability(bookID string, userID *uuid.UUID, role string) (*BookDetailWithStability, error)
	GetBookDetailForAdmin(bookID string) (*entities.Book, error)
	GetBookTree(bookID string, userID *uuid.UUID, role string) (*BookTreeResponse, error)
	UpdateBook(bookID string, ownerID uuid.UUID, title, description, coverImage string) (*entities.Book, error)
	DeleteBook(bookID string, ownerID uuid.UUID) error

	// Publish workflow
	RequestPublish(bookID string, ownerID uuid.UUID, isEditable bool) error
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
	AddItem(bookID string, moduleID *uuid.UUID, ownerID uuid.UUID, title, content, answer string, order int, estimateVal int, estimateUnit string, imageURL string) (*entities.BookItem, error)
	UpdateItem(itemID string, ownerID uuid.UUID, title, content, answer string, order int, estimateVal int, estimateUnit string, imageURL string, removeImage bool) (*entities.BookItem, error)
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

	// Book Item Overrides
	GetMyOverride(userID uuid.UUID, bookItemID string) (*entities.BookItemOverride, error)
	RemoveMyOverride(userID uuid.UUID, bookItemID string) error

	// Reorder
	ReorderBookStructure(bookID string, ownerID uuid.UUID, req ReorderBookStructureRequest) error
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
// calculateStability returns the interval in days between last_review_at and next_review_at.
// Both dates are normalized to midnight so the result is always a whole number of days.
// This value is fixed and only changes when next_review_at is updated after a review.
// Returns "item belum masuk ujian" if item hasn't been reviewed yet.
func calculateStability(item *entities.Item) string {
	if item == nil {
		return "item belum masuk ujian"
	}

	if item.Status == entities.ItemStatusStart {
		return "item belum masuk ujian"
	}

	if item.NextReviewAt != nil && item.LastReviewAt != nil {
		loc := item.NextReviewAt.Location()

		// Normalize both to midnight so partial hours don't affect the day count
		last := time.Date(
			item.LastReviewAt.Year(), item.LastReviewAt.Month(), item.LastReviewAt.Day(),
			0, 0, 0, 0, loc,
		)
		next := time.Date(
			item.NextReviewAt.Year(), item.NextReviewAt.Month(), item.NextReviewAt.Day(),
			0, 0, 0, 0, loc,
		)

		interval := next.Sub(last).Hours() / 24
		return fmt.Sprintf("%.0f", interval)
	}

	return "item belum masuk ujian"
}

// BookTreeResponse represents hierarchical modules and items for a book
type BookTreeResponse struct {
	BookID  string                  `json:"book_id"`
	Title   string                  `json:"title"`
	Items   []BookItemWithStability `json:"items"` // items directly under book
	Modules []ModuleNodeWithItems   `json:"modules"`
}

type ModuleNodeWithItems struct {
	ID          string                  `json:"id"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Order       int                     `json:"order"`
	Items       []BookItemWithStability `json:"items"`
	Children    []ModuleNodeWithItems   `json:"children"`
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
	BookID      string `json:"book_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CoverImage  string `json:"cover_image,omitempty"`
	OwnerName   string `json:"owner_name,omitempty"`
	ItemCount   int    `json:"item_count"`
	AddedAt     string `json:"added_at"`
}

// PublishedBookWithStats wraps a Book with additional stats for the published listing
type PublishedBookWithStats struct {
	entities.Book
	TotalAdded int64 `json:"total_added"`
}

type bookService struct {
	bookRepo          repositories.BookRepository
	bookModuleRepo    repositories.BookModuleRepository
	bookItemRepo      repositories.BookItemRepository
	classBookRepo     repositories.ClassBookRepository
	itemRepo          *repositories.ItemRepository
	userRepo          repositories.UserRepository
	updateRequestRepo *repositories.BookUpdateRequestRepository
	overrideRepo      repositories.BookItemOverrideRepository
}

func NewBookService(
	bookRepo repositories.BookRepository,
	bookModuleRepo repositories.BookModuleRepository,
	bookItemRepo repositories.BookItemRepository,
	classBookRepo repositories.ClassBookRepository,
	itemRepo *repositories.ItemRepository,
	userRepo repositories.UserRepository,
	updateRequestRepo *repositories.BookUpdateRequestRepository,
	overrideRepo repositories.BookItemOverrideRepository,
) BookService {
	return &bookService{
		bookRepo:          bookRepo,
		bookModuleRepo:    bookModuleRepo,
		bookItemRepo:      bookItemRepo,
		classBookRepo:     classBookRepo,
		itemRepo:          itemRepo,
		userRepo:          userRepo,
		updateRequestRepo: updateRequestRepo,
		overrideRepo:      overrideRepo,
	}
}

func (s *bookService) canAccessClassBook(bookID string, ownerID uuid.UUID, userID *uuid.UUID) bool {
	if userID == nil {
		return false
	}

	// Book owner always has access
	if ownerID == *userID {
		return true
	}

	if s.classBookRepo == nil {
		return false
	}

	// Student who joined a class containing this book
	memberOK, err := s.classBookRepo.IsBookAccessibleByMember(bookID, userID.String())
	if err == nil && memberOK {
		return true
	}

	// Teacher (guru) of a class that contains this book
	teacherOK, err := s.classBookRepo.IsBookAccessibleByTeacher(bookID, userID.String())
	return err == nil && teacherOK
}

func (s *bookService) canViewBook(book *entities.Book, userID *uuid.UUID, role string) bool {
	if role == "admin" {
		return true
	}

	if s.classBookRepo != nil {
		isClassBook, err := s.classBookRepo.IsBookAssignedToClass(book.ID.String())
		if err == nil && isClassBook {
			return s.canAccessClassBook(book.ID.String(), book.OwnerID, userID)
		}
	}

	return book.Status == entities.BookStatusPublished || s.canAccessClassBook(book.ID.String(), book.OwnerID, userID)
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

func (s *bookService) GetPublishedBooks() ([]PublishedBookWithStats, error) {
	books, err := s.bookRepo.FindPublished()
	if err != nil {
		return nil, err
	}

	result := make([]PublishedBookWithStats, 0, len(books))
	for _, book := range books {
		count, _ := s.itemRepo.CountDistinctOwnersByBookID(book.ID.String())
		result = append(result, PublishedBookWithStats{
			Book:       book,
			TotalAdded: count,
		})
	}
	return result, nil
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

	if !s.canViewBook(book, userID, role) {
		return nil, errors.New("you don't have access to this book")
	}

	return book, nil
}

// GetBookDetailWithStability returns book detail with stability information for each item
func (s *bookService) GetBookDetailWithStability(bookID string, userID *uuid.UUID, role string) (*BookDetailWithStability, error) {
	book, err := s.bookRepo.FindByIDWithRelations(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	if !s.canViewBook(book, userID, role) {
		return nil, errors.New("you don't have access to this book")
	}

	// Load all modules and items for this book (same as GetBookTree)
	modules, err := s.bookModuleRepo.FindByBookID(bookID)
	if err != nil {
		return nil, err
	}
	// Show importer's personal items alongside canonical ones when user is not the owner.
	var items []entities.BookItem
	if userID != nil && book.OwnerID != *userID {
		items, err = s.bookItemRepo.FindByBookIDForImporter(bookID, *userID)
	} else {
		items, err = s.bookItemRepo.FindByBookID(bookID)
	}
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

	if !s.canViewBook(book, userID, role) {
		return nil, errors.New("you don't have access to this book")
	}

	// Load all modules and items for this book
	modules, err := s.bookModuleRepo.FindByBookID(bookID)
	if err != nil {
		return nil, err
	}
	// Show importer's personal items alongside canonical ones when user is not the owner.
	var items []entities.BookItem
	if userID != nil && book.OwnerID != *userID {
		items, err = s.bookItemRepo.FindByBookIDForImporter(bookID, *userID)
	} else {
		items, err = s.bookItemRepo.FindByBookID(bookID)
	}
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

	if len(bookItems) == 0 {
		return nil, errors.New("book has no items")
	}

	isImported, err := s.classBookRepo.IsBookImportedByUser(bookID, userID.String())
	if err == nil && isImported {
		return &AddPublishedBookToMyBookResult{
			BookID:       bookID,
			AddedCount:   0,
			SkippedCount: len(bookItems),
		}, nil
	}

	if err := s.classBookRepo.CreateImportedBook(userID.String(), bookID); err != nil {
		return nil, err
	}

	return &AddPublishedBookToMyBookResult{
		BookID:     bookID,
		AddedCount: len(bookItems),
	}, nil
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
		IsEditable:  srcBook.IsEditable, // inherit editable flag from source book
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

	// Delete memorization Item rows created from this book before deleting the
	// book structure itself.
	if err := s.itemRepo.DeleteBookItemsByBookID(bookID); err != nil {
		return err
	}

	// Delete all related book items and modules
	if err := s.bookItemRepo.DeleteByBookID(bookID); err != nil {
		return err
	}
	if err := s.bookModuleRepo.DeleteByBookID(bookID); err != nil {
		return err
	}

	return s.bookRepo.Delete(bookID)
}

// ==================== PUBLISH WORKFLOW ====================

func (s *bookService) RequestPublish(bookID string, ownerID uuid.UUID, isEditable bool) error {
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

	// Save the editable flag before changing status
	book.IsEditable = isEditable
	if err := s.bookRepo.Update(book); err != nil {
		return err
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

	// Delete memorization Item rows created from this book before deleting the
	// book structure itself.
	if err := s.itemRepo.DeleteBookItemsByBookID(bookID); err != nil {
		return err
	}

	// Delete all related book items and modules
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

	// For published books: allow anyone if is_editable=true, owner always can edit
	// For non-published books: only owner can edit
	if book.Status == entities.BookStatusPublished {
		if !book.IsEditable && book.OwnerID != ownerID {
			return nil, errors.New("this book is not editable")
		}
	} else {
		if book.OwnerID != ownerID {
			return nil, errors.New("you don't have permission to add module to this book")
		}
	}

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

	// Published books: allow anyone if is_editable=true, owner always can edit
	// Non-published books: only owner can edit
	if book.Status == entities.BookStatusPublished {
		if !book.IsEditable && book.OwnerID != ownerID {
			return nil, errors.New("this book is not editable")
		}
	} else {
		if book.OwnerID != ownerID {
			return nil, errors.New("you don't have permission to update this module")
		}
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

	// Published books: allow anyone if is_editable=true, owner always can delete
	// Non-published books: only owner can delete
	if book.Status == entities.BookStatusPublished {
		if !book.IsEditable && book.OwnerID != ownerID {
			return errors.New("this book is not editable")
		}
	} else {
		if book.OwnerID != ownerID {
			return errors.New("you don't have permission to delete this module")
		}
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

// ResolvedBookItem wraps a BookItem with resolved content (override applied if exists).
type ResolvedBookItem struct {
	entities.BookItem
	HasOverride bool `json:"has_override"` // true if user has personal override for this item
}

// ResolveBookItemContent resolves a BookItem's content for a specific user.
// If the user has a BookItemOverride for this item, the override values shadow
// the canonical BookItem fields. Otherwise, returns the canonical item as-is.
//
// This is the single source of truth for reading BookItem content across the app:
// all handlers/services that serve title/content/answer to the user MUST call this.
func ResolveBookItemContent(
	canonical *entities.BookItem,
	userID *uuid.UUID,
	overrideRepo repositories.BookItemOverrideRepository,
) ResolvedBookItem {
	resolved := ResolvedBookItem{
		BookItem:    *canonical,
		HasOverride: false,
	}

	if userID == nil || overrideRepo == nil {
		return resolved
	}

	override, err := overrideRepo.FindByUserAndBookItemID(*userID, canonical.ID)
	if err != nil || override == nil {
		return resolved
	}

	// Apply override: non-empty fields from override replace canonical values.
	resolved.HasOverride = true
	if override.Title != "" {
		resolved.Title = override.Title
	}
	if override.Content != "" {
		resolved.Content = override.Content
	}
	if override.Answer != "" {
		resolved.Answer = override.Answer
	}
	if override.ImageURL != "" {
		resolved.ImageURL = override.ImageURL
	}
	if override.EstimatedReviewSeconds > 0 {
		resolved.EstimatedReviewSeconds = override.EstimatedReviewSeconds
	}

	return resolved
}

// ResolveBatchBookItemContent batch-resolves multiple BookItems for a user.
// Returns a map keyed by book_item_id string for O(1) lookup.
func ResolveBatchBookItemContent(
	canonicals []entities.BookItem,
	userID *uuid.UUID,
	overrideRepo repositories.BookItemOverrideRepository,
) map[uuid.UUID]ResolvedBookItem {
	result := make(map[uuid.UUID]ResolvedBookItem, len(canonicals))

	if len(canonicals) == 0 || userID == nil || overrideRepo == nil {
		// No overrides possible; return canonicals as-is.
		for _, item := range canonicals {
			result[item.ID] = ResolvedBookItem{BookItem: item, HasOverride: false}
		}
		return result
	}

	// Batch-fetch all overrides for this user + these book_item_ids.
	bookItemIDs := make([]uuid.UUID, len(canonicals))
	for i, item := range canonicals {
		bookItemIDs[i] = item.ID
	}
	overrideMap, err := overrideRepo.FindByUserAndBookItemIDs(*userID, bookItemIDs)
	if err != nil {
		overrideMap = make(map[uuid.UUID]*entities.BookItemOverride)
	}

	// Build resolved map.
	for _, canonical := range canonicals {
		resolved := ResolvedBookItem{
			BookItem:    canonical,
			HasOverride: false,
		}

		if override, exists := overrideMap[canonical.ID]; exists && override != nil {
			resolved.HasOverride = true
			if override.Title != "" {
				resolved.Title = override.Title
			}
			if override.Content != "" {
				resolved.Content = override.Content
			}
			if override.Answer != "" {
				resolved.Answer = override.Answer
			}
			if override.ImageURL != "" {
				resolved.ImageURL = override.ImageURL
			}
			if override.EstimatedReviewSeconds > 0 {
				resolved.EstimatedReviewSeconds = override.EstimatedReviewSeconds
			}
		}

		result[canonical.ID] = resolved
	}

	return result
}

// normalizeEstSeconds converts estimateVal + estimateUnit to seconds.
func normalizeEstSeconds(estimateVal int, estimateUnit string) int {
	if estimateVal <= 0 {
		return 0
	}
	switch strings.ToLower(estimateUnit) {
	case "minutes", "minute", "min", "m":
		return estimateVal * 60
	default:
		return estimateVal
	}
}

func (s *bookService) AddItem(bookID string, moduleID *uuid.UUID, ownerID uuid.UUID, title, content, answer string, order int, estimateVal int, estimateUnit string, imageURL string) (*entities.BookItem, error) {
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	// Published books: allow anyone if is_editable=true, owner always can edit
	// Non-published books: only owner can edit
	if book.Status == entities.BookStatusPublished {
		if !book.IsEditable && book.OwnerID != ownerID {
			return nil, errors.New("this book is not editable")
		}
	} else {
		if book.OwnerID != ownerID {
			return nil, errors.New("you don't have permission to add item to this book")
		}
	}

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

	estSeconds := normalizeEstSeconds(estimateVal, estimateUnit)

	// ── Non-owner of a published book → create personal BookItem with ImporterID ─
	// BookItem dengan importer_id terisi TIDAK akan muncul di FindByBookID (canonical),
	// sehingga pemilik buku dan importer lain tidak melihatnya sama sekali.
	if book.Status == entities.BookStatusPublished && book.OwnerID != ownerID {
		importerItem := &entities.BookItem{
			BookID:                 uuid.MustParse(bookID),
			ModuleID:               moduleID,
			ImporterID:             &ownerID, // tandai sebagai milik importer ini
			Title:                  title,
			Content:                content,
			Answer:                 answer,
			Order:                  order,
			EstimatedReviewSeconds: estSeconds,
			ImageURL:               imageURL,
			CreatedAt:              time.Now().In(config.AppLocation),
			UpdatedAt:              time.Now().In(config.AppLocation),
		}
		if err := s.bookItemRepo.Create(importerItem); err != nil {
			return nil, err
		}

		// Buat Item (memorization row) agar muncul di daily feed dan koleksi importer.
		contentRef := "book:" + bookID + ":item:" + importerItem.ID.String()
		memItem := &entities.Item{
			OwnerID:                ownerID,
			SourceType:             "book",
			ContentRef:             contentRef,
			Status:                 entities.ItemStatusMenghafal,
			EstimatedReviewSeconds: estSeconds,
		}
		_ = s.itemRepo.Create(memItem)

		return importerItem, nil
	}

	// ── Owner (or draft book) → write directly to book_items ───────────────
	item := &entities.BookItem{
		BookID:                 uuid.MustParse(bookID),
		ModuleID:               moduleID,
		Title:                  title,
		Content:                content,
		Answer:                 answer,
		Order:                  order,
		EstimatedReviewSeconds: estSeconds,
		ImageURL:               imageURL,
		CreatedAt:              time.Now().In(config.AppLocation),
		UpdatedAt:              time.Now().In(config.AppLocation),
	}

	if err := s.bookItemRepo.Create(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *bookService) UpdateItem(itemID string, ownerID uuid.UUID, title, content, answer string, order int, estimateVal int, estimateUnit string, imageURL string, removeImage bool) (*entities.BookItem, error) {
	item, err := s.bookItemRepo.FindByID(itemID)
	if err != nil {
		return nil, errors.New("item not found")
	}

	book, err := s.bookRepo.FindByID(item.BookID.String())
	if err != nil {
		return nil, errors.New("book not found")
	}

	// Published books: allow anyone if is_editable=true, owner always can edit
	// Non-published books: only owner can edit
	if book.Status == entities.BookStatusPublished {
		if !book.IsEditable && book.OwnerID != ownerID {
			return nil, errors.New("this book is not editable")
		}
	} else {
		if book.OwnerID != ownerID {
			return nil, errors.New("you don't have permission to update this item")
		}
	}

	// ── Non-owner of a published book ──────────────────────────────────────
	if book.Status == entities.BookStatusPublished && book.OwnerID != ownerID {
		// Case A: item adalah milik importer ini sendiri (importer_id = ownerID)
		//         → edit langsung BookItem mereka.
		if item.ImporterID != nil && *item.ImporterID == ownerID {
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
			if estimateVal > 0 {
				item.EstimatedReviewSeconds = normalizeEstSeconds(estimateVal, estimateUnit)
			}
			if imageURL != "" {
				item.ImageURL = imageURL
			} else if removeImage {
				_ = utils.DeleteFromSupabase(item.ImageURL)
				item.ImageURL = ""
			}
			item.UpdatedAt = time.Now().In(config.AppLocation)
			if err := s.bookItemRepo.Update(item); err != nil {
				return nil, err
			}
			return item, nil
		}

		// Case B: item adalah canonical (importer_id IS NULL), artinya importer
		//         ingin meng-override konten canonical → simpan ke BookItemOverride.
		if item.ImporterID != nil {
			return nil, errors.New("you don't have permission to update this item")
		}

		if s.overrideRepo == nil {
			return nil, errors.New("override repository not available")
		}

		existing, _ := s.overrideRepo.FindByUserAndBookItemID(ownerID, item.ID)
		var base entities.BookItemOverride
		if existing != nil {
			base = *existing
		} else {
			base = entities.BookItemOverride{
				UserID:                 ownerID,
				BookItemID:             item.ID,
				Title:                  item.Title,
				Content:                item.Content,
				Answer:                 item.Answer,
				ImageURL:               item.ImageURL,
				EstimatedReviewSeconds: item.EstimatedReviewSeconds,
			}
		}
		if title != "" {
			base.Title = title
		}
		if content != "" {
			base.Content = content
		}
		if answer != "" {
			base.Answer = answer
		}
		if imageURL != "" {
			base.ImageURL = imageURL
		} else if removeImage {
			_ = utils.DeleteFromSupabase(base.ImageURL)
			base.ImageURL = ""
		}
		if estimateVal > 0 {
			base.EstimatedReviewSeconds = normalizeEstSeconds(estimateVal, estimateUnit)
		}
		base.UpdatedAt = time.Now().In(config.AppLocation)
		if err := s.overrideRepo.Upsert(&base); err != nil {
			return nil, err
		}
		result := *item
		result.Title = base.Title
		result.Content = base.Content
		result.Answer = base.Answer
		result.ImageURL = base.ImageURL
		result.EstimatedReviewSeconds = base.EstimatedReviewSeconds
		return &result, nil
	}

	// ── Owner (or draft book) → mutate the canonical BookItem ──────────────
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
	if estimateVal > 0 {
		item.EstimatedReviewSeconds = normalizeEstSeconds(estimateVal, estimateUnit)
	}
	if imageURL != "" {
		item.ImageURL = imageURL
	} else if removeImage {
		// Delete the old image from Supabase Storage before clearing the field
		_ = utils.DeleteFromSupabase(item.ImageURL)
		item.ImageURL = ""
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

	// Published books: allow anyone if is_editable=true, owner always can delete
	// Non-published books: only owner can delete
	if book.Status == entities.BookStatusPublished {
		if !book.IsEditable && book.OwnerID != ownerID {
			return errors.New("this book is not editable")
		}
	} else {
		if book.OwnerID != ownerID {
			return errors.New("you don't have permission to delete this item")
		}
	}

	// ── Non-owner of a published book → only remove their personal BookItem
	//    (importer_id = ownerID), their override, and their memorization Item row.
	//    The canonical BookItem (importer_id IS NULL) stays intact.
	if book.Status == entities.BookStatusPublished && book.OwnerID != ownerID {
		// Only allowed to delete items they personally created (importer_id = ownerID).
		if item.ImporterID == nil || *item.ImporterID != ownerID {
			return errors.New("you don't have permission to delete this item")
		}

		// Remove personal override if present.
		if s.overrideRepo != nil {
			_ = s.overrideRepo.DeleteByUserAndBookItemID(ownerID, item.ID)
		}

		// Remove only this user's memorization Item row.
		contentRef := "book:" + item.BookID.String() + ":item:" + itemID
		userItems, err := s.itemRepo.FindByOwnerAndContentRef(ownerID, contentRef)
		if err == nil {
			for _, ui := range userItems {
				_ = s.itemRepo.DeleteByID(ui.ID)
			}
		}

		// Delete the personal BookItem itself.
		return s.bookItemRepo.Delete(itemID)
	}

	// ── Owner → delete the canonical BookItem (and all memorization rows for
	//    everyone pointing to it, plus any overrides).
	contentRef := "book:" + item.BookID.String() + ":item:" + itemID
	existingItems, err := s.itemRepo.FindByContentRef(contentRef)
	if err == nil && len(existingItems) > 0 {
		for _, existingItem := range existingItems {
			if err := s.itemRepo.DeleteByID(existingItem.ID); err != nil {
				// Log error but continue with BookItem deletion
			}
		}
	}

	// Clean up any personal overrides pointing to this BookItem.
	if s.overrideRepo != nil {
		bookItemUUID, parseErr := uuid.Parse(itemID)
		if parseErr == nil {
			_ = s.overrideRepo.DeleteByUserAndBookItemID(ownerID, bookItemUUID)
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

	isOwner := book.OwnerID == userID
	var isClassroomMember bool
	if s.classBookRepo != nil {
		isClassBook, err := s.classBookRepo.IsBookAssignedToClass(bookID)
		if err == nil && isClassBook {
			isClassroomMember = s.canAccessClassBook(bookID, book.OwnerID, &userID)
		}
	}
	var isImporter bool
	if s.classBookRepo != nil {
		isImported, err := s.classBookRepo.IsBookImportedByUser(bookID, userID.String())
		if err == nil && isImported && book.Status == entities.BookStatusPublished {
			isImporter = true
		}
	}

	if !isOwner && !isClassroomMember && !isImporter {
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
	importedBooks, err := s.classBookRepo.FindImportedBooksByUserID(userID.String())
	if err != nil {
		return nil, err
	}

	result := make([]BookCollectionItem, 0, len(importedBooks))
	for _, ib := range importedBooks {
		bookID := ib.BookID.String()

		book, err := s.bookRepo.FindByID(bookID)
		if err != nil {
			continue
		}

		ownerName := ""
		owner, err := s.userRepo.FindByID(book.OwnerID.String())
		if err == nil && owner != nil {
			ownerName = owner.FullName
		}

		items, err := s.itemRepo.FindByOwnerAndSourceType(userID, "book")
		itemCount := 0
		if err == nil {
			for _, item := range items {
				parts := strings.Split(item.ContentRef, ":")
				if len(parts) == 4 && parts[0] == "book" && parts[1] == bookID {
					itemCount++
				}
			}
		}

		result = append(result, BookCollectionItem{
			BookID:      bookID,
			Title:       book.Title,
			Description: book.Description,
			CoverImage:  book.CoverImage,
			OwnerName:   ownerName,
			ItemCount:   itemCount,
			AddedAt:     ib.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return result, nil
}

func (s *bookService) RemoveFromMyBookCollection(userID uuid.UUID, bookID string) error {
	_, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return errors.New("book not found")
	}

	_ = s.classBookRepo.DeleteImportedBook(userID.String(), bookID)

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

// ==================== BOOK ITEM OVERRIDES ====================

// GetMyOverride retrieves the user's personal override for a BookItem.
// Returns nil if no override exists (user sees canonical content).
func (s *bookService) GetMyOverride(userID uuid.UUID, bookItemID string) (*entities.BookItemOverride, error) {
	bookItemUUID, err := uuid.Parse(bookItemID)
	if err != nil {
		return nil, errors.New("invalid book_item_id")
	}

	// Verify the BookItem exists first.
	_, err = s.bookItemRepo.FindByID(bookItemID)
	if err != nil {
		return nil, errors.New("book item not found")
	}

	if s.overrideRepo == nil {
		return nil, errors.New("override repository not available")
	}

	override, err := s.overrideRepo.FindByUserAndBookItemID(userID, bookItemUUID)
	if err != nil {
		return nil, err
	}
	if override == nil {
		return nil, errors.New("no override found for this item")
	}

	return override, nil
}

// RemoveMyOverride deletes the user's personal override for a BookItem,
// restoring their view to the canonical content.
func (s *bookService) RemoveMyOverride(userID uuid.UUID, bookItemID string) error {
	bookItemUUID, err := uuid.Parse(bookItemID)
	if err != nil {
		return errors.New("invalid book_item_id")
	}

	// Verify the BookItem exists first.
	_, err = s.bookItemRepo.FindByID(bookItemID)
	if err != nil {
		return errors.New("book item not found")
	}

	if s.overrideRepo == nil {
		return errors.New("override repository not available")
	}

	return s.overrideRepo.DeleteByUserAndBookItemID(userID, bookItemUUID)
}

type ModuleReorderInput struct {
	ID       uuid.UUID  `json:"id"`
	Order    int        `json:"order"`
	ParentID *uuid.UUID `json:"parent_id"`
}

type ItemReorderInput struct {
	ID       uuid.UUID  `json:"id"`
	Order    int        `json:"order"`
	ModuleID *uuid.UUID `json:"module_id"`
}

type ReorderBookStructureRequest struct {
	Modules []ModuleReorderInput `json:"modules"`
	Items   []ItemReorderInput   `json:"items"`
}

func (s *bookService) ReorderBookStructure(bookID string, ownerID uuid.UUID, req ReorderBookStructureRequest) error {
	bookUUID, err := uuid.Parse(bookID)
	if err != nil {
		return errors.New("invalid book_id")
	}

	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return errors.New("book not found")
	}

	if book.OwnerID != ownerID {
		return errors.New("unauthorized: you don't own this book")
	}

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update modules order and parent_id
	for _, m := range req.Modules {
		updates := map[string]interface{}{
			"order":     m.Order,
			"parent_id": m.ParentID,
		}
		if err := tx.Model(&entities.BookModule{}).Where("id = ? AND book_id = ?", m.ID, bookUUID).Updates(updates).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Update items order and module_id
	for _, it := range req.Items {
		updates := map[string]interface{}{
			"order":     it.Order,
			"module_id": it.ModuleID,
		}
		if err := tx.Model(&entities.BookItem{}).Where("id = ? AND book_id = ?", it.ID, bookUUID).Updates(updates).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

