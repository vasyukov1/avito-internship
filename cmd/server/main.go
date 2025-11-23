package main

import (
	_ "avito-internship/docs"
	"avito-internship/internal/config"
	"avito-internship/internal/delivery"
	"avito-internship/internal/logger"
	"avito-internship/internal/repository"
	"avito-internship/internal/usecase"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"os"
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
	logger.Init(cfg.LogLevel)
	logger.Log.WithFields(logrus.Fields{
		"port":    cfg.Port,
		"db_host": os.Getenv("DB_HOST"),
	}).Info("Configuration loaded")

	// Database setup
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Log.WithError(err).Fatal("DB connection failed")
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		logger.Log.WithError(err).Fatal("DB ping failed")
	}

	logger.Log.Info("Database connection established")

	storage := repository.NewPgStorage(dbPool, logger.Log)
	service := usecase.NewServer(storage, logger.Log)
	handler := delivery.NewHandler(service, logger.Log)
	router := delivery.NewRouter(handler)

	logger.Log.WithFields(logrus.Fields{
		"port":    cfg.Port,
		"swagger": "http://localhost" + cfg.Port + "/swagger/index.html",
	}).Info("Starting HTTP server")

	if err := router.Run(cfg.Port); err != nil {
		logger.Log.WithError(err).Fatal("Server failed to start")
	}
}
