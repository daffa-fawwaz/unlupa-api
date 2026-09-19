package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"hifzhun-api/pkg/entities"
	"hifzhun-api/pkg/fsrs"
)

type QuranPageSummary struct {
	PageNumber     int        `json:"page_number"`
	JuzNumber      int        `json:"juz_number"`
	Status         string     `json:"status"` // new | learning | review | mapan
	Stability      float64    `json:"stability"`
	Difficulty     float64    `json:"difficulty"`
	LastReviewedAt *time.Time `json:"last_reviewed_at,omitempty"`
	NextReviewAt   *time.Time `json:"next_review_at,omitempty"`
	ReviewCount    int        `json:"review_count"`
	IsDue          bool       `json:"is_due"`
}

type QuranPagesStats struct {
	TotalPages     int     `json:"total_pages"`
	MapanPages     int     `json:"mapan_pages"`
	ReviewPages    int     `json:"review_pages"`
	LearningPages  int     `json:"learning_pages"`
	NewPages       int     `json:"new_pages"`
	DuePagesToday  int     `json:"due_pages_today"`
	MapanPercent   float64 `json:"mapan_percent"`
}

type QuranPagesResponse struct {
	Stats []QuranPagesStats   `json:"stats,omitempty"`
	Pages []QuranPageSummary `json:"pages"`
}

type SurahJuz30Info struct {
	Number       int    `json:"number"`
	Name         string `json:"name"`
	EnglishName  string `json:"english_name"`
	StartPage    int    `json:"start_page"`
	EndPage      int    `json:"end_page"`
	AyahCount    int    `json:"ayah_count"`
	Status       string `json:"status"` // mapan | partial | new
}

type Juz30ProgressResponse struct {
	JuzNumber    int                `json:"juz_number"`
	TotalPages   int                `json:"total_pages"`
	MapanPages   int                `json:"mapan_pages"`
	DuePages     int                `json:"due_pages"`
	MapanPercent float64            `json:"mapan_percent"`
	Pages        []QuranPageSummary `json:"pages"`
	Surahs       []SurahJuz30Info   `json:"surahs"`
}

type QuranPageService struct {
	db *gorm.DB
}

func NewQuranPageService(db *gorm.DB) *QuranPageService {
	return &QuranPageService{db: db}
}

// GetAllPagesProgress returns progress for all 604 pages of Mushaf Madani
func (s *QuranPageService) GetAllPagesProgress(ctx context.Context, userID uuid.UUID) ([]QuranPageSummary, *QuranPagesStats, error) {
	var records []entities.QuranPageProgress
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&records).Error; err != nil {
		return nil, nil, err
	}

	progressMap := make(map[int]entities.QuranPageProgress, len(records))
	for _, r := range records {
		progressMap[r.PageNumber] = r
	}

	now := time.Now()
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())

	pages := make([]QuranPageSummary, 604)
	stats := &QuranPagesStats{
		TotalPages: 604,
	}

	for p := 1; p <= 604; p++ {
		juz := entities.GetJuzForPage(p)
		record, exists := progressMap[p]

		if !exists {
			pages[p-1] = QuranPageSummary{
				PageNumber: p,
				JuzNumber:  juz,
				Status:     entities.QuranPageStatusNew,
				Difficulty: 5.0,
			}
			stats.NewPages++
		} else {
			isDue := false
			if record.NextReviewAt != nil && !record.NextReviewAt.After(todayEnd) {
				isDue = true
				stats.DuePagesToday++
			}

			stab := record.Stability
			if math.IsNaN(stab) || math.IsInf(stab, 0) || stab < 0 {
				stab = 0
			}
			diff := record.Difficulty
			if math.IsNaN(diff) || math.IsInf(diff, 0) || diff < 0 {
				diff = 5.0
			}

			pages[p-1] = QuranPageSummary{
				PageNumber:     p,
				JuzNumber:      juz,
				Status:         record.Status,
				Stability:      math.Round(stab*100) / 100,
				Difficulty:     math.Round(diff*100) / 100,
				LastReviewedAt: record.LastReviewedAt,
				NextReviewAt:   record.NextReviewAt,
				ReviewCount:    record.ReviewCount,
				IsDue:          isDue,
			}

			switch record.Status {
			case entities.QuranPageStatusMapan:
				stats.MapanPages++
			case entities.QuranPageStatusReview:
				stats.ReviewPages++
			case entities.QuranPageStatusLearning:
				stats.LearningPages++
			default:
				stats.NewPages++
			}
		}
	}

	stats.MapanPercent = math.Round((float64(stats.MapanPages)/604.0)*1000) / 10.0

	return pages, stats, nil
}

