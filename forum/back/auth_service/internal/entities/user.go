package entities

import (
	"errors"
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

// ToSafe конвертирует User в SafeUser, скрывая чувствительные данные
func (u *User) ToSafe() *SafeUser {
	if u == nil {
		return nil
	}
	return &SafeUser{
		ID:        u.ID,
		Username:  u.Username,
		CreatedAt: u.CreatedAt,
		IsAdmin:   u.IsAdmin,
	}
}

// Validate проверяет корректность данных пользователя
func (u *User) Validate() error {
	if u.Username == "" {
		return errors.New("username cannot be empty")
	}
	if len(u.Username) < 3 || len(u.Username) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}
	if len(u.PasswordHash) == 0 {
		return errors.New("password hash cannot be empty")
	}
	return nil
}

// BeforeCreate подготавливает пользователя к созданию
func (u *User) BeforeCreate() {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
}
