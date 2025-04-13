package models

import "github.com/google/uuid"

// User model struct
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Role         string
}
