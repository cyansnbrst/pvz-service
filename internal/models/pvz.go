package models

import (
	"time"

	"github.com/google/uuid"
)

// PVZ model struct
type PVZ struct {
	ID               uuid.UUID
	City             string
	RegistrationDate time.Time
}

// PVZ with receptions struct
type PVZWithReceptions struct {
	PVZ        PVZ
	Receptions []*ReceptionWithProducts
}
