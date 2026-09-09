package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/services"
)

func TestGraduateDailyTaskScheduling(t *testing.T) {
	db := setupTestPostgresDB(t)

	itemRepo := repositories.NewItemRepository(db)
	reviewStateRepo := repositories.NewReviewStateRepository(db)
	dailyTaskRepo := repositories.NewDailyTaskRepository(db)
	juzRepo := repositories.NewJuzRepository(db)
	juzItemRepo := repositories.NewJuzItemRepository(db)

	dailyTaskSvc := services.NewDailyTaskService(
		reviewStateRepo,
		dailyTaskRepo,
		itemRepo,
		nil,
		nil,
		juzRepo,
		juzItemRepo,
	)

	userID := uuid.New()
	now := time.Now().In(config.AppLocation)
	targetJuzIndex := 30

	// Cleanup test artifacts for this user
	defer func() {
		_ = db.Where("user_id = ?", userID).Delete(&entities.DailyTask{}).Error
		_ = db.Where("owner_id = ?", userID).Delete(&entities.Item{}).Error
		_ = db.Where("user_id = ?", userID).Delete(&entities.Juz{}).Error
	}()

	// Create user's active Juz 30
	juz := entities.Juz{
		ID:       uuid.New(),
		UserID:   userID,
		Index:    targetJuzIndex,
		IsActive: true,
	}
	if err := db.Create(&juz).Error; err != nil {
		t.Fatalf("Failed to create test juz: %v", err)
	}

	// -------------------------------------------------------------------------
	// Case 1: Graduate item with next_review_at in the future -> NO Daily Task
	// -------------------------------------------------------------------------
	futureReview := now.AddDate(0, 0, 16)
	futureItem := entities.Item{
		ID:           uuid.New(),
		OwnerID:      userID,
		SourceType:   "quran",
		ContentRef:   "surah:98:1-8",
		Status:       entities.ItemStatusGraduate,
		NextReviewAt: &futureReview,
	}
	if err := db.Create(&futureItem).Error; err != nil {
		t.Fatalf("Failed to create futureItem: %v", err)
	}
	if err := db.Create(&entities.JuzItem{ID: uuid.New(), JuzID: juz.ID, ItemID: futureItem.ID}).Error; err != nil {
		t.Fatalf("Failed to link futureItem to juz: %v", err)
	}

	// -------------------------------------------------------------------------
	// Case 2: Graduate item with next_review_at today -> 1 Daily Task
	// -------------------------------------------------------------------------
	todayReview := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, config.AppLocation)
	dueTodayItem := entities.Item{
		ID:           uuid.New(),
		OwnerID:      userID,
		SourceType:   "quran",
		ContentRef:   "surah:99:1-8",
		Status:       entities.ItemStatusGraduate,
		NextReviewAt: &todayReview,
	}
	if err := db.Create(&dueTodayItem).Error; err != nil {
		t.Fatalf("Failed to create dueTodayItem: %v", err)
	}
	if err := db.Create(&entities.JuzItem{ID: uuid.New(), JuzID: juz.ID, ItemID: dueTodayItem.ID}).Error; err != nil {
		t.Fatalf("Failed to link dueTodayItem to juz: %v", err)
	}

	// -------------------------------------------------------------------------
	// Case 3: Graduate item with next_review_at in the past -> 1 Daily Task
	// -------------------------------------------------------------------------
	pastReview := now.AddDate(0, 0, -3)
	overdueItem := entities.Item{
		ID:           uuid.New(),
		OwnerID:      userID,
		SourceType:   "quran",
		ContentRef:   "surah:100:1-11",
		Status:       entities.ItemStatusGraduate,
		NextReviewAt: &pastReview,
	}
	if err := db.Create(&overdueItem).Error; err != nil {
		t.Fatalf("Failed to create overdueItem: %v", err)
	}
	if err := db.Create(&entities.JuzItem{ID: uuid.New(), JuzID: juz.ID, ItemID: overdueItem.ID}).Error; err != nil {
		t.Fatalf("Failed to link overdueItem to juz: %v", err)
	}

	// -------------------------------------------------------------------------
	// Case 4: Graduate item with next_review_at = NULL -> NO Daily Task
	// -------------------------------------------------------------------------
	nullReviewItem := entities.Item{
		ID:           uuid.New(),
		OwnerID:      userID,
		SourceType:   "quran",
		ContentRef:   "surah:101:1-11",
		Status:       entities.ItemStatusGraduate,
		NextReviewAt: nil,
	}
	if err := db.Create(&nullReviewItem).Error; err != nil {
		t.Fatalf("Failed to create nullReviewItem: %v", err)
	}
	if err := db.Create(&entities.JuzItem{ID: uuid.New(), JuzID: juz.ID, ItemID: nullReviewItem.ID}).Error; err != nil {
		t.Fatalf("Failed to link nullReviewItem to juz: %v", err)
	}

	// Run GenerateToday
	tasks, err := dailyTaskSvc.GenerateToday(context.Background(), userID, now, 0)
	if err != nil {
		t.Fatalf("GenerateToday failed: %v", err)
	}

	taskMap := make(map[uuid.UUID]entities.DailyTask)
	for _, tsk := range tasks {
		taskMap[tsk.ItemID] = tsk
	}

	// Assertions for Cases 1-4
	if _, exists := taskMap[futureItem.ID]; exists {
		t.Fatalf("Case 1 FAILED: Future graduate item %s generated a daily task prematurely", futureItem.ID)
	}

	if tsk, exists := taskMap[dueTodayItem.ID]; !exists {
		t.Fatalf("Case 2 FAILED: Due today graduate item %s did not generate a daily task", dueTodayItem.ID)
	} else if tsk.Source != "graduate" || tsk.State != "pending" {
		t.Fatalf("Case 2 FAILED: Due today task has invalid source=%s state=%s", tsk.Source, tsk.State)
	}

	if tsk, exists := taskMap[overdueItem.ID]; !exists {
		t.Fatalf("Case 3 FAILED: Overdue graduate item %s did not generate a daily task", overdueItem.ID)
	} else if tsk.Source != "graduate" || tsk.State != "pending" {
		t.Fatalf("Case 3 FAILED: Overdue task has invalid source=%s state=%s", tsk.Source, tsk.State)
	}

	if _, exists := taskMap[nullReviewItem.ID]; exists {
		t.Fatalf("Case 4 FAILED: NULL next_review_at item %s generated a daily task", nullReviewItem.ID)
	}

	// -------------------------------------------------------------------------
	// Case 5: Duplicate item in single batch to UpsertDailyTasks -> Exactly 1 inserted
	// -------------------------------------------------------------------------
	taskDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, config.AppLocation)
	duplicateBatchItemID := uuid.New()
	batchTasks := []entities.DailyTask{
		{ID: uuid.New(), UserID: userID, ItemID: duplicateBatchItemID, TaskDate: taskDate, Source: "graduate", State: "pending"},
		{ID: uuid.New(), UserID: userID, ItemID: duplicateBatchItemID, TaskDate: taskDate, Source: "graduate", State: "pending"},
	}

	if err := dailyTaskRepo.UpsertDailyTasks(context.Background(), userID, taskDate, batchTasks); err != nil {
		t.Fatalf("Case 5 FAILED: UpsertDailyTasks error: %v", err)
	}

	var insertedCount int64
	db.Model(&entities.DailyTask{}).Where("user_id = ? AND item_id = ? AND task_date = ?", userID, duplicateBatchItemID, taskDate).Count(&insertedCount)
	if insertedCount != 1 {
		t.Fatalf("Case 5 FAILED: Expected exactly 1 row for duplicate batch input, got %d", insertedCount)
	}

	// -------------------------------------------------------------------------
	// Case 6: Repeated GenerateToday calls -> Idempotent, no duplicates
	// -------------------------------------------------------------------------
	secondRunTasks, err := dailyTaskSvc.GenerateToday(context.Background(), userID, now, 0)
	if err != nil {
		t.Fatalf("Case 6 FAILED: Second GenerateToday failed: %v", err)
	}

	if len(secondRunTasks) != len(tasks) {
		t.Fatalf("Case 6 FAILED: Expected %d tasks in second run, got %d", len(tasks), len(secondRunTasks))
	}

	// Verify no duplicate row for dueTodayItem in DB
	var dueTodayCount int64
	db.Model(&entities.DailyTask{}).Where("user_id = ? AND item_id = ? AND task_date = ?", userID, dueTodayItem.ID, taskDate).Count(&dueTodayCount)
	if dueTodayCount != 1 {
		t.Fatalf("Case 6 FAILED: Expected exactly 1 DB row for dueTodayItem after repeated GenerateToday, got %d", dueTodayCount)
	}
}