// ReviewPage records an FSRS review for a specific Mushaf page
func (s *QuranPageService) ReviewPage(
	ctx context.Context,
	userID uuid.UUID,
	pageNumber int,
	rating fsrs.Rating,
	now time.Time,
) (*entities.QuranPageProgress, error) {
	if pageNumber < 1 || pageNumber > 604 {
		return nil, errors.New("nomor halaman harus antara 1 dan 604")
	}
	if rating < fsrs.Again || rating > fsrs.Easy {
		return nil, errors.New("rating tidak valid (1-4)")
	}

	var progress entities.QuranPageProgress
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND page_number = ?", userID, pageNumber).
		First(&progress).Error

	isFirstReview := false
	if errors.Is(err, gorm.ErrRecordNotFound) {
		isFirstReview = true
		progress = entities.QuranPageProgress{
			ID:         uuid.New(),
			UserID:     userID,
			PageNumber: pageNumber,
			JuzNumber:  entities.GetJuzForPage(pageNumber),
			Status:     entities.QuranPageStatusLearning,
			Stability:  0.4,
			Difficulty: 5.0,
		}
	} else if err != nil {
		return nil, err
	}

	if isFirstReview {
		progress.Stability = 0.4
		progress.Difficulty = 5.0
	}
	if math.IsNaN(progress.Stability) || math.IsInf(progress.Stability, 0) || progress.Stability <= 0 {
		progress.Stability = 0.4
	}
	if math.IsNaN(progress.Difficulty) || math.IsInf(progress.Difficulty, 0) || progress.Difficulty <= 0 {
		progress.Difficulty = 5.0
	}

	var lastReview time.Time
	if progress.LastReviewedAt != nil {
		lastReview = *progress.LastReviewedAt
	}

	prevState := fsrs.CardState{
		Stability:  progress.Stability,
		Difficulty: progress.Difficulty,
		LastReview: lastReview,
	}

	weights := fsrs.DefaultWeights()
	result := fsrs.Review(prevState, rating, now, weights)

	progress.Stability = result.NewState.Stability
	progress.Difficulty = result.NewState.Difficulty
	if math.IsNaN(progress.Stability) || math.IsInf(progress.Stability, 0) || progress.Stability <= 0 {
		progress.Stability = 0.01
	}
	if math.IsNaN(progress.Difficulty) || math.IsInf(progress.Difficulty, 0) || progress.Difficulty <= 0 {
		progress.Difficulty = 5.0
	}

	progress.ReviewCount++
	progress.LastReviewedAt = &now

	nextReview := now.Add(result.Interval)
	nextReview = time.Date(nextReview.Year(), nextReview.Month(), nextReview.Day(), 0, 0, 0, 0, nextReview.Location())
	progress.NextReviewAt = &nextReview

	// Status transition
	if progress.Stability >= 30.0 {
		progress.Status = entities.QuranPageStatusMapan
	} else if rating == fsrs.Again {
		progress.Status = entities.QuranPageStatusLearning
	} else {
		progress.Status = entities.QuranPageStatusReview
	}

	// Upsert to DB
	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "page_number"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"status", "stability", "difficulty", "last_reviewed_at", "next_review_at", "review_count", "updated_at",
		}),
	}).Save(&progress).Error; err != nil {
		return nil, err
	}

	// Also sync user's Item record in items table
	contentRef := fmt.Sprintf("page:%d", pageNumber)
	intervalDays := int(result.Interval.Hours() / 24)
	if intervalDays < 1 {
		intervalDays = 1
	}

	itemStatus := entities.ItemStatusFSRSActive
	if progress.Stability >= 30.0 || intervalDays >= 30 {
		itemStatus = entities.ItemStatusGraduate
	}

	var existingItem entities.Item
	err = s.db.WithContext(ctx).Where("owner_id = ? AND source_type = ? AND content_ref = ?", userID, "quran", contentRef).First(&existingItem).Error
	if err == nil {
		existingItem.Status = itemStatus
		existingItem.Stability = progress.Stability
		existingItem.Difficulty = progress.Difficulty
		existingItem.ReviewCount = progress.ReviewCount
		existingItem.LastReviewAt = progress.LastReviewedAt
		existingItem.NextReviewAt = progress.NextReviewAt
		existingItem.IntervalDays = intervalDays
		_ = s.db.WithContext(ctx).Save(&existingItem).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		juzNum := entities.GetJuzForPage(pageNumber)
		var personalJuz entities.Juz
		_ = s.db.WithContext(ctx).Where("user_id = ? AND index = ? AND class_id IS NULL", userID, juzNum).First(&personalJuz).Error
		if personalJuz.ID == uuid.Nil {
			personalJuz = entities.Juz{
				UserID:   userID,
				Index:    juzNum,
				IsActive: true,
			}
			_ = s.db.WithContext(ctx).Create(&personalJuz).Error
		}
		newItem := entities.Item{
			OwnerID:      userID,
			SourceType:   "quran",
			ContentRef:   contentRef,
			Status:       itemStatus,
			Stability:    progress.Stability,
			Difficulty:   progress.Difficulty,
			ReviewCount:  progress.ReviewCount,
			LastReviewAt: progress.LastReviewedAt,
			NextReviewAt: progress.NextReviewAt,
			IntervalDays: intervalDays,
		}
		if err := s.db.WithContext(ctx).Create(&newItem).Error; err == nil && personalJuz.ID != uuid.Nil {
			juzItem := entities.JuzItem{
				ID:     uuid.New(),
				JuzID:  personalJuz.ID,
				ItemID: newItem.ID,
			}
			_ = s.db.WithContext(ctx).Create(&juzItem).Error
		}
	}

	return &progress, nil
}

