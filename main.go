package main

import (
	"log"
	"os"

	"calendar/internal/agent"
	"calendar/internal/database"
	"calendar/internal/server"
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
	if err := database.Open(dbPath); err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	// Create the agent
	a, err := agent.New()
	if err != nil {
		log.Fatalf("agent: %v", err)
	}

	// start and run the http server
	server.StartServer(a)
}
