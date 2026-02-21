package apierror

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
)

// envelope every error goes through, clients can rely on this shape.
type errorResponse struct {
	Success bool      `json:"success"`
	Error   errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// error from any handler or middleware flows through a single codepath.
func GlobalErrorHandler(c *fiber.Ctx, err error) error {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.HTTPStatus).JSON(errorResponse{
			Success: false,
			Error: errorBody{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
		})
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(errorResponse{
			Success: false,
			Error: errorBody{
				Code:    "FIBER_ERROR",
				Message: fiberErr.Message,
			},
		})
	}

	// unknown errors, log the real cause, return a safe generic response
	log.Printf("ERROR: unhandled error: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(errorResponse{
		Success: false,
		Error: errorBody{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "an unexpected error occurred",
		},
	})
}
