package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"

	"github.com/google/uuid"
)

type ClassService interface {
	// Teacher methods
	CreateClass(teacherID uuid.UUID, name, description, classType, coverImage string) (*entities.Class, error)
	GetMyClasses(teacherID uuid.UUID) ([]entities.Class, error)
	GetClassDetail(classID string, userID uuid.UUID) (*entities.Class, error)
	UpdateClass(classID string, teacherID uuid.UUID, name, description, coverImage string, isActive *bool) (*entities.Class, error)
	DeleteClass(classID string, teacherID uuid.UUID) error
	CreateBookInClass(classID string, teacherID uuid.UUID, title, description, coverImage string, order int) (*entities.ClassBook, error)
	AddBookToClass(classID string, teacherID uuid.UUID, bookID string, order int) (*entities.ClassBook, error)
	RemoveBookFromClass(classID string, teacherID uuid.UUID, bookID string) error
	GetStudentProgress(classID string, teacherID uuid.UUID) ([]StudentProgress, error)
	GetClassBookStudentProgress(classID, bookID string, teacherID uuid.UUID) (*ClassBookStudentProgress, error)
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

// ItemDetail represents detailed information about a single class item
// @Description Detail of a single class-scoped item including its current phase/status
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
	// Stability in days (selisih next_review_at - last_review_at). "item belum masuk ujian" if not yet reviewed.
	Stability string `json:"stability,omitempty" example:"14"`
	// When the item was created
	CreatedAt time.Time `json:"created_at"`
}

// StudentProgress represents the progress of a student in a class
// @Description Progress data for a student in a class, showing class-scoped item statistics and item details
type StudentProgress struct {
	// UUID of the student
	UserID uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// Student's email address
	Email string `json:"email" example:"student@example.com"`
	// Student's full name
	FullName string `json:"full_name" example:"Ahmad Abdullah"`
	// Total number of hafalan items the student has
	TotalItems int `json:"total_items" example:"30"`
	// Number of book items in 'start' status
	Start int `json:"start" example:"3"`
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
	// Number of book items in 'inactive' status
	Inactive int `json:"inactive" example:"1"`
	// Overall progress percentage (graduate / total_items * 100)
	ProgressPct float64 `json:"progress_pct" example:"16.67"`
	// Detailed list of all hafalan items with their current status
	Items []ItemDetail `json:"items"`
}

// ClassBookStudentProgress is a teacher-facing progress report for one book in a class.
// AverageStability is the total valid stability across a student's items divided by
// every item defined in the book. Items a student has not started contribute zero.
type ClassBookStudentProgress struct {
	ClassID        uuid.UUID             `json:"class_id"`
	BookID         uuid.UUID             `json:"book_id"`
	BookTitle      string                `json:"book_title"`
	TotalBookItems int                   `json:"total_book_items"`
	Students       []StudentBookProgress `json:"students"`
}

type StudentBookProgress struct {
	UserID          uuid.UUID `json:"user_id"`
	Email           string    `json:"email"`
	FullName        string    `json:"full_name"`
	TotalItems      int       `json:"total_items"`
	Start           int       `json:"start"`
	Menghafal       int       `json:"menghafal"`
	Interval        int       `json:"interval"`
	FSRSActive      int       `json:"fsrs_active"`
	PendingGraduate int       `json:"pending_graduate"`
	Graduate        int       `json:"graduate"`
	Inactive        int       `json:"inactive"`
	TotalUnreviewed int       `json:"total_unreviewed"`
	TotalFSRSActive int       `json:"total_fsrs_active"`
	TotalInactive   int       `json:"total_inactive"`
	// AverageStability divides total review-interval days by every item in the
	// book. Unstarted items contribute zero, making it a completion-aware value.
	AverageStability        float64                   `json:"average_stability"`
	AverageStartedStability float64                   `json:"average_started_stability"`
	Items                   []StudentBookItemProgress `json:"items"`
}

