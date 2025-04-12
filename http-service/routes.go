package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	shouldReturn := returnIfBreak(r, w)
	if shouldReturn {
		return
	}

	// Handle delay parameter
	setDelayIfPresent(r)

	resp := Response{
		Status:    "success",
		Message:   "Service is healthy (Gorilla Mux)",
		Timestamp: time.Now(),
	}
	sendJSON(w, resp)
}

func returnIfBreak(r *http.Request, w http.ResponseWriter) bool {
	if breakParam := r.URL.Query().Get("break"); breakParam != "" {
		// Log the error
		log.Printf("Break parameter detected, returning internal server error")

		// Set HTTP status code to 500 Internal Server Error
		w.WriteHeader(http.StatusInternalServerError)

		// Return error response
		resp := Response{
			Status:    "error",
			Message:   "Internal server error",
			Timestamp: time.Now(),
		}

		sendJSON(w, resp)
		return true
	}
	return false
}

func setDelayIfPresent(r *http.Request) {
	delayParam := r.URL.Query().Get("delay")
	if delayParam != "" {
		if parsedDelay, err := strconv.Atoi(delayParam); err == nil && parsedDelay > 0 {
			delay := parsedDelay
			// Log the delay we're using
			log.Printf("Using delay of %d ms", delay)

			// Simulate processing delay
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	// Set CORS headers to allow requests from any origin
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
