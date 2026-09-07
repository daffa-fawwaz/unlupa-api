package services_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/fsrs"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
)

func TestClassroomStudentBookReviewFlow(t *testing.T) {
	db := setupTestPostgresDB(t)
	now := time.Now().In(config.AppLocation)

	// 0. Setup Users
	teacher := &entities.User{
		Email:    fmt.Sprintf("teacher_%s@example.com", uuid.New().String()[:8]),
		FullName: "Teacher",
		Role:     "teacher",
	}
	student1 := &entities.User{
		Email:    fmt.Sprintf("student1_%s@example.com", uuid.New().String()[:8]),
		FullName: "Student 1",
		Role:     "student",
	}
	student2 := &entities.User{
		Email:    fmt.Sprintf("student2_%s@example.com", uuid.New().String()[:8]),
		FullName: "Student 2",
		Role:     "student",
	}
	if err := db.Create(teacher).Error; err != nil {
		t.Fatalf("failed to create teacher: %v", err)
	}
	if err := db.Create(student1).Error; err != nil {
		t.Fatalf("failed to create student1: %v", err)
	}
	if err := db.Create(student2).Error; err != nil {
		t.Fatalf("failed to create student2: %v", err)
	}
	defer db.Delete(teacher)
	defer db.Delete(student1)
	defer db.Delete(student2)

	teacherID := teacher.ID
	student1ID := student1.ID
	student2ID := student2.ID

	// 1. Setup Class and Members
	class := &entities.Class{
		GuruID:    teacherID,
		Name:      "Test Book Class",
		ClassCode: fmt.Sprintf("CODE_%s", uuid.New().String()[:8]),
		Type:      entities.ClassTypeBook,
		IsActive:  true,
	}
	if err := db.Create(class).Error; err != nil {
		t.Fatalf("failed to create class: %v", err)
	}
	defer db.Delete(class)
	classID := class.ID

	member1 := &entities.ClassMember{
		ClassID: classID,
		UserID:  student1ID,
	}
	member2 := &entities.ClassMember{
		ClassID: classID,
		UserID:  student2ID,
	}
	_ = db.Create(member1)
	_ = db.Create(member2)
	defer db.Delete(member1)
	defer db.Delete(member2)

	// 2. Setup Book, BookItem, and ClassBook
	book := &entities.Book{
		OwnerID: teacherID,
		Title:   "Classroom Test Book",
	}
	if err := db.Create(book).Error; err != nil {
		t.Fatalf("failed to create book: %v", err)
	}
	defer db.Delete(book)
	bookID := book.ID

	bookItem := &entities.BookItem{
		BookID:  bookID,
		Title:   "Item 1",
		Content: "Question 1",
		Answer:  "Answer 1",
	}
	if err := db.Create(bookItem).Error; err != nil {
		t.Fatalf("failed to create book item: %v", err)
	}
	defer db.Delete(bookItem)
	bookItemID := bookItem.ID

	classBook := &entities.ClassBook{
		ClassID: classID,
		BookID:  bookID,
	}
	if err := db.Create(classBook).Error; err != nil {
		t.Fatalf("failed to create class book: %v", err)
	}
	defer db.Delete(classBook)

	// 3. Setup student items with status = "start"
	contentRef := fmt.Sprintf("book:%s:item:%s", bookID.String(), bookItemID.String())
	student1Item := &entities.Item{
		ID:         uuid.New(),
		OwnerID:    student1ID,
		SourceType: "book",
		ContentRef: contentRef,
		Status:     entities.ItemStatusStart,
		Stability:  0,
		Difficulty: 5.0,
		CreatedAt:  now,
	}
	if err := db.Create(student1Item).Error; err != nil {
		t.Fatalf("failed to create student1 item: %v", err)
	}
	defer db.Delete(student1Item)

	student2Item := &entities.Item{
		ID:         uuid.New(),
		OwnerID:    student2ID,
		SourceType: "book",
		ContentRef: contentRef,
		Status:     entities.ItemStatusStart,
		Stability:  0,
		Difficulty: 5.0,
		CreatedAt:  now,
	}
	if err := db.Create(student2Item).Error; err != nil {
		t.Fatalf("failed to create student2 item: %v", err)
	}
	defer db.Delete(student2Item)

	itemRepo := repositories.NewItemRepository(db)
	classMemberRepo := repositories.NewClassMemberRepository(db)
	classRepo := repositories.NewClassRepository(db)
	classBookRepo := repositories.NewClassBookRepository(db)

	reviewService := services.NewItemReviewService(
		itemRepo,
		nil,
		nil,
		classMemberRepo,
		classRepo,
		classBookRepo,
		nil,
	)

	// 4. Test unauthorized access: student1 cannot review student2's item
	_, err := reviewService.ReviewItem(student1ID, student2Item.ID, fsrs.Good, now)
	if err == nil {
		t.Errorf("Expected error when student1 reviews student2's item, got nil")
	}

	// 5. Test successful review: student1 reviews their own classroom item (status = "start")
	result, err := reviewService.ReviewItem(student1ID, student1Item.ID, fsrs.Good, now)
	if err != nil {
		t.Fatalf("Expected student1 review to succeed, got error: %v", err)
	}

	// 6. Assertions on reviewed item:
	// - Status transitioned from "start" -> "fsrs_active"
	// - FSRS state is initialized (Stability > 0, NextReviewAt != nil)
	if result.Item.Status != entities.ItemStatusFSRSActive {
		t.Errorf("Expected status '%s', got '%s'", entities.ItemStatusFSRSActive, result.Item.Status)
	}
	if result.Item.NextReviewAt == nil {
		t.Errorf("Expected NextReviewAt to be non-nil after FSRS review")
	}
	if result.Item.ReviewCount != 1 {
		t.Errorf("Expected ReviewCount = 1, got %d", result.Item.ReviewCount)
	}

	// 7. Verify student2's item state remains untouched (isolated state)
	var student2ItemCheck entities.Item
	if err := db.Where("id = ?", student2Item.ID).First(&student2ItemCheck).Error; err != nil {
		t.Fatalf("failed to reload student2 item: %v", err)
	}
	if student2ItemCheck.Status != entities.ItemStatusStart {
		t.Errorf("Expected student2 item status to remain 'start', got '%s'", student2ItemCheck.Status)
	}
	if student2ItemCheck.ReviewCount != 0 {
		t.Errorf("Expected student2 item review count to be 0, got %d", student2ItemCheck.ReviewCount)
	}
}

