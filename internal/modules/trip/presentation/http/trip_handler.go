package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/aliwert/go-ride/internal/modules/trip/application/dto"
	"github.com/aliwert/go-ride/internal/modules/trip/application/usecase"
	"github.com/aliwert/go-ride/internal/modules/trip/infrastructure/persistence"
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// user's ID as the rider, ignoring any value from the body
	req.RiderID = c.Locals("userID").(string)

	resp, err := h.tripUC.RequestTrip(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err)
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
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *TripHandler) CompleteTrip(c *fiber.Ctx) error {
	req := &dto.CompleteTripRequest{
		TripID: c.Params("id"),
	}

	resp, err := h.tripUC.CompleteTrip(c.Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// handleError maps known use-case / persistence errors to proper HTTP codes.
// everything else falls through as 500 to avoid leaking internals.
func (h *TripHandler) handleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, usecase.ErrTripNotFound),
		errors.Is(err, persistence.ErrTripNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "trip not found"})

	case errors.Is(err, usecase.ErrTripAlreadyAccepted),
		errors.Is(err, persistence.ErrTripAlreadyAccepted):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "trip has already been accepted"})

	case errors.Is(err, usecase.ErrInvalidTripStatus):
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})

	case errors.Is(err, usecase.ErrInvalidCoordinates),
		errors.Is(err, usecase.ErrInvalidRiderID),
		errors.Is(err, usecase.ErrInvalidDriverID),
		errors.Is(err, usecase.ErrInvalidTripID):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})

	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
