package trip

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aliwert/go-ride/internal/modules/trip/application/usecase"
	"github.com/aliwert/go-ride/internal/modules/trip/infrastructure/persistence"
	triphttp "github.com/aliwert/go-ride/internal/modules/trip/presentation/http"
)

func InitModule(router fiber.Router, dbPool *pgxpool.Pool, authMiddleware fiber.Handler) {
	tripGroup := router.Group("/trip", authMiddleware)

	tripRepo := persistence.NewPostgresTripRepository(dbPool)
	tripUC := usecase.NewTripUseCase(tripRepo)
	tripHandler := triphttp.NewTripHandler(tripUC)

	triphttp.RegisterRoutes(tripGroup, tripHandler)
}
