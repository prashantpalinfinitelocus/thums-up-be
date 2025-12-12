package entities

import (
	"time"
)

type LanguageKey struct {
	ID             int                `gorm:"primaryKey;column:id"`
	Name           string             `gorm:"column:name;not null"`
	KeyTypeID      int                `gorm:"column:key_type_id;not null"`
	IsActive       bool               `gorm:"column:is_active;not null"`
	IsDeleted      bool               `gorm:"column:is_deleted;not null"`
	CreatedBy      string             `gorm:"column:created_by;not null"`
	CreatedOn      time.Time          `gorm:"column:created_on;not null"`
	LastModifiedBy *string            `gorm:"column:last_modified_by"`
	LastModifiedOn *time.Time         `gorm:"column:last_modified_on"`
	KeyValues      []LanguageKeyValue `gorm:"foreignKey:LanguageKeyID"`
}

type LanguageKeyValue struct {
	ID               int            `gorm:"primaryKey;column:id"`
	LanguageKeyID    int            `gorm:"column:language_key_id;not null"`
	LanguageMasterID int            `gorm:"column:language_master_id;not null"`
	KeyValue         string         `gorm:"column:key_value;not null"`
	IsActive         bool           `gorm:"column:is_active;not null"`
	IsDeleted        bool           `gorm:"column:is_deleted;not null"`
	CreatedBy        string         `gorm:"column:created_by;not null"`
	CreatedOn        time.Time      `gorm:"column:created_on;not null"`
	LastModifiedBy   *string        `gorm:"column:last_modified_by"`
	LastModifiedOn   *time.Time     `gorm:"column:last_modified_on"`
	LanguageKey      LanguageKey    `gorm:"foreignKey:LanguageKeyID"`
	LanguageMaster   LanguageMaster `gorm:"foreignKey:LanguageMasterID"`
}

type LanguageMaster struct {
	ID             int                `gorm:"primaryKey;column:id"`
	Name           string             `gorm:"column:name;not null"`
	LanguageCode   *string            `gorm:"column:language_code"`
	PlatformID     int                `gorm:"column:platform_id;not null"`
	IsActive       bool               `gorm:"column:is_active;not null"`
	IsDeleted      bool               `gorm:"column:is_deleted;not null"`
	CreatedBy      string             `gorm:"column:created_by;not null"`
	CreatedOn      time.Time          `gorm:"column:created_on;not null"`
	LastModifiedBy *string            `gorm:"column:last_modified_by"`
	LastModifiedOn *time.Time         `gorm:"column:last_modified_on"`
	KeyValues      []LanguageKeyValue `gorm:"foreignKey:LanguageMasterID"`
}

type Language struct {
	ID             int        `gorm:"primaryKey;column:id"`
	Name           string     `gorm:"column:name;not null"`
	LanguageKey    string     `gorm:"column:language_key;not null"`
	IsActive       bool       `gorm:"column:is_active;not null"`
	IsDeleted      bool       `gorm:"column:is_deleted;not null"`
	CreatedBy      string     `gorm:"column:created_by;not null"`
	CreatedOn      time.Time  `gorm:"column:created_on;not null"`
	LastModifiedBy *string    `gorm:"column:last_modified_by"`
	LastModifiedOn *time.Time `gorm:"column:last_modified_on"`
}

func (LanguageKey) TableName() string {
	return "language_key"
}

func (LanguageKeyValue) TableName() string {
	return "language_key_values"
}

func (LanguageMaster) TableName() string {
	return "language_master"
}

func (Language) TableName() string {
	return "languages"
}
