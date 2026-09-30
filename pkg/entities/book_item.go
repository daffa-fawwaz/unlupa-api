package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookItem struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	BookID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"book_id"`
	ModuleID *uuid.UUID `gorm:"type:uuid;index" json:"module_id,omitempty"` // null jika langsung di book

	// ImporterID: null untuk item milik pemilik buku (canonical).
	// Terisi dengan user_id importer jika item dibuat oleh non-owner pada
	// published book — item tersebut HANYA terlihat untuk importer tersebut.
	ImporterID *uuid.UUID `gorm:"type:uuid;index" json:"importer_id,omitempty"`

	Title       string `gorm:"size:200;not null" json:"title"`
	Content     string `gorm:"type:text" json:"content"`     // materi konten
	Answer      string `gorm:"type:text" json:"answer"`      // jawaban
	Explanation string `gorm:"type:text" json:"explanation"` // penjelasan (opsional)
	Order       int    `gorm:"not null;default:0" json:"order"`

	// Gambar item (hanya untuk user premium)
	ImageURL string `gorm:"size:500" json:"image_url,omitempty"`

	// Estimasi waktu review (detik) opsional untuk item buku.
	EstimatedReviewSeconds int `gorm:"default:0" json:"estimated_review_seconds"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (i *BookItem) BeforeCreate(tx *gorm.DB) error {
	i.ID = uuid.New()
	return nil
}
