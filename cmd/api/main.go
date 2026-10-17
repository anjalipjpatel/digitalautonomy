package main

import (
	"log"
	"net/http"
	"os"

	"digitalautonomy/internal/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/health", api.HealthHandler)
	http.HandleFunc("/payments", api.NewPaymentHandler(os.Getenv("PAYMENT_API_URL")))

	log.Println("Server running on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
