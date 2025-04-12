package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func startServer(router *mux.Router) {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		// TLSConfig: &tls.Config{
		// 	MinVersion: tls.VersionTLS12,
		// },
	}

	// Log startup
	// log.Printf("HTTP service starting on TLS port %s", port)

	// go func() {
	// 	for {
	// 		log.Printf("HTTP service is running with TLS...")
	// 		time.Sleep(5 * time.Second)
	// 	}
	// }()

	// Check if TLS is enabled
	// certFile := os.Getenv("TLS_CERT")
	// keyFile := os.Getenv("TLS_KEY")

	// Start the server with TLS if certificates are provided
	// if certFile != "" && keyFile != "" {
	// 	log.Printf("Starting with TLS using cert: %s and key: %s", certFile, keyFile)
	// 	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
	// 		log.Fatalf("Failed to start TLS server: %v", err)
	// 	}
	// } else {
	// 	log.Printf("TLS certificates not provided, starting without TLS")
	// 	if err := server.ListenAndServe(); err != nil {
	// 		log.Fatalf("Failed to start server: %v", err)
	// 	}
	// }

	// Log startup
	log.Printf("HTTP service starting on port %s with Gorilla Mux", port)

	// Start the server
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
