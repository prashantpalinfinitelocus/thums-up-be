package entities

import "time"

type UserQuestionAnswer struct {
	ID               int        `gorm:"column:id;primaryKey;autoIncrement"`
	UserID           string     `gorm:"column:user_id;not null"`
	QuestionMasterID int        `gorm:"column:question_master_id;not null"`
	OptionID         int        `gorm:"column:option_id;not null"`
	SelectedAnswer   bool       `gorm:"column:selected_answer;not null"`
	IsActive         bool       `gorm:"column:is_active;not null"`
	IsDeleted        bool       `gorm:"column:is_deleted;not null"`
	CreatedBy        string     `gorm:"column:created_by;not null"`
	CreatedOn        time.Time  `gorm:"column:created_on;not null"`
	LastModifiedBy   *string    `gorm:"column:last_modified_by"`
	LastModifiedOn   *time.Time `gorm:"column:last_modified_on"`
}

// TableName specifies the table name for UserQuestionAnswer
func (UserQuestionAnswer) TableName() string {
	return "user_question_answer"
}
