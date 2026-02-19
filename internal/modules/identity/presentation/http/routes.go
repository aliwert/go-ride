package http

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *AuthHandler) {
	router.Post("/register", handler.Register)
	router.Post("/login", handler.Login)
}
