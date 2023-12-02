package ds

import "time"

type Role int

const (
	Guest     Role = iota // 0
	Usr                   // 1
	Moderator             // 2
)

type Session struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	UserID    int       `json:"-"`
	Role      Role      `json:"-"`
}

type Credentials struct {
	Password []byte `json:"password"`
	Email    string `json:"email"`
}

type SessionContext struct {
	UserID int
	Role   Role
}
