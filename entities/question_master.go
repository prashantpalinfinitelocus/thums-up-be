package entities

import "time"

type QuestionMaster struct {
	ID             int        `gorm:"column:id;primaryKey;autoIncrement"`
	QuestionText   string     `gorm:"column:question_text;type:text"`
	QuesPoint      int        `gorm:"column:ques_point;not null"`
	LanguageID     int        `gorm:"column:language_id;not null"`
	IsActive       bool       `gorm:"column:is_active;not null"`
	IsDeleted      bool       `gorm:"column:is_deleted;not null"`
	ProfileOnly    bool       `gorm:"column:profile_only"`
	CreatedBy      string     `gorm:"column:created_by;not null"`
	CreatedOn      time.Time  `gorm:"column:created_on;not null"`
	LastModifiedBy *string    `gorm:"column:last_modified_by"`
	LastModifiedOn *time.Time `gorm:"column:last_modified_on"`
}

// TableName specifies the table name for QuestionMaster
func (QuestionMaster) TableName() string {
	return "question_master"
}
