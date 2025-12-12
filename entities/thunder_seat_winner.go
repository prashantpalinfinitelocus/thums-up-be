package entities

import "time"

type ThunderSeatWinner struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        string    `gorm:"type:uuid;not null;index" json:"user_id"`
	ThunderSeatID int       `gorm:"column:thunder_seat_id;not null" json:"thunder_seat_id"`
	WeekNumber    int       `gorm:"column:week_number;not null" json:"week_number"`
	CreatedBy     string    `gorm:"type:uuid;not null" json:"created_by"`
	CreatedOn     time.Time `gorm:"autoCreateTime" json:"created_on"`
}

func (ThunderSeatWinner) TableName() string {
	return "thunder_seat_winner"
}
