package services

import (
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"

	"github.com/google/uuid"
)

type ClassService interface {
	// Teacher methods
	CreateClass(teacherID uuid.UUID, name, description, classType string) (*entities.Class, error)
	GetMyClasses(teacherID uuid.UUID) ([]entities.Class, error)
	GetClassDetail(classID string, userID uuid.UUID) (*entities.Class, error)
	UpdateClass(classID string, teacherID uuid.UUID, name, description string, isActive *bool) (*entities.Class, error)
	DeleteClass(classID string, teacherID uuid.UUID) error
	AddBookToClass(classID string, teacherID uuid.UUID, bookID string, order int) (*entities.ClassBook, error)
	RemoveBookFromClass(classID string, teacherID uuid.UUID, bookID string) error
	GetStudentProgress(classID string, teacherID uuid.UUID) ([]StudentProgress, error)
	GetPendingGraduations(classID string, teacherID uuid.UUID) ([]PendingGraduation, error)
	ApproveGraduation(classID string, teacherID uuid.UUID, itemID string) error
	RejectGraduation(classID string, teacherID uuid.UUID, itemID string) error

	// Student methods
	JoinClass(userID uuid.UUID, classCode string) (*entities.Class, error)
	LeaveClass(userID uuid.UUID, classID string) error
	GetMyJoinedClasses(userID uuid.UUID) ([]entities.Class, error)
	GetClassBooks(classID string, userID uuid.UUID) ([]entities.ClassBook, error)
	GetClassMembers(classID string, userID uuid.UUID) ([]MemberInfo, error)
}

