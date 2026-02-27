package services

import (
	"errors"
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"

	"github.com/google/uuid"
)

type TeacherRequestService interface {
	RequestTeacher(userID uuid.UUID, message string) error
	GetMyRequest(userID uuid.UUID) (*entities.TeacherRequest, error)
	GetPendingRequests() ([]entities.TeacherRequest, error)
	ApproveRequest(id string) error
	RejectRequest(id string) error
	GetStats() (*repositories.TeacherRequestStats, error)
}

type teacherRequestService struct {
	teacherReqRepo repositories.TeacherRequestRepository
	userRepo       repositories.UserRepository
}

func NewTeacherRequestService(
	teacherReqRepo repositories.TeacherRequestRepository,
	userRepo repositories.UserRepository,
) TeacherRequestService {
	return &teacherRequestService{
		teacherReqRepo: teacherReqRepo,
		userRepo:       userRepo,
	}
}

func (s *teacherRequestService) RequestTeacher(userID uuid.UUID, message string) error {
	user, err := s.userRepo.FindByID(userID.String())
	if err != nil {
		return errors.New("user not found")
	}

	if user.Role == "teacher" {
		return errors.New("you are already a teacher")
	}

	if user.Role == "admin" {
		return errors.New("admin cannot request to become teacher")
	}

	req := &entities.TeacherRequest{
		UserID:  userID,
		Message: message,
		Status:  "pending",
	}

	return s.teacherReqRepo.Create(req)
}

func (s *teacherRequestService) GetMyRequest(userID uuid.UUID) (*entities.TeacherRequest, error) {
	return s.teacherReqRepo.FindByUserID(userID)
}

func (s *teacherRequestService) GetPendingRequests() ([]entities.TeacherRequest, error) {
	return s.teacherReqRepo.GetPendingRequests()
}

func (s *teacherRequestService) ApproveRequest(id string) error {
	req, err := s.teacherReqRepo.FindByID(id)
	if err != nil {
		return errors.New("teacher request not found")
	}

	if req.Status != "pending" {
		return errors.New("request already processed")
	}

	if err := s.teacherReqRepo.UpdateStatus(id, "approved"); err != nil {
		return err
	}

	if err := s.userRepo.UpdateRole(req.UserID.String(), "teacher"); err != nil {
		return err
	}

	return s.userRepo.ActivateUser(req.UserID.String())
}

func (s *teacherRequestService) RejectRequest(id string) error {
	req, err := s.teacherReqRepo.FindByID(id)
	if err != nil {
		return errors.New("teacher request not found")
	}

	if req.Status != "pending" {
		return errors.New("request already processed")
	}

	return s.teacherReqRepo.UpdateStatus(id, "rejected")
}

func (s *teacherRequestService) GetStats() (*repositories.TeacherRequestStats, error) {
	return s.teacherReqRepo.GetStats()
}

