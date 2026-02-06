package repositories

import (
	"hifzhun-api/pkg/entities"

	"gorm.io/gorm"
)

type BookItemRepository interface {
	Create(item *entities.BookItem) error
	FindByID(id string) (*entities.BookItem, error)
	FindByBookID(bookID string) ([]entities.BookItem, error)
	FindByModuleID(moduleID string) ([]entities.BookItem, error)
	Update(item *entities.BookItem) error
	Delete(id string) error
	DeleteByBookID(bookID string) error
	DeleteByModuleID(moduleID string) error
}

type bookItemRepository struct {
	db *gorm.DB
}

func NewBookItemRepository(db *gorm.DB) BookItemRepository {
	return &bookItemRepository{db}
}

func (r *bookItemRepository) Create(item *entities.BookItem) error {
	return r.db.Create(item).Error
}

func (r *bookItemRepository) FindByID(id string) (*entities.BookItem, error) {
	var item entities.BookItem
	err := r.db.Where("id = ?", id).First(&item).Error
	return &item, err
}

func (r *bookItemRepository) FindByBookID(bookID string) ([]entities.BookItem, error) {
	var items []entities.BookItem
	err := r.db.
		Where("book_id = ?", bookID).
		Order("\"order\" ASC").
		Find(&items).Error
	return items, err
}

func (r *bookItemRepository) FindByModuleID(moduleID string) ([]entities.BookItem, error) {
	var items []entities.BookItem
	err := r.db.
		Where("module_id = ?", moduleID).
		Order("\"order\" ASC").
		Find(&items).Error
	return items, err
}

func (r *bookItemRepository) Update(item *entities.BookItem) error {
	return r.db.Save(item).Error
}

func (r *bookItemRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&entities.BookItem{}).Error
}

func (r *bookItemRepository) DeleteByBookID(bookID string) error {
	return r.db.Where("book_id = ?", bookID).Delete(&entities.BookItem{}).Error
}

func (r *bookItemRepository) DeleteByModuleID(moduleID string) error {
	return r.db.Where("module_id = ?", moduleID).Delete(&entities.BookItem{}).Error
}