// GetJuz30Progress returns detailed page and surah metrics for Juz 30 (pages 582-604)
func (s *QuranPageService) GetJuz30Progress(ctx context.Context, userID uuid.UUID) (*Juz30ProgressResponse, error) {
	var records []entities.QuranPageProgress
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND juz_number = 30", userID).
		Find(&records).Error; err != nil {
		return nil, err
	}

	pageMap := make(map[int]entities.QuranPageProgress, len(records))
	for _, r := range records {
		pageMap[r.PageNumber] = r
	}

	now := time.Now()
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())

	pages := make([]QuranPageSummary, 0, 23)
	mapanCount := 0
	dueCount := 0

	for p := 582; p <= 604; p++ {
		record, exists := pageMap[p]
		if !exists {
			pages = append(pages, QuranPageSummary{
				PageNumber: p,
				JuzNumber:  30,
				Status:     entities.QuranPageStatusNew,
				Difficulty: 5.0,
			})
		} else {
			isDue := record.NextReviewAt != nil && !record.NextReviewAt.After(todayEnd)
			if isDue {
				dueCount++
			}
			if record.Status == entities.QuranPageStatusMapan {
				mapanCount++
			}
			pages = append(pages, QuranPageSummary{
				PageNumber:     p,
				JuzNumber:      30,
				Status:         record.Status,
				Stability:      record.Stability,
				Difficulty:     record.Difficulty,
				LastReviewedAt: record.LastReviewedAt,
				NextReviewAt:   record.NextReviewAt,
				ReviewCount:    record.ReviewCount,
				IsDue:          isDue,
			})
		}
	}

	surahs := GetJuz30SurahsMetadata()
	for i, surah := range surahs {
		surahMapan := true
		hasAnyLearning := false
		for p := surah.StartPage; p <= surah.EndPage; p++ {
			rec, ok := pageMap[p]
			if !ok || rec.Status != entities.QuranPageStatusMapan {
				surahMapan = false
			}
			if ok && (rec.Status == entities.QuranPageStatusLearning || rec.Status == entities.QuranPageStatusReview) {
				hasAnyLearning = true
			}
		}
		if surahMapan {
			surahs[i].Status = "mapan"
		} else if hasAnyLearning {
			surahs[i].Status = "learning"
		} else {
			surahs[i].Status = "new"
		}
	}

	mapanPercent := math.Round((float64(mapanCount)/23.0)*1000) / 10.0

	return &Juz30ProgressResponse{
		JuzNumber:    30,
		TotalPages:   23,
		MapanPages:   mapanCount,
		DuePages:     dueCount,
		MapanPercent: mapanPercent,
		Pages:        pages,
		Surahs:       surahs,
	}, nil
}

