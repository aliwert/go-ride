package port

import (
	"context"

	"github.com/google/uuid"
)

// abstracts the location module so matching never depends on its internals
type LocationPort interface {
	FindNearbyDrivers(ctx context.Context, lat, lon float64, radiusKm float64) ([]uuid.UUID, error)
}