type StudentBookItemProgress struct {
	ItemID     uuid.UUID `json:"item_id"`
	BookItemID uuid.UUID `json:"book_item_id"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	// Stability remains the existing formatted review interval for compatibility.
	Stability          string     `json:"stability"`
	ReviewIntervalDays *float64   `json:"review_interval_days,omitempty"`
	FSRSStabilityDays  float64    `json:"fsrs_stability_days"`
	ReviewCount        int        `json:"review_count"`
	LastReviewAt       *time.Time `json:"last_review_at,omitempty"`
	NextReviewAt       *time.Time `json:"next_review_at,omitempty"`
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
	// Stability in days (selisih next_review_at - last_review_at). "item belum masuk ujian" if not yet reviewed.
	Stability string `json:"stability" example:"35"`
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
	juzRepo         *repositories.JuzRepository
	juzItemRepo     *repositories.JuzItemRepository
	dailyTaskRepo   repositories.DailyTaskRepository
	dailyTaskSvc    DailyTaskService
}

// calculateItemStability returns the interval in days between last_review_at and next_review_at.
// Both dates are normalized to midnight so the result is always a whole number of days.
// Returns "item belum masuk ujian" if the item hasn't been reviewed yet.
func calculateItemStability(item *entities.Item) string {
	if item == nil {
		return "item belum masuk ujian"
	}

	if item.Status == entities.ItemStatusStart || item.Status == entities.ItemStatusMenghafal {
		return "item belum masuk ujian"
	}

	if item.NextReviewAt != nil && item.LastReviewAt != nil {
		loc := item.NextReviewAt.Location()
		last := time.Date(
			item.LastReviewAt.Year(), item.LastReviewAt.Month(), item.LastReviewAt.Day(),
			0, 0, 0, 0, loc,
		)
		next := time.Date(
			item.NextReviewAt.Year(), item.NextReviewAt.Month(), item.NextReviewAt.Day(),
			0, 0, 0, 0, loc,
		)
		interval := next.Sub(last).Hours() / 24
		return fmt.Sprintf("%.0f", interval)
	}

	return "item belum masuk ujian"
}

func itemStabilityDays(item *entities.Item) (float64, bool) {
	if item == nil || item.NextReviewAt == nil || item.LastReviewAt == nil {
		return 0, false
	}

	loc := item.NextReviewAt.Location()
	last := time.Date(item.LastReviewAt.Year(), item.LastReviewAt.Month(), item.LastReviewAt.Day(), 0, 0, 0, 0, loc)
	next := time.Date(item.NextReviewAt.Year(), item.NextReviewAt.Month(), item.NextReviewAt.Day(), 0, 0, 0, 0, loc)
	days := next.Sub(last).Hours() / 24
	if days < 0 {
		return 0, false
	}
	return days, true
}

func itemBelongsToClassBooks(item entities.Item, classBooks []entities.ClassBook) bool {
	if item.SourceType != "book" {
		return false
	}

	bookID, ok := bookIDFromItemContentRef(item.ContentRef)
	if !ok {
		return false
	}

	for _, classBook := range classBooks {
		if classBook.BookID.String() == bookID {
			return true
		}
	}

	return false
}

func (s *classService) classQuranItemIDSet(userID uuid.UUID, classID string) (map[uuid.UUID]bool, error) {
	itemSet := map[uuid.UUID]bool{}

	classJuzs, err := s.juzRepo.FindByUserAndClass(userID.String(), classID)
	if err != nil {
		return itemSet, err
	}
	if len(classJuzs) == 0 {
		return itemSet, nil
	}

	juzIDs := make([]string, 0, len(classJuzs))
	for _, juz := range classJuzs {
		juzIDs = append(juzIDs, juz.ID.String())
	}

	itemIDs, err := s.juzItemRepo.FindItemIDsByJuzIDs(juzIDs)
	if err != nil {
		return itemSet, err
	}
	for _, itemID := range itemIDs {
		parsedItemID, err := uuid.Parse(itemID)
		if err == nil {
			itemSet[parsedItemID] = true
		}
	}

	return itemSet, nil
}

func (s *classService) enrichClassSummary(class *entities.Class) error {
	teacher, err := s.userRepo.FindByID(class.GuruID.String())
	if err == nil {
		class.OwnerName = teacher.FullName
		if class.OwnerName == "" {
			class.OwnerName = teacher.Email
		}
	}

	studentCount, err := s.classMemberRepo.CountByClassID(class.ID.String())
	if err != nil {
		return err
	}
	class.StudentCount = studentCount

	bookCount, err := s.classBookRepo.CountByClassID(class.ID.String())
	if err != nil {
		return err
	}
	class.BookCount = bookCount

	return nil
}

func NewClassService(
	classRepo repositories.ClassRepository,
	classMemberRepo repositories.ClassMemberRepository,
	classBookRepo repositories.ClassBookRepository,
	bookRepo repositories.BookRepository,
	userRepo repositories.UserRepository,
	itemRepo *repositories.ItemRepository,
	juzRepo *repositories.JuzRepository,
	juzItemRepo *repositories.JuzItemRepository,
	dailyTaskRepo repositories.DailyTaskRepository,
	dailyTaskSvc DailyTaskService,
) ClassService {
	return &classService{
		classRepo:       classRepo,
		classMemberRepo: classMemberRepo,
		classBookRepo:   classBookRepo,
		bookRepo:        bookRepo,
		userRepo:        userRepo,
		itemRepo:        itemRepo,
		juzRepo:         juzRepo,
		juzItemRepo:     juzItemRepo,
		dailyTaskRepo:   dailyTaskRepo,
		dailyTaskSvc:    dailyTaskSvc,
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

func (s *classService) CreateClass(teacherID uuid.UUID, name, description, classType, coverImage string) (*entities.Class, error) {
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
		CoverImage:  coverImage,
		ClassCode:   classCode,
		Type:        classType,
		IsActive:    true,
	}

	if err := s.classRepo.Create(class); err != nil {
		return nil, err
	}

	if err := s.enrichClassSummary(class); err != nil {
		return nil, err
	}

	return class, nil
}

func (s *classService) GetMyClasses(teacherID uuid.UUID) ([]entities.Class, error) {
	classes, err := s.classRepo.FindByTeacher(teacherID.String())
	if err != nil {
		return nil, err
	}

	for i := range classes {
		if err := s.enrichClassSummary(&classes[i]); err != nil {
			return nil, err
		}
	}

	return classes, nil
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

	if err := s.enrichClassSummary(class); err != nil {
		return nil, err
	}

	return class, nil
}

func (s *classService) UpdateClass(classID string, teacherID uuid.UUID, name, description, coverImage string, isActive *bool) (*entities.Class, error) {
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
	if coverImage != "" {
		class.CoverImage = coverImage
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

// CreateBookInClass creates a draft book owned by the class teacher and assigns it
// to the class immediately. Modules and items are managed through the regular book
// endpoints, so their validation and structure stay identical to personal books.
func (s *classService) CreateBookInClass(classID string, teacherID uuid.UUID, title, description, coverImage string, order int) (*entities.ClassBook, error) {
	class, err := s.classRepo.FindByID(classID)
	if err != nil {
		return nil, errors.New("class not found")
	}

	if class.GuruID != teacherID {
		return nil, errors.New("you don't have permission to create a book in this class")
	}

	if class.Type != entities.ClassTypeBook {
		return nil, errors.New("can only create books in book-type classes")
	}

	if title == "" {
		return nil, errors.New("book title is required")
	}

	book := &entities.Book{
		OwnerID:     teacherID,
		Title:       title,
		Description: description,
		CoverImage:  coverImage,
		Status:      entities.BookStatusDraft,
	}
	if err := s.bookRepo.Create(book); err != nil {
		return nil, err
	}

	classBook := &entities.ClassBook{
		ClassID: class.ID,
		BookID:  book.ID,
		Order:   order,
	}
	if err := s.classBookRepo.Create(classBook); err != nil {
		// The book has no modules or items yet. Remove it so a failed request does
		// not leave an inaccessible personal draft behind.
		_ = s.bookRepo.Delete(book.ID.String())
		return nil, err
	}

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

	var classBooks []entities.ClassBook
	if class.Type == entities.ClassTypeBook {
		classBooks, err = s.classBookRepo.FindByClassID(classID)
		if err != nil {
			return nil, err
		}
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

		classQuranItemIDs := map[uuid.UUID]bool{}
		if class.Type == entities.ClassTypeQuran {
			classQuranItemIDs, err = s.classQuranItemIDSet(member.UserID, classID)
			if err != nil {
				continue
			}
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
			if class.Type == entities.ClassTypeQuran {
				if !classQuranItemIDs[item.ID] {
					continue
				}
			} else if !itemBelongsToClassBooks(item, classBooks) {
				continue
			}

			progress.TotalItems++

			itemDetail := ItemDetail{
				ItemID:     item.ID,
				ContentRef: item.ContentRef,
				Status:     item.Status,
				CreatedAt:  item.CreatedAt,
			}

			switch item.Status {
			case entities.ItemStatusStart:
				progress.Start++
			case entities.ItemStatusMenghafal:
				progress.Menghafal++
			case entities.ItemStatusInterval:
				progress.Interval++
				itemDetail.IntervalDays = item.IntervalDays
				itemDetail.IntervalEndAt = item.IntervalEndAt
			case entities.ItemStatusFSRSActive:
				progress.FSRSActive++
				itemDetail.NextReviewAt = item.NextReviewAt
				itemDetail.Stability = calculateItemStability(&item)
			case entities.ItemStatusPendingGraduate:
				progress.PendingGraduate++
				itemDetail.NextReviewAt = item.NextReviewAt
				itemDetail.Stability = calculateItemStability(&item)
			case entities.ItemStatusGraduate:
				progress.Graduate++
			case entities.ItemStatusInactive:
				progress.Inactive++
			}

			progress.Items = append(progress.Items, itemDetail)
		}

		if progress.TotalItems > 0 {
			progress.ProgressPct = float64(progress.Graduate) / float64(progress.TotalItems) * 100
		}

		progressList = append(progressList, progress)
	}

	return progressList, nil
}

// GetClassBookStudentProgress returns progress for every student in one book.
// It deliberately scopes the report to a single class-book relation so a book
// assigned to another class cannot be queried through this class.
func (s *classService) GetClassBookStudentProgress(classID, bookID string, teacherID uuid.UUID) (*ClassBookStudentProgress, error) {
	class, err := s.classRepo.FindByID(classID)
	if err != nil {
		return nil, errors.New("class not found")
	}
	if class.GuruID != teacherID {
		return nil, errors.New("you don't have permission to view this class progress")
	}
	if class.Type != entities.ClassTypeBook {
		return nil, errors.New("this progress endpoint is only available for book-type classes")
	}
	if _, err := s.classBookRepo.FindByClassAndBook(classID, bookID); err != nil {
		return nil, errors.New("book is not assigned to this class")
	}

	book, err := s.bookRepo.FindByIDWithRelations(bookID)
	if err != nil {
		return nil, errors.New("book not found")
	}

	bookItems := make(map[uuid.UUID]entities.BookItem)
	for _, item := range book.Items {
		bookItems[item.ID] = item
	}
	for _, module := range book.Modules {
		for _, item := range module.Items {
			bookItems[item.ID] = item
		}
	}

	members, err := s.classMemberRepo.FindByClassID(classID)
	if err != nil {
		return nil, err
	}

	result := &ClassBookStudentProgress{
		ClassID:        class.ID,
		BookID:         book.ID,
		BookTitle:      book.Title,
		TotalBookItems: len(bookItems),
		Students:       make([]StudentBookProgress, 0, len(members)),
	}

	now := time.Now().In(config.AppLocation)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	for _, member := range members {
		user, err := s.userRepo.FindByID(member.UserID.String())
		if err != nil {
			continue
		}

		items, err := s.itemRepo.FindByOwnerAndBookIDs(member.UserID, []string{bookID})
		if err != nil {
			return nil, err
		}

		// Fetch or generate student's daily tasks snapshot for today
		dailyTaskState := make(map[uuid.UUID]string)
		if s.dailyTaskRepo != nil {
			tasks, err := s.dailyTaskRepo.ListByUserAndDate(context.Background(), member.UserID, now)
			if err == nil && len(tasks) == 0 && s.dailyTaskSvc != nil {
				tasks, _ = s.dailyTaskSvc.GenerateToday(context.Background(), member.UserID, now, 0)
			}
			for _, t := range tasks {
				dailyTaskState[t.ItemID] = t.State
			}
		}

		student := StudentBookProgress{
			UserID:   member.UserID,
			Email:    user.Email,
			FullName: user.FullName,
			Items:    make([]StudentBookItemProgress, 0, len(items)),
		}
		stabilityTotal := 0.0
		startedStabilityCount := 0
		totalUnreviewed := 0
		totalInactive := 0

		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		for _, item := range items {
			_, bookItemID, ok := strings.Cut(item.ContentRef, "book:"+bookID+":item:")
			if !ok {
				continue
			}
			parsedBookItemID, err := uuid.Parse(bookItemID)
			if err != nil {
				continue
			}
			bookItem, exists := bookItems[parsedBookItemID]
			if !exists {
				continue
			}

			student.TotalItems++
			switch item.Status {
			case entities.ItemStatusStart:
				student.Start++
			case entities.ItemStatusMenghafal:
				student.Menghafal++
			case entities.ItemStatusInterval:
				student.Interval++
			case entities.ItemStatusFSRSActive:
				student.FSRSActive++
			case entities.ItemStatusPendingGraduate:
				student.PendingGraduate++
			case entities.ItemStatusGraduate:
				student.Graduate++
			case entities.ItemStatusInactive:
				student.Inactive++
			}

			if item.Status == entities.ItemStatusGraduate || item.Status == entities.ItemStatusInactive {
				totalInactive++
			} else {
				taskState, hasTask := dailyTaskState[item.ID]
				isTaskCompleted := hasTask && (taskState == "done" || taskState == "completed")
				isReviewedToday := isTaskCompleted || (item.LastReviewAt != nil && !item.LastReviewAt.Before(startOfDay) && !item.LastReviewAt.After(endOfDay))

				// Item is strictly due if it has an uncompleted daily task or scheduled deadline <= endOfDay
				isDueBySchedule := (item.NextReviewAt != nil && !item.NextReviewAt.After(endOfDay)) || (item.IntervalNextReviewAt != nil && !item.IntervalNextReviewAt.After(endOfDay))
				isDue := (hasTask && !isTaskCompleted) || isDueBySchedule

				// Outstanding review obligation for today: must be due and not reviewed today
				if isDue && !isReviewedToday {
					totalUnreviewed++
				}
			}

			stability := calculateItemStability(&item)
			var reviewIntervalDays *float64
			if days, valid := itemStabilityDays(&item); valid {
				stabilityTotal += days
				startedStabilityCount++
				reviewIntervalDays = &days
			}
			student.Items = append(student.Items, StudentBookItemProgress{
				ItemID:             item.ID,
				BookItemID:         bookItem.ID,
				Title:              bookItem.Title,
				Status:             item.Status,
				Stability:          stability,
				ReviewIntervalDays: reviewIntervalDays,
				FSRSStabilityDays:  item.Stability,
				ReviewCount:        item.ReviewCount,
				LastReviewAt:       item.LastReviewAt,
				NextReviewAt:       item.NextReviewAt,
			})
		}

		student.TotalUnreviewed = totalUnreviewed
		student.TotalInactive = totalInactive

		if result.TotalBookItems > 0 {
			student.AverageStability = math.Round((stabilityTotal/float64(result.TotalBookItems))*100) / 100
		}
		if startedStabilityCount > 0 {
			student.AverageStartedStability = math.Round((stabilityTotal/float64(startedStabilityCount))*100) / 100
		}
		result.Students = append(result.Students, student)
	}

	return result, nil
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
		JoinedAt: time.Now().In(config.AppLocation),
	}

	if err := s.classMemberRepo.Create(member); err != nil {
		return nil, err
	}

	if err := s.enrichClassSummary(class); err != nil {
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
		if err := s.enrichClassSummary(class); err == nil {
			classes = append(classes, *class)
		}
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

	classBooks, err := s.classBookRepo.FindByClassID(classID)
	if err != nil {
		return nil, err
	}

	for i := range classBooks {
		owner, err := s.userRepo.FindByID(classBooks[i].Book.OwnerID.String())
		if err == nil {
			classBooks[i].OwnerName = owner.FullName
			if classBooks[i].OwnerName == "" {
				classBooks[i].OwnerName = owner.Email
			}
		}
	}

	return classBooks, nil
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

		classQuranItemIDs, err := s.classQuranItemIDSet(member.UserID, classID)
		if err != nil {
			continue
		}

		// Get pending graduate items for this user
		items, err := s.itemRepo.FindByOwnerAndStatus(member.UserID, entities.ItemStatusPendingGraduate)
		if err != nil {
			continue
		}

		for _, item := range items {
			if item.SourceType == "quran" && classQuranItemIDs[item.ID] {
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
					Stability:        calculateItemStability(&item),
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

	classQuranItemIDs, err := s.classQuranItemIDSet(item.OwnerID, classID)
	if err != nil || !classQuranItemIDs[item.ID] {
		return errors.New("item is not part of this class")
	}

	// Approve graduation
	now := time.Now().In(config.AppLocation)
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

	classQuranItemIDs, err := s.classQuranItemIDSet(item.OwnerID, classID)
	if err != nil || !classQuranItemIDs[item.ID] {
		return errors.New("item is not part of this class")
	}

	// Reject - return to fsrs_active
	item.Status = entities.ItemStatusFSRSActive

	return s.itemRepo.Update(item)
}
