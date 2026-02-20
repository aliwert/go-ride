package entity

import (
	"time"

	"github.com/google/uuid"
)

type TripStatus string

const (
	TripStatusRequested  TripStatus = "REQUESTED"
	TripStatusAccepted   TripStatus = "ACCEPTED"
	TripStatusInProgress TripStatus = "IN_PROGRESS"
	TripStatusCompleted  TripStatus = "COMPLETED"
	TripStatusCancelled  TripStatus = "CANCELLED"
)

func (s TripStatus) IsValid() bool {
	switch s {
	case TripStatusRequested, TripStatusAccepted, TripStatusInProgress, TripStatusCompleted, TripStatusCancelled:
		return true
	}
	return false
}

type Trip struct {
	ID         uuid.UUID
	RiderID    uuid.UUID
	DriverID   *uuid.UUID
	PickupLat  float64
	PickupLon  float64
	DropoffLat float64
	DropoffLon float64
	Status     TripStatus
	Fare       *float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewTrip(riderID uuid.UUID, pickupLat, pickupLon, dropoffLat, dropoffLon float64) *Trip {
	now := time.Now()
	return &Trip{
		ID:         uuid.New(),
		RiderID:    riderID,
		PickupLat:  pickupLat,
		PickupLon:  pickupLon,
		DropoffLat: dropoffLat,
		DropoffLon: dropoffLon,
		Status:     TripStatusRequested,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
