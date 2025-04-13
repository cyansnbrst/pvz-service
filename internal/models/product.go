package models

import (
	"time"

	"github.com/google/uuid"
)

// Product model struct
type Product struct {
	ID          uuid.UUID
	Type        string
	DateTime    time.Time
	ReceptionID uuid.UUID
}
