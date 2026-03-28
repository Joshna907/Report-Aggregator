package main

import (
	"log"

	"github.com/fossology/report-aggregator/internal/api"
	"github.com/fossology/report-aggregator/internal/db"
)

func main() {
	database, err := db.InitDB("aggregator.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	server := api.NewServer(database)

	log.Println("Starting Aggregator API Service (localhost:8080)...")
	if err := server.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
