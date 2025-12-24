package entities

import "time"

type UserAadharCard struct {
	ID             int        `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         string     `gorm:"type:uuid;not null;index" json:"user_id"`
	AadharNumber   string     `gorm:"type:varchar(20);not null" json:"aadhar_number"`
	AadharFrontKey string     `gorm:"type:text;not null" json:"aadhar_front"`
	AadharBackKey  string     `gorm:"type:text;not null" json:"aadhar_back"`
	IsDeleted      bool       `gorm:"default:false" json:"is_deleted"`
	CreatedBy      string     `gorm:"type:uuid" json:"created_by"`
	CreatedOn      time.Time  `gorm:"autoCreateTime" json:"created_on"`
	LastModifiedBy *string    `gorm:"type:uuid" json:"last_modified_by,omitempty"`
	LastModifiedOn *time.Time `json:"last_modified_on,omitempty"`
}

func (UserAadharCard) TableName() string {
	return "user_adhar_cards"
}