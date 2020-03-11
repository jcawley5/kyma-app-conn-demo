package utils

import (
	"encoding/json"
	"net/http"
)

//ReturnError -
func ReturnError(errMsg string, w http.ResponseWriter) {
	response := map[string]string{"error": errMsg}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

//ReturnSuccess -
func ReturnSuccess(message string, w http.ResponseWriter) {
	response := map[string]string{"message": message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
