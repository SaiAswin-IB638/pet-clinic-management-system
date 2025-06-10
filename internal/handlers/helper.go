package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

func (h *handlerService) respond(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	if data != nil {
		switch errData := data.(type) {
		case error:
			jsonError := map[string]string{"error": errData.Error()}
			w.WriteHeader(statusCode)
			if err := json.NewEncoder(w).Encode(jsonError); err != nil {
				http.Error(w, "{\"error\": \"Failed to encode error response\"}", http.StatusInternalServerError)
			}
		default:
			w.WriteHeader(statusCode)
			if err := json.NewEncoder(w).Encode(data); err != nil {
				http.Error(w, "{\"error\": \"Failed to encode response\"}", http.StatusInternalServerError)
			}
		}
	} else {
		w.WriteHeader(statusCode)
	}
}

func (h *handlerService) petIDValidate(vars *map[string]string) (uint, error) {
	petIDStr, ok := (*vars)["id"]
	if !ok {
		return 0, errors.New("pet id not provided")
	}
	petID64, err := strconv.ParseUint(petIDStr, 10, 32)
	petID := uint(petID64)
	if err != nil {
		return 0, errors.New("pet id is not valid")
	}
	return petID, nil
}

func (h *handlerService) appointmentIDValidate(vars *map[string]string) (uint, error) {
	appointmentIDStr, ok := (*vars)["id"]
	if !ok {
		return 0, errors.New("appointment id not provided")
	}
	appointmentID64, err := strconv.ParseUint(appointmentIDStr, 10, 32)
	appointmentID := uint(appointmentID64)
	if err != nil {
		return 0, errors.New("appointment id is not valid")
	}
	return appointmentID, nil
}
