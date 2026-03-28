package main

import (
	"log"

	"os"

	"github.com/fossology/report-aggregator/internal/api"
	"github.com/fossology/report-aggregator/internal/db"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "aggregator.db"
	}

	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Aggregator API Service on port %s...\n", port)
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
