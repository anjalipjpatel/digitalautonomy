package api

import (
	"context"
	"digitalautonomy/internal/user"
	"encoding/json"
	"errors"
	"net/http"
)

type userCreator interface {
	Create(ctx context.Context, email, name, password, role string) (*user.User, error)
}

type createUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UserHandler struct {
	users userCreator
}

func NewUserHandler(users userCreator) *UserHandler {
	return &UserHandler{
		users: users,
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	created, err := h.users.Create(r.Context(), req.Email, req.Name, req.Password, req.Role)
	if errors.Is(err, user.ErrInvalidInput) {
		http.Error(w, "invalid user details", http.StatusBadRequest)
		return
	}

	if errors.Is(err, user.ErrEmailTaken) {
		http.Error(w, "email already registered", http.StatusConflict)
		return
	}

	if err != nil {
		http.Error(w, "could not create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}
