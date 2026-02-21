package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/aliwert/go-ride/internal/modules/matching/application/port"
)

var (
	ErrNoDriversAvailable = errors.New("no drivers available nearby")
)

type MatchingUseCase struct {
	locationPort     port.LocationPort
	notificationPort port.NotificationPort
}

func NewMatchingUseCase(loc port.LocationPort, notif port.NotificationPort) *MatchingUseCase {
	return &MatchingUseCase{
		locationPort:     loc,
		notificationPort: notif,
	}
}

const defaultSearchRadiusKm = 5.0

func (uc *MatchingUseCase) ProcessTrip(ctx context.Context, tripID uuid.UUID, lat, lon float64) error {
	drivers, err := uc.locationPort.FindNearbyDrivers(ctx, lat, lon, defaultSearchRadiusKm)
	if err != nil {
		return fmt.Errorf("matching: find nearby drivers: %w", err)
	}

	if len(drivers) == 0 {
		return ErrNoDriversAvailable
	}

	// fan-out notifications to every nearby driver; log failures but don't abort
	// the whole batch, a single unreachable device shouldn't block others
	for _, driverID := range drivers {
		if err := uc.notificationPort.NotifyDriver(ctx, driverID, tripID); err != nil {
			fmt.Printf("WARN: failed to notify driver %s for trip %s: %v\n", driverID, tripID, err)
		}
	}

	return nil
}
