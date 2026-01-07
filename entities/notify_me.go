package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotifyMe struct {
	ID          string     `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string     `gorm:"type:varchar(255);not null" json:"name"`
	PhoneNumber string     `gorm:"type:varchar(15);index;not null" json:"phone_number"`
	Email       *string    `gorm:"type:varchar(255)" json:"email,omitempty"`
	IsNotified  bool       `gorm:"default:false;index:idx_notify_me_is_notified" json:"is_notified"`
	NotifiedAt  *time.Time `json:"notified_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (nm *NotifyMe) BeforeCreate(tx *gorm.DB) error {
	if nm.ID == "" {
		nm.ID = uuid.New().String()
	}
	return nil
}

func (NotifyMe) TableName() string {
	return "notify_me"
}
