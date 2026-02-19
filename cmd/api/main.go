package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aliwert/go-ride/internal/platform/config"
	"github.com/aliwert/go-ride/internal/platform/database"
)

func main() {
	// ---------------------------------------------------------------
	// 1. configuration
	// ---------------------------------------------------------------
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("FATAL: failed to load configuration: %v", err)
	}

	log.Printf("INFO: environment=%s port=%s", cfg.AppEnv, cfg.Port)

	// ---------------------------------------------------------------
	// 2. db — PostgreSQL connection pool
	// ---------------------------------------------------------------
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	pg, err := database.NewPostgresDB(dbCtx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("FATAL: failed to connect to PostgreSQL: %v", err)
	}
	log.Println("INFO: PostgreSQL connection pool initialized successfully")

	// ---------------------------------------------------------------
	// 3. (later i'll) Initialize Redis, Kafka, Fiber, gRPC here
	// ---------------------------------------------------------------

	// ---------------------------------------------------------------
	// 4. Graceful Shutdown
	// ---------------------------------------------------------------
	// create a channel that receives OS signals.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// block the main goroutine until a shutdown signal arrives.
	sig := <-quit
	log.Printf("INFO: received signal %s — initiating graceful shutdown…", sig)

	// release infrastructure resources in reverse order of creation
	pg.Close()
	log.Println("INFO: PostgreSQL connection pool closed")

	log.Println("INFO: go-ride API server stopped gracefully")
}
