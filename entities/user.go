package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           string     `gorm:"type:uuid;primaryKey" json:"id"`
	PhoneNumber  string     `gorm:"type:varchar(15);uniqueIndex;not null" json:"phone_number"`
	Name         *string    `gorm:"type:varchar(255)" json:"name,omitempty"`
	Email        *string    `gorm:"type:varchar(255);uniqueIndex" json:"email,omitempty"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	IsVerified   bool       `gorm:"default:false" json:"is_verified"`
	ReferralCode *string    `gorm:"type:varchar(20);uniqueIndex" json:"referral_code,omitempty"`
	ReferredBy   *string    `gorm:"type:varchar(20)" json:"referred_by,omitempty"`
	DeviceToken  *string    `gorm:"type:text" json:"device_token,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

func (User) TableName() string {
	return "users"
}

