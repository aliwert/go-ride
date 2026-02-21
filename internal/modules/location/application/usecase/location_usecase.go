package usecase

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"

	"github.com/aliwert/go-ride/internal/modules/location/application/dto"
	"github.com/aliwert/go-ride/internal/modules/location/domain/repository"
	trackingport "github.com/aliwert/go-ride/internal/modules/tracking/application/port"
)

var (
	ErrInvalidCoordinates = errors.New("invalid coordinates")
	ErrInvalidRadius      = errors.New("radius must be greater than zero")
)

type LocationUseCase struct {
	locationRepo repository.LocationRepository
	broadcaster  trackingport.BroadcasterPort
}

func NewLocationUseCase(repo repository.LocationRepository, broadcaster trackingport.BroadcasterPort) *LocationUseCase {
	return &LocationUseCase{locationRepo: repo, broadcaster: broadcaster}
}

func (uc *LocationUseCase) UpdateLocation(ctx context.Context, req *dto.UpdateLocationRequest) error {
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		return ErrInvalidCoordinates
	}

	if err := uc.locationRepo.UpdateLocation(ctx, req.DriverID, req.Latitude, req.Longitude); err != nil {
		return err
	}

	// if the driver is on an active trip, push the location to connected riders
	if req.TripID != nil && uc.broadcaster != nil {
		go func() {
			if err := uc.broadcaster.BroadcastLocation(ctx, *req.TripID, req.Latitude, req.Longitude); err != nil {
				log.Printf("WARN: broadcast location failed for trip %s: %v", req.TripID, err)
			}
		}()
	}

	return nil
}

func (uc *LocationUseCase) FindNearbyDrivers(ctx context.Context, req *dto.FindNearbyRequest) ([]uuid.UUID, error) {
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		return nil, ErrInvalidCoordinates
	}
	if req.RadiusKm <= 0 {
		return nil, ErrInvalidRadius
	}

	return uc.locationRepo.FindNearbyDrivers(ctx, req.Latitude, req.Longitude, req.RadiusKm)
}
