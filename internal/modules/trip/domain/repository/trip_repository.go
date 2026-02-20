package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/aliwert/go-ride/internal/modules/trip/domain/entity"
)

type TripRepository interface {
	Create(ctx context.Context, trip *entity.Trip) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Trip, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TripStatus) error
	// AssignDriver atomically sets the driver and transitions status to ACCEPTED
	// must only succeed when the current status is REQUESTED to prevent race conditions
	AssignDriver(ctx context.Context, tripID, driverID uuid.UUID) error
}
