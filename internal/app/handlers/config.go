package handlers

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/database"
)

// Api Config struct
type ApiConfig struct {
	Db          *database.Queries
	Platform    string
	TokenSecret string
	S3Bucket    string
	S3Region    string
	S3Client    *s3.Client
}
