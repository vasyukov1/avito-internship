package main

import (
	_ "avito-internship/docs"
	"avito-internship/internal/config"
	"avito-internship/internal/delivery"
	"avito-internship/internal/infrastructure"
	"avito-internship/internal/repository"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

// @title PR Service
// @version 1.0
// @description This is PR Service.
// @contact.name API Support
// @contact.url http://localhost:8080/swagger/index.html
// @host localhost:8080
// @BasePath /
func main() {
	ctx := context.Background()

	cfg := config.Load()

	// Database setup
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatal("DB ping failed:", err)
	}

	storage := repository.NewPgStorage(dbPool)
	service := infrastructure.NewServer(storage)
	handler := delivery.NewHandler(service)
	router := delivery.NewRouter(handler)

	log.Printf("Starting HTTP server on %s\n", cfg.Port)
	log.Printf("Swagger: http://localhost%s/swagger/index.html\n", cfg.Port)

	if err := router.Run(cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
