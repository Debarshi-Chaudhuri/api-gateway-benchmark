package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Response struct {
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// Set up routes
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/api/data", handleData)
	http.HandleFunc("/health", handleHealth)

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("HTTP service starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	resp := Response{
		Status:    "success",
		Message:   "Welcome to the API",
		Timestamp: time.Now(),
	}

	sendJSON(w, resp)
}

func handleData(w http.ResponseWriter, r *http.Request) {
	// Get delay parameter for testing
	delayParam := r.URL.Query().Get("delay")
	if delayParam != "" {
		delay, err := strconv.Atoi(delayParam)
		if err == nil && delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	} else {
		// Default delay
		time.Sleep(50 * time.Millisecond)
	}

	data := map[string]interface{}{
		"items": []string{"item1", "item2", "item3"},
		"count": 3,
	}

	resp := Response{
		Status:    "success",
		Message:   "Data retrieved successfully",
		Timestamp: time.Now(),
		Data:      data,
	}

	sendJSON(w, resp)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := Response{
		Status:    "success",
		Message:   "Service is healthy",
		Timestamp: time.Now(),
	}

	sendJSON(w, resp)
}

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}
