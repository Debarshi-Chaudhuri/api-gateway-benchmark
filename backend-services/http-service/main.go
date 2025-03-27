package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

type Response struct {
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

var (
	requestCount uint64
	serviceReady = true
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	// Set up routes
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/api/data", handleData)
	http.HandleFunc("/api/protected", handleProtected)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/toggle-ready", handleToggleReady)

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("HTTP service starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if !serviceReady {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	id := atomic.AddUint64(&requestCount, 1)

	// Simulate some load
	time.Sleep(50 * time.Millisecond)

	resp := Response{
		Status:    "success",
		Message:   "Welcome to the API",
		RequestID: fmt.Sprintf("req-%d", id),
		Timestamp: time.Now(),
	}

	sendJSON(w, resp)
}

func handleData(w http.ResponseWriter, r *http.Request) {
	if !serviceReady {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	id := atomic.AddUint64(&requestCount, 1)

	// Get delay parameter for resilience testing
	delayStr := r.URL.Query().Get("delay")
	if delayStr != "" {
		delay, err := strconv.Atoi(delayStr)
		if err == nil && delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}

	data := map[string]interface{}{
		"items": []string{"item1", "item2", "item3"},
		"count": 3,
	}

	resp := Response{
		Status:    "success",
		Message:   "Data retrieved successfully",
		RequestID: fmt.Sprintf("req-%d", id),
		Data:      data,
		Timestamp: time.Now(),
	}

	sendJSON(w, resp)
}

func handleProtected(w http.ResponseWriter, r *http.Request) {
	if !serviceReady {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	id := atomic.AddUint64(&requestCount, 1)

	// Check for authorization header - in a real app, this would be handled by the gateway
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		resp := Response{
			Status:    "error",
			Message:   "Authorization required",
			RequestID: fmt.Sprintf("req-%d", id),
			Timestamp: time.Now(),
		}
		sendJSON(w, resp)
		return
	}

	// Log headers for debugging
	log.Printf("Request headers: %v", r.Header)

	resp := Response{
		Status:    "success",
		Message:   "Access granted to protected resource",
		RequestID: fmt.Sprintf("req-%d", id),
		Timestamp: time.Now(),
		Data: map[string]string{
			"user_id":    "user-123",
			"user_email": "user@example.com",
		},
	}

	sendJSON(w, resp)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if !serviceReady {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	resp := Response{
		Status:    "success",
		Message:   "Service is healthy",
		Timestamp: time.Now(),
	}

	sendJSON(w, resp)
}

func handleToggleReady(w http.ResponseWriter, r *http.Request) {
	// Toggle service readiness for resilience testing
	serviceReady = !serviceReady

	status := "ready"
	if !serviceReady {
		status = "not ready"
	}

	resp := Response{
		Status:    "success",
		Message:   fmt.Sprintf("Service is now %s", status),
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
