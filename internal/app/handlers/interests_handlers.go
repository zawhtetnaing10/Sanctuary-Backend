package handlers

import (
	"net/http"
	"time"
)

type interestResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Get All Interests
func (cfg *ApiConfig) GetAllInterests(writer http.ResponseWriter, request *http.Request) {
	interests, err := cfg.Db.GetAllInterests(request.Context())
	if err != nil {
		RespondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	response := []interestResponse{}

	for _, interestDb := range interests {
		interestResponse := interestResponse{
			ID:        interestDb.ID,
			Name:      interestDb.Name,
			CreatedAt: interestDb.CreatedAt.Time,
			UpdatedAt: interestDb.UpdatedAt.Time,
		}
		response = append(response, interestResponse)
	}

	RespondWithJson(writer, http.StatusOK, response)
}
