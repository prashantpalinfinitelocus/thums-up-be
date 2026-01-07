package entities

import (
	"time"
)

type OTPLog struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	PhoneNumber string     `gorm:"type:varchar(15);index;not null" json:"phone_number"`
	OTP         string     `gorm:"type:varchar(6);not null" json:"otp"`
	ExpiresAt   time.Time  `gorm:"not null" json:"expires_at"`
	IsVerified  bool       `gorm:"default:false" json:"is_verified"`
	VerifiedAt  *time.Time `json:"verified_at,omitempty"`
	Attempts    int        `gorm:"default:0" json:"attempts"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (OTPLog) TableName() string {
	return "otp_logs"
}
