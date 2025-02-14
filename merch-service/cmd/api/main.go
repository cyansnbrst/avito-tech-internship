package main

import (
	"log"

	"go.uber.org/zap"

	"cyansnbrst/merch-service/config"
	"cyansnbrst/merch-service/internal/server"
	"cyansnbrst/merch-service/pkg/db/postgres"
)

func main() {
	log.Println("starting merch-service server")

	// Load configuration
	cfg, err := config.LoadConfig("config/config-local.yml")
	if err != nil {
		log.Fatalf("loadConfig: %v", err)
	}

	// Set up logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("failed to sync logger: %v", err)
		}
	}()

	// Connect to database
	psqlDB, err := postgres.OpenDB(cfg)
	if err != nil {
		logger.Fatal("failed to init storage", zap.String("error", err.Error()))
	}
	defer psqlDB.Close()
	logger.Info("database connected")

	// Start the server
	s := server.NewServer(cfg, logger, psqlDB)
	if err = s.Run(); err != nil {
		logger.Fatal("an error occurred", zap.String("error", err.Error()))
	}
}
