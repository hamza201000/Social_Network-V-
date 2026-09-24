package handler

import (
	"backend/internal/models"
	"encoding/json"
	"net/http"
)

func SendJSON(w http.ResponseWriter, status int, data models.Res) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

