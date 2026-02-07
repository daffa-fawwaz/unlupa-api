package repositories

import (
	"hifzhun-api/pkg/entities"

	"gorm.io/gorm"
)

type ClassMemberRepository interface {
	Create(member *entities.ClassMember) error
	FindByClassID(classID string) ([]entities.ClassMember, error)
	FindByUserID(userID string) ([]entities.ClassMember, error)
	FindByClassAndUser(classID, userID string) (*entities.ClassMember, error)
	IsMember(classID, userID string) (bool, error)
	Delete(id string) error
	DeleteByClassAndUser(classID, userID string) error
	DeleteByClassID(classID string) error
}

type classMemberRepository struct {
	db *gorm.DB
}

func NewClassMemberRepository(db *gorm.DB) ClassMemberRepository {
	return &classMemberRepository{db}
}

func (r *classMemberRepository) Create(member *entities.ClassMember) error {
	return r.db.Create(member).Error
}

func (r *classMemberRepository) FindByClassID(classID string) ([]entities.ClassMember, error) {
	var members []entities.ClassMember
	err := r.db.
		Where("class_id = ?", classID).
		Order("joined_at ASC").
		Find(&members).Error
	return members, err
}

func (r *classMemberRepository) FindByUserID(userID string) ([]entities.ClassMember, error) {
	var members []entities.ClassMember
	err := r.db.
		Where("user_id = ?", userID).
		Find(&members).Error
	return members, err
}

func (r *classMemberRepository) FindByClassAndUser(classID, userID string) (*entities.ClassMember, error) {
	var member entities.ClassMember
	err := r.db.
		Where("class_id = ? AND user_id = ?", classID, userID).
		First(&member).Error
	return &member, err
}

func (r *classMemberRepository) IsMember(classID, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&entities.ClassMember{}).
		Where("class_id = ? AND user_id = ?", classID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *classMemberRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&entities.ClassMember{}).Error
}

func (r *classMemberRepository) DeleteByClassAndUser(classID, userID string) error {
	return r.db.
		Where("class_id = ? AND user_id = ?", classID, userID).
		Delete(&entities.ClassMember{}).Error
}

func (r *classMemberRepository) DeleteByClassID(classID string) error {
	return r.db.Where("class_id = ?", classID).Delete(&entities.ClassMember{}).Error
}
