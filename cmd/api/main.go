package main

import (
	"context"
	"log"
	"time"

	"github.com/aliwert/go-ride/internal/platform/cache"
	"github.com/aliwert/go-ride/internal/platform/config"
	"github.com/aliwert/go-ride/internal/platform/database"
	"github.com/aliwert/go-ride/internal/platform/server"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("FATAL: failed to load configuration: %v", err)
	}

	log.Printf("INFO: environment=%s port=%s", cfg.AppEnv, cfg.Port)

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	pg, err := database.NewPostgresDB(dbCtx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("FATAL: failed to connect to PostgreSQL: %v", err)
	}
	log.Println("INFO: PostgreSQL connection pool initialized successfully")

	redisClient, err := cache.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("FATAL: failed to connect to Redis: %v", err)
	}
	log.Println("INFO: Redis client initialized successfully")

	srv := server.NewServer(&cfg, pg, redisClient)
	srv.MountHandlers()
	srv.Run()
}
