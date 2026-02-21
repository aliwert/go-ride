package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/aliwert/go-ride/internal/modules/matching/application/usecase"
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid trip id",
		})
	}

	var req matchRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.matchingUC.ProcessTrip(c.Context(), tripID, req.Lat, req.Lon); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "drivers notified successfully",
	})
}

func (h *MatchingHandler) handleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, usecase.ErrNoDriversAvailable):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
