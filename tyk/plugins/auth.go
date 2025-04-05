// File: tyk/plugins/logger.go
package main

import (
	"api-gateway-benchmark/tyk/helper"
	"fmt"
	"net/http"
)

func Auth(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Auth")
	helper.Hello()
}
