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
	dbPath := "db/calendar.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		_, err = os.OpenFile(dbPath, os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("failed to create database file: %v", err)
		}
	} else if err != nil {
		log.Fatalf("failed to check database: %v", err)
	}

	// Open and migrate database
	db, err := database.Open(dbPath, "db/migrations")
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
