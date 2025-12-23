package domain

import (
	"errors"
	"time"
)

var (
	ErrURLNotFound     = errors.New("url not found")
	ErrInvalidURL      = errors.New("invalid url")
	ErrShortCodeExists = errors.New("short code already exists")
)

type URL struct {
	ID        string
	ShortCode string
	LongURL   string
	CreatedAt time.Time
	ExpiresAt *time.Time
}

func (u *URL) IsExpired() bool {
	if u.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.ExpiresAt)
}

func (u *URL) Validate() error {
	if u.LongURL == "" {
		return ErrInvalidURL
	}
	if u.ShortCode == "" {
		return ErrInvalidURL
	}
	return nil
}
