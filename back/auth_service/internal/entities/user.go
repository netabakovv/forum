package entities

import (
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Изменено с Password на PasswordHash для безопасности
	CreatedAt    time.Time `json:"created_at"`
	IsAdmin      bool      `json:"is_admin"`   // Добавлено для контроля прав
	UpdatedAt    time.Time `json:"updated_at"` // Добавлено для аудита
}

type SafeUser struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	IsAdmin   bool      `json:"is_admin"`
}
