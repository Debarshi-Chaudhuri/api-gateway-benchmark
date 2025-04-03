package main

import (
	"encoding/json"
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

	// Create a more robust HTTP server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Log startup
	log.Printf("HTTP service starting on TLS port %s", port)

	// Print a message every 5 seconds to show the service is alive
	go func() {
		for {
			log.Printf("HTTP service is running with TLS...")
			time.Sleep(5 * time.Second)
		}
	}()

	// Check if TLS is enabled
	certFile := os.Getenv("TLS_CERT")
	keyFile := os.Getenv("TLS_KEY")

	// Start the server with TLS if certificates are provided
	if certFile != "" && keyFile != "" {
		log.Printf("Starting with TLS using cert: %s and key: %s", certFile, keyFile)
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
			log.Fatalf("Failed to start TLS server: %v", err)
		}
	} else {
		log.Printf("TLS certificates not provided, starting without TLS")
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)

	resp := Response{
		Status:    "success",
		Message:   "Welcome to the API (TLS Enabled)",
		Timestamp: time.Now(),
	}

	sendJSON(w, resp)
}

func handleData(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received data request: %s %s", r.Method, r.URL.Path)

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
	log.Printf("Received health check request")

	resp := Response{
		Status:    "success",
		Message:   "Service is healthy (TLS Enabled)",
		Timestamp: time.Now(),
	}

	sendJSON(w, resp)
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
