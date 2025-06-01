package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/app"
)

// Helper function to respond with json
func RespondWithJson(writer http.ResponseWriter, code int, payload any) {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		RespondWithError(writer, http.StatusInternalServerError, err.Error())
		return
	}

	writer.Header().Set(app.CONTENT_TYPE, app.APPLICATION_JSON)
	writer.WriteHeader(code)
	writer.Write(payloadData)
}

// Helper function to respond with error
func RespondWithError(writer http.ResponseWriter, code int, msg string) {
	type errorVals struct {
		Error string `json:"error"`
	}

	errStruct := errorVals{
		Error: msg,
	}

	errData, err := json.Marshal(errStruct)
	if err != nil {
		/// If Encoding fails, sent the server error as plain text
		writer.Header().Set(app.CONTENT_TYPE, app.TEXT_PLAIN)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.Header().Set(app.CONTENT_TYPE, app.APPLICATION_JSON)
	writer.WriteHeader(code)
	writer.Write(errData)
}
