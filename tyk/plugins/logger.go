// File: tyk/plugins/logger.go
package main

import (
	"fmt"
	"net/http"
)

// RequestLogger logs the incoming request details
func RequestLogger(w http.ResponseWriter, r *http.Request) {
	fmt.Println("RequestLogger")
}

// ResponseLogger logs the response details
func ResponseLogger(w http.ResponseWriter, res *http.Response, r *http.Request) {
	fmt.Println("ResponseLogger")

}

func main() {}
