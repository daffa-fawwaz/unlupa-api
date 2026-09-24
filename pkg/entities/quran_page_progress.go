package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	QuranPageStatusNew      = "new"
	QuranPageStatusLearning = "learning"
	QuranPageStatusReview   = "review"
	QuranPageStatusMapan    = "mapan"
)

type QuranPageProgress struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_user_page,priority:1" json:"user_id"`
	PageNumber int        `gorm:"not null;uniqueIndex:idx_user_page,priority:2" json:"page_number"` // 1 - 604
	JuzNumber  int        `gorm:"not null;index" json:"juz_number"`                                 // 1 - 30
	Status     string     `gorm:"size:20;not null;default:'new'" json:"status"`                     // new | learning | review | mapan

	// FSRS Parameters
	Stability    float64    `gorm:"default:0" json:"stability"`
	Difficulty   float64    `gorm:"default:5.0" json:"difficulty"`
	LastReviewedAt *time.Time `gorm:"type:timestamp" json:"last_reviewed_at,omitempty"`
	NextReviewAt *time.Time `gorm:"type:timestamp;index" json:"next_review_at,omitempty"`
	ReviewCount  int        `gorm:"default:0" json:"review_count"`
	HasReachedMapan bool    `gorm:"default:false" json:"has_reached_mapan"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *QuranPageProgress) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.Status == "" {
		p.Status = QuranPageStatusNew
	}
	return nil
}

// GetJuzForPage returns the Juz number (1-30) for a given standard Mushaf Madani page (1-604)
func GetJuzForPage(page int) int {
	if page < 1 {
		return 1
	}
	if page > 604 {
		return 30
	}
	if page <= 21 {
		return 1
	}
	if page >= 582 {
		return 30
	}
	// Pages 22 to 581 are evenly 20 pages per Juz
	return ((page - 22) / 20) + 2
}
