// @title Digital Autonomy API
// @version 1.0
// @description API for the digital autonomy prototype.
// @BasePath /
package main

import (
	"log"
	"net/http"
	"os"

	_ "digitalautonomy/docs"
	"digitalautonomy/internal/api"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.Handle("/swagger/", httpSwagger.WrapHandler) // For Swagger UI
	http.HandleFunc("/health", api.HealthHandler)

	log.Println("Server running on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
