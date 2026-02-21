package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/aliwert/go-ride/internal/modules/matching/application/usecase"
	"github.com/aliwert/go-ride/internal/platform/apierror"
)

type MatchingHandler struct {
	matchingUC *usecase.MatchingUseCase
}

func NewMatchingHandler(matchingUC *usecase.MatchingUseCase) *MatchingHandler {
	return &MatchingHandler{matchingUC: matchingUC}
}

type matchRequest struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func (h *MatchingHandler) Match(c *fiber.Ctx) error {
	tripID, err := uuid.Parse(c.Params("trip_id"))
	if err != nil {
		return apierror.NewBadRequest("INVALID_TRIP_ID", "invalid trip id")
	}

	var req matchRequest
	if err := c.BodyParser(&req); err != nil {
		return apierror.NewBadRequest("INVALID_REQUEST_BODY", "invalid request body")
	}

	if err := h.matchingUC.ProcessTrip(c.Context(), tripID, req.Lat, req.Lon); err != nil {
		return mapMatchingError(err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "drivers notified successfully",
	})
}

func mapMatchingError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrNoDriversAvailable):
		return apierror.NewNotFound("NO_DRIVERS_AVAILABLE", "no drivers available nearby")
	default:
		return err
	}
}
