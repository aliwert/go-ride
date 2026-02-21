package http

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *MatchingHandler) {
	router.Post("/:trip_id/match", handler.Match)
}
