package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoginCount struct {
	ID          string    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      string    `gorm:"type:uuid;index;not null" json:"user_id"`
	PhoneNumber string    `gorm:"type:varchar(15);index;not null" json:"phone_number"`
	Count       int       `gorm:"default:0" json:"count"`
	LastLogin   time.Time `gorm:"not null" json:"last_login"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (l *LoginCount) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}

func (LoginCount) TableName() string {
	return "login_counts"
}
