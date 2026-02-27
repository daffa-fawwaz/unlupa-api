package services

import (
	"context"
	"time"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
	"hifzhun-api/pkg/utils"

	"github.com/google/uuid"
)

type DailyTaskService interface {
	GenerateToday(
		ctx context.Context,
		userID uuid.UUID,
		now time.Time,
		limit int,
	) ([]entities.DailyTask, error)

	ListToday(
		ctx context.Context,
		userID uuid.UUID,
		now time.Time,
	) ([]entities.DailyTask, error)
}

type dailyTaskService struct {
	reviewStateRepo repositories.ReviewStateRepository
	dailyTaskRepo   repositories.DailyTaskRepository
	itemRepo        *repositories.ItemRepository
	classMemberRepo repositories.ClassMemberRepository
	classRepo       repositories.ClassRepository
	juzRepo         *repositories.JuzRepository
}

func NewDailyTaskService(
	reviewStateRepo repositories.ReviewStateRepository,
	dailyTaskRepo repositories.DailyTaskRepository,
	itemRepo *repositories.ItemRepository,
	classMemberRepo repositories.ClassMemberRepository,
	classRepo repositories.ClassRepository,
	juzRepo *repositories.JuzRepository,
) DailyTaskService {
	return &dailyTaskService{
		reviewStateRepo: reviewStateRepo,
		dailyTaskRepo:   dailyTaskRepo,
		itemRepo:        itemRepo,
		classMemberRepo: classMemberRepo,
		classRepo:       classRepo,
		juzRepo:         juzRepo,
	}
}

func (s *dailyTaskService) isUserInQuranClass(userID uuid.UUID) bool {
	classes, err := s.classMemberRepo.FindByUserID(userID.String())
	if err != nil || len(classes) == 0 {
		return false
	}

	for _, membership := range classes {
		class, err := s.classRepo.FindByID(membership.ClassID.String())
		if err != nil {
			continue
		}
		if class.Type == entities.ClassTypeQuran && class.IsActive {
			return true
		}
	}
	return false
}

