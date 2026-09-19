package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"hifzhun-api/pkg/config"
	"hifzhun-api/pkg/entities"
)

type JuzCatalogWithUserStats struct {
	JuzNumber     int    `json:"juz_number"`
	NameAr        string `json:"name_ar"`
	NameEn        string `json:"name_en"`
	StartPage     int    `json:"start_page"`
	EndPage       int    `json:"end_page"`
	TotalPages    int    `json:"total_pages"`
	StartSurah    string `json:"start_surah"`
	EndSurah      string `json:"end_surah"`
	SurahSpan     string `json:"surah_span"`
	AyahSpan      string `json:"ayah_span"`
	ActivePages   int    `json:"active_pages"`
	MasteredPages int    `json:"mastered_pages"`
	DueToday      int    `json:"due_today"`
}

type PageCatalogWithUserStatus struct {
	MushafPage       int        `json:"mushaf_page"`
	JuzNumber        int        `json:"juz_number"`
	PageNumberInJuz  int        `json:"page_number_in_juz"`
	SurahNameEn      string     `json:"surah_name_en"`
	SurahNameAr      string     `json:"surah_name_ar"`
	SurahNumber      int        `json:"surah_number"`
	AyahRange        string     `json:"ayah_range"`
	StartVerseKey    string     `json:"start_verse_key"`
	EndVerseKey      string     `json:"end_verse_key"`
	ContentRef       string     `json:"content_ref"`
	IsActivated      bool       `json:"is_activated"`
	ActivationStatus string     `json:"activation_status"` // not_activated | active | mastered
	ReviewStatus     string     `json:"review_status"`     // menghafal | interval | fsrs_active | graduate | ""
	ItemID           *uuid.UUID `json:"item_id,omitempty"`
	NextReviewAt     *time.Time `json:"next_review_at,omitempty"`
	LastReviewAt     *time.Time `json:"last_review_at,omitempty"`
	Stability        float64    `json:"stability"`
	Difficulty       float64    `json:"difficulty"`
	ReviewCount      int        `json:"review_count"`
	IsDue            bool       `json:"is_due"`
}

type QuranCatalogRepository interface {
	FindAllJuzWithUserStats(ctx context.Context, userID uuid.UUID, now time.Time) ([]JuzCatalogWithUserStats, error)
	FindJuzPagesWithUserStatus(ctx context.Context, juzNumber int, userID uuid.UUID, now time.Time) ([]PageCatalogWithUserStatus, error)
	FindPageByMushafPage(ctx context.Context, mushafPage int) (*entities.QuranPageCatalog, error)
	FindJuzByNumber(ctx context.Context, juzNumber int) (*entities.QuranJuzCatalog, error)
	GetDB() *gorm.DB
}

type quranCatalogRepository struct {
	db *gorm.DB
}

func NewQuranCatalogRepository(db *gorm.DB) QuranCatalogRepository {
	return &quranCatalogRepository{db: db}
}

func (r *quranCatalogRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *quranCatalogRepository) FindPageByMushafPage(ctx context.Context, mushafPage int) (*entities.QuranPageCatalog, error) {
	var page entities.QuranPageCatalog
	err := r.db.WithContext(ctx).Where("mushaf_page = ?", mushafPage).First(&page).Error
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *quranCatalogRepository) FindJuzByNumber(ctx context.Context, juzNumber int) (*entities.QuranJuzCatalog, error) {
	var juz entities.QuranJuzCatalog
	err := r.db.WithContext(ctx).Where("juz_number = ?", juzNumber).First(&juz).Error
	if err != nil {
		return nil, err
	}
	return &juz, nil
}

func (r *quranCatalogRepository) FindAllJuzWithUserStats(ctx context.Context, userID uuid.UUID, now time.Time) ([]JuzCatalogWithUserStats, error) {
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	// Single query with aggregation
	query := `
		SELECT 
			j.juz_number,
			j.name_ar,
			j.name_en,
			j.start_page,
			j.end_page,
			j.total_pages,
			j.start_surah,
			j.end_surah,
			j.surah_span,
			j.ayah_span,
			COUNT(DISTINCT i.id) AS active_pages,
			COUNT(DISTINCT CASE WHEN i.status = 'graduate' THEN i.id END) AS mastered_pages,
			COUNT(DISTINCT CASE 
				WHEN i.status = 'fsrs_active' AND i.next_review_at IS NOT NULL AND i.next_review_at <= ? THEN i.id
				WHEN i.status = 'interval' AND i.interval_next_review_at IS NOT NULL AND i.interval_next_review_at <= ? THEN i.id
				ELSE NULL
			END) AS due_today
		FROM quran_juz_catalogs j
		LEFT JOIN quran_page_catalogs p ON p.juz_number = j.juz_number
		LEFT JOIN items i ON i.content_ref = p.content_ref AND i.owner_id = ? AND i.source_type = 'quran'
		GROUP BY j.juz_number, j.name_ar, j.name_en, j.start_page, j.end_page, j.total_pages, j.start_surah, j.end_surah, j.surah_span, j.ayah_span
		ORDER BY j.juz_number ASC
	`

	var results []JuzCatalogWithUserStats
	err := r.db.WithContext(ctx).Raw(query, endOfDay, endOfDay, userID).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// Fallback to static catalog if results are empty
	if len(results) == 0 {
		var juzs []entities.QuranJuzCatalog
		if err := r.db.WithContext(ctx).Order("juz_number ASC").Find(&juzs).Error; err == nil {
			for _, j := range juzs {
				results = append(results, JuzCatalogWithUserStats{
					JuzNumber:   j.JuzNumber,
					NameAr:      j.NameAr,
					NameEn:      j.NameEn,
					StartPage:   j.StartPage,
					EndPage:     j.EndPage,
					TotalPages:  j.TotalPages,
					StartSurah:  j.StartSurah,
					EndSurah:    j.EndSurah,
					SurahSpan:   j.SurahSpan,
					AyahSpan:    j.AyahSpan,
					ActivePages: 0,
				})
			}
		}
	}

	return results, nil
}

