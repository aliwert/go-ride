package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/aliwert/go-ride/internal/modules/trip/application/dto"
	"github.com/aliwert/go-ride/internal/modules/trip/domain/entity"
	"github.com/aliwert/go-ride/internal/modules/trip/domain/repository"
)

var (
	ErrTripNotFound        = errors.New("trip not found")
	ErrTripAlreadyAccepted = errors.New("trip has already been accepted by another driver")
	ErrInvalidTripStatus   = errors.New("invalid trip status for this operation")
	ErrInvalidCoordinates  = errors.New("invalid pickup or dropoff coordinates")
	ErrInvalidRiderID      = errors.New("invalid rider id")
	ErrInvalidDriverID     = errors.New("invalid driver id")
	ErrInvalidTripID       = errors.New("invalid trip id")
)

type TripUseCase struct {
	tripRepo repository.TripRepository
}

func NewTripUseCase(repo repository.TripRepository) *TripUseCase {
	return &TripUseCase{tripRepo: repo}
}

func (uc *TripUseCase) RequestTrip(ctx context.Context, req *dto.CreateTripRequest) (*dto.TripResponse, error) {
	riderID, err := uuid.Parse(req.RiderID)
	if err != nil {
		return nil, ErrInvalidRiderID
	}

	if !validCoordinates(req.PickupLat, req.PickupLon) || !validCoordinates(req.DropoffLat, req.DropoffLon) {
		return nil, ErrInvalidCoordinates
	}

	trip := entity.NewTrip(riderID, req.PickupLat, req.PickupLon, req.DropoffLat, req.DropoffLon)

	if err := uc.tripRepo.Create(ctx, trip); err != nil {
		return nil, err
	}

	return buildTripResponse(trip), nil
}

func (uc *TripUseCase) AcceptTrip(ctx context.Context, req *dto.AcceptTripRequest) (*dto.TripResponse, error) {
	tripID, err := uuid.Parse(req.TripID)
	if err != nil {
		return nil, ErrInvalidTripID
	}

	driverID, err := uuid.Parse(req.DriverID)
	if err != nil {
		return nil, ErrInvalidDriverID
	}

	// AssignDriver uses row-level WHERE status = 'REQUESTED' to prevent two drivers
	// from accepting the same trip, the loser gets ErrTripAlreadyAccepted
	if err := uc.tripRepo.AssignDriver(ctx, tripID, driverID); err != nil {
		return nil, err
	}

	trip, err := uc.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	return buildTripResponse(trip), nil
}

func (uc *TripUseCase) CompleteTrip(ctx context.Context, req *dto.CompleteTripRequest) (*dto.TripResponse, error) {
	tripID, err := uuid.Parse(req.TripID)
	if err != nil {
		return nil, ErrInvalidTripID
	}

	// only an IN_PROGRESS trip can be completed; the repo enforces this at the SQL level
	trip, err := uc.tripRepo.FindByID(ctx, tripID)
	if err != nil {
		return nil, err
	}

	if trip.Status != entity.TripStatusInProgress {
		return nil, ErrInvalidTripStatus
	}

	if err := uc.tripRepo.UpdateStatus(ctx, tripID, entity.TripStatusCompleted); err != nil {
		return nil, err
	}

	trip.Status = entity.TripStatusCompleted
	return buildTripResponse(trip), nil
}

func validCoordinates(lat, lon float64) bool {
	return lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180
}

func buildTripResponse(trip *entity.Trip) *dto.TripResponse {
	resp := &dto.TripResponse{
		ID:         trip.ID.String(),
		RiderID:    trip.RiderID.String(),
		PickupLat:  trip.PickupLat,
		PickupLon:  trip.PickupLon,
		DropoffLat: trip.DropoffLat,
		DropoffLon: trip.DropoffLon,
		Status:     string(trip.Status),
		Fare:       trip.Fare,
		CreatedAt:  trip.CreatedAt,
		UpdatedAt:  trip.UpdatedAt,
	}

	if trip.DriverID != nil {
		driverStr := trip.DriverID.String()
		resp.DriverID = &driverStr
	}

	return resp
}