func (s *dailyTaskService) GenerateToday(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
	limit int,
) ([]entities.DailyTask, error) {

	// 📌 SNAPSHOT DATE (WAJIB)
	taskDate := now.Truncate(24 * time.Hour)

	// ========== 0️⃣ Auto-graduation for FSRS items ==========
	eligibleItemsByDays, err := s.itemRepo.FindEligibleForGraduation(userID, entities.GraduationIntervalDays, now)
	if err == nil {
		for _, item := range eligibleItemsByDays {
			if s.isUserInQuranClass(userID) {
				item.Status = entities.ItemStatusPendingGraduate
			} else {
				item.Status = entities.ItemStatusGraduate
				nextRev := now.AddDate(0, 0, entities.GraduateReviewDays)
				nextRev = time.Date(nextRev.Year(), nextRev.Month(), nextRev.Day(), 0, 0, 0, 0, nextRev.Location())
				item.NextReviewAt = &nextRev
			}
			s.itemRepo.Update(&item)
		}
	}
	eligibleItemsByStab, err := s.itemRepo.FindEligibleForGraduationByStability(userID, entities.GraduateStabilityThreshold)
	if err == nil {
		for _, item := range eligibleItemsByStab {
			if s.isUserInQuranClass(userID) {
				item.Status = entities.ItemStatusPendingGraduate
			} else {
				item.Status = entities.ItemStatusGraduate
				nextRev := now.AddDate(0, 0, entities.GraduateReviewDays)
				nextRev = time.Date(nextRev.Year(), nextRev.Month(), nextRev.Day(), 0, 0, 0, 0, nextRev.Location())
				item.NextReviewAt = &nextRev
			}
			s.itemRepo.Update(&item)
		}
	}

	tasks := make([]entities.DailyTask, 0)

	// ========== 1️⃣ Items dari Interval yang sudah deadline ==========
	intervalItems, err := s.itemRepo.FindIntervalDeadlineReached(now)
	if err != nil {
		return nil, err
	}

	for _, item := range intervalItems {
		// Filter by owner
		if item.OwnerID != userID {
			continue
		}

		tasks = append(tasks, entities.DailyTask{
			ID:        uuid.New(),
			UserID:    userID,
			ItemID:    item.ID,
			CardID:    uuid.Nil, // No card yet for interval items
			TaskDate:  taskDate,
			Source:    "interval", // Mark as interval source
			State:     "pending",
			CreatedAt: now,
		})

		// Update item status to fsrs_active
		item.Status = entities.ItemStatusFSRSActive
		// Set the time when item entered fsrs_active for graduation tracking
		if item.FSRSStartAt == nil {
			t := now
			item.FSRSStartAt = &t
		}
		// Ensure FSRS params are initialized when promoted automatically.
		if item.Stability <= 0 {
			item.Stability = 0.4
		}
		if item.Difficulty <= 0 {
			item.Difficulty = 5.0
		}
		s.itemRepo.Update(&item)
	}

	// ========== 1.5️⃣ Items Interval yang due untuk recurring review ==========
	intervalReviewDueItems, err := s.itemRepo.FindIntervalReviewDue(userID, now)
	if err != nil {
		return nil, err
	}

	for _, item := range intervalReviewDueItems {
		tasks = append(tasks, entities.DailyTask{
			ID:        uuid.New(),
			UserID:    userID,
			ItemID:    item.ID,
			CardID:    uuid.Nil,
			TaskDate:  taskDate,
			Source:    "interval_review", // Mark as interval recurring review
			State:     "pending",
			CreatedAt: now,
		})
	}

	// ========== 2️⃣ Items FSRS Active yang due untuk review ==========
	fsrsItems, err := s.itemRepo.FindFSRSDueItems(userID, now)
	if err != nil {
		return nil, err
	}

	for _, item := range fsrsItems {
		tasks = append(tasks, entities.DailyTask{
			ID:        uuid.New(),
			UserID:    userID,
			ItemID:    item.ID,
			CardID:    uuid.Nil, // Item-based, no card
			TaskDate:  taskDate,
			Source:    "quran", // or item.SourceType
			State:     "pending",
			CreatedAt: now,
		})
	}

	// ========== 3️⃣ Cards dari FSRS Review State (existing logic) ==========
	candidates, err := s.reviewStateRepo.FindDueByUser(
		ctx,
		userID,
		now,
		limit,
	)
	if err != nil {
		return nil, err
	}

	for _, c := range candidates {
		tasks = append(tasks, entities.DailyTask{
			ID:        uuid.New(),
			UserID:    c.UserID,
			ItemID:    c.ItemID,
			CardID:    c.CardID,
			TaskDate:  taskDate,
			Source:    c.Source,
			State:     "pending",
			CreatedAt: now,
		})
	}

	// ========== 4️⃣ Graduate items untuk review bulanan (by juz) ==========
	// Default: Juz i → tanggal i. Jika sebagian juz non-aktif, gunakan antrian aktif harian.
	dayOfMonth := now.Day()
	targetJuzIndex := 0

	activeJuzs, err := s.juzRepo.FindActiveByUser(userID.String())
	if err == nil && len(activeJuzs) > 0 {
		// Round-robin berdasarkan daftar juz aktif
		pos := (dayOfMonth - 1) % len(activeJuzs)
		targetJuzIndex = activeJuzs[pos].Index
	} else {
		// Fallback ke mapping tanggal → juz (1..30)
		if dayOfMonth <= 30 {
			targetJuzIndex = dayOfMonth
		}
	}

	if targetJuzIndex > 0 {
		gradItems, err := s.itemRepo.FindGraduateItemsByJuzDay(userID, targetJuzIndex)
		if err != nil {
			return nil, err
		}
		for _, item := range gradItems {
			tasks = append(tasks, entities.DailyTask{
				ID:        uuid.New(),
				UserID:    userID,
				ItemID:    item.ID,
				CardID:    uuid.Nil,
				TaskDate:  taskDate,
				Source:    "graduate",
				State:     "pending",
				CreatedAt: now,
			})
		}
	}

	// Apply limit if needed
	if limit > 0 && len(tasks) > limit {
		tasks = tasks[:limit]
	}

	// 5️⃣ Simpan snapshot (IDEMPOTENT)
	if err := s.dailyTaskRepo.UpsertDailyTasks(
		ctx,
		userID,
		taskDate,
		tasks,
	); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *dailyTaskService) ListToday(
	ctx context.Context,
	userID uuid.UUID,
	now time.Time,
) ([]entities.DailyTask, error) {

	taskDate := utils.NormalizeDate(now)

	return s.dailyTaskRepo.ListByUserAndDate(
		ctx,
		userID,
		taskDate,
	)
}
