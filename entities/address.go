package entities

import "time"

type Address struct {
	ID              int        `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          string     `gorm:"type:uuid;not null;index" json:"user_id"`
	Address1        string     `gorm:"type:varchar(500);not null" json:"address1"`
	Address2        *string    `gorm:"type:varchar(500)" json:"address2,omitempty"`
	Pincode         int        `gorm:"not null" json:"pincode"`
	PinCodeID       int        `gorm:"not null" json:"pin_code_id"`
	CityID          int        `gorm:"not null" json:"city_id"`
	StateID         int        `gorm:"not null" json:"state_id"`
	NearestLandmark *string    `gorm:"type:varchar(255)" json:"nearest_landmark,omitempty"`
	ShippingMobile  *string    `gorm:"type:varchar(15)" json:"shipping_mobile,omitempty"`
	IsDefault       bool       `gorm:"default:false" json:"is_default"`
	IsActive        bool       `gorm:"default:true" json:"is_active"`
	IsDeleted       bool       `gorm:"default:false" json:"is_deleted"`
	CreatedBy       string     `gorm:"type:uuid" json:"created_by"`
	CreatedOn       time.Time  `gorm:"autoCreateTime" json:"created_on"`
	LastModifiedBy  *string    `gorm:"type:uuid" json:"last_modified_by,omitempty"`
	LastModifiedOn  *time.Time `json:"last_modified_on,omitempty"`
}

func (Address) TableName() string {
	return "address"
}
