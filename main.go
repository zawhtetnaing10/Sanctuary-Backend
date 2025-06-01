package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {

	// Load environment file
	godotenv.Load()

	// Get dburl from env
	dbURL := os.Getenv("DB_URL")

	// Open DB
	_, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error loading .env file: %w", err)
	}

	// New http server mux
	mux := http.NewServeMux()

	// Config
	// apiCfg := handlers.ApiConfig{
	// 	Db:          database.New(db),
	// 	TokenSecret: os.Getenv("TOKEN_SECRET"),
	// 	Platform:    os.Getenv("PLATFORM"),
	// }

	// TODO: - Add Handlers

	// New http server
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start the server
	server.ListenAndServe()
}
