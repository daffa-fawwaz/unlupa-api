package services

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/repositories"
)

type WeakSpotItem struct {
	ItemID       uuid.UUID  `json:"item_id"`
	Title        string     `json:"title"`
	Subtitle     string     `json:"subtitle"`
	SourceType   string     `json:"source_type"`
	Status       string     `json:"status"`
	Stability    float64    `json:"stability"`
	Difficulty   float64    `json:"difficulty"`
	ReviewCount  int        `json:"review_count"`
	LastReviewAt *time.Time `json:"last_review_at,omitempty"`
	NextReviewAt *time.Time `json:"next_review_at,omitempty"`
}

type DailyActivity struct {
	Date  string `json:"date"`  // YYYY-MM-DD
	Count int    `json:"count"` // Number of reviews on that day
}

type UpcomingReviewForecast struct {
	Date  string `json:"date"`  // YYYY-MM-DD
	Count int    `json:"count"` // Due items on that date
}

type DashboardStatsResponse struct {
	TotalItems             int64                    `json:"total_items"`
	TotalMemorized         int64                    `json:"total_memorized"`
	TotalLearning          int64                    `json:"total_learning"`
	DueReviewsToday        int64                    `json:"due_reviews_today"`
	CompletedReviewsToday  int64                    `json:"completed_reviews_today"`
	CurrentStreak          int                      `json:"current_streak"`
	LongestStreak          int                      `json:"longest_streak"`
	EstimatedFocusMinutes  int                      `json:"estimated_focus_minutes"`
	ConsistencyHeatmap     []DailyActivity          `json:"consistency_heatmap"`
	WeakSpots              []WeakSpotItem           `json:"weak_spots"`
	UpcomingForecast       []UpcomingReviewForecast `json:"upcoming_forecast"`
}

type DashboardService struct {
	db           *gorm.DB
	itemRepo     *repositories.ItemRepository
	juzItemRepo  *repositories.JuzItemRepository
	bookItemRepo repositories.BookItemRepository
	bookRepo     repositories.BookRepository
}

func NewDashboardService(
	db *gorm.DB,
	itemRepo *repositories.ItemRepository,
	juzItemRepo *repositories.JuzItemRepository,
	bookItemRepo repositories.BookItemRepository,
	bookRepo repositories.BookRepository,
) *DashboardService {
	return &DashboardService{
		db:           db,
		itemRepo:     itemRepo,
		juzItemRepo:  juzItemRepo,
		bookItemRepo: bookItemRepo,
		bookRepo:     bookRepo,
	}
}

