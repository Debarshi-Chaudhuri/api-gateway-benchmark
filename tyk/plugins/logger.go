// File: tyk/plugins/logger.go
package main

import (
	"fmt"
	"net/http"
	"time"
)

// RequestLogger logs the incoming request details
func RequestLogger(w http.ResponseWriter, r *http.Request) {
	// Log basic request info
	fmt.Printf("[TYK-REQUEST] %s | %s %s | Client: %s | Headers: %d\n",
		time.Now().Format(time.RFC3339),
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		len(r.Header))
}

// ResponseLogger logs the response details
func ResponseLogger(w http.ResponseWriter, res *http.Response, r *http.Request) {
	// Log basic response info
	fmt.Printf("[TYK-RESPONSE] %s | Status: %d | Size: %d | Path: %s\n",
		time.Now().Format(time.RFC3339),
		res.StatusCode,
		res.ContentLength,
		r.URL.Path)
}

// Export the plugin functions
func main() {}
