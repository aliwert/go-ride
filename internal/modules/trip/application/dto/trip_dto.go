package dto

import "time"

type CreateTripRequest struct {
	RiderID    string  `json:"rider_id"`
	PickupLat  float64 `json:"pickup_lat"`
	PickupLon  float64 `json:"pickup_lon"`
	DropoffLat float64 `json:"dropoff_lat"`
	DropoffLon float64 `json:"dropoff_lon"`
}

type AcceptTripRequest struct {
	TripID   string `json:"trip_id" params:"id"`
	DriverID string `json:"driver_id"`
}

type CompleteTripRequest struct {
	TripID string `json:"trip_id" params:"id"`
}

type TripResponse struct {
	ID         string    `json:"id"`
	RiderID    string    `json:"rider_id"`
	DriverID   *string   `json:"driver_id,omitempty"`
	PickupLat  float64   `json:"pickup_lat"`
	PickupLon  float64   `json:"pickup_lon"`
	DropoffLat float64   `json:"dropoff_lat"`
	DropoffLon float64   `json:"dropoff_lon"`
	Status     string    `json:"status"`
	Fare       *float64  `json:"fare,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
