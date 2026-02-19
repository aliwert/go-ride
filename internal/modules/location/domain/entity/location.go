package entity

import (
	"time"

	"github.com/google/uuid"
)

type DriverLocation struct {
	DriverID  uuid.UUID
	Latitude  float64
	Longitude float64
	UpdatedAt time.Time
}
