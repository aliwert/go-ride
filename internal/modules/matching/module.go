package matching

import (
	"github.com/gofiber/fiber/v2"

	locationuc "github.com/aliwert/go-ride/internal/modules/location/application/usecase"
	"github.com/aliwert/go-ride/internal/modules/matching/application/usecase"
	"github.com/aliwert/go-ride/internal/modules/matching/infrastructure/adapter"
	"github.com/aliwert/go-ride/internal/modules/matching/infrastructure/service"
	matchinghttp "github.com/aliwert/go-ride/internal/modules/matching/presentation/http"
)

func InitModule(router fiber.Router, locUC *locationuc.LocationUseCase, authMiddleware fiber.Handler) {
	matchingGroup := router.Group("/matching", authMiddleware)

	locAdapter := adapter.NewLocationAdapter(locUC)
	notifService := service.NewMockNotificationService()
	matchingUC := usecase.NewMatchingUseCase(locAdapter, notifService)
	matchingHandler := matchinghttp.NewMatchingHandler(matchingUC)

	matchinghttp.RegisterRoutes(matchingGroup, matchingHandler)
}
