package main

import (
	"log"
	"os"

	"fake.com/nilspcarlson/internal/agent"
	"fake.com/nilspcarlson/internal/dal"
	"fake.com/nilspcarlson/internal/database"
	"fake.com/nilspcarlson/internal/server"
)

func main() {
	// load mysql dsn from environment
	dsn := os.Getenv("NILSPCARLSON_MYSQL_DSN")
	if dsn == "" {
		log.Fatal("NILSPCARLSON_MYSQL_DSN environment variable not set")
	}

	// Open and migrate database
	db, err := database.Open(dsn, "db/migrations/mysql")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// set dal database
	dal.DB = db

	// Create the agent
	a, err := agent.New()
	if err != nil {
		log.Fatalf("agent: %v", err)
	}

    // set server ui path
    server.UiPath = "ui"

	// return the route muxer
	server.StartServer(a)
}
