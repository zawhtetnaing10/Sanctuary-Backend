package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/app"
	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/app/validators"
	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/database"
)

type emailAndPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userWithTokenResponse struct {
	ID              int64              `json:"id"`
	Email           string             `json:"email"`
	UserName        string             `json:"user_name"`
	FullName        string             `json:"full_name"`
	ProfileImageUrl string             `json:"profile_image_url"`
	Dob             string             `json:"dob"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	AccessToken     string             `json:"access_token"`
	Interests       []interestResponse `json:"interests"`
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
}

// Update user
func (cfg *ApiConfig) UpdateUserHandler(writer http.ResponseWriter, request *http.Request) {

	// Validate the request first before doing anything
	errCode, validationErr := validators.ValidateUpdateUserRequest(request, cfg.Db)
	if validationErr != nil {
		RespondWithError(writer, errCode, validationErr.Error())
		return
	}

	// Parse dob
	dobParsed, parseErr := time.Parse(app.TIME_PARSE_LAYOUT, request.FormValue("dob"))
	if parseErr != nil {
		cfg.LogError(SERVER_MSG_ERROR_PARSING_DOB, parseErr)
		RespondWithError(writer, http.StatusInternalServerError, CLIENT_MSG_ERROR_UPDATE_USER)
		return
	}

	// Get the bearer token from the request
	token, tokenErr := GetBearerToken(request.Header)
	if tokenErr != nil {
		cfg.LogError(SERVER_MSG_ERROR_GET_BEARER_TOKEN, tokenErr)
		RespondWithError(writer, http.StatusUnauthorized, CLIENT_MSG_ERROR_UPDATE_USER)
		return
	}

	// Verify the bearer token and get the id
	userId, jwtErr := ValidateJWT(token, cfg.TokenSecret)
	if jwtErr != nil {
		cfg.LogError(SERVER_MSG_JWT_VALIDATION_FAILED, jwtErr)
		RespondWithError(writer, http.StatusUnauthorized, CLIENT_MSG_ERROR_UPDATE_USER)
		return
	}

	// Get the file from request
	const maxMemory = 10 << 30
	request.Body = http.MaxBytesReader(writer, request.Body, maxMemory)

	// Parse multipart form
	request.ParseMultipartForm(maxMemory)

	// Upload it to AWS Server and get the download url
	downloadUrl, uploadErr := UploadFileToAWS(
		"profile",
		"image/",
		request,
		cfg.S3Client,
		cfg.S3Bucket,
		cfg.S3Region,
	)
	if uploadErr != nil {
		cfg.LogError(SERVER_MSG_ERROR_UPLOADING_PHOTO, uploadErr)
		RespondWithError(writer, http.StatusInternalServerError, CLIENT_MSG_ERROR_UPLOADING_PROFILE_PICTURE)
		return
	}

	// Create params to update user
	params := database.UpdateUserProfileParams{
		ID:       userId,
		FullName: request.FormValue("full_name"),
		UserName: request.FormValue("user_name"),
		Dob: pgtype.Date{
			Time:  dobParsed,
			Valid: true,
		},
		ProfileImageUrl: pgtype.Text{
			String: downloadUrl,
			Valid:  true,
		},
	}

	// Update the user in db using the id from bearer token
	updatedUser, updateErr := cfg.Db.UpdateUserProfile(request.Context(), params)
	if updateErr != nil {
		cfg.LogError(SERVER_MSG_UPDATE_USER_ERROR, updateErr)
		RespondWithError(writer, http.StatusInternalServerError, CLIENT_MSG_ERROR_UPDATE_USER)
		return
	}

	// Bulk inserting into users_has_interests
	// Get Duplicate interest ids
	interestsIds := request.MultipartForm.Value["ids"]
	parsedInterestIdsFromRequest := []int64{}
	for _, interestId := range interestsIds {
		parsedId, parseErr := strconv.ParseInt(interestId, 10, 64)
		if parseErr != nil {
			cfg.LogError(SERVER_MSG_INTEREST_BULK_INSERT_ERROR, parseErr)
			RespondWithError(writer, http.StatusBadRequest, CLIENT_MSG_ERROR_UPDATE_USER)
			return
		}
		parsedInterestIdsFromRequest = append(parsedInterestIdsFromRequest, parsedId)
	}

	getDupInterstIdParams := database.GetDuplicateInterestIdsParams{
		UserID:  updatedUser.ID,
		Column2: parsedInterestIdsFromRequest,
	}
	duplicateInterstIds, dupInterestIdsErr := cfg.Db.GetDuplicateInterestIds(request.Context(), getDupInterstIdParams)
	if dupInterestIdsErr != nil {
		cfg.LogError(SERVER_MSG_GET_DUPLICATE_IDS, dupInterestIdsErr)
		RespondWithError(writer, http.StatusInternalServerError, CLIENT_MSG_ERROR_UPDATE_USER)
		return
	}

	// Remove the duplicate interest ids from request
	duplicateInterstIdsMap := map[int64]bool{}
	for _, dupInterestId := range duplicateInterstIds {
		duplicateInterstIdsMap[dupInterestId] = true
	}

	interestIdsWithoutDuplicates := []int64{}
	for _, interestIdFromRequest := range parsedInterestIdsFromRequest {
		_, ok := duplicateInterstIdsMap[interestIdFromRequest]
		if !ok {
			interestIdsWithoutDuplicates = append(interestIdsWithoutDuplicates, interestIdFromRequest)
		}
	}

	// Bulk Insert the Interest ids into users_has_interests
	usersHasInterestParamsSlice := []database.CreateUsersHasInterestsParams{}
	for _, interestId := range interestIdsWithoutDuplicates {
		params := database.CreateUsersHasInterestsParams{
			UserID:     userId,
			InterestID: interestId,
			CreatedAt: pgtype.Timestamp{
				Time:  time.Now(),
				Valid: true,
			},
			UpdatedAt: pgtype.Timestamp{
				Time:  time.Now(),
				Valid: true,
			},
		}
		usersHasInterestParamsSlice = append(usersHasInterestParamsSlice, params)
	}
	_, createUsersHasInterestErr := cfg.Db.CreateUsersHasInterests(request.Context(), usersHasInterestParamsSlice)
	if createUsersHasInterestErr != nil {
		cfg.LogError(SERVER_MSG_CREATE_USER_HAS_INTEREST_ERROR, createUsersHasInterestErr)
		RespondWithError(writer, http.StatusInternalServerError, CLIENT_MSG_ERROR_UPDATE_USER)
		return
	}

	// Get Interests For User
	interests, getInterestsErr := getInterestsForUser(userId, request, cfg.Db)
	if getInterestsErr != nil {
		cfg.LogError(SERVER_MSG_CREATE_USER_HAS_INTEREST_ERROR, getInterestsErr)
		RespondWithError(writer, http.StatusInternalServerError, CLIENT_MSG_ERROR_UPDATE_USER)
		return
	}

	// Create JWT
	tokenString, jwtErr := MakeJWT(updatedUser.ID, cfg.TokenSecret, app.JWT_EXPIRE_TIME)
	if jwtErr != nil {
		cfg.LogError(SERVER_MSG_MAKE_JWT_FAILED, jwtErr)
		RespondWithError(writer, http.StatusInternalServerError, jwtErr.Error())
		return
	}

	// Create the response
	response := userWithTokenResponse{
		ID:              updatedUser.ID,
		Email:           updatedUser.Email,
		UserName:        updatedUser.UserName,
		FullName:        updatedUser.FullName,
		ProfileImageUrl: updatedUser.ProfileImageUrl.String,
		Dob:             FormatNullDobString(updatedUser.Dob.Time),
		CreatedAt:       updatedUser.CreatedAt.Time,
		UpdatedAt:       updatedUser.UpdatedAt.Time,
		Interests:       interests,
		AccessToken:     tokenString,
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
		RespondWithError(writer, http.StatusBadRequest, CLIENT_MSG_EMAIL_CANNOT_BE_EMPTY)
		return
	}

	if loginRequest.Password == "" {
		RespondWithError(writer, http.StatusBadRequest, CLIENT_MSG_PASSWORD_CANNOT_BE_EMPTY)
		return
	}

	// Hash the password
	password := loginRequest.Password
	hashedPassword, hashErr := HashPassword(password)
	if hashErr != nil {
		cfg.LogError(SERVER_MSG_PASSWORD_HASH_FAILED, hashErr)
		RespondWithError(writer, http.StatusInternalServerError, CLIENT_MSG_CREATE_USER_ERROR)
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
		cfg.LogError(SERVER_MSG_CREATE_USER_FAILED, createUserErr)
		RespondWithError(writer, http.StatusInternalServerError, CLIENT_MSG_CREATE_USER_ERROR)
		return
	}

	// Create JWT
	tokenString, jwtErr := MakeJWT(createdUser.ID, cfg.TokenSecret, app.JWT_EXPIRE_TIME)
	if jwtErr != nil {
		cfg.LogError(SERVER_MSG_MAKE_JWT_FAILED, jwtErr)
		RespondWithError(writer, http.StatusInternalServerError, jwtErr.Error())
		return
	}

	// Get Interests For User
	interests, getInterestsErr := getInterestsForUser(createdUser.ID, request, cfg.Db)
	if getInterestsErr != nil {
		cfg.LogError(SERVER_MSG_GETTING_INTEREST_FOR_USER_FAILED, getInterestsErr)
		RespondWithError(writer, http.StatusInternalServerError, CLIENT_MSG_CREATE_USER_ERROR)
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
		CreatedAt:       createdUser.CreatedAt.Time,
		UpdatedAt:       createdUser.UpdatedAt.Time,
		AccessToken:     tokenString,
		Interests:       interests,
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
		RespondWithError(writer, http.StatusBadRequest, CLIENT_MSG_EMAIL_CANNOT_BE_EMPTY)
		return
	}

	if loginRequest.Password == "" {
		RespondWithError(writer, http.StatusBadRequest, CLIENT_MSG_PASSWORD_CANNOT_BE_EMPTY)
		return
	}

	// Find user by email
	userFromDb, getUserErr := cfg.Db.GetUserByEmail(request.Context(), loginRequest.Email)
	if getUserErr != nil {
		if errors.Is(getUserErr, sql.ErrNoRows) {
			RespondWithError(writer, http.StatusNotFound, CLIENT_MSG_INCORRECT_EMAIL_OR_PASSWORD)
			return
		}
		cfg.LogError(SERVER_MSG_LOGIN_FAILED, getUserErr)
		RespondWithError(writer, http.StatusNotFound, getUserErr.Error())
		return
	}

	// Verify password
	password := loginRequest.Password
	if checkPassErr := CheckPasswordHash(userFromDb.HashedPassword, password); checkPassErr != nil {
		RespondWithError(writer, http.StatusNotFound, CLIENT_MSG_INCORRECT_EMAIL_OR_PASSWORD)
		return
	}

	// Create JWT
	tokenString, jwtErr := MakeJWT(userFromDb.ID, cfg.TokenSecret, app.JWT_EXPIRE_TIME)
	if jwtErr != nil {
		cfg.LogError(SERVER_MSG_MAKE_JWT_FAILED, jwtErr)
		RespondWithError(writer, http.StatusInternalServerError, jwtErr.Error())
		return
	}

	// Get Interests For User
	interests, getInterestsErr := getInterestsForUser(userFromDb.ID, request, cfg.Db)
	if getInterestsErr != nil {
		cfg.LogError(SERVER_MSG_GETTING_INTEREST_FOR_USER_FAILED, getInterestsErr)
		RespondWithError(writer, http.StatusInternalServerError, CLIENT_MSG_INCORRECT_EMAIL_OR_PASSWORD)
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
		CreatedAt:       userFromDb.CreatedAt.Time,
		UpdatedAt:       userFromDb.UpdatedAt.Time,
		AccessToken:     tokenString,
		Interests:       interests,
	}

	RespondWithJson(writer, http.StatusOK, response)
}

// Get Interests for user
func getInterestsForUser(userId int64, request *http.Request, db *database.Queries) ([]interestResponse, error) {
	interestsFromDb, err := db.GetInterestsForUser(request.Context(), userId)
	if err != nil {
		return []interestResponse{}, err
	}

	response := []interestResponse{}
	for _, interest := range interestsFromDb {
		interestResponse := interestResponse{
			ID:        interest.ID,
			Name:      interest.Name,
			CreatedAt: interest.CreatedAt.Time,
			UpdatedAt: interest.UpdatedAt.Time,
		}
		response = append(response, interestResponse)
	}

	return response, nil
}
