package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/aliwert/go-ride/internal/modules/identity/application/dto"
	"github.com/aliwert/go-ride/internal/modules/identity/application/usecase"
	"github.com/aliwert/go-ride/internal/modules/identity/infrastructure/persistence"
	"github.com/aliwert/go-ride/internal/platform/apierror"
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
		return apierror.NewBadRequest("INVALID_REQUEST_BODY", "invalid request body")
	}

	resp, err := h.authUC.Register(c.Context(), &req)
	if err != nil {
		return mapIdentityError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return apierror.NewBadRequest("INVALID_REQUEST_BODY", "invalid request body")
	}

	resp, err := h.authUC.Login(c.Context(), &req)
	if err != nil {
		return mapIdentityError(err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// translates domain/persistence errors into structured AppErrors
// the global error handler takes care of serialisation
func mapIdentityError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrInvalidCredentials):
		return apierror.NewUnauthorized("INVALID_CREDENTIALS", "invalid credentials")

	case errors.Is(err, usecase.ErrAccountSuspended):
		return apierror.NewForbidden("ACCOUNT_SUSPENDED", "account suspended")

	case errors.Is(err, usecase.ErrInvalidRole):
		return apierror.NewBadRequest("INVALID_ROLE", "invalid role")

	case errors.Is(err, usecase.ErrEmailAlreadyTaken),
		errors.Is(err, persistence.ErrDuplicateEmail):
		return apierror.NewConflict("EMAIL_ALREADY_TAKEN", "a user with this email already exists")

	default:
		return err
	}
}
