package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/aliwert/go-ride/internal/modules/location/application/dto"
	"github.com/aliwert/go-ride/internal/modules/location/domain/repository"
)

var (
	ErrInvalidCoordinates = errors.New("invalid coordinates")
	ErrInvalidRadius      = errors.New("radius must be greater than zero")
)

type LocationUseCase struct {
	locationRepo repository.LocationRepository
}

func NewLocationUseCase(repo repository.LocationRepository) *LocationUseCase {
	return &LocationUseCase{locationRepo: repo}
}

func (uc *LocationUseCase) UpdateLocation(ctx context.Context, req *dto.UpdateLocationRequest) error {
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		return ErrInvalidCoordinates
	}

	return uc.locationRepo.UpdateLocation(ctx, req.DriverID, req.Latitude, req.Longitude)
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
