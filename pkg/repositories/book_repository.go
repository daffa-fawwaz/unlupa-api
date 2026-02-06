package repositories

import (
	"time"

	"hifzhun-api/pkg/entities"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookRepository interface {
	Create(book *entities.Book) error
	FindByID(id string) (*entities.Book, error)
	FindByIDWithRelations(id string) (*entities.Book, error)
	FindByOwner(ownerID string) ([]entities.Book, error)
	FindPublished() ([]entities.Book, error)
	FindPendingPublish() ([]entities.Book, error)
	Update(book *entities.Book) error
	UpdateStatus(id, status string) error
	Delete(id string) error
}

type bookRepository struct {
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) BookRepository {
	return &bookRepository{db}
}

func (r *bookRepository) Create(book *entities.Book) error {
	return r.db.Create(book).Error
}

func (r *bookRepository) FindByID(id string) (*entities.Book, error) {
	var book entities.Book
	err := r.db.Where("id = ?", id).First(&book).Error
	return &book, err
}

func (r *bookRepository) FindByIDWithRelations(id string) (*entities.Book, error) {
	var book entities.Book
	err := r.db.
		Preload("Modules", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Preload("Modules.Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Where("module_id IS NULL").Order("\"order\" ASC")
		}).
		Where("id = ?", id).
		First(&book).Error
	return &book, err
}

func (r *bookRepository) FindByOwner(ownerID string) ([]entities.Book, error) {
	var books []entities.Book
	err := r.db.
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Find(&books).Error
	return books, err
}

func (r *bookRepository) FindPublished() ([]entities.Book, error) {
	var books []entities.Book
	err := r.db.
		Where("status = ?", entities.BookStatusPublished).
		Order("published_at DESC").
		Find(&books).Error
	return books, err
}

func (r *bookRepository) FindPendingPublish() ([]entities.Book, error) {
	var books []entities.Book
	err := r.db.
		Where("status = ?", entities.BookStatusPending).
		Order("created_at ASC").
		Find(&books).Error
	return books, err
}

func (r *bookRepository) Update(book *entities.Book) error {
	return r.db.Save(book).Error
}

func (r *bookRepository) UpdateStatus(id, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == entities.BookStatusPublished {
		now := time.Now()
		updates["published_at"] = &now
	}

	return r.db.Model(&entities.Book{}).
		Where("id = ?", uuid.MustParse(id)).
		Updates(updates).Error
}

func (r *bookRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&entities.Book{}).Error
}
