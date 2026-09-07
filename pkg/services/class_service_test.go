package services_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/entities"
)

func TestClassBookStudentProgressCategorization(t *testing.T) {
	config.InitAppLocation()
	now := time.Now().In(config.AppLocation)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	tests := []struct {
		name               string
		status             string
		nextReviewAt       *time.Time
		lastReviewAt       *time.Time
		taskState          string
		hasTask            bool
		expectedUnreviewed bool // true if categorized as Belum di-review (totalUnreviewed++)
		expectedInactive   bool // true if categorized as Nonaktif (totalInactive++)
	}{
		{
			name:               "due + belum review -> Belum di-review +1",
			status:             entities.ItemStatusFSRSActive,
			nextReviewAt:       &now,
			taskState:          "pending",
			hasTask:            true,
			expectedUnreviewed: true,
			expectedInactive:   false,
		},
		{
			name:               "due + sudah review -> Belum di-review +0",
			status:             entities.ItemStatusFSRSActive,
			nextReviewAt:       &now,
			lastReviewAt:       &now,
			taskState:          "done",
			hasTask:            true,
			expectedUnreviewed: false,
			expectedInactive:   false,
		},
		{
			name:               "overdue + belum review -> Belum di-review +1",
			status:             entities.ItemStatusFSRSActive,
			nextReviewAt:       &yesterday,
			taskState:          "pending",
			hasTask:            true,
			expectedUnreviewed: true,
			expectedInactive:   false,
		},
		{
			name:               "future/not due -> Belum di-review +0",
			status:             entities.ItemStatusFSRSActive,
			nextReviewAt:       &tomorrow,
			taskState:          "",
			hasTask:            false,
			expectedUnreviewed: false,
			expectedInactive:   false,
		},
		{
			name:               "graduate -> Nonaktif +1",
			status:             entities.ItemStatusGraduate,
			nextReviewAt:       nil,
			taskState:          "",
			hasTask:            false,
			expectedUnreviewed: false,
			expectedInactive:   true,
		},
		{
			name:               "inactive -> Nonaktif +1",
			status:             entities.ItemStatusInactive,
			nextReviewAt:       nil,
			taskState:          "",
			hasTask:            false,
			expectedUnreviewed: false,
			expectedInactive:   true,
		},
		{
			name:               "graduate + existing daily task -> tetap Nonaktif dan tidak dihitung sebagai Belum di-review",
			status:             entities.ItemStatusGraduate,
			nextReviewAt:       &now,
			taskState:          "pending",
			hasTask:            true,
			expectedUnreviewed: false,
			expectedInactive:   true,
		},
		{
			name:               "start item without deadline -> not due, Belum di-review +0",
			status:             entities.ItemStatusStart,
			nextReviewAt:       nil,
			taskState:          "",
			hasTask:            false,
			expectedUnreviewed: false,
			expectedInactive:   false,
		},
		{
			name:               "menghafal item without deadline -> not due, Belum di-review +0",
			status:             entities.ItemStatusMenghafal,
			nextReviewAt:       nil,
			taskState:          "",
			hasTask:            false,
			expectedUnreviewed: false,
			expectedInactive:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := entities.Item{
				ID:           uuid.New(),
				Status:       tt.status,
				NextReviewAt: tt.nextReviewAt,
				LastReviewAt: tt.lastReviewAt,
			}

			dailyTaskState := make(map[uuid.UUID]string)
			if tt.hasTask {
				dailyTaskState[item.ID] = tt.taskState
			}

			isInactive := false
			isUnreviewed := false

			if item.Status == entities.ItemStatusGraduate || item.Status == entities.ItemStatusInactive {
				isInactive = true
			} else {
				taskState, hasTask := dailyTaskState[item.ID]
				isTaskCompleted := hasTask && (taskState == "done" || taskState == "completed")
				isReviewedToday := isTaskCompleted || (item.LastReviewAt != nil && !item.LastReviewAt.Before(startOfDay) && !item.LastReviewAt.After(endOfDay))

				isDueBySchedule := (item.NextReviewAt != nil && !item.NextReviewAt.After(endOfDay)) || (item.IntervalNextReviewAt != nil && !item.IntervalNextReviewAt.After(endOfDay))
				isDue := (hasTask && !isTaskCompleted) || isDueBySchedule

				if isDue && !isReviewedToday {
					isUnreviewed = true
				}
			}

			if isUnreviewed != tt.expectedUnreviewed {
				t.Errorf("expected isUnreviewed=%v, got %v", tt.expectedUnreviewed, isUnreviewed)
			}
			if isInactive != tt.expectedInactive {
				t.Errorf("expected isInactive=%v, got %v", tt.expectedInactive, isInactive)
			}
		})
	}
}
