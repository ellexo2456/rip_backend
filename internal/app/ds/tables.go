package ds

import "time"

type Alpinist struct {
	Id          uint `gorm:"primarykey"`
	Name        string
	Age         int
	Country     string
	ImageRef    string
	Description string
	Status      string
}

type User struct {
	Id       uint `gorm:"primarykey"`
	Login    string
	Password string
	ImageRef string
}

type Expedition struct {
	Id        uint `gorm:"primarykey"`
	Name      string
	Year      int
	Status    string
	CreatedAt time.Time `json:"start_date"`
	FormedAt  time.Time `json:"start_date"`
	ClosedAt  time.Time `json:"start_date"`
	moderator User
}

type AlpinistExpedition struct {
	Id         uint `gorm:"primarykey"`
	Alpinist   Alpinist
	Expedition Expedition
}
