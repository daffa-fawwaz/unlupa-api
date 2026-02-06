package repositories

import (
	"hifzhun-api/pkg/entities"

	"gorm.io/gorm"
)

type BookModuleRepository interface {
	Create(module *entities.BookModule) error
	FindByID(id string) (*entities.BookModule, error)
	FindByBookID(bookID string) ([]entities.BookModule, error)
	Update(module *entities.BookModule) error
	Delete(id string) error
	DeleteByBookID(bookID string) error
}

type bookModuleRepository struct {
	db *gorm.DB
}

func NewBookModuleRepository(db *gorm.DB) BookModuleRepository {
	return &bookModuleRepository{db}
}

func (r *bookModuleRepository) Create(module *entities.BookModule) error {
	return r.db.Create(module).Error
}

func (r *bookModuleRepository) FindByID(id string) (*entities.BookModule, error) {
	var module entities.BookModule
	err := r.db.Where("id = ?", id).First(&module).Error
	return &module, err
}

func (r *bookModuleRepository) FindByBookID(bookID string) ([]entities.BookModule, error) {
	var modules []entities.BookModule
	err := r.db.
		Where("book_id = ?", bookID).
		Order("\"order\" ASC").
		Find(&modules).Error
	return modules, err
}

func (r *bookModuleRepository) Update(module *entities.BookModule) error {
	return r.db.Save(module).Error
}

func (r *bookModuleRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&entities.BookModule{}).Error
}

func (r *bookModuleRepository) DeleteByBookID(bookID string) error {
	return r.db.Where("book_id = ?", bookID).Delete(&entities.BookModule{}).Error
}
