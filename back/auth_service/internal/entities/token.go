package entities

import "time"

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// AccessToken представляет токен доступа
type AccessToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Методы для проверки состояния access токена
func (at *AccessToken) IsExpired() bool {
	return time.Now().After(at.ExpiresAt)
}

func (at *AccessToken) IsValid() bool {
	return !at.IsExpired()
}

type TokenClaims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	IsAdmin   bool   `json:"is_admin"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	TokenType string `json:"typ"`
}

// RefreshToken представляет токен обновления в базе данных
type RefreshToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Revoked   bool      `json:"revoked"`
}

// Методы для проверки состояния refresh токена
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

func (rt *RefreshToken) IsValid() bool {
	return !rt.Revoked && !rt.IsExpired()
}
