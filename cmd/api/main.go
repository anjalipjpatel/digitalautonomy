// @title Digital Autonomy API
// @version 1.0
// @description API for the digital autonomy prototype.
// @BasePath /
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	_ "digitalautonomy/docs"
	"digitalautonomy/internal/api"
	"digitalautonomy/internal/delegation"
	"digitalautonomy/internal/session"
	"digitalautonomy/internal/user"

	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {

	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create background context with 5 second timeout for operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	// Initialise database connection
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	userService := user.NewService(user.NewRepository(pool))
	sessionService := session.NewService(session.NewRepository(pool))
	secureCookies := os.Getenv("COOKIE_SECURE") == "true"
	userHandler := api.NewUserHandler(userService, sessionService, secureCookies)
	authHandler := api.NewAuthHandler(userService, sessionService, secureCookies)
	delegationService := delegation.NewService(delegation.NewRepository(pool), userService)
	delegationHandler := api.NewDelegationHandler(sessionService, delegationService)

	mux := http.NewServeMux()
	mux.Handle("/swagger/", httpSwagger.WrapHandler) // For Swagger UI
	mux.HandleFunc("/health", api.HealthHandler)
	mux.HandleFunc("POST /users", userHandler.Create) // Creating new users
	mux.HandleFunc("POST /login", authHandler.Login)  // Login mechanism
	mux.HandleFunc("POST /logout", userHandler.Logout)
	mux.HandleFunc("GET /me", userHandler.Me) // Return current user
	mux.HandleFunc("POST /delegation-links", delegationHandler.Link)
	mux.HandleFunc("GET /delegation-links", delegationHandler.List)
	mux.HandleFunc("GET /delegation-links/pending", delegationHandler.Pending)
	mux.HandleFunc("POST /delegation-links/{id}/accept", delegationHandler.Accept)

	log.Println("Server running on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
