package entities

import "time"

type ThunderSeat struct {
	ID             int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         string    `gorm:"type:uuid;not null;index" json:"user_id"`
	QuestionID     int       `gorm:"column:question_id;not null" json:"question_id"`
	WeekNumber     int       `gorm:"column:week_number;not null" json:"week_number"`
	Answer         string    `gorm:"column:answer;type:text" json:"answer"`
	CreatedBy      string    `gorm:"type:uuid;not null" json:"created_by"`
	CreatedOn      time.Time `gorm:"autoCreateTime" json:"created_on"`
}

func (ThunderSeat) TableName() string {
	return "thunder_seat"
}