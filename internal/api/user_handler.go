package api

import (
	"context"
	"digitalautonomy/internal/session"
	"digitalautonomy/internal/user"
	"encoding/json"
	"errors"
	"net/http"
)

type users interface {
	Create(ctx context.Context, email, name, password, role string) (*user.User, error)
	GetByID(ctx context.Context, id string) (*user.User, error)
}

type sessionValidator interface {
	GetUserIDByToken(ctx context.Context, token string) (string, error)
}

type createUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UserHandler struct {
	users    users
	sessions sessionValidator
}

func NewUserHandler(users users, sessions sessionValidator) *UserHandler {
	return &UserHandler{
		users:    users,
		sessions: sessions,
	}
}

// Create godoc
// @Summary Create a user
// @Description Creates a user account.
// @Tags users
// @Accept json
// @Produce json
// @Param request body createUserRequest true "User details"
// @Success 201 {object} user.User "User created"
// @Failure 400 {string} string "Invalid JSON body or user details"
// @Failure 409 {string} string "Email already registered"
// @Failure 500 {string} string "Could not create user"
// @Router /users [post]
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

// Me godoc
// @Summary Return the current user.
// @Tags users
// @Produce json
// @Success 200 {object} user.User "User"
// @Failure 401 {string} string "Invalid cookie"
// @Failure 404 {string} string "Unknown user"
// @Router /me [get]
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		http.Error(w, "missing session cookie", http.StatusUnauthorized)
		return
	}

	userID, err := h.sessions.GetUserIDByToken(r.Context(), cookie.Value)
	if errors.Is(err, session.ErrInvalidSession) || errors.Is(err, session.ErrTokenExpired) || errors.Is(err, session.ErrTokenRevoked) {
		http.Error(w, "unauthorised", http.StatusUnauthorized)
		return
	}

	if err != nil {
		http.Error(w, "could not validate session", http.StatusInternalServerError)
		return
	}

	profile, err := h.users.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "could not load user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(profile)
}
