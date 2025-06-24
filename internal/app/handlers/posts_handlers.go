package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/database"
)

// Post Response
type PostResponse struct {
	ID           int64                    `json:"id"`
	Content      string                   `json:"content"`
	MediaUrl     string                   `json:"media_url"`
	IsLiked      bool                     `json:"is_liked"`
	LikeCount    int                      `json:"like_count"`
	CommentCount int                      `json:"comment_count"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
	User         userWithoutTokenResponse `json:"user"`
}

// Create post handler.
func (cfg *ApiConfig) CreatePostHandler(writer http.ResponseWriter, request *http.Request) {
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

	// Get the user from db
	postUser, userErr := cfg.Db.GetUserById(request.Context(), userId)
	if userErr != nil {
		cfg.LogError(userErr.Error(), userErr)
		RespondWithError(writer, http.StatusInternalServerError, "Something went wrong while creating post. Please try again.")
		return
	}

	// Validate the request
	content := request.FormValue("content")
	if content == "" {
		cfg.LogError("Please provide content for the post.", errors.New("content missing for post"))
		RespondWithError(writer, http.StatusBadRequest, "Please provide content for the post.")
		return
	}

	// If "file" is not null, upload the file to aws and get the download url
	const maxMemory = 10 << 30
	request.Body = http.MaxBytesReader(writer, request.Body, maxMemory)
	request.ParseMultipartForm(maxMemory)
	defer request.MultipartForm.RemoveAll()

	file, fileHeader, getFileErr := request.FormFile("file")
	if getFileErr != nil && getFileErr != http.ErrMissingFile {
		// Actually return the error because the file is actually provided and there's an error parsing it.

		cfg.LogError(getFileErr.Error(), getFileErr)
		RespondWithError(writer, http.StatusBadRequest, "Error uploading the image for the post.")
		return
	}
	if getFileErr == nil {
		defer file.Close()
	}

	// Add post to the db
	params := database.CreatePostParams{
		Content: content,
		UserID:  userId,
	}

	createdPost, createPostErr := cfg.Db.CreatePost(request.Context(), params)
	if createPostErr != nil {
		cfg.LogError(createPostErr.Error(), createPostErr)
		RespondWithError(writer, http.StatusInternalServerError, "Something went wrong while creating post. Please try again.")
		return
	}

	// Media url. Will be empty by default
	mediaUrl := ""

	// If getFileErr is nil and fileHeader.size is not zero, upload the file and add download url to post media table
	if getFileErr == nil && fileHeader.Size != 0 {

		// Check file size before uploading. The file size limit is 1GB
		const maxIndividualFileSize = 1 << 30
		if fileHeader.Size > maxIndividualFileSize {
			cfg.LogError("File too large", errors.New("file too large"))
			RespondWithError(writer, http.StatusBadRequest, "File too large.")
			return
		}

		// Upload the file and get the download url
		downloadUrl, uploadErr := UploadFileToAWS(
			"file",
			"image/",
			request,
			cfg.S3Client,
			cfg.S3Bucket,
			cfg.S3Region,
		)
		if uploadErr != nil {
			cfg.LogError(uploadErr.Error(), uploadErr)
			RespondWithError(writer, http.StatusInternalServerError, "Something went wrong while uploading the file. Please try again.")
			return
		}

		// Update the post_media table.
		postMediaParams := database.CreatePostMediaParams{
			MediaUrl:   downloadUrl,
			OrderIndex: 0,
			PostID:     createdPost.ID,
		}
		_, createPostMediaErr := cfg.Db.CreatePostMedia(request.Context(), postMediaParams)
		if createPostMediaErr != nil {
			cfg.LogError(createPostMediaErr.Error(), createPostMediaErr)
			RespondWithError(writer, http.StatusInternalServerError, "There's something wrong while creating the post. Please try again.")
			return
		}

		// update media url
		mediaUrl = downloadUrl
	}

	postUserResponse := userWithoutTokenResponse{
		ID:              postUser.ID,
		Email:           postUser.Email,
		UserName:        postUser.UserName,
		FullName:        postUser.FullName,
		ProfileImageUrl: postUser.ProfileImageUrl.String,
		Dob:             FormatNullDobString(postUser.Dob.Time),
		CreatedAt:       postUser.CreatedAt.Time,
		UpdatedAt:       postUser.UpdatedAt.Time,
	}

	// Return the post. Need join statement for post_media and user.
	response := PostResponse{
		ID:           createdPost.ID,
		Content:      createdPost.Content,
		IsLiked:      false,
		LikeCount:    0,
		CommentCount: 0,
		CreatedAt:    createdPost.CreatedAt.Time,
		UpdatedAt:    createdPost.UpdatedAt.Time,
		User:         postUserResponse,
		MediaUrl:     mediaUrl,
	}

	RespondWithJson(writer, http.StatusCreated, response)
}
