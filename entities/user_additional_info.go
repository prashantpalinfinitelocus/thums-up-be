package entities

import "time"

type UserAdditionalInfo struct {
	ID             int        `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         string     `gorm:"type:uuid;not null;index" json:"user_id"`
	City1          string     `gorm:"type:text" json:"city1"`
	City2          string     `gorm:"type:text" json:"city2"`
	City3          string     `gorm:"type:text" json:"city3"`
	IsDeleted      bool       `gorm:"default:false" json:"is_deleted"`
	CreatedBy      string     `gorm:"type:uuid" json:"created_by"`
	CreatedOn      time.Time  `gorm:"autoCreateTime" json:"created_on"`
	LastModifiedBy *string    `gorm:"type:uuid" json:"last_modified_by,omitempty"`
	LastModifiedOn *time.Time `json:"last_modified_on,omitempty"`
}

func (UserAdditionalInfo) TableName() string {
	return "user_additional_info"
}

