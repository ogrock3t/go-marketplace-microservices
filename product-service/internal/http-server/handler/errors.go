package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
)

const (
	ErrInternalServer = "internal server error"
	ErrInvalidJSON    = "invalid json"
)

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func parsePathID(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(r.PathValue(name), 10, 64)
}