func TestPersonalBookReviewFlowAndIntervalRejection(t *testing.T) {
	db := setupTestPostgresDB(t)
	now := time.Now().In(config.AppLocation)

	user := &entities.User{
		Email:    fmt.Sprintf("user_%s@example.com", uuid.New().String()[:8]),
		FullName: "Personal User",
		Role:     "student",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	defer db.Delete(user)
	userID := user.ID

	book := &entities.Book{
		OwnerID: userID,
		Title:   "Personal Book",
	}
	if err := db.Create(book).Error; err != nil {
		t.Fatalf("failed to create book: %v", err)
	}
	defer db.Delete(book)

	bookItem := &entities.BookItem{
		BookID:  book.ID,
		Title:   "Item P1",
		Content: "Q",
		Answer:  "A",
	}
	if err := db.Create(bookItem).Error; err != nil {
		t.Fatalf("failed to create book item: %v", err)
	}
	defer db.Delete(bookItem)

	contentRef := fmt.Sprintf("book:%s:item:%s", book.ID.String(), bookItem.ID.String())
	item := &entities.Item{
		OwnerID:    userID,
		SourceType: "book",
		ContentRef: contentRef,
		Status:     entities.ItemStatusStart,
		Stability:  0,
		Difficulty: 5.0,
		CreatedAt:  now,
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("failed to create item: %v", err)
	}
	defer db.Delete(item)

	itemRepo := repositories.NewItemRepository(db)
	intervalRepo := repositories.NewIntervalReviewLogRepository(db)
	classBookRepo := repositories.NewClassBookRepository(db)
	taskActionRepo := repositories.NewDailyTaskActionRepository(db)

	statusService := services.NewItemStatusService(
		itemRepo,
		intervalRepo,
		classBookRepo,
		taskActionRepo,
	)
	reviewService := services.NewItemReviewService(
		itemRepo,
		nil,
		nil,
		nil,
		nil,
		classBookRepo,
		nil,
	)

	// 1. ReviewInterval on status 'start' MUST fail with error
	_, err := statusService.ReviewInterval(item.ID, userID, 2)
	if err == nil {
		t.Errorf("Expected ReviewInterval on item with status 'start' to fail, got nil")
	} else if err.Error() != "item must be in 'interval' status to review" {
		t.Errorf("Expected error 'item must be in 'interval' status to review', got '%s'", err.Error())
	}

	// 2. ReviewItem (FSRS) on status 'start' MUST succeed and transition to 'fsrs_active'
	res, err := reviewService.ReviewItem(userID, item.ID, fsrs.Good, now)
	if err != nil {
		t.Fatalf("Expected ReviewItem to succeed on personal book item: %v", err)
	}
	if res.Item.Status != entities.ItemStatusFSRSActive {
		t.Errorf("Expected status to be 'fsrs_active', got '%s'", res.Item.Status)
	}
}

