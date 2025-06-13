package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/app/handlers"
	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/database"
	"go.uber.org/zap"
)

func main() {

	// Load environment file
	godotenv.Load()

	// Get dburl from env
	dbURL := os.Getenv("DB_URL")

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Unable to connect to db %w", err)
	}

	// // Open DB
	// dbConn, err := sql.Open("pgx", dbURL)
	// if err != nil {
	// 	log.Fatal("Error loading .env file: %w", err)
	// }

	// AWS Set up
	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		log.Fatal("S3_BUCKET environment variable is not set")
	}

	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		log.Fatal("S3_REGION environment variable is not set")
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(s3Region))
	if err != nil {
		log.Fatal("error configuring aws s3")
	}

	// Logger
	logger, loggerInitErr := zap.NewDevelopment()
	if loggerInitErr != nil {
		os.Exit(1)
	}

	// Config
	apiCfg := handlers.ApiConfig{
		Db:          database.New(pool),
		TokenSecret: os.Getenv("TOKEN_SECRET"),
		Platform:    os.Getenv("PLATFORM"),
		S3Bucket:    s3Bucket,
		S3Region:    s3Region,
		S3Client:    s3.NewFromConfig(awsCfg),
		Logger:      logger,
	}

	// New http server mux
	mux := http.NewServeMux()

	// Add Handlers
	mux.HandleFunc("POST /api/register", apiCfg.RegisterHandler)
	mux.HandleFunc("POST /api/reset", apiCfg.ResetHandler)
	mux.HandleFunc("POST /api/login", apiCfg.LoginHandler)
	mux.HandleFunc("POST /api/updateUser", apiCfg.UpdateUserHandler)
	mux.HandleFunc("GET /api/interests", apiCfg.GetAllInterests)

	// New http server
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start the server
	server.ListenAndServe()
}
