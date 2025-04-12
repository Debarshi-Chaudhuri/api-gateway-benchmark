package main

import (
	"time"

	"github.com/gorilla/mux"
)

type Response struct {
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	initialize(router.PathPrefix("").Subrouter())

	startServer(router)
}

func initialize(router *mux.Router) {
	initializeMiddleware(router)
	initializeRoutes(router)
}

func initializeMiddleware(router *mux.Router) {
	router.Use(loggingMiddleware)
}

func initializeRoutes(router *mux.Router) {
	router.HandleFunc("/health", handleHealth).Methods("GET")
}
