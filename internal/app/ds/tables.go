package ds

import "time"

type Expedition struct {
	ID          uint `gorm:"primarykey;autoIncrement"`
	Name        string
	Year        int
	Description string
	Status      string
	CreatedAt   time.Time `json:"start_date"`
	FormedAt    time.Time `json:"start_date"`
	ClosedAt    time.Time `json:"start_date"`
	UserID      uint
	Alpinists   []Alpinist `gorm:"many2many:alpinist_expedition"`
}

type User struct {
	ID          uint `gorm:"primarykey;autoIncrement"`
	Login       string
	Password    string
	ImageRef    string
	Expeditions []Expedition
}

type Alpinist struct {
	ID          uint `gorm:"primarykey;autoIncrement"`
	Name        string
	Lifetime    string
	Country     string
	ImageRef    string
	BigImageRef string
	Description string
	Status      string
	Expeditions []Expedition `gorm:"many2many:alpinist_expedition"`
}
