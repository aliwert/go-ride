package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/aliwert/go-ride/internal/modules/identity/application/dto"
	"github.com/aliwert/go-ride/internal/modules/identity/application/usecase"
	"github.com/aliwert/go-ride/internal/modules/identity/infrastructure/persistence"
)

type AuthHandler struct {
	authUC *usecase.AuthUseCase
}

func NewAuthHandler(authUC *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	resp, err := h.authUC.Register(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	resp, err := h.authUC.Login(c.Context(), &req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// handleError maps known use-case / persistence errors to proper HTTP codes
// everything else falls through as 500 to avoid leaking internals
func (h *AuthHandler) handleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, usecase.ErrInvalidCredentials):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})

	case errors.Is(err, usecase.ErrAccountSuspended):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})

	case errors.Is(err, usecase.ErrInvalidRole):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})

	case errors.Is(err, usecase.ErrEmailAlreadyTaken),
		errors.Is(err, persistence.ErrDuplicateEmail):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already taken"})

	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