func (r *quranCatalogRepository) FindJuzPagesWithUserStatus(ctx context.Context, juzNumber int, userID uuid.UUID, now time.Time) ([]PageCatalogWithUserStatus, error) {
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, config.AppLocation)

	type pageQueryResult struct {
		MushafPage           int
		JuzNumber            int
		PageNumberInJuz      int
		SurahNameEn          string
		SurahNameAr          string
		SurahNumber          int
		AyahRange            string
		StartVerseKey        string
		EndVerseKey          string
		ContentRef           string
		ItemID               *uuid.UUID
		Status               *string
		Stability            *float64
		Difficulty           *float64
		ReviewCount          *int
		NextReviewAt         *time.Time
		IntervalNextReviewAt *time.Time
		LastReviewAt         *time.Time
	}

	query := `
		SELECT 
			p.mushaf_page,
			p.juz_number,
			p.page_number_in_juz,
			p.surah_name_en,
			p.surah_name_ar,
			p.surah_number,
			p.ayah_range,
			p.start_verse_key,
			p.end_verse_key,
			p.content_ref,
			i.id AS item_id,
			i.status,
			i.stability,
			i.difficulty,
			i.review_count,
			i.next_review_at,
			i.interval_next_review_at,
			i.last_review_at
		FROM quran_page_catalogs p
		LEFT JOIN items i ON i.content_ref = p.content_ref AND i.owner_id = ? AND i.source_type = 'quran'
		WHERE p.juz_number = ?
		ORDER BY p.mushaf_page ASC
	`

	var queryResults []pageQueryResult
	err := r.db.WithContext(ctx).Raw(query, userID, juzNumber).Scan(&queryResults).Error
	if err != nil {
		return nil, err
	}

	output := make([]PageCatalogWithUserStatus, len(queryResults))
	for idx, row := range queryResults {
		isActivated := row.ItemID != nil
		activationStatus := "not_activated"
		reviewStatus := ""
		isDue := false
		stability := 0.0
		difficulty := 5.0
		reviewCount := 0
		var nextReview *time.Time

		if isActivated && row.Status != nil {
			reviewStatus = *row.Status
			if reviewStatus == entities.ItemStatusGraduate {
				activationStatus = "mastered"
			} else {
				activationStatus = "active"
			}

			if row.Stability != nil {
				stability = *row.Stability
			}
			if row.Difficulty != nil {
				difficulty = *row.Difficulty
			}
			if row.ReviewCount != nil {
				reviewCount = *row.ReviewCount
			}

			if reviewStatus == entities.ItemStatusInterval {
				nextReview = row.IntervalNextReviewAt
				if nextReview != nil && !nextReview.After(endOfDay) {
					isDue = true
				}
			} else {
				nextReview = row.NextReviewAt
				if nextReview != nil && !nextReview.After(endOfDay) {
					isDue = true
				}
			}
		}

		output[idx] = PageCatalogWithUserStatus{
			MushafPage:       row.MushafPage,
			JuzNumber:        row.JuzNumber,
			PageNumberInJuz:  row.PageNumberInJuz,
			SurahNameEn:      row.SurahNameEn,
			SurahNameAr:      row.SurahNameAr,
			SurahNumber:      row.SurahNumber,
			AyahRange:        row.AyahRange,
			StartVerseKey:    row.StartVerseKey,
			EndVerseKey:      row.EndVerseKey,
			ContentRef:       row.ContentRef,
			IsActivated:      isActivated,
			ActivationStatus: activationStatus,
			ReviewStatus:     reviewStatus,
			ItemID:           row.ItemID,
			NextReviewAt:     nextReview,
			LastReviewAt:     row.LastReviewAt,
			Stability:        stability,
			Difficulty:       difficulty,
			ReviewCount:      reviewCount,
			IsDue:            isDue,
		}
	}

	return output, nil
}
