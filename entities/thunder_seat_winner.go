package entities

import "time"

type ThunderSeatWinner struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        string    `gorm:"type:uuid;not null;index" json:"user_id"`
	ThunderSeatID int       `gorm:"column:thunder_seat_id;not null" json:"thunder_seat_id"`
	QRCode        string    `gorm:"column:qr_code;not null" json:"qr_code"`
	WeekNumber    int       `gorm:"column:week_number;not null" json:"week_number"`
	HasViewed     bool      `gorm:"column:has_viewed;default:false" json:"has_viewed"`
	CreatedBy     string    `gorm:"type:uuid;not null" json:"created_by"`
	CreatedOn     time.Time `gorm:"autoCreateTime" json:"created_on"`
	User          User      `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

func (ThunderSeatWinner) TableName() string {
	return "thunder_seat_winner"
}
