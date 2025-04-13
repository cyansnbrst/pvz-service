package models

import (
	"time"

	"github.com/google/uuid"
)

// Reception model struct
type Reception struct {
	ID       uuid.UUID
	DateTime time.Time
	PvzID    uuid.UUID
	Status   string
}

// Reception with products struct
type ReceptionWithProducts struct {
	Reception Reception
	Products  []*Product
}
