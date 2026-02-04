package httperr

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func RespondWithError(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errResp := ErrorResponse{
		Message:    err.Error(),
		StatusCode: statusCode,
	}

	json.NewEncoder(w).Encode(errResp)
}

func WriteJSON(w http.ResponseWriter, data interface{}, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}
