package dto

import "github.com/google/uuid"

type UpdateLocationRequest struct {
	DriverID  uuid.UUID `json:"driver_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
}

type FindNearbyRequest struct {
	Latitude  float64 `json:"latitude" query:"latitude"`
	Longitude float64 `json:"longitude" query:"longitude"`
	RadiusKm  float64 `json:"radius_km" query:"radius_km"`
}

type NearbyDriversResponse struct {
	Drivers []string `json:"drivers"`
}
