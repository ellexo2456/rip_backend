package ds

import "time"

const (
	UserID      = 1
	ModeratorID = 2
)

type Expedition struct {
	ID            uint       `gorm:"primarykey;autoIncrement" json:"id"`
	Name          string     `gorm:"type:varchar(90)" json:"name"`
	Year          int        `json:"year"`
	Status        string     `gorm:"type:varchar(90)" json:"status"`
	CreatedAt     time.Time  `json:"-"`
	FormedAt      time.Time  `json:"-"`
	ClosedAt      time.Time  `json:"-"`
	UserID        uint       `json:"userId"`
	Usr           User       `gorm:"foreignkey:UserID" json:"-"`
	ModeratorUser User       `gorm:"foreignkey:ModeratorID;" json:"-"`
	ModeratorID   uint       `json:"moderatorId"`
	Alpinists     []Alpinist `gorm:"many2many:alpinist_expedition" json:"alpinists"`
}

type User struct {
	ID       uint   `gorm:"primarykey;autoIncrement"`
	Login    string `gorm:"type:varchar(90); unique"`
	Password string `gorm:"type:varchar(90)"`
	ImageRef string `gorm:"type:varchar(90)"`
	//Expeditions []Expedition `gorm:"foreignkey:UserID;foreignkey:ModeratorID;"`
}

type Alpinist struct {
	ID          uint         `gorm:"primarykey;autoIncrement" json:"id"`
	Name        string       `gorm:"type:varchar(90)" json:"name"`
	Lifetime    string       `gorm:"type:varchar(90)" json:"lifetime"`
	Country     string       `gorm:"type:varchar(90)" json:"country"`
	ImageRef    string       `gorm:"type:varchar(90)" json:"imageRef"`
	BigImageRef string       `gorm:"type:varchar(90)" json:"bigImageRef"`
	Description string       `json:"description"`
	Status      string       `gorm:"type:varchar(90)"`
	Expeditions []Expedition `gorm:"many2many:alpinist_expedition"`
}
