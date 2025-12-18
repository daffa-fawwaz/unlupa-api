package services

import (
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
)

type UserService interface {
	GetAllUsers(role string) ([]entities.User, error)
	ActivateUser(id string) error
	DeactivateUser(id string) error
}

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

// ================= GET ALL USERS =================
func (s *userService) GetAllUsers(role string) ([]entities.User, error) {
	return s.userRepo.GetAllUsers(role)
}

// ================= ACTIVATE USER =================
func (s *userService) ActivateUser(id string) error {
	return s.userRepo.ActivateUser(id)
}

// ================= DEACTIVATE USER =================
func (s *userService) DeactivateUser(id string) error {
	return s.userRepo.DeactivateUser(id)
}
