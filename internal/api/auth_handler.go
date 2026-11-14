package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"digitalautonomy/internal/user"
)

type userAuthenticator interface {
	Login(ctx context.Context, email, password string) (*user.User, error)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthHandler struct {
	users userAuthenticator
}

func NewAuthHandler(users userAuthenticator) *AuthHandler {
	return &AuthHandler{users: users}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	loggedIn, err := h.users.Login(r.Context(), req.Email, req.Password)
	if errors.Is(err, user.ErrInvalidCredentials) {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "could not log in", http.StatusInternalServerError)
		return
	}
	if loggedIn == nil {
		http.Error(w, "could not log in", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(loggedIn)
}
