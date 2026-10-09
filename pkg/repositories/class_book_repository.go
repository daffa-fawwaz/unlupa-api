package repositories

import (
	"hifzhun-api/pkg/entities"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClassBookRepository interface {
	Create(classBook *entities.ClassBook) error
	FindByID(id string) (*entities.ClassBook, error)
	FindByClassID(classID string) ([]entities.ClassBook, error)
	FindByClassAndBook(classID, bookID string) (*entities.ClassBook, error)
	IsBookAssignedToClass(bookID string) (bool, error)
	IsBookAccessibleByMember(bookID, userID string) (bool, error)
	// IsBookAccessibleByTeacher returns true when userID is the guru of any class that contains bookID.
	IsBookAccessibleByTeacher(bookID, userID string) (bool, error)
	// IsBookOwner returns true when userID is the owner of the book.
	IsBookOwner(bookID, userID string) (bool, error)
	// IsBookPublished returns true when status of the book is published.
	IsBookPublished(bookID string) (bool, error)
	CreateImportedBook(userID, bookID string) error
	IsBookImportedByUser(bookID, userID string) (bool, error)
	FindImportedBooksByUserID(userID string) ([]entities.ImportedBook, error)
	DeleteImportedBook(userID, bookID string) error
	CountByClassID(classID string) (int64, error)
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
	classUUID, err := uuid.Parse(classID)
	if err != nil {
		return nil, err
	}
	bookUUID, err := uuid.Parse(bookID)
	if err != nil {
		return nil, err
	}
	var classBook entities.ClassBook
	err = r.db.
		Where("class_id = ? AND book_id = ?", classUUID, bookUUID).
		First(&classBook).Error
	return &classBook, err
}

func (r *classBookRepository) IsBookAssignedToClass(bookID string) (bool, error) {
	bookUUID, err := uuid.Parse(bookID)
	if err != nil {
		return false, err
	}
	var count int64
	err = r.db.Model(&entities.ClassBook{}).
		Where("book_id = ?", bookUUID).
		Count(&count).Error
	return count > 0, err
}

func (r *classBookRepository) IsBookAccessibleByMember(bookID, userID string) (bool, error) {
	bookUUID, err := uuid.Parse(bookID)
	if err != nil {
		return false, err
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}
	var count int64
	err = r.db.Model(&entities.ClassBook{}).
		Joins("JOIN class_members ON class_members.class_id = class_books.class_id").
		Where("class_books.book_id = ? AND class_members.user_id = ?", bookUUID, userUUID).
		Count(&count).Error
	return count > 0, err
}

// IsBookAccessibleByTeacher returns true when userID is the guru_id of any class that contains bookID.
func (r *classBookRepository) IsBookAccessibleByTeacher(bookID, userID string) (bool, error) {
	bookUUID, err := uuid.Parse(bookID)
	if err != nil {
		return false, err
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}
	var count int64
	err = r.db.Model(&entities.ClassBook{}).
		Joins("JOIN classes ON classes.id = class_books.class_id").
		Where("class_books.book_id = ? AND classes.guru_id = ?", bookUUID, userUUID).
		Count(&count).Error
	return count > 0, err
}

// IsBookOwner returns true when userID is the owner_id of the book.
func (r *classBookRepository) IsBookOwner(bookID, userID string) (bool, error) {
	bookUUID, err := uuid.Parse(bookID)
	if err != nil {
		return false, err
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}
	var count int64
	err = r.db.Model(&entities.Book{}).
		Where("id = ? AND owner_id = ?", bookUUID, userUUID).
		Count(&count).Error
	return count > 0, err
}

// IsBookPublished returns true when status of the book is published.
func (r *classBookRepository) IsBookPublished(bookID string) (bool, error) {
	bookUUID, err := uuid.Parse(bookID)
	if err != nil {
		return false, err
	}
	var count int64
	err = r.db.Model(&entities.Book{}).
		Where("id = ? AND status = ?", bookUUID, "published").
		Count(&count).Error
	return count > 0, err
}

func (r *classBookRepository) CreateImportedBook(userID, bookID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	bookUUID, err := uuid.Parse(bookID)
	if err != nil {
		return err
	}
	imported := &entities.ImportedBook{
		ID:     uuid.New(),
		UserID: userUUID,
		BookID: bookUUID,
	}
	return r.db.Create(imported).Error
}

func (r *classBookRepository) IsBookImportedByUser(bookID, userID string) (bool, error) {
	bookUUID, err := uuid.Parse(bookID)
	if err != nil {
		return false, err
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}
	var count int64
	err = r.db.Model(&entities.ImportedBook{}).
		Where("book_id = ? AND user_id = ?", bookUUID, userUUID).
		Count(&count).Error
	return count > 0, err
}

func (r *classBookRepository) FindImportedBooksByUserID(userID string) ([]entities.ImportedBook, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	var imported []entities.ImportedBook
	err = r.db.Where("user_id = ?", userUUID).Find(&imported).Error
	return imported, err
}

func (r *classBookRepository) DeleteImportedBook(userID, bookID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	bookUUID, err := uuid.Parse(bookID)
	if err != nil {
		return err
	}
	return r.db.Where("user_id = ? AND book_id = ?", userUUID, bookUUID).Delete(&entities.ImportedBook{}).Error
}

func (r *classBookRepository) CountByClassID(classID string) (int64, error) {
	classUUID, err := uuid.Parse(classID)
	if err != nil {
		return 0, err
	}
	var count int64
	err = r.db.Model(&entities.ClassBook{}).
		Where("class_id = ?", classUUID).
		Count(&count).Error
	return count, err
}

func (r *classBookRepository) Delete(id string) error {
	itemUUID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return r.db.Where("id = ?", itemUUID).Delete(&entities.ClassBook{}).Error
}

func (r *classBookRepository) DeleteByClassID(classID string) error {
	classUUID, err := uuid.Parse(classID)
	if err != nil {
		return err
	}
	return r.db.Where("class_id = ?", classUUID).Delete(&entities.ClassBook{}).Error
}

func (r *classBookRepository) DeleteByClassAndBook(classID, bookID string) error {
	classUUID, err := uuid.Parse(classID)
	if err != nil {
		return err
	}
	bookUUID, err := uuid.Parse(bookID)
	if err != nil {
		return err
	}
	return r.db.
		Where("class_id = ? AND book_id = ?", classUUID, bookUUID).
		Delete(&entities.ClassBook{}).Error
}
