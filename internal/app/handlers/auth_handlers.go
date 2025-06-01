package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/app"
	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/database"
)

type emailAndPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userWithTokenResponse struct {
	ID              int64     `json:"id"`
	Email           string    `json:"email"`
	UserName        string    `json:"user_name"`
	FullName        string    `json:"full_name"`
	ProfileImageUrl string    `json:"profile_image_url"`
	Dob             time.Time `json:"dob"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       time.Time `json:"deleted_at"`
	AccessToken     string    `json:"access_token"`
}

func (cfg *ApiConfig) RegisterHandler(writer http.ResponseWriter, request *http.Request) {
	// Decode the request
	decoder := json.NewDecoder(request.Body)
	loginRequest := emailAndPasswordRequest{}
	if decodeErr := decoder.Decode(&loginRequest); decodeErr != nil {
		RespondWithError(writer, http.StatusBadRequest, decodeErr.Error())
		return
	}

	// Hash the password
	password := loginRequest.Password
	hashedPassword, hashErr := HashPassword(password)
	if hashErr != nil {
		RespondWithError(writer, http.StatusInternalServerError, hashErr.Error())
		return
	}

	// Insert the user into db
	createUserParams := database.CreateUserParams{
		Email:          loginRequest.Email,
		HashedPassword: hashedPassword,
		FullName:       "",
		UserName:       "",
	}
	createdUser, createUserErr := cfg.Db.CreateUser(request.Context(), createUserParams)
	if createUserErr != nil {
		RespondWithError(writer, http.StatusInternalServerError, createUserErr.Error())
		return
	}

	// Create JWT
	tokenString, jwtErr := MakeJWT(createdUser.ID, cfg.TokenSecret, app.JWT_EXPIRE_TIME)
	if jwtErr != nil {
		RespondWithError(writer, http.StatusInternalServerError, jwtErr.Error())
		return
	}

	// Create the response
	response := userWithTokenResponse{
		ID:              createdUser.ID,
		Email:           createdUser.Email,
		UserName:        createdUser.UserName,
		FullName:        createdUser.FullName,
		ProfileImageUrl: createdUser.ProfileImageUrl.String,
		Dob:             createdUser.Dob.Time,
		CreatedAt:       createdUser.CreatedAt,
		UpdatedAt:       createdUser.UpdatedAt,
		DeletedAt:       createdUser.DeletedAt.Time,
		AccessToken:     tokenString,
	}

	RespondWithJson(writer, http.StatusCreated, response)
}