func (s *DashboardService) GetDashboardStats(ctx context.Context, userID uuid.UUID) (*DashboardStatsResponse, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)

	resp := &DashboardStatsResponse{
		ConsistencyHeatmap: []DailyActivity{},
		WeakSpots:          []WeakSpotItem{},
		UpcomingForecast:   []UpcomingReviewForecast{},
	}

	// 1. Total Items & Counts
	if err := s.db.WithContext(ctx).Model(&entities.Item{}).
		Where("owner_id = ?", userID).
		Count(&resp.TotalItems).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&entities.Item{}).
		Where("owner_id = ? AND status IN (?, ?)", userID, entities.ItemStatusGraduate, "graduated").
		Count(&resp.TotalMemorized).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&entities.Item{}).
		Where("owner_id = ? AND status IN (?, ?, ?, ?)", userID, entities.ItemStatusFSRSActive, entities.ItemStatusMenghafal, entities.ItemStatusStart, entities.ItemStatusInterval).
		Count(&resp.TotalLearning).Error; err != nil {
		return nil, err
	}

	// 2. Due Reviews Today
	if err := s.db.WithContext(ctx).Model(&entities.Item{}).
		Where("owner_id = ? AND (next_review_at <= ? OR (last_review_at IS NULL AND status IN (?, ?))) AND status != ?",
			userID, todayEnd, entities.ItemStatusStart, entities.ItemStatusMenghafal, entities.ItemStatusInactive).
		Count(&resp.DueReviewsToday).Error; err != nil {
		return nil, err
	}

	// 3. Completed Reviews Today from review_logs
	if err := s.db.WithContext(ctx).Model(&entities.ReviewLog{}).
		Where("user_id = ? AND reviewed_at >= ? AND reviewed_at < ?", userID, todayStart, todayEnd).
		Count(&resp.CompletedReviewsToday).Error; err != nil {
		resp.CompletedReviewsToday = 0
	}

	// 4. Consistency Heatmap (Past 365 Days)
	pastYear := todayStart.AddDate(-1, 0, 0)
	type DateCount struct {
		DateStr string `gorm:"column:d"`
		Count   int    `gorm:"column:c"`
	}
	var logsCounts []DateCount
	// PostgreSQL date formatting
	err := s.db.WithContext(ctx).Raw(`
		SELECT TO_CHAR(reviewed_at, 'YYYY-MM-DD') AS d, COUNT(*) AS c
		FROM review_logs
		WHERE user_id = ? AND reviewed_at >= ?
		GROUP BY TO_CHAR(reviewed_at, 'YYYY-MM-DD')
		ORDER BY d ASC
	`, userID, pastYear).Scan(&logsCounts).Error

	activityMap := make(map[string]int)
	if err == nil {
		for _, lc := range logsCounts {
			activityMap[lc.DateStr] = lc.Count
		}
	}

	// Calculate streaks
	currentStreak := 0
	longestStreak := 0
	tempStreak := 0

	// Check consecutive days starting from today or yesterday
	checkDay := todayStart
	if activityMap[checkDay.Format("2006-01-02")] == 0 {
		checkDay = checkDay.AddDate(0, 0, -1)
	}

	for {
		dStr := checkDay.Format("2006-01-02")
		if activityMap[dStr] > 0 {
			currentStreak++
			checkDay = checkDay.AddDate(0, 0, -1)
		} else {
			break
		}
	}
	resp.CurrentStreak = currentStreak

	// Longest streak in past year
	for d := pastYear; !d.After(todayStart); d = d.AddDate(0, 0, 1) {
		dStr := d.Format("2006-01-02")
		c := activityMap[dStr]
		resp.ConsistencyHeatmap = append(resp.ConsistencyHeatmap, DailyActivity{
			Date:  dStr,
			Count: c,
		})
		if c > 0 {
			tempStreak++
			if tempStreak > longestStreak {
				longestStreak = tempStreak
			}
		} else {
			tempStreak = 0
		}
	}
	resp.LongestStreak = longestStreak

	// 5. Weak Spots (Items with high difficulty or low stability needing attention)
	var weakItems []entities.Item
	s.db.WithContext(ctx).
		Where("owner_id = ? AND status IN (?, ?, ?) AND status != ?",
			userID, entities.ItemStatusFSRSActive, entities.ItemStatusGraduate, entities.ItemStatusMenghafal, entities.ItemStatusInactive).
		Order("difficulty DESC, stability ASC, review_count DESC").
		Limit(6).
		Find(&weakItems)

	for _, it := range weakItems {
		title, subtitle := s.resolveItemTitleAndSubtitle(it)
		stab := it.Stability
		if math.IsNaN(stab) || math.IsInf(stab, 0) || stab < 0 {
			stab = 0
		}
		diff := it.Difficulty
		if math.IsNaN(diff) || math.IsInf(diff, 0) || diff < 0 {
			diff = 5.0
		}

		resp.WeakSpots = append(resp.WeakSpots, WeakSpotItem{
			ItemID:       it.ID,
			Title:        title,
			Subtitle:     subtitle,
			SourceType:   it.SourceType,
			Status:       it.Status,
			Stability:    math.Round(stab*100) / 100,
			Difficulty:   math.Round(diff*100) / 100,
			ReviewCount:  it.ReviewCount,
			LastReviewAt: it.LastReviewAt,
			NextReviewAt: it.NextReviewAt,
		})
	}

	// 6. Upcoming 7-Day Forecast
	type ForecastCount struct {
		DateStr string `gorm:"column:d"`
		Count   int    `gorm:"column:c"`
	}
	var forecasts []ForecastCount
	nextWeek := todayStart.AddDate(0, 0, 7)
	s.db.WithContext(ctx).Raw(`
		SELECT TO_CHAR(next_review_at, 'YYYY-MM-DD') AS d, COUNT(*) AS c
		FROM items
		WHERE owner_id = ? AND next_review_at >= ? AND next_review_at <= ? AND status != 'inactive'
		GROUP BY TO_CHAR(next_review_at, 'YYYY-MM-DD')
		ORDER BY d ASC
	`, userID, todayStart, nextWeek).Scan(&forecasts)

	forecastMap := make(map[string]int)
	for _, fc := range forecasts {
		forecastMap[fc.DateStr] = fc.Count
	}
	for i := 0; i < 7; i++ {
		d := todayStart.AddDate(0, 0, i)
		dStr := d.Format("2006-01-02")
		resp.UpcomingForecast = append(resp.UpcomingForecast, UpcomingReviewForecast{
			Date:  dStr,
			Count: forecastMap[dStr],
		})
	}

	// 7. Estimated Focus Time
	resp.EstimatedFocusMinutes = int(resp.DueReviewsToday*3 + resp.CompletedReviewsToday*3)

	return resp, nil
}

func (s *DashboardService) resolveItemTitleAndSubtitle(item entities.Item) (string, string) {
	if item.SourceType == "book" {
		parts := strings.Split(item.ContentRef, ":")
		if len(parts) >= 4 && parts[0] == "book" && parts[2] == "item" {
			bookItemID := parts[3]
			if bi, err := s.bookItemRepo.FindByID(bookItemID); err == nil && bi != nil {
				subtitle := "Materi Kitab"
				if b, err := s.bookRepo.FindByID(parts[1]); err == nil && b != nil {
					subtitle = b.Title
				}
				return bi.Title, subtitle
			}
		}
		return "Materi Kitab", "Personal Book"
	}

	// Quran items
	parts := strings.Split(item.ContentRef, ":")
	if len(parts) >= 2 {
		return fmt.Sprintf("Al-Qur'an (%s)", parts[len(parts)-1]), "Hafalan Al-Qur'an"
	}

	return "Item Hafalan", "Al-Qur'an"
}
