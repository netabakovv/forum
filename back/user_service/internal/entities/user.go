package entities

import "time"

type User struct {
	ID        int64
	Username  string
	Password  string
	CreatedAt time.Time
}

type SafeUser struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

func (u *User) ToSafe() *SafeUser {
	return &SafeUser{
		ID:        u.ID,
		Username:  u.Username,
		CreatedAt: u.CreatedAt,
	}
}
