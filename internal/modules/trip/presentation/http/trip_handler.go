package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/aliwert/go-ride/internal/modules/trip/application/dto"
	"github.com/aliwert/go-ride/internal/modules/trip/application/usecase"
	"github.com/aliwert/go-ride/internal/modules/trip/infrastructure/persistence"
	"github.com/aliwert/go-ride/internal/platform/apierror"
)

type TripHandler struct {
	tripUC *usecase.TripUseCase
}

func NewTripHandler(tripUC *usecase.TripUseCase) *TripHandler {
	return &TripHandler{tripUC: tripUC}
}

func (h *TripHandler) RequestTrip(c *fiber.Ctx) error {
	var req dto.CreateTripRequest
	if err := c.BodyParser(&req); err != nil {
		return apierror.NewBadRequest("INVALID_REQUEST_BODY", "invalid request body")
	}

	// rider identity from JWT, ignoring any value from the body
	req.RiderID = c.Locals("userID").(string)

	resp, err := h.tripUC.RequestTrip(c.Context(), &req)
	if err != nil {
		return mapTripError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *TripHandler) AcceptTrip(c *fiber.Ctx) error {
	var req dto.AcceptTripRequest
	req.TripID = c.Params("id")

	// driver ID comes from the JWT token, not the body
	req.DriverID = c.Locals("userID").(string)

	resp, err := h.tripUC.AcceptTrip(c.Context(), &req)
	if err != nil {
		return mapTripError(err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *TripHandler) CompleteTrip(c *fiber.Ctx) error {
	req := &dto.CompleteTripRequest{
		TripID: c.Params("id"),
	}

	resp, err := h.tripUC.CompleteTrip(c.Context(), req)
	if err != nil {
		return mapTripError(err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// translates domain/persistence errors into structured AppErrors
// unknown errors pass through to the global handler which logs and returns 500
func mapTripError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrTripNotFound),
		errors.Is(err, persistence.ErrTripNotFound):
		return apierror.NewNotFound("TRIP_NOT_FOUND", "trip not found")

	case errors.Is(err, usecase.ErrTripAlreadyAccepted),
		errors.Is(err, persistence.ErrTripAlreadyAccepted):
		return apierror.NewConflict("TRIP_ALREADY_ACCEPTED", "trip has already been accepted")

	case errors.Is(err, usecase.ErrInvalidTripStatus):
		return apierror.NewUnprocessable("INVALID_TRIP_STATUS", "invalid trip status for this operation")

	case errors.Is(err, usecase.ErrInvalidCoordinates):
		return apierror.NewBadRequest("INVALID_COORDINATES", "invalid pickup or dropoff coordinates")

	case errors.Is(err, usecase.ErrInvalidRiderID):
		return apierror.NewBadRequest("INVALID_RIDER_ID", "invalid rider id")

	case errors.Is(err, usecase.ErrInvalidDriverID):
		return apierror.NewBadRequest("INVALID_DRIVER_ID", "invalid driver id")

	case errors.Is(err, usecase.ErrInvalidTripID):
		return apierror.NewBadRequest("INVALID_TRIP_ID", "invalid trip id")

	default:
		return err
	}
}
