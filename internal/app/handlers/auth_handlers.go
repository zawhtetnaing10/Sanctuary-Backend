package handlers

import (
	"database/sql"
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

type userWithoutTokenResponse struct {
	ID              int64     `json:"id"`
	Email           string    `json:"email"`
	UserName        string    `json:"user_name"`
	FullName        string    `json:"full_name"`
	ProfileImageUrl string    `json:"profile_image_url"`
	Dob             string    `json:"dob"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       string    `json:"deleted_at"`
}

type userWithTokenResponse struct {
	ID              int64     `json:"id"`
	Email           string    `json:"email"`
	UserName        string    `json:"user_name"`
	FullName        string    `json:"full_name"`
	ProfileImageUrl string    `json:"profile_image_url"`
	Dob             string    `json:"dob"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       string    `json:"deleted_at"`
	AccessToken     string    `json:"access_token"`
}

// Update user
func (cfg *ApiConfig) UpdateUserHandler(writer http.ResponseWriter, request *http.Request) {

	// Validate the request first before doing anything
	fullName := request.FormValue("full_name")
	userName := request.FormValue("user_name")
	dob := request.FormValue("dob")

	if fullName == "" {
		RespondWithError(writer, http.StatusBadRequest, "full name is requred.")
		return
	}

	if userName == "" {
		RespondWithError(writer, http.StatusBadRequest, "user name is requred.")
		return
	}

	if dob == "" {
		RespondWithError(writer, http.StatusBadRequest, "dob is requred.")
		return
	}

	// Parse dob
	dobParsed, parseErr := time.Parse(app.TIME_PARSE_LAYOUT, dob)
	if parseErr != nil {
		RespondWithError(writer, http.StatusInternalServerError, parseErr.Error())
		return
	}

	// Get the bearer token from the request
	token, tokenErr := GetBearerToken(request.Header)
	if tokenErr != nil {
		RespondWithError(writer, http.StatusUnauthorized, tokenErr.Error())
		return
	}

	// Verify the bearer token and get the id
	userId, jwtErr := ValidateJWT(token, cfg.TokenSecret)
	if jwtErr != nil {
		RespondWithError(writer, http.StatusUnauthorized, jwtErr.Error())
		return
	}

	// Get the file from request
	const maxMemory = 10 << 30
	request.Body = http.MaxBytesReader(writer, request.Body, maxMemory)

	request.ParseMultipartForm(maxMemory)

	// Upload it to AWS Server and get the download url
	// TODO: - Continue here
	downloadUrl, uploadErr := UploadFileToAWS(
		"profile",
		"image/",
		request,
		cfg.S3Client,
		cfg.S3Bucket,
		cfg.S3Region,
	)
	if uploadErr != nil {
		RespondWithError(writer, http.StatusInternalServerError, uploadErr.Error())
		return
	}

	// Create params to update user
	params := database.UpdateUserProfileParams{
		ID:       userId,
		FullName: fullName,
		UserName: userName,
		Dob: sql.NullTime{
			Time:  dobParsed,
			Valid: true,
		},
		ProfileImageUrl: sql.NullString{
			String: downloadUrl,
			Valid:  true,
		},
	}

	// Update the user in db using the id from bearer token
	updatedUser, updateErr := cfg.Db.UpdateUserProfile(request.Context(), params)
	if updateErr != nil {
		RespondWithError(writer, http.StatusInternalServerError, updateErr.Error())
		return
	}

	// Create the response
	response := userWithoutTokenResponse{
		ID:              updatedUser.ID,
		Email:           updatedUser.Email,
		UserName:        updatedUser.UserName,
		FullName:        updatedUser.FullName,
		ProfileImageUrl: updatedUser.ProfileImageUrl.String,
		Dob:             FormatNullDobString(updatedUser.Dob.Time),
		CreatedAt:       updatedUser.CreatedAt,
		UpdatedAt:       updatedUser.UpdatedAt,
		DeletedAt:       FormatNullDobString(updatedUser.DeletedAt.Time),
	}

	RespondWithJson(writer, http.StatusOK, response)
}

func (cfg *ApiConfig) RegisterHandler(writer http.ResponseWriter, request *http.Request) {
	// Decode the request
	decoder := json.NewDecoder(request.Body)
	loginRequest := emailAndPasswordRequest{}
	if decodeErr := decoder.Decode(&loginRequest); decodeErr != nil {
		RespondWithError(writer, http.StatusBadRequest, decodeErr.Error())
		return
	}

	if loginRequest.Email == "" {
		RespondWithError(writer, http.StatusBadRequest, "email cannot be empty")
		return
	}

	if loginRequest.Password == "" {
		RespondWithError(writer, http.StatusBadRequest, "password cannot be empty")
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
		Dob:             FormatNullDobString(createdUser.Dob.Time),
		CreatedAt:       createdUser.CreatedAt,
		UpdatedAt:       createdUser.UpdatedAt,
		DeletedAt:       FormatNullDobString(createdUser.DeletedAt.Time),
		AccessToken:     tokenString,
	}

	RespondWithJson(writer, http.StatusCreated, response)
}

func (cfg *ApiConfig) LoginHandler(writer http.ResponseWriter, request *http.Request) {
	// Decode the request
	decoder := json.NewDecoder(request.Body)
	loginRequest := emailAndPasswordRequest{}
	if decodeErr := decoder.Decode(&loginRequest); decodeErr != nil {
		RespondWithError(writer, http.StatusBadRequest, decodeErr.Error())
		return
	}

	if loginRequest.Email == "" {
		RespondWithError(writer, http.StatusBadRequest, "email cannot be empty")
		return
	}

	if loginRequest.Password == "" {
		RespondWithError(writer, http.StatusBadRequest, "password cannot be empty")
		return
	}

	// Find user by email
	userFromDb, getUserErr := cfg.Db.GetUserByEmail(request.Context(), loginRequest.Email)
	if getUserErr != nil {
		RespondWithError(writer, http.StatusNotFound, getUserErr.Error())
		return
	}

	// Verify password
	password := loginRequest.Password
	if checkPassErr := CheckPasswordHash(userFromDb.HashedPassword, password); checkPassErr != nil {
		RespondWithError(writer, http.StatusBadRequest, checkPassErr.Error())
		return
	}

	// Create JWT
	tokenString, jwtErr := MakeJWT(userFromDb.ID, cfg.TokenSecret, app.JWT_EXPIRE_TIME)
	if jwtErr != nil {
		RespondWithError(writer, http.StatusInternalServerError, jwtErr.Error())
		return
	}

	// Create the response
	response := userWithTokenResponse{
		ID:              userFromDb.ID,
		Email:           userFromDb.Email,
		UserName:        userFromDb.UserName,
		FullName:        userFromDb.FullName,
		ProfileImageUrl: userFromDb.ProfileImageUrl.String,
		Dob:             FormatNullDobString(userFromDb.Dob.Time),
		CreatedAt:       userFromDb.CreatedAt,
		UpdatedAt:       userFromDb.UpdatedAt,
		DeletedAt:       FormatNullDobString(userFromDb.DeletedAt.Time),
		AccessToken:     tokenString,
	}

	RespondWithJson(writer, http.StatusOK, response)
}
