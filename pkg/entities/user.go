package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	Username string `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email    string `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Password string `gorm:"size:255;not null" json:"-"`

	Role     string `gorm:"size:10;not null" json:"role"`
	IsActive bool   `gorm:"not null;default:false" json:"is_active"`

	// ===== SAAS READY =====
	Plan               string     `gorm:"size:20;not null;default:free" json:"plan"`
	PlanExpiredAt      *time.Time `json:"plan_expired_at"`
	SubscriptionStatus string     `gorm:"size:20;default:inactive" json:"subscription_status"`
	LastPaymentRef     string     `gorm:"size:100" json:"last_payment_ref"`

	// ===== PROFILE =====
	FullName  string `gorm:"size:100" json:"full_name"`
	School    string `gorm:"size:100" json:"school"`
	Domicile  string `gorm:"size:100" json:"domicile"`
	AvatarURL string `gorm:"size:255" json:"avatar_url"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}


func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}
