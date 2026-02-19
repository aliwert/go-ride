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

	"github.com/aliwert/go-ride/internal/modules/identity"
	"github.com/aliwert/go-ride/internal/platform/config"
	"github.com/aliwert/go-ride/internal/platform/database"
)

type Server struct {
	fiberApp *fiber.App
	cfg      *config.Config
	db       *database.Postgres
}

func NewServer(cfg *config.Config, db *database.Postgres) *Server {
	app := fiber.New(fiber.Config{
		AppName:      "go-ride",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	})

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New())

	return &Server{
		fiberApp: app,
		cfg:      cfg,
		db:       db,
	}
}

// bootstraps every business module and mounts its routes.
// the server only owns the version prefix, each module defines its own domain path.
func (s *Server) MountHandlers() {
	v1 := s.fiberApp.Group("/api/v1")

	identity.InitModule(v1, s.db.Pool, s.cfg.JWTSecret)
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

	log.Println("INFO: go-ride API server stopped gracefully")
}
