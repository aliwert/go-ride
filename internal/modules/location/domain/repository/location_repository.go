package repository

import (
	"context"

	"github.com/google/uuid"
)

type LocationRepository interface {
	UpdateLocation(ctx context.Context, driverID uuid.UUID, lat, lon float64) error
	FindNearbyDrivers(ctx context.Context, lat, lon float64, radiusInKm float64) ([]uuid.UUID, error)
}
