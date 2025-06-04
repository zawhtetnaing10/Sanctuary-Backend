package validators

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/database"
)

// Validate update user request
func ValidateUpdateUserRequest(request *http.Request, db *database.Queries) (int, error) {
	fullName := request.FormValue("full_name")
	userName := request.FormValue("user_name")
	dob := request.FormValue("dob")

	if fullName == "" {
		return http.StatusBadRequest, errors.New("full name is requred")
	}

	if userName == "" {
		return http.StatusBadRequest, errors.New("user name is requred")
	}

	if dob == "" {
		return http.StatusBadRequest, errors.New("dob is requred")
	}

	// Validate interest ids from request
	// Receive interests_id from the request and update user_has_interests table.
	interestsIds := request.MultipartForm.Value["ids"]
	parsedInterestIdsFromRequest := []int64{}
	// Parse the interestids to int64
	for _, interestId := range interestsIds {
		parsedId, parseErr := strconv.ParseInt(interestId, 10, 64)
		if parseErr != nil {
			return http.StatusBadRequest, parseErr
		}
		parsedInterestIdsFromRequest = append(parsedInterestIdsFromRequest, parsedId)
	}

	existingInterestIds, getExistingInterestIdsErr := db.GetExistingInterestIds(request.Context(), parsedInterestIdsFromRequest)
	if getExistingInterestIdsErr != nil {
		return http.StatusInternalServerError, getExistingInterestIdsErr
	}

	// Create a map to cross reference
	existingIdsMap := map[int64]bool{}
	for _, existingId := range existingInterestIds {
		existingIdsMap[existingId] = true
	}

	for _, interestId := range parsedInterestIdsFromRequest {
		_, ok := existingIdsMap[interestId]
		if !ok {
			return http.StatusNotFound, fmt.Errorf("interest id does not exist : %v", interestId)
		}
	}

	return 0, nil
}
