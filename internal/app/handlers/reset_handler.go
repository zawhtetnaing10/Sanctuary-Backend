package handlers

import (
	"net/http"
)

func (cfg *ApiConfig) ResetHandler(writer http.ResponseWriter, request *http.Request) {
	if cfg.Platform == "DEV" {
		if delUsersErr := cfg.Db.DeleteAllUsers(request.Context()); delUsersErr != nil {
			RespondWithError(writer, http.StatusInternalServerError, "error deleting users")
			return
		}
		if delUsersHasInterestErr := cfg.Db.DeleteAllUsersHasInterests(request.Context()); delUsersHasInterestErr != nil {
			RespondWithError(writer, http.StatusInternalServerError, "error deleting users has interests")
			return
		}
		if delPostsErr := cfg.Db.DeleteAllPosts(request.Context()); delPostsErr != nil {
			RespondWithError(writer, http.StatusInternalServerError, "error deleting posts.")
			return
		}
	} else {
		RespondWithError(writer, http.StatusMethodNotAllowed, "you can only reset data in dev mode")
	}
}
