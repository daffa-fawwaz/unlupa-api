package repositories

import (
	"hifzhun-api/pkg/entities"

	"gorm.io/gorm"
)

type ClassRepository interface {
	Create(class *entities.Class) error
	FindByID(id string) (*entities.Class, error)
	FindByIDWithRelations(id string) (*entities.Class, error)
	FindByCode(code string) (*entities.Class, error)
	FindByTeacher(teacherID string) ([]entities.Class, error)
	FindByTeacherAndType(teacherID, classType string) ([]entities.Class, error)
	Update(class *entities.Class) error
	Delete(id string) error
	IsCodeExists(code string) (bool, error)
}

type classRepository struct {
	db *gorm.DB
}

func NewClassRepository(db *gorm.DB) ClassRepository {
	return &classRepository{db}
}

func (r *classRepository) Create(class *entities.Class) error {
	return r.db.Create(class).Error
}

func (r *classRepository) FindByID(id string) (*entities.Class, error) {
	var class entities.Class
	err := r.db.Where("id = ?", id).First(&class).Error
	return &class, err
}

func (r *classRepository) FindByIDWithRelations(id string) (*entities.Class, error) {
	var class entities.Class
	err := r.db.
		Preload("Members").
		Preload("Books").
		Preload("Books.Book").
		Where("id = ?", id).
		First(&class).Error
	return &class, err
}

func (r *classRepository) FindByCode(code string) (*entities.Class, error) {
	var class entities.Class
	err := r.db.Where("class_code = ?", code).First(&class).Error
	return &class, err
}

func (r *classRepository) FindByTeacher(teacherID string) ([]entities.Class, error) {
	var classes []entities.Class
	err := r.db.
		Where("guru_id = ?", teacherID).
		Order("created_at DESC").
		Find(&classes).Error
	return classes, err
}

func (r *classRepository) FindByTeacherAndType(teacherID, classType string) ([]entities.Class, error) {
	var classes []entities.Class
	err := r.db.
		Where("guru_id = ? AND type = ?", teacherID, classType).
		Order("created_at DESC").
		Find(&classes).Error
	return classes, err
}

func (r *classRepository) Update(class *entities.Class) error {
	return r.db.Save(class).Error
}

func (r *classRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&entities.Class{}).Error
}

func (r *classRepository) IsCodeExists(code string) (bool, error) {
	var count int64
	err := r.db.Model(&entities.Class{}).Where("class_code = ?", code).Count(&count).Error
	return count > 0, err
}
