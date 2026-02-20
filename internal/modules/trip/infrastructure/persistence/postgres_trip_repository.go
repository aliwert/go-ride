package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aliwert/go-ride/internal/modules/trip/domain/entity"
)

var (
	ErrTripNotFound        = errors.New("trip not found")
	ErrTripAlreadyAccepted = errors.New("trip has already been accepted by another driver")
)

type PostgresTripRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTripRepository(pool *pgxpool.Pool) *PostgresTripRepository {
	return &PostgresTripRepository{pool: pool}
}

func (r *PostgresTripRepository) Create(ctx context.Context, trip *entity.Trip) error {
	query := `
		INSERT INTO trips (id, rider_id, pickup_lat, pickup_lon, dropoff_lat, dropoff_lon, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pool.Exec(ctx, query,
		trip.ID,
		trip.RiderID,
		trip.PickupLat,
		trip.PickupLon,
		trip.DropoffLat,
		trip.DropoffLon,
		string(trip.Status),
		trip.CreatedAt,
		trip.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres: create trip: %w", err)
	}

	return nil
}

func (r *PostgresTripRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Trip, error) {
	query := `
		SELECT id, rider_id, driver_id, pickup_lat, pickup_lon, dropoff_lat, dropoff_lon, status, fare, created_at, updated_at
		FROM trips WHERE id = $1`

	return r.scanTrip(ctx, query, id)
}

func (r *PostgresTripRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TripStatus) error {
	query := `UPDATE trips SET status = $1 WHERE id = $2`

	tag, err := r.pool.Exec(ctx, query, string(status), id)
	if err != nil {
		return fmt.Errorf("postgres: update trip status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTripNotFound
	}

	return nil
}

// AssignDriver atomically assigns a driver and moves the trip to ACCEPTED.
// the WHERE clause enforces that status must be REQUESTED, if another driver
// already accepted (status changed), RowsAffected will be 0 and we surface
// ErrTripAlreadyAccepted. this is optimistic concurrency at the row level.
func (r *PostgresTripRepository) AssignDriver(ctx context.Context, tripID, driverID uuid.UUID) error {
	query := `
		UPDATE trips
		SET driver_id = $1, status = 'ACCEPTED'
		WHERE id = $2 AND status = 'REQUESTED'`

	tag, err := r.pool.Exec(ctx, query, driverID, tripID)
	if err != nil {
		return fmt.Errorf("postgres: assign driver: %w", err)
	}

	if tag.RowsAffected() == 0 {
		// distinguish between "trip doesn't exist" and "already accepted"
		exists, _ := r.exists(ctx, tripID)
		if !exists {
			return ErrTripNotFound
		}
		return ErrTripAlreadyAccepted
	}

	return nil
}

// exists is a lightweight check so AssignDriver can differentiate not-found from race-lost
func (r *PostgresTripRepository) exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var found bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM trips WHERE id = $1)`, id).Scan(&found)
	return found, err
}

// scanTrip centralises row-to-entity mapping so every finder stays DRY
func (r *PostgresTripRepository) scanTrip(ctx context.Context, query string, args ...any) (*entity.Trip, error) {
	var t entity.Trip
	var status string

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&t.ID,
		&t.RiderID,
		&t.DriverID,
		&t.PickupLat,
		&t.PickupLon,
		&t.DropoffLat,
		&t.DropoffLon,
		&status,
		&t.Fare,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTripNotFound
		}
		return nil, fmt.Errorf("postgres: scan trip: %w", err)
	}

	t.Status = entity.TripStatus(status)
	return &t, nil
}
