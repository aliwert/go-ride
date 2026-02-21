package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/aliwert/go-ride/internal/modules/location/application/dto"
	"github.com/aliwert/go-ride/internal/modules/location/application/usecase"
)

type LocationHandler struct {
	locationUC *usecase.LocationUseCase
}

func NewLocationHandler(uc *usecase.LocationUseCase) *LocationHandler {
	return &LocationHandler{locationUC: uc}
}

func (h *LocationHandler) UpdateLocation(c *fiber.Ctx) error {
	var req dto.UpdateLocationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	// driver identity must come from the JWT, not the request body never trust the client
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing user identity"})
	}

	driverID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid driver id in token"})
	}
	req.DriverID = driverID

	if err := h.locationUC.UpdateLocation(c.Context(), &req); err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "location updated"})
}

func (h *LocationHandler) FindNearby(c *fiber.Ctx) error {
	var req dto.FindNearbyRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid query parameters"})
	}

	drivers, err := h.locationUC.FindNearbyDrivers(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err)
	}

	ids := make([]string, len(drivers))
	for i, d := range drivers {
		ids[i] = d.String()
	}

	return c.Status(fiber.StatusOK).JSON(dto.NearbyDriversResponse{Drivers: ids})
}

func (h *LocationHandler) handleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, usecase.ErrInvalidCoordinates):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, usecase.ErrInvalidRadius):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
