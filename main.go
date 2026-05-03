package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"calendar/internal/agent"
	"calendar/internal/database"
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

	// Single conversation session for the REPL
	conv := agent.NewConversation()

	fmt.Println("Calendar Agent ready. Type your message (Ctrl+C to exit).")
	fmt.Println("---")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		reply, err := a.Reply(context.Background(), conv, line)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("Agent: %s\n\n", reply)
	}
}
