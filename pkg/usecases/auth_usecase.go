package usecases

import (
	"errors"
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
)

type AuthUsecase interface {
	Register(user *entities.User) error
	Login(email, password string) (*entities.User, string, error)
}

type authUsecase struct {
	userRepo repositories.UserRepository
	authSvc  services.AuthService
}

func NewAuthUsecase(
	userRepo repositories.UserRepository,
	authSvc services.AuthService,
) AuthUsecase {
	return &authUsecase{userRepo, authSvc}
}

func (u *authUsecase) Register(user *entities.User) error {
	hashed, err := u.authSvc.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashed

	return u.userRepo.Create(user)
}

func (u *authUsecase) Login(email, password string) (*entities.User, string, error) {
	user, err := u.userRepo.FindByEmail(email)
	if err != nil {
		return nil, "", errors.New("email not found")
	}

	if !user.IsActive {
		return nil, "", errors.New("account not active, waiting admin approval")
	}

	if err := u.authSvc.CheckPassword(user.Password, password); err != nil {
		return nil, "", errors.New("wrong password")
	}

	token, err := u.authSvc.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return user, token, nil
}

