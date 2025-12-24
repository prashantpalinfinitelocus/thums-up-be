package entities

import "time"

type OptionMasterLanguage struct {
	ID             int        `gorm:"column:id;primaryKey;autoIncrement"`
	OptionMasterID int        `gorm:"column:option_master_id;not null"`
	LanguageID     int        `gorm:"column:language_id;not null"`
	OptionText     string     `gorm:"column:option_text;type:text;not null"`
	IsActive       bool       `gorm:"column:is_active;not null"`
	IsDeleted      bool       `gorm:"column:is_deleted;not null"`
	CreatedBy      string     `gorm:"column:created_by;not null"`
	CreatedOn      time.Time  `gorm:"column:created_on;not null"`
	LastModifiedBy *string    `gorm:"column:last_modified_by"`
	LastModifiedOn *time.Time `gorm:"column:last_modified_on"`
}

// TableName specifies the table name for OptionMasterLanguage
func (OptionMasterLanguage) TableName() string {
	return "option_master_languages"
}

