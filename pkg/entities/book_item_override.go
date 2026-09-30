package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BookItemOverride menyimpan versi personal konten BookItem untuk user tertentu.
// Ketika user yang bukan pemilik buku mengedit/menambah item, perubahan disimpan
// di sini — bukan di book_items — sehingga pemilik buku dan user lain tidak
// terpengaruh. Pemilik buku selalu menulis langsung ke book_items.
//
// Format content_ref yang merujuk ke override tetap sama:
//
//	"book:{book_id}:item:{book_item_id}"
//
// Sehingga semua layer yang parse content_ref tidak perlu berubah; hanya layer
// yang *membaca konten* (title/content/answer) perlu memanggil
// resolveBookItemContent agar override diutamakan.
type BookItemOverride struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	BookItemID uuid.UUID `gorm:"type:uuid;not null;index" json:"book_item_id"`

	Title       string `gorm:"size:200" json:"title"`
	Content     string `gorm:"type:text" json:"content"`
	Answer      string `gorm:"type:text" json:"answer"`
	Explanation string `gorm:"type:text" json:"explanation"`
	ImageURL    string `gorm:"size:500" json:"image_url,omitempty"`

	EstimatedReviewSeconds int `gorm:"default:0" json:"estimated_review_seconds"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (o *BookItemOverride) BeforeCreate(tx *gorm.DB) error {
	o.ID = uuid.New()
	return nil
}