// GetJuz30SurahsMetadata returns metadata for Surahs in Juz 30 (Surah 78-114)
func GetJuz30SurahsMetadata() []SurahJuz30Info {
	return []SurahJuz30Info{
		{Number: 78, Name: "An-Naba'", EnglishName: "The Tidings", StartPage: 582, EndPage: 583, AyahCount: 40},
		{Number: 79, Name: "An-Nazi'at", EnglishName: "Those Who Drag Forth", StartPage: 583, EndPage: 584, AyahCount: 46},
		{Number: 80, Name: "'Abasa", EnglishName: "He Frowned", StartPage: 585, EndPage: 586, AyahCount: 42},
		{Number: 81, Name: "At-Takwir", EnglishName: "The Overthrowing", StartPage: 586, EndPage: 586, AyahCount: 29},
		{Number: 82, Name: "Al-Infitar", EnglishName: "The Cleaving", StartPage: 587, EndPage: 587, AyahCount: 19},
		{Number: 83, Name: "Al-Mutaffifin", EnglishName: "Defrauding", StartPage: 587, EndPage: 589, AyahCount: 36},
		{Number: 84, Name: "Al-Inshiqaq", EnglishName: "The Sundering", StartPage: 589, EndPage: 590, AyahCount: 25},
		{Number: 85, Name: "Al-Buruj", EnglishName: "The Mansions of the Stars", StartPage: 590, EndPage: 590, AyahCount: 22},
		{Number: 86, Name: "At-Tariq", EnglishName: "The Morning Star", StartPage: 591, EndPage: 591, AyahCount: 17},
		{Number: 87, Name: "Al-A'la", EnglishName: "The Most High", StartPage: 591, EndPage: 592, AyahCount: 19},
		{Number: 88, Name: "Al-Ghashiyah", EnglishName: "The Overwhelming Event", StartPage: 592, EndPage: 593, AyahCount: 26},
		{Number: 89, Name: "Al-Fajr", EnglishName: "The Dawn", StartPage: 593, EndPage: 594, AyahCount: 30},
		{Number: 90, Name: "Al-Balad", EnglishName: "The City", StartPage: 594, EndPage: 595, AyahCount: 20},
		{Number: 91, Name: "Ash-Shams", EnglishName: "The Sun", StartPage: 595, EndPage: 595, AyahCount: 15},
		{Number: 92, Name: "Al-Layl", EnglishName: "The Night", StartPage: 595, EndPage: 596, AyahCount: 21},
		{Number: 93, Name: "Ad-Duha", EnglishName: "The Morning Hours", StartPage: 596, EndPage: 596, AyahCount: 11},
		{Number: 94, Name: "Ash-Sharh", EnglishName: "The Relief", StartPage: 596, EndPage: 597, AyahCount: 8},
		{Number: 95, Name: "At-Tin", EnglishName: "The Fig", StartPage: 597, EndPage: 597, AyahCount: 8},
		{Number: 96, Name: "Al-'Alaq", EnglishName: "The Clot", StartPage: 597, EndPage: 598, AyahCount: 19},
		{Number: 97, Name: "Al-Qadr", EnglishName: "The Power", StartPage: 598, EndPage: 598, AyahCount: 5},
		{Number: 98, Name: "Al-Bayyinah", EnglishName: "The Clear Proof", StartPage: 598, EndPage: 599, AyahCount: 8},
		{Number: 99, Name: "Az-Zalzalah", EnglishName: "The Earthquake", StartPage: 599, EndPage: 599, AyahCount: 8},
		{Number: 100, Name: "Al-'Adiyat", EnglishName: "The Courser", StartPage: 599, EndPage: 600, AyahCount: 11},
		{Number: 101, Name: "Al-Qari'ah", EnglishName: "The Calamity", StartPage: 600, EndPage: 600, AyahCount: 11},
		{Number: 102, Name: "At-Takathur", EnglishName: "The Rivalry in World Increase", StartPage: 600, EndPage: 600, AyahCount: 8},
		{Number: 103, Name: "Al-'Asr", EnglishName: "The Declining Day", StartPage: 601, EndPage: 601, AyahCount: 3},
		{Number: 104, Name: "Al-Humazah", EnglishName: "The Traducer", StartPage: 601, EndPage: 601, AyahCount: 9},
		{Number: 105, Name: "Al-Fil", EnglishName: "The Elephant", StartPage: 601, EndPage: 601, AyahCount: 5},
		{Number: 106, Name: "Quraysh", EnglishName: "Quraysh", StartPage: 602, EndPage: 602, AyahCount: 4},
		{Number: 107, Name: "Al-Ma'un", EnglishName: "The Small Kindness", StartPage: 602, EndPage: 602, AyahCount: 7},
		{Number: 108, Name: "Al-Kawthar", EnglishName: "The Abundance", StartPage: 602, EndPage: 602, AyahCount: 3},
		{Number: 109, Name: "Al-Kafirun", EnglishName: "The Disbelievers", StartPage: 603, EndPage: 603, AyahCount: 6},
		{Number: 110, Name: "An-Nasr", EnglishName: "The Divine Support", StartPage: 603, EndPage: 603, AyahCount: 3},
		{Number: 111, Name: "Al-Masad", EnglishName: "The Palm Fiber", StartPage: 603, EndPage: 603, AyahCount: 5},
		{Number: 112, Name: "Al-Ikhlas", EnglishName: "The Sincerity", StartPage: 604, EndPage: 604, AyahCount: 4},
		{Number: 113, Name: "Al-Falaq", EnglishName: "The Daybreak", StartPage: 604, EndPage: 604, AyahCount: 5},
		{Number: 114, Name: "An-Nas", EnglishName: "Mankind", StartPage: 604, EndPage: 604, AyahCount: 6},
	}
}
