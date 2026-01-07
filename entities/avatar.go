package entities

import "time"

type Avatar struct {
	ID             int        `json:"id" gorm:"primaryKey;autoIncrement"`
	Name           string     `json:"name" gorm:"type:text;not null"`
	ImageKey       string     `json:"image_key" gorm:"type:text;not null"`
	IsPublished    bool       `json:"is_published" gorm:"type:boolean;not null;default:false"`
	PublishedBy    *string    `json:"published_by" gorm:"type:text"`
	PublishedOn    *time.Time `json:"published_on" gorm:"type:timestamp"`
	IsActive       bool       `json:"is_active" gorm:"type:boolean;not null;default:true"`
	IsDeleted      bool       `json:"is_deleted" gorm:"type:boolean;not null;default:false"`
	CreatedBy      string     `json:"created_by" gorm:"type:text;not null"`
	CreatedOn      time.Time  `json:"created_on" gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	LastModifiedBy *string    `json:"last_modified_by" gorm:"type:text"`
	LastModifiedOn *time.Time `json:"last_modified_on" gorm:"type:timestamp"`
}

func (Avatar) TableName() string {
	return "avatar"
}
