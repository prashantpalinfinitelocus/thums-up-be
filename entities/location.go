package entities

import (
	"time"
)

type State struct {
	ID             int        `json:"id" gorm:"primaryKey;autoIncrement"`
	CountryID      int        `json:"country_id" gorm:"type:int;not null"`
	ShortName      string     `json:"short_name" gorm:"type:text;not null"`
	Name           string     `json:"name" gorm:"type:text;not null"`
	IsActive       bool       `json:"is_active" gorm:"type:bool;not null"`
	IsDeleted      bool       `json:"is_deleted" gorm:"type:bool;not null"`
	CreatedBy      string     `json:"created_by" gorm:"type:text;not null"`
	CreatedOn      time.Time  `json:"created_on" gorm:"type:timestamp;not null"`
	LastModifiedBy *string    `json:"last_modified_by" gorm:"type:text"`
	LastModifiedOn *time.Time `json:"last_modified_on" gorm:"type:timestamp"`
}

type City struct {
	ID             int        `json:"id" gorm:"primaryKey;autoIncrement"`
	Name           string     `json:"name" gorm:"type:text;not null"`
	StateID        int        `json:"state_id" gorm:"type:int;not null"`
	IsActive       bool       `json:"is_active" gorm:"type:bool;not null"`
	IsDeleted      bool       `json:"is_deleted" gorm:"type:bool;not null"`
	CreatedBy      string     `json:"created_by" gorm:"type:text;not null"`
	CreatedOn      time.Time  `json:"created_on" gorm:"type:timestamp;not null"`
	LastModifiedBy *string    `json:"last_modified_by" gorm:"type:text"`
	LastModifiedOn *time.Time `json:"last_modified_on" gorm:"type:timestamp"`
}

type PinCode struct {
	ID              int        `json:"id" gorm:"primaryKey;autoIncrement"`
	Pincode         int        `json:"pincode" gorm:"type:int;not null"`
	IsBlackList     bool       `json:"is_black_list" gorm:"type:bool;not null"`
	CityID          int        `json:"city_id" gorm:"type:int;not null"`
	IsDeliverable   bool       `json:"is_deliverable" gorm:"type:bool;not null"`
	DeliveryCharges float64    `json:"delivery_charges" gorm:"type:float8;not null"`
	IsActive        bool       `json:"is_active" gorm:"type:bool;not null"`
	IsDeleted       bool       `json:"is_deleted" gorm:"type:bool;not null"`
	CreatedBy       string     `json:"created_by" gorm:"type:text;not null"`
	CreatedOn       time.Time  `json:"created_on" gorm:"type:timestamp;not null"`
	LastModifiedBy  *string    `json:"last_modified_by" gorm:"type:text"`
	LastModifiedOn  *time.Time `json:"last_modified_on" gorm:"type:timestamp"`
}
