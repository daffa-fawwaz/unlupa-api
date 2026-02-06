package services

import (
	"errors"
	"time"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"

	"github.com/google/uuid"
)

type BookService interface {
	// Book CRUD
	CreateBook(ownerID uuid.UUID, title, description, coverImage string) (*entities.Book, error)
	GetMyBooks(ownerID uuid.UUID) ([]entities.Book, error)
	GetPublishedBooks() ([]entities.Book, error)
	GetBookDetail(bookID string, userID *uuid.UUID) (*entities.Book, error)
	UpdateBook(bookID string, ownerID uuid.UUID, title, description, coverImage string) (*entities.Book, error)
	DeleteBook(bookID string, ownerID uuid.UUID) error

	// Publish workflow
	RequestPublish(bookID string, ownerID uuid.UUID) error
	GetPendingBooks() ([]entities.Book, error)
	ApproveBook(bookID string) error
	RejectBook(bookID string) error

	// Module CRUD
	AddModule(bookID string, ownerID uuid.UUID, title, description string, order int, parentID *uuid.UUID) (*entities.BookModule, error)
	UpdateModule(moduleID string, ownerID uuid.UUID, title, description string, order int) (*entities.BookModule, error)
	DeleteModule(moduleID string, ownerID uuid.UUID) error

	// Item CRUD
	AddItem(bookID string, moduleID *uuid.UUID, ownerID uuid.UUID, title, content string, order int) (*entities.BookItem, error)
	UpdateItem(itemID string, ownerID uuid.UUID, title, content string, order int) (*entities.BookItem, error)
	DeleteItem(itemID string, ownerID uuid.UUID) error
}

type bookService struct {
	bookRepo       repositories.BookRepository
	bookModuleRepo repositories.BookModuleRepository
	bookItemRepo   repositories.BookItemRepository
}

func NewBookService(
	bookRepo repositories.BookRepository,
	bookModuleRepo repositories.BookModuleRepository,
	bookItemRepo repositories.BookItemRepository,
) BookService {
	return &bookService{
		bookRepo:       bookRepo,
		bookModuleRepo: bookModuleRepo,
		bookItemRepo:   bookItemRepo,
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

func (s *bookService) GetBookDetail(bookID string, userID *uuid.UUID) (*entities.Book, error) {
	book, err := s.bookRepo.FindByIDWithRelations(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	// If not published, only owner can view
	if book.Status != entities.BookStatusPublished {
		if userID == nil || book.OwnerID != *userID {
			return nil, errors.New("you don't have access to this book")
		}
	}

	return book, nil
}

func (s *bookService) UpdateBook(bookID string, ownerID uuid.UUID, title, description, coverImage string) (*entities.Book, error) {
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	if book.OwnerID != ownerID {
		return nil, errors.New("you don't have permission to update this book")
	}

	if book.Status == entities.BookStatusPublished {
		return nil, errors.New("cannot update published book")
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

// ==================== MODULE CRUD ====================

func (s *bookService) AddModule(bookID string, ownerID uuid.UUID, title, description string, order int, parentID *uuid.UUID) (*entities.BookModule, error) {
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	if book.OwnerID != ownerID {
		return nil, errors.New("you don't have permission to add module to this book")
	}

	if book.Status == entities.BookStatusPublished {
		return nil, errors.New("cannot add module to published book")
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

	// Delete all items in this module
	if err := s.bookItemRepo.DeleteByModuleID(moduleID); err != nil {
		return err
	}

	return s.bookModuleRepo.Delete(moduleID)
}

// ==================== ITEM CRUD ====================

func (s *bookService) AddItem(bookID string, moduleID *uuid.UUID, ownerID uuid.UUID, title, content string, order int) (*entities.BookItem, error) {
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	if book.OwnerID != ownerID {
		return nil, errors.New("you don't have permission to add item to this book")
	}

	if book.Status == entities.BookStatusPublished {
		return nil, errors.New("cannot add item to published book")
	}

	if title == "" {
		return nil, errors.New("item title is required")
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

	item := &entities.BookItem{
		BookID:    uuid.MustParse(bookID),
		ModuleID:  moduleID,
		Title:     title,
		Content:   content,
		Order:     order,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.bookItemRepo.Create(item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *bookService) UpdateItem(itemID string, ownerID uuid.UUID, title, content string, order int) (*entities.BookItem, error) {
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
	if order > 0 {
		item.Order = order
	}
	item.UpdatedAt = time.Now()

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

	return s.bookItemRepo.Delete(itemID)
}
