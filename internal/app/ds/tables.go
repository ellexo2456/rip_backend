package ds

import "time"

type Expedition struct {
	ID          uint   `gorm:"primarykey;autoIncrement"`
	Name        string `gorm:"type:varchar(90)"`
	Year        int
	Status      string    `gorm:"type:varchar(90)"`
	CreatedAt   time.Time `json:"start_date"`
	FormedAt    time.Time `json:"start_date"`
	ClosedAt    time.Time `json:"start_date"`
	UserID      uint
	ModeratorID uint
	Alpinists   []Alpinist `gorm:"many2many:alpinist_expedition"`
}

type User struct {
	ID          uint         `gorm:"primarykey;autoIncrement"`
	Login       string       `gorm:"type:varchar(90); unique"`
	Password    string       `gorm:"type:varchar(90)"`
	ImageRef    string       `gorm:"type:varchar(90)"`
	Expeditions []Expedition `gorm:"foreignkey:UserID;foreignkey:ModeratorID;"`
}

type Alpinist struct {
	ID          uint   `gorm:"primarykey;autoIncrement"`
	Name        string `gorm:"type:varchar(90)"`
	Lifetime    string `gorm:"type:varchar(90)"`
	Country     string `gorm:"type:varchar(90)"`
	ImageRef    string `gorm:"type:varchar(90)"`
	BigImageRef string `gorm:"type:varchar(90)"`
	Description string
	Status      string       `gorm:"type:varchar(90)"`
	Expeditions []Expedition `gorm:"many2many:alpinist_expedition"`
}
