package entities

import "time"

type QuestionMasterLanguage struct {
	ID               int        `gorm:"column:id;primaryKey;autoIncrement"`
	QuestionMasterID int        `gorm:"column:question_master_id;not null"`
	LanguageID       int        `gorm:"column:language_id;not null"`
	QuestionText     string     `gorm:"column:question_text;type:text;not null"`
	IsActive         bool       `gorm:"column:is_active;not null"`
	IsDeleted        bool       `gorm:"column:is_deleted;not null"`
	CreatedBy        string     `gorm:"column:created_by;not null"`
	CreatedOn        time.Time  `gorm:"column:created_on;not null"`
	LastModifiedBy   *string    `gorm:"column:last_modified_by"`
	LastModifiedOn   *time.Time `gorm:"column:last_modified_on"`
}

// TableName specifies the table name for QuestionMasterLanguage
func (QuestionMasterLanguage) TableName() string {
	return "question_master_languages"
}
