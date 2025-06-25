package handlers

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/database"
	"go.uber.org/zap"
)

// Api Config struct
type ApiConfig struct {
	Db          *database.Queries
	Platform    string
	TokenSecret string
	S3Bucket    string
	S3Region    string
	S3Client    *s3.Client
	Logger      *zap.Logger
}

// Get Base url
func (cfg *ApiConfig) GetBaseUrl() string {
	if cfg.Platform == "DEV" {
		return "http://localhost:8080"
	} else {
		// TODO: - replace with production url
		return ""
	}
}

// Util function to log error
func (cfg *ApiConfig) LogError(message string, err error) {
	cfg.Logger.Error("Create user error", zap.Error(err))
}
