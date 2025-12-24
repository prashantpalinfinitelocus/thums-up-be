package entities

import "time"

type OptionMaster struct {
	ID               int        `gorm:"column:id;primaryKey;autoIncrement"`
	QuestionMasterID int        `gorm:"column:question_master_id;not null"`
	OptionText       string     `gorm:"column:option_text;type:text"`
	DisplayOrder     int        `gorm:"column:display_order;not null"`
	IsActive         bool       `gorm:"column:is_active;not null"`
	IsDeleted        bool       `gorm:"column:is_deleted;not null"`
	CreatedBy        string     `gorm:"column:created_by;not null"`
	CreatedOn        time.Time  `gorm:"column:created_on;not null"`
	LastModifiedBy   *string    `gorm:"column:last_modified_by"`
	LastModifiedOn   *time.Time `gorm:"column:last_modified_on"`
}

// TableName specifies the table name for OptionMaster
func (OptionMaster) TableName() string {
	return "option_master"
}

