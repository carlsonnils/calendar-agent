// package main // dev
package nilspcarlson

import (
	"log"
    "net/http"
	"os"

	"fake.com/nilspcarlson/internal/agent"
	"fake.com/nilspcarlson/internal/dal"
	"fake.com/nilspcarlson/internal/database"
	"fake.com/nilspcarlson/internal/jwt"
	"fake.com/nilspcarlson/internal/server"
)

// func main() {    //dev
func NewServeMuxer() http.Handler {
	// set jwt hmac key
	envHmacKey := os.Getenv("NILSPCARLSON_HMAC_KEY")
	if envHmacKey == "" {
		log.Fatal("NILSPCARLSON_HMAC_KEY environment variable not set")
	}
	jwt.HmacKey = []byte(envHmacKey)

	// load mysql dsn from environment
	dsn := os.Getenv("NILSPCARLSON_MYSQL_DSN")
	if dsn == "" {
		log.Fatal("NILSPCARLSON_MYSQL_DSN environment variable not set")
	}

	// Open and migrate database
	// db, err := database.Open(dsn, "db/migrations/mysql") // dev
	db, err := database.Open(dsn, "nilspcarlson/db/migrations/mysql")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	// defer database.Close()   // dev

	// set dal database
	dal.DB = db

	// Create the agent
	a, err := agent.New()
	if err != nil {
		log.Fatalf("new agent: %v", err)
	}

	// Load conversation
	conv, err := agent.LoadConversation("nilspcarlson_calendar_0")
	if err != nil {
		log.Fatalf("loading conversation: %v", err)
	}

    // set server ui path
    // server.UiPath = "ui" // dev
    server.UiPath = "nilspcarlson/ui"

	// return the route muxer
    // server.StartServer(a, conv)  // dev
	return server.BuildMuxer(a, conv)
}

func CloseDB() {
    database.Close()
}
