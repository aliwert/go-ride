package server

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/redis/go-redis/v9"

	"github.com/aliwert/go-ride/internal/modules/identity"
	"github.com/aliwert/go-ride/internal/modules/location"
	"github.com/aliwert/go-ride/internal/modules/matching"
	"github.com/aliwert/go-ride/internal/modules/tracking"
	"github.com/aliwert/go-ride/internal/modules/trip"
	"github.com/aliwert/go-ride/internal/platform/apierror"
	"github.com/aliwert/go-ride/internal/platform/config"
	"github.com/aliwert/go-ride/internal/platform/database"
	"github.com/aliwert/go-ride/internal/platform/middleware"
	platformws "github.com/aliwert/go-ride/internal/platform/websocket"
)

type Server struct {
	fiberApp    *fiber.App
	cfg         *config.Config
	db          *database.Postgres
	redisClient *redis.Client
}

func NewServer(cfg *config.Config, db *database.Postgres, redisClient *redis.Client) *Server {
	app := fiber.New(fiber.Config{
		AppName:      "go-ride",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
		ErrorHandler: apierror.GlobalErrorHandler,
	})

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New())

	return &Server{
		fiberApp:    app,
		cfg:         cfg,
		db:          db,
		redisClient: redisClient,
	}
}

// bootstraps every business module and mounts its routes.
// the server only owns the version prefix, each module defines its own domain path.
func (s *Server) MountHandlers() {
	v1 := s.fiberApp.Group("/api/v1")

	// auth middleware protects modules that require a valid JWT; identity stays public
	authMid := middleware.RequireAuth(s.cfg.JWTSecret)

	// tracking module is initialized first so location can broadcast via its port
	hub := platformws.NewHub()
	broadcaster := tracking.InitModule(s.fiberApp, hub)

	identity.InitModule(v1, s.db.Pool, s.cfg.JWTSecret)
	locUC := location.InitModule(v1, s.redisClient, authMid, broadcaster)
	trip.InitModule(v1, s.db.Pool, authMid)
	matching.InitModule(v1, locUC, authMid)
}

// starts the server and blocks until a termination signal arrives.
// shutdown order: drain HTTP → close DB pool.
func (s *Server) Run() {
	go func() {
		addr := ":" + s.cfg.Port
		if err := s.fiberApp.Listen(addr); err != nil {
			log.Fatalf("FATAL: fiber server failed: %v", err)
		}
	}()

	log.Printf("INFO: fiber listening on :%s", s.cfg.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Printf("INFO: received signal %s — initiating graceful shutdown…", sig)

	if err := s.fiberApp.Shutdown(); err != nil {
		log.Printf("WARN: fiber shutdown error: %v", err)
	}
	log.Println("INFO: fiber server stopped")

	s.db.Close()
	log.Println("INFO: PostgreSQL connection pool closed")

	if s.redisClient != nil {
		s.redisClient.Close()
	}
	log.Println("INFO: Redis client closed")

	log.Println("INFO: go-ride API server stopped gracefully")
}
