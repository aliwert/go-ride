package identity

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aliwert/go-ride/internal/modules/identity/application/usecase"
	"github.com/aliwert/go-ride/internal/modules/identity/infrastructure/persistence"
	"github.com/aliwert/go-ride/internal/modules/identity/infrastructure/security"
	identityhttp "github.com/aliwert/go-ride/internal/modules/identity/presentation/http"
)

// wires the full identity dependency graph and mounts routes.
// the module owns its own path prefix — the server only passes the version group
func InitModule(router fiber.Router, dbPool *pgxpool.Pool, jwtSecret string) {
	identityGroup := router.Group("/identity")

	userRepo := persistence.NewPostgresUserRepository(dbPool)
	tokenGen := security.NewJwtTokenGenerator(jwtSecret)
	authUC := usecase.NewAuthUseCase(userRepo, tokenGen)
	authHandler := identityhttp.NewAuthHandler(authUC)

	identityhttp.RegisterRoutes(identityGroup, authHandler)
}
