package location

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	"github.com/aliwert/go-ride/internal/modules/location/application/usecase"
	"github.com/aliwert/go-ride/internal/modules/location/infrastructure/persistence"
	locationhttp "github.com/aliwert/go-ride/internal/modules/location/presentation/http"
)

func InitModule(router fiber.Router, redisClient *redis.Client, authMiddleware fiber.Handler) {
	// apply auth middleware at the group level so every location route requires a valid token
	locationGroup := router.Group("/location", authMiddleware)

	locationRepo := persistence.NewRedisLocationRepository(redisClient)
	locationUC := usecase.NewLocationUseCase(locationRepo)
	locationHandler := locationhttp.NewLocationHandler(locationUC)

	locationhttp.RegisterRoutes(locationGroup, locationHandler)
}
