package entities

type State struct {
	ID        int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string `gorm:"type:varchar(255);not null" json:"name"`
	IsActive  bool   `gorm:"default:true" json:"is_active"`
	IsDeleted bool   `gorm:"default:false" json:"is_deleted"`
}

func (State) TableName() string {
	return "state"
}

type City struct {
	ID        int    `gorm:"primaryKey;autoIncrement" json:"id"`
	StateID   int    `gorm:"not null;index" json:"state_id"`
	Name      string `gorm:"type:varchar(255);not null" json:"name"`
	IsActive  bool   `gorm:"default:true" json:"is_active"`
	IsDeleted bool   `gorm:"default:false" json:"is_deleted"`
}

func (City) TableName() string {
	return "city"
}

type PinCode struct {
	ID            int  `gorm:"primaryKey;autoIncrement" json:"id"`
	CityID        int  `gorm:"not null;index" json:"city_id"`
	Pincode       int  `gorm:"not null" json:"pincode"`
	IsDeliverable bool `gorm:"default:false" json:"is_deliverable"`
	IsActive      bool `gorm:"default:true" json:"is_active"`
	IsDeleted     bool `gorm:"default:false" json:"is_deleted"`
}

func (PinCode) TableName() string {
	return "pin_code"
}
