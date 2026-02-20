package http

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *TripHandler) {
	router.Post("/request", handler.RequestTrip)
	router.Put("/:id/accept", handler.AcceptTrip)
	router.Put("/:id/complete", handler.CompleteTrip)
}
