package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/app"
	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/database"
)

// Post List Response
type PostListResponse struct {
	Data []PostResponse `json:"data"`
	Meta MetaResponse   `json:"meta"`
}

// Meta Response
type MetaResponse struct {
	CurrentPage int    `json:"current_page"`
	NextPageUrl string `json:"next_page_url"`
}

// Post Response
type PostResponse struct {
	ID           int64                    `json:"id"`
	Content      string                   `json:"content"`
	MediaUrl     string                   `json:"media_url"`
	LikedByUser  bool                     `json:"liked_by_user"`
	LikeCount    int                      `json:"like_count"`
	CommentCount int                      `json:"comment_count"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
	User         userWithoutTokenResponse `json:"user"`
}

// Post Like
func (cfg *ApiConfig) PostLikeHandler(writer http.ResponseWriter, request *http.Request) {
	// Get the bearer token from the request
	token, tokenErr := GetBearerToken(request.Header)
	if tokenErr != nil {
		cfg.LogError(SERVER_MSG_ERROR_GET_BEARER_TOKEN, tokenErr)
		RespondWithError(writer, http.StatusUnauthorized, "You are not authorized to like the post.")
		return
	}

	// Verify the bearer token and get the id
	userId, jwtErr := ValidateJWT(token, cfg.TokenSecret)
	if jwtErr != nil {
		cfg.LogError(SERVER_MSG_JWT_VALIDATION_FAILED, jwtErr)
		RespondWithError(writer, http.StatusUnauthorized, "You are not authorized to like the post.")
		return
	}

	// Parse post id
	postIdStr := request.FormValue("post_id")
	if postIdStr == "" {
		cfg.LogError("Post id empty", errors.New("post id empty"))
		RespondWithError(writer, http.StatusBadRequest, "Post id cannot be empty.")
		return
	}
	postId, postIdErr := strconv.Atoi(postIdStr)
	if postIdErr != nil {
		cfg.LogError(postIdErr.Error(), postIdErr)
		RespondWithError(writer, http.StatusBadRequest, "Post id must be a number")
		return
	}

	// Like and Unlike
	getPostLikeParams := database.GetPostLikeParams{
		PostID: int64(postId),
		UserID: userId,
	}
	_, getPostLikeErr := cfg.Db.GetPostLike(request.Context(), getPostLikeParams)
	if getPostLikeErr != nil {
		if errors.Is(getPostLikeErr, sql.ErrNoRows) {
			// Post not yet liked. Insert post like
			// Insert post likes
			params := database.CreatePostLikeParams{
				UserID: userId,
				PostID: int64(postId),
			}
			_, postLikeErr := cfg.Db.CreatePostLike(request.Context(), params)
			if postLikeErr != nil {
				cfg.LogError(postLikeErr.Error(), postLikeErr)
				RespondWithError(writer, http.StatusInternalServerError, "Something went wrong when liking the post.")
				return
			}

			// Return empty response
			writer.WriteHeader(http.StatusOK)
			return
		} else {
			// Real database error. Return error response
			cfg.LogError(getPostLikeErr.Error(), getPostLikeErr)
			RespondWithError(writer, http.StatusInternalServerError, "Something went wrong while modifying likes.")
			return
		}
	} else {
		// Already Liked. Delete post like
		deletePostLikeParams := database.DeletePostLikeParams{
			UserID: userId,
			PostID: int64(postId),
		}
		if postUnlikeError := cfg.Db.DeletePostLike(request.Context(), deletePostLikeParams); postUnlikeError != nil {
			cfg.LogError(postUnlikeError.Error(), postUnlikeError)
			RespondWithError(writer, http.StatusInternalServerError, "Something went wrong while modifying likes.")
			return
		}
		// Return empty response
		writer.WriteHeader(http.StatusOK)
		return
	}
}

// Get All Posts Handler
func (cfg *ApiConfig) GetAllPostsHandler(writer http.ResponseWriter, request *http.Request) {
	// Get the bearer token from the request
	token, tokenErr := GetBearerToken(request.Header)
	if tokenErr != nil {
		cfg.LogError(SERVER_MSG_ERROR_GET_BEARER_TOKEN, tokenErr)
		RespondWithError(writer, http.StatusUnauthorized, "You are not authorized to get all posts.")
		return
	}

	// Verify the bearer token and get the id
	userId, jwtErr := ValidateJWT(token, cfg.TokenSecret)
	if jwtErr != nil {
		cfg.LogError(SERVER_MSG_JWT_VALIDATION_FAILED, jwtErr)
		RespondWithError(writer, http.StatusUnauthorized, "You are not authorized to get all posts.")
		return
	}

	// Get the page and calculate the offset
	var page int
	pageStr := request.URL.Query().Get("page")
	if pageStr == "" {
		// If no page number is provided, default to 1
		page = 1
	} else {
		// Convert the pageStr to int. If conversion fails, default to 1
		pageNumber, err := strconv.Atoi(pageStr)
		if err != nil || pageNumber < 1 {
			pageNumber = 1
		}
		page = pageNumber
	}
	offset := (page - 1) * app.PAGE_SIZE

	// Get all posts
	params := database.GetAllPostsParams{
		UserID: userId,
		Offset: int32(offset),
		Limit:  app.PAGE_SIZE,
	}
	posts, getAllPostsErr := cfg.Db.GetAllPosts(request.Context(), params)
	if getAllPostsErr != nil {
		cfg.LogError(getAllPostsErr.Error(), getAllPostsErr)
		RespondWithError(writer, http.StatusInternalServerError, "Error retrieving all posts")
		return
	}

	postList := []PostResponse{}

	for _, postFromDb := range posts {

		// If media url exists, get the first media url.
		mediaUrl := ""
		if len(postFromDb.MediaUrlsArray) > 0 {
			mediaUrl = postFromDb.MediaUrlsArray[0]
		}

		postResponse := PostResponse{
			ID:           postFromDb.ID,
			Content:      postFromDb.Content,
			MediaUrl:     mediaUrl,
			LikedByUser:  postFromDb.LikedByUser,
			LikeCount:    int(postFromDb.LikeCount),
			CommentCount: int(postFromDb.CommentCount),
			CreatedAt:    postFromDb.CreatedAt.Time,
			UpdatedAt:    postFromDb.UpdatedAt.Time,
			User: userWithoutTokenResponse{
				ID:              postFromDb.AuthorID,
				Email:           postFromDb.AuthorEmail,
				UserName:        postFromDb.AuthorUserName,
				FullName:        postFromDb.AuthorFullName,
				ProfileImageUrl: postFromDb.AuthorProfileImageUrl.String,
				Dob:             FormatNullDobString(postFromDb.AuthorDob.Time),
				CreatedAt:       postFromDb.AuthorCreatedAt.Time,
				UpdatedAt:       postFromDb.AuthorUpdatedAt.Time,
			},
		}

		postList = append(postList, postResponse)
	}

	// Get Total posts
	totalCount, totalCountErr := cfg.Db.GetPostsCount(request.Context())
	if totalCountErr != nil {
		cfg.LogError(totalCountErr.Error(), totalCountErr)
		RespondWithError(writer, http.StatusInternalServerError, "Something went wrong while fetching posts.")
		return
	}

	// Construct Meta Response
	totalPages := (totalCount + app.PAGE_SIZE - 1) / app.PAGE_SIZE
	var nextPageUrl string
	if page < int(totalPages) {
		nextPageUrl = fmt.Sprintf("%v/api/posts?page=%v", cfg.GetBaseUrl(), page+1)
	} else {
		nextPageUrl = ""
	}
	metaResponse := MetaResponse{
		CurrentPage: page,
		NextPageUrl: nextPageUrl,
	}

	response := PostListResponse{
		Data: postList,
		Meta: metaResponse,
	}

	RespondWithJson(writer, http.StatusOK, response)
}

// Get Post Details Handler
func (cfg *ApiConfig) GetPostById(writer http.ResponseWriter, request *http.Request) {
	// Get the bearer token from the request
	token, tokenErr := GetBearerToken(request.Header)
	if tokenErr != nil {
		cfg.LogError(SERVER_MSG_ERROR_GET_BEARER_TOKEN, tokenErr)
		RespondWithError(writer, http.StatusUnauthorized, "You are not authorized to get the post.")
		return
	}

	// Verify the bearer token and get the id
	userId, jwtErr := ValidateJWT(token, cfg.TokenSecret)
	if jwtErr != nil {
		cfg.LogError(SERVER_MSG_JWT_VALIDATION_FAILED, jwtErr)
		RespondWithError(writer, http.StatusUnauthorized, "You are not authorized to get the post.")
		return
	}

	// Parse post id from request
	postIdStr := request.PathValue("post_id")
	if postIdStr == "" {
		cfg.LogError("Post id empty", errors.New("post id empty"))
		RespondWithError(writer, http.StatusBadRequest, "Post id cannot be empty.")
		return
	}
	postId, postIdErr := strconv.Atoi(postIdStr)
	if postIdErr != nil {
		cfg.LogError(postIdErr.Error(), postIdErr)
		RespondWithError(writer, http.StatusBadRequest, "Post id must be a number")
		return
	}

	// Get post by id db call
	params := database.GetPostByIdParams{
		ID:     int64(postId),
		UserID: userId,
	}
	postFromDb, postDetailsErr := cfg.Db.GetPostById(request.Context(), params)
	if postDetailsErr != nil {
		cfg.LogError(postDetailsErr.Error(), postDetailsErr)
		RespondWithError(writer, http.StatusInternalServerError, "Something went wrong while getting the post details.")
		return
	}

	// If media url exists, get the first media url.
	mediaUrl := ""
	if len(postFromDb.MediaUrlsArray) > 0 {
		mediaUrl = postFromDb.MediaUrlsArray[0]
	}

	// Parse the response.
	response := PostResponse{
		ID:           postFromDb.ID,
		Content:      postFromDb.Content,
		MediaUrl:     mediaUrl,
		LikedByUser:  postFromDb.LikedByUser,
		LikeCount:    int(postFromDb.LikeCount),
		CommentCount: int(postFromDb.CommentCount),
		CreatedAt:    postFromDb.CreatedAt.Time,
		UpdatedAt:    postFromDb.UpdatedAt.Time,
		User: userWithoutTokenResponse{
			ID:              postFromDb.AuthorID,
			Email:           postFromDb.AuthorEmail,
			UserName:        postFromDb.AuthorUserName,
			FullName:        postFromDb.AuthorFullName,
			ProfileImageUrl: postFromDb.AuthorProfileImageUrl.String,
			Dob:             FormatNullDobString(postFromDb.AuthorDob.Time),
			CreatedAt:       postFromDb.AuthorCreatedAt.Time,
			UpdatedAt:       postFromDb.AuthorUpdatedAt.Time,
		},
	}

	RespondWithJson(writer, http.StatusOK, response)
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
		LikedByUser:  false,
		LikeCount:    0,
		CommentCount: 0,
		CreatedAt:    createdPost.CreatedAt.Time,
		UpdatedAt:    createdPost.UpdatedAt.Time,
		User:         postUserResponse,
		MediaUrl:     mediaUrl,
	}

	RespondWithJson(writer, http.StatusCreated, response)
}
