package repositories

import (
	"errors"
	"strings"

	"hifzhun-api/pkg/entities"
	"gorm.io/gorm"
)

var ErrEmailAlreadyExists = errors.New("email already exists")
var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(user *entities.User) error
	FindByEmail(email string) (*entities.User, error)
	FindByID(id string) (*entities.User, error)
	GetAllUsers(role string) ([]entities.User, error)
	UpdateRole(id string, role string) error
	ActivateUser(id string) error
	DeactivateUser(id string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) Create(user *entities.User) error {
	err := r.db.Create(user).Error
	if err != nil {
		// PostgreSQL unique constraint
		if strings.Contains(err.Error(), "idx_users_email") {
			return ErrEmailAlreadyExists
		}
		return err
	}
	return nil
}

func (r *userRepository) FindByEmail(email string) (*entities.User, error) {
	var user entities.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *userRepository) FindByID(id string) (*entities.User, error) {
	var user entities.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *userRepository) GetAllUsers(role string) ([]entities.User, error) {
	var users []entities.User
	query := r.db.Order("created_at DESC")

	if role != "" {
		query = query.Where("role = ?", role)
	}

	err := query.Find(&users).Error
	return users, err
}

func (r *userRepository) UpdateRole(id string, role string) error {
	return r.db.Model(&entities.User{}).
		Where("id = ?", id).
		Update("role", role).
		Error
}

func (r *userRepository) ActivateUser(id string) error {
	res := r.db.Model(&entities.User{}).
		Where("id = ?", id).
		Update("is_active", true)

	if res.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return res.Error
}

func (r *userRepository) DeactivateUser(id string) error {
	res := r.db.Model(&entities.User{}).
		Where("id = ?", id).
		Update("is_active", false)

	if res.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return res.Error
}
