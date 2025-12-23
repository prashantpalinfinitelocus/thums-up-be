package entities

import "time"

type ContestWeek struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`
	WeekNumber  int       `gorm:"column:week_number;not null;unique" json:"week_number"`
	StartDate   time.Time `gorm:"column:start_date;not null" json:"start_date"`
	EndDate     time.Time `gorm:"column:end_date;not null" json:"end_date"`
	WinnerCount int       `gorm:"column:winner_count;not null" json:"winner_count"`
	IsActive    bool      `gorm:"column:is_active;default:false" json:"is_active"`
	CreatedBy   string    `gorm:"type:varchar(255);not null" json:"created_by"`
	CreatedOn   time.Time `gorm:"autoCreateTime" json:"created_on"`
	UpdatedBy   string    `gorm:"type:varchar(255)" json:"updated_by"`
	UpdatedOn   time.Time `gorm:"autoUpdateTime" json:"updated_on"`
}

func (ContestWeek) TableName() string {
	return "contest_week"
}
