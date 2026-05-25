package repositories

import (
	"hifzhun-api/pkg/entities"

	"gorm.io/gorm"
)

type ClassBookRepository interface {
	Create(classBook *entities.ClassBook) error
	FindByID(id string) (*entities.ClassBook, error)
	FindByClassID(classID string) ([]entities.ClassBook, error)
	FindByClassAndBook(classID, bookID string) (*entities.ClassBook, error)
	IsBookAssignedToClass(bookID string) (bool, error)
	IsBookAccessibleByMember(bookID, userID string) (bool, error)
	Delete(id string) error
	DeleteByClassID(classID string) error
	DeleteByClassAndBook(classID, bookID string) error
}

type classBookRepository struct {
	db *gorm.DB
}

func NewClassBookRepository(db *gorm.DB) ClassBookRepository {
	return &classBookRepository{db}
}

func (r *classBookRepository) Create(classBook *entities.ClassBook) error {
	return r.db.Create(classBook).Error
}

func (r *classBookRepository) FindByID(id string) (*entities.ClassBook, error) {
	var classBook entities.ClassBook
	err := r.db.
		Preload("Book").
		Where("id = ?", id).
		First(&classBook).Error
	return &classBook, err
}

func (r *classBookRepository) FindByClassID(classID string) ([]entities.ClassBook, error) {
	var classBooks []entities.ClassBook
	err := r.db.
		Preload("Book").
		Where("class_id = ?", classID).
		Order("\"order\" ASC").
		Find(&classBooks).Error
	return classBooks, err
}

func (r *classBookRepository) FindByClassAndBook(classID, bookID string) (*entities.ClassBook, error) {
	var classBook entities.ClassBook
	err := r.db.
		Where("class_id = ? AND book_id = ?", classID, bookID).
		First(&classBook).Error
	return &classBook, err
}

func (r *classBookRepository) IsBookAssignedToClass(bookID string) (bool, error) {
	var count int64
	err := r.db.Model(&entities.ClassBook{}).
		Where("book_id = ?", bookID).
		Count(&count).Error
	return count > 0, err
}

func (r *classBookRepository) IsBookAccessibleByMember(bookID, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&entities.ClassBook{}).
		Joins("JOIN class_members ON class_members.class_id = class_books.class_id").
		Where("class_books.book_id = ? AND class_members.user_id = ?", bookID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *classBookRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&entities.ClassBook{}).Error
}

func (r *classBookRepository) DeleteByClassID(classID string) error {
	return r.db.Where("class_id = ?", classID).Delete(&entities.ClassBook{}).Error
}

func (r *classBookRepository) DeleteByClassAndBook(classID, bookID string) error {
	return r.db.
		Where("class_id = ? AND book_id = ?", classID, bookID).
		Delete(&entities.ClassBook{}).Error
}
