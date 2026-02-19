package http

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *LocationHandler) {
	router.Post("/update", handler.UpdateLocation)
	router.Get("/nearby", handler.FindNearby)
}
