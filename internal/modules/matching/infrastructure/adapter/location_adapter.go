package adapter

import (
	"context"

	"github.com/google/uuid"

	locationdto "github.com/aliwert/go-ride/internal/modules/location/application/dto"
	locationuc "github.com/aliwert/go-ride/internal/modules/location/application/usecase"
)

// bridges the matching module to the location module
// without leaking any internal types across module boundaries
type LocationAdapter struct {
	locUC *locationuc.LocationUseCase
}

func NewLocationAdapter(locUC *locationuc.LocationUseCase) *LocationAdapter {
	return &LocationAdapter{locUC: locUC}
}

func (a *LocationAdapter) FindNearbyDrivers(ctx context.Context, lat, lon float64, radiusKm float64) ([]uuid.UUID, error) {
	req := &locationdto.FindNearbyRequest{
		Latitude:  lat,
		Longitude: lon,
		RadiusKm:  radiusKm,
	}
	return a.locUC.FindNearbyDrivers(ctx, req)
}