// ItemDetail represents detailed information about a single hafalan item
// @Description Detail of a single hafalan item including its current phase/status
type ItemDetail struct {
	// UUID of the item
	ItemID uuid.UUID `json:"item_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Content reference (e.g., "surah:78:1-10" or "page:582")
	ContentRef string `json:"content_ref" example:"surah:78:1-10"`
	// Current status/phase: menghafal, interval, fsrs_active, graduate
	Status string `json:"status" example:"interval"`
	// Interval days (only for interval phase)
	IntervalDays int `json:"interval_days,omitempty" example:"7"`
	// When interval will end (only for interval phase)
	IntervalEndAt *time.Time `json:"interval_end_at,omitempty"`
	// Next review date (only for fsrs_active phase)
	NextReviewAt *time.Time `json:"next_review_at,omitempty"`
	// FSRS stability score (only for fsrs_active phase)
	Stability float64 `json:"stability,omitempty" example:"5.5"`
	// When the item was created
	CreatedAt time.Time `json:"created_at"`
}

// StudentProgress represents the progress of a student in a Quran class
// @Description Progress data for a student in a Quran-type class, showing hafalan statistics and item details
type StudentProgress struct {
	// UUID of the student
	UserID uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Student's email address
	Email string `json:"email" example:"student@example.com"`
	// Student's full name
	FullName string `json:"full_name" example:"Ahmad Abdullah"`
	// Total number of hafalan items the student has
	TotalItems int `json:"total_items" example:"30"`
	// Number of items in 'menghafal' status (currently memorizing)
	Menghafal int `json:"menghafal" example:"5"`
	// Number of items in 'interval' status (waiting for interval period to complete)
	Interval int `json:"interval" example:"8"`
	// Number of items in 'fsrs_active' status (actively being reviewed with FSRS algorithm)
	FSRSActive int `json:"fsrs_active" example:"12"`
	// Number of items in 'pending_graduate' status (waiting for teacher approval)
	PendingGraduate int `json:"pending_graduate" example:"2"`
	// Number of items in 'graduate' status (mastered/completed memorization)
	Graduate int `json:"graduate" example:"5"`
	// Overall progress percentage (graduate / total_items * 100)
	ProgressPct float64 `json:"progress_pct" example:"16.67"`
	// Detailed list of all hafalan items with their current status
	Items []ItemDetail `json:"items"`
}

// MemberInfo represents basic information about a class member
// @Description Basic information about a student who joined the class
type MemberInfo struct {
	// UUID of the member
	UserID uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Member's email address
	Email string `json:"email" example:"student@example.com"`
	// Member's full name
	FullName string `json:"full_name" example:"Ahmad Abdullah"`
	// Date and time when the member joined the class
	JoinedAt time.Time `json:"joined_at" example:"2026-02-01T10:30:00Z"`
}

// PendingGraduation represents an item waiting for teacher approval to graduate
// @Description Item pending teacher approval for graduation
type PendingGraduation struct {
	// UUID of the item
	ItemID uuid.UUID `json:"item_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Content reference (e.g., "surah:78:1-10" or "page:582")
	ContentRef string `json:"content_ref" example:"surah:78:1-10"`
	// Student who owns this item
	StudentID uuid.UUID `json:"student_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	// Student's email
	StudentEmail string `json:"student_email" example:"student@example.com"`
	// Student's full name
	StudentName string `json:"student_name" example:"Ahmad Abdullah"`
	// When the item was created
	CreatedAt time.Time `json:"created_at"`
	// FSRS stability score
	Stability float64 `json:"stability" example:"35.5"`
	// Last interval days before pending
	LastIntervalDays int `json:"last_interval_days" example:"32"`
}

type classService struct {
	classRepo       repositories.ClassRepository
	classMemberRepo repositories.ClassMemberRepository
	classBookRepo   repositories.ClassBookRepository
	bookRepo        repositories.BookRepository
	userRepo        repositories.UserRepository
	itemRepo        *repositories.ItemRepository
}

func NewClassService(
	classRepo repositories.ClassRepository,
	classMemberRepo repositories.ClassMemberRepository,
	classBookRepo repositories.ClassBookRepository,
	bookRepo repositories.BookRepository,
	userRepo repositories.UserRepository,
	itemRepo *repositories.ItemRepository,
) ClassService {
	return &classService{
		classRepo:       classRepo,
		classMemberRepo: classMemberRepo,
		classBookRepo:   classBookRepo,
		bookRepo:        bookRepo,
		userRepo:        userRepo,
		itemRepo:        itemRepo,
	}
}

// generateClassCode generates a unique 6-character alphanumeric code
func (s *classService) generateClassCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 6

	for attempts := 0; attempts < 10; attempts++ {
		code := make([]byte, codeLength)
		for i := range code {
			num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return "", err
			}
			code[i] = charset[num.Int64()]
		}

		codeStr := string(code)
		exists, err := s.classRepo.IsCodeExists(codeStr)
		if err != nil {
			return "", err
		}
		if !exists {
			return codeStr, nil
		}
	}

	return "", errors.New("failed to generate unique class code")
}

// ==================== TEACHER METHODS ====================

func (s *classService) CreateClass(teacherID uuid.UUID, name, description, classType string) (*entities.Class, error) {
	if name == "" {
		return nil, errors.New("class name is required")
	}

	if classType != entities.ClassTypeQuran && classType != entities.ClassTypeBook {
		return nil, errors.New("invalid class type, must be 'quran' or 'book'")
	}

	classCode, err := s.generateClassCode()
	if err != nil {
		return nil, err
	}

	class := &entities.Class{
		GuruID:      teacherID,
		Name:        name,
		Description: description,
		ClassCode:   classCode,
		Type:        classType,
		IsActive:    true,
	}

	if err := s.classRepo.Create(class); err != nil {
		return nil, err
	}

	return class, nil
}

func (s *classService) GetMyClasses(teacherID uuid.UUID) ([]entities.Class, error) {
	return s.classRepo.FindByTeacher(teacherID.String())
}

func (s *classService) GetClassDetail(classID string, userID uuid.UUID) (*entities.Class, error) {
	class, err := s.classRepo.FindByIDWithRelations(classID)
	if err != nil {
		return nil, errors.New("class not found")
	}

	// Check if user is teacher or member
	if class.GuruID != userID {
		_, err := s.classMemberRepo.FindByClassAndUser(classID, userID.String())
		if err != nil {
			return nil, errors.New("you don't have access to this class")
		}
	}

	return class, nil
}

func (s *classService) UpdateClass(classID string, teacherID uuid.UUID, name, description string, isActive *bool) (*entities.Class, error) {
	class, err := s.classRepo.FindByID(classID)
	if err != nil {
		return nil, errors.New("class not found")
	}

	if class.GuruID != teacherID {
		return nil, errors.New("you don't have permission to update this class")
	}

	if name != "" {
		class.Name = name
	}
	if description != "" {
		class.Description = description
	}
	if isActive != nil {
		class.IsActive = *isActive
	}

	if err := s.classRepo.Update(class); err != nil {
		return nil, err
	}

	return class, nil
}

func (s *classService) DeleteClass(classID string, teacherID uuid.UUID) error {
	class, err := s.classRepo.FindByID(classID)
	if err != nil {
		return errors.New("class not found")
	}

	if class.GuruID != teacherID {
		return errors.New("you don't have permission to delete this class")
	}

	// Delete all members and books
	if err := s.classMemberRepo.DeleteByClassID(classID); err != nil {
		return err
	}
	if err := s.classBookRepo.DeleteByClassID(classID); err != nil {
		return err
	}

	return s.classRepo.Delete(classID)
}

func (s *classService) AddBookToClass(classID string, teacherID uuid.UUID, bookID string, order int) (*entities.ClassBook, error) {
	class, err := s.classRepo.FindByID(classID)
	if err != nil {
		return nil, errors.New("class not found")
	}

	if class.GuruID != teacherID {
		return nil, errors.New("you don't have permission to add book to this class")
	}

	if class.Type != entities.ClassTypeBook {
		return nil, errors.New("can only add books to book-type classes")
	}

	// Verify book exists and belongs to teacher
	book, err := s.bookRepo.FindByID(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	if book.OwnerID != teacherID {
		return nil, errors.New("you can only add your own books to class")
	}

	// Check if book already added
	_, err = s.classBookRepo.FindByClassAndBook(classID, bookID)
	if err == nil {
		return nil, errors.New("book already added to this class")
	}

	classBook := &entities.ClassBook{
		ClassID: uuid.MustParse(classID),
		BookID:  uuid.MustParse(bookID),
		Order:   order,
	}

	if err := s.classBookRepo.Create(classBook); err != nil {
		return nil, err
	}

	// Reload with Book relation
	return s.classBookRepo.FindByID(classBook.ID.String())
}

func (s *classService) RemoveBookFromClass(classID string, teacherID uuid.UUID, bookID string) error {
	class, err := s.classRepo.FindByID(classID)
	if err != nil {
		return errors.New("class not found")
	}

	if class.GuruID != teacherID {
		return errors.New("you don't have permission to remove book from this class")
	}

	return s.classBookRepo.DeleteByClassAndBook(classID, bookID)
}

func (s *classService) GetStudentProgress(classID string, teacherID uuid.UUID) ([]StudentProgress, error) {
	class, err := s.classRepo.FindByID(classID)
	if err != nil {
		return nil, errors.New("class not found")
	}

	if class.GuruID != teacherID {
		return nil, errors.New("you don't have permission to view this class progress")
	}

	if class.Type != entities.ClassTypeQuran {
		return nil, errors.New("progress tracking only available for quran-type classes")
	}

	members, err := s.classMemberRepo.FindByClassID(classID)
	if err != nil {
		return nil, err
	}

	var progressList []StudentProgress
	for _, member := range members {
		user, err := s.userRepo.FindByID(member.UserID.String())
		if err != nil {
			continue
		}

		// Get item stats for this user
		items, err := s.itemRepo.FindByOwner(member.UserID.String())
		if err != nil {
			continue
		}

		progress := StudentProgress{
			UserID:     member.UserID,
			Email:      user.Email,
			FullName:   user.FullName,
			TotalItems: 0,
			Items:      []ItemDetail{},
		}

		for _, item := range items {
			if item.SourceType == "quran" {
				progress.TotalItems++

				// Add item detail
				itemDetail := ItemDetail{
					ItemID:     item.ID,
					ContentRef: item.ContentRef,
					Status:     item.Status,
					CreatedAt:  item.CreatedAt,
				}

				// Add phase-specific fields
				switch item.Status {
				case entities.ItemStatusMenghafal:
					progress.Menghafal++
				case entities.ItemStatusInterval:
					progress.Interval++
					itemDetail.IntervalDays = item.IntervalDays
					itemDetail.IntervalEndAt = item.IntervalEndAt
				case entities.ItemStatusFSRSActive:
					progress.FSRSActive++
					itemDetail.NextReviewAt = item.NextReviewAt
					itemDetail.Stability = item.Stability
				case entities.ItemStatusPendingGraduate:
					progress.PendingGraduate++
					itemDetail.NextReviewAt = item.NextReviewAt
					itemDetail.Stability = item.Stability
				case entities.ItemStatusGraduate:
					progress.Graduate++
				}

				progress.Items = append(progress.Items, itemDetail)
			}
		}

		if progress.TotalItems > 0 {
			progress.ProgressPct = float64(progress.Graduate) / float64(progress.TotalItems) * 100
		}

		progressList = append(progressList, progress)
	}

	return progressList, nil
}

// ==================== STUDENT METHODS ====================

func (s *classService) JoinClass(userID uuid.UUID, classCode string) (*entities.Class, error) {
	class, err := s.classRepo.FindByCode(classCode)
	if err != nil {
		return nil, errors.New("invalid class code")
	}

	if !class.IsActive {
		return nil, errors.New("class is not active")
	}

	// Check if already a member
	_, err = s.classMemberRepo.FindByClassAndUser(class.ID.String(), userID.String())
	if err == nil {
		return nil, errors.New("you are already a member of this class")
	}

	// Can't join own class
	if class.GuruID == userID {
		return nil, errors.New("you cannot join your own class")
	}

	member := &entities.ClassMember{
		ClassID:  class.ID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}

	if err := s.classMemberRepo.Create(member); err != nil {
		return nil, err
	}

	return class, nil
}

func (s *classService) LeaveClass(userID uuid.UUID, classID string) error {
	_, err := s.classMemberRepo.FindByClassAndUser(classID, userID.String())
	if err != nil {
		return errors.New("you are not a member of this class")
	}

	return s.classMemberRepo.DeleteByClassAndUser(classID, userID.String())
}

func (s *classService) GetMyJoinedClasses(userID uuid.UUID) ([]entities.Class, error) {
	members, err := s.classMemberRepo.FindByUserID(userID.String())
	if err != nil {
		return nil, err
	}

	var classes []entities.Class
	for _, member := range members {
		class, err := s.classRepo.FindByID(member.ClassID.String())
		if err != nil {
			continue
		}
		classes = append(classes, *class)
	}

	return classes, nil
}

func (s *classService) GetClassBooks(classID string, userID uuid.UUID) ([]entities.ClassBook, error) {
	class, err := s.classRepo.FindByID(classID)
	if err != nil {
		return nil, errors.New("class not found")
	}

	// Check access
	if class.GuruID != userID {
		_, err := s.classMemberRepo.FindByClassAndUser(classID, userID.String())
		if err != nil {
			return nil, errors.New("you don't have access to this class")
		}
	}

	if class.Type != entities.ClassTypeBook {
		return nil, errors.New("this class does not contain books")
	}

	return s.classBookRepo.FindByClassID(classID)
}

func (s *classService) GetClassMembers(classID string, userID uuid.UUID) ([]MemberInfo, error) {
	class, err := s.classRepo.FindByID(classID)
	if err != nil {
		return nil, errors.New("class not found")
	}

	// Only teacher can see members
	if class.GuruID != userID {
		return nil, errors.New("only teacher can view class members")
	}

	members, err := s.classMemberRepo.FindByClassID(classID)
	if err != nil {
		return nil, err
	}

	var memberInfos []MemberInfo
	for _, member := range members {
		user, err := s.userRepo.FindByID(member.UserID.String())
		if err != nil {
			continue
		}

		memberInfos = append(memberInfos, MemberInfo{
			UserID:   member.UserID,
			Email:    user.Email,
			FullName: user.FullName,
			JoinedAt: member.JoinedAt,
		})
	}

	return memberInfos, nil
}

// ==================== GRADUATION APPROVAL METHODS ====================

func (s *classService) GetPendingGraduations(classID string, teacherID uuid.UUID) ([]PendingGraduation, error) {
	class, err := s.classRepo.FindByID(classID)
	if err != nil {
		return nil, errors.New("class not found")
	}

	if class.GuruID != teacherID {
		return nil, errors.New("you don't have permission to view this class")
	}

	if class.Type != entities.ClassTypeQuran {
		return nil, errors.New("graduation approval only available for quran-type classes")
	}

	// Get all members
	members, err := s.classMemberRepo.FindByClassID(classID)
	if err != nil {
		return nil, err
	}

	var pendingList []PendingGraduation
	for _, member := range members {
		user, err := s.userRepo.FindByID(member.UserID.String())
		if err != nil {
			continue
		}

		// Get pending graduate items for this user
		items, err := s.itemRepo.FindByOwnerAndStatus(member.UserID, entities.ItemStatusPendingGraduate)
		if err != nil {
			continue
		}

		for _, item := range items {
			if item.SourceType == "quran" {
				// Calculate last interval days
				intervalDays := 0
				if item.NextReviewAt != nil && item.LastReviewAt != nil {
					duration := item.NextReviewAt.Sub(*item.LastReviewAt)
					intervalDays = int(duration.Hours() / 24)
				}

				pendingList = append(pendingList, PendingGraduation{
					ItemID:           item.ID,
					ContentRef:       item.ContentRef,
					StudentID:        member.UserID,
					StudentEmail:     user.Email,
					StudentName:      user.FullName,
					CreatedAt:        item.CreatedAt,
					Stability:        item.Stability,
					LastIntervalDays: intervalDays,
				})
			}
		}
	}

	return pendingList, nil
}

func (s *classService) ApproveGraduation(classID string, teacherID uuid.UUID, itemID string) error {
	class, err := s.classRepo.FindByID(classID)
	if err != nil {
		return errors.New("class not found")
	}

	if class.GuruID != teacherID {
		return errors.New("you don't have permission to approve graduations in this class")
	}

	if class.Type != entities.ClassTypeQuran {
		return errors.New("graduation approval only available for quran-type classes")
	}

	// Get the item
	itemUUID, err := uuid.Parse(itemID)
	if err != nil {
		return errors.New("invalid item ID")
	}

	item, err := s.itemRepo.GetByID(itemUUID)
	if err != nil {
		return errors.New("item not found")
	}

	// Verify item is pending graduate
	if item.Status != entities.ItemStatusPendingGraduate {
		return errors.New("item is not pending graduation")
	}

	// Verify item owner is a member of this class
	isMember, err := s.classMemberRepo.IsMember(classID, item.OwnerID.String())
	if err != nil || !isMember {
		return errors.New("item owner is not a member of this class")
	}

	// Approve graduation
	now := time.Now()
	item.Status = entities.ItemStatusGraduate
	item.ApprovedBy = &teacherID
	item.ApprovedAt = &now

	return s.itemRepo.Update(item)
}

func (s *classService) RejectGraduation(classID string, teacherID uuid.UUID, itemID string) error {
	class, err := s.classRepo.FindByID(classID)
	if err != nil {
		return errors.New("class not found")
	}

	if class.GuruID != teacherID {
		return errors.New("you don't have permission to reject graduations in this class")
	}

	if class.Type != entities.ClassTypeQuran {
		return errors.New("graduation rejection only available for quran-type classes")
	}

	// Get the item
	itemUUID, err := uuid.Parse(itemID)
	if err != nil {
		return errors.New("invalid item ID")
	}

	item, err := s.itemRepo.GetByID(itemUUID)
	if err != nil {
		return errors.New("item not found")
	}

	// Verify item is pending graduate
	if item.Status != entities.ItemStatusPendingGraduate {
		return errors.New("item is not pending graduation")
	}

	// Verify item owner is a member of this class
	isMember, err := s.classMemberRepo.IsMember(classID, item.OwnerID.String())
	if err != nil || !isMember {
		return errors.New("item owner is not a member of this class")
	}

	// Reject - return to fsrs_active
	item.Status = entities.ItemStatusFSRSActive

	return s.itemRepo.Update(item)
}

