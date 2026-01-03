package entities

import "time"

type ThunderSeat struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     string    `gorm:"type:uuid;not null;index" json:"user_id"`
	WeekNumber int       `gorm:"column:week_number;not null" json:"week_number"`
	Answer     string    `gorm:"column:answer;type:text" json:"answer"`
	MediaURL   *string   `gorm:"column:media_url;type:text" json:"media_url,omitempty"`
	MediaKey   *string   `gorm:"column:media_key;type:text" json:"media_key,omitempty"`
	MediaType  *string   `gorm:"column:media_type;type:varchar(50)" json:"media_type,omitempty"`
	CreatedBy  string    `gorm:"type:uuid;not null" json:"created_by"`
	CreatedOn  time.Time `gorm:"autoCreateTime" json:"created_on"`
}

func (ThunderSeat) TableName() string {
	return "thunder_seat"
}
