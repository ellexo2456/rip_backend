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
	CreatedAt     time.Time  `json:"createdAt"`
	FormedAt      time.Time  `json:"formedAt"`
	ClosedAt      time.Time  `json:"closedAt"`
	UserID        uint       `json:"-"`
	Usr           *User      `gorm:"foreignkey:UserID" json:"user,omitempty"`
	ModeratorUser *User      `gorm:"foreignkey:ModeratorID;" json:"moderator,omitempty"`
	ModeratorID   uint       `json:"-"`
	Alpinists     []Alpinist `gorm:"many2many:alpinist_expedition" json:"alpinists,omitempty"`
}

type User struct {
	ID       uint   `gorm:"primarykey;autoIncrement"`
	Email    string `gorm:"type:varchar(90); unique"`
	Password []byte `gorm:"type:bytea" json:"password,omitempty"`
	ImageRef string `gorm:"type:varchar(90)" json:"imageRef,omitempty"Z`
	Role     Role   `gorm:"type:int;" json:"role,omitempty"`
	//Expeditions []Expedition `gorm:"foreignkey:UserID;foreignkey:ModeratorID;"`
}

type Alpinist struct {
	ID          uint         `gorm:"primarykey;autoIncrement" json:"id"`
	Name        string       `gorm:"type:varchar(90)" json:"name"`
	Lifetime    string       `gorm:"type:varchar(90)" json:"lifetime"`
	Country     string       `gorm:"type:varchar(90)" json:"country"`
	ImageRef    string       `gorm:"type:varchar(90)" json:"imageRef"`
	ImageName   string       `gorm:"type:varchar(90)" json:"-"`
	Description string       `json:"description"`
	Status      string       `gorm:"type:varchar(90)"`
	Expeditions []Expedition `gorm:"many2many:alpinist_expedition"`
}
