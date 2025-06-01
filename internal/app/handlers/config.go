package handlers

import "github.com/zawhtetnaing10/Sanctuary-Backend/internal/database"

// Api Config struct
type ApiConfig struct {
	Db          *database.Queries
	Platform    string
	TokenSecret string
}
