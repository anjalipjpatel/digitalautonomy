package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"digitalautonomy/internal/user"
)

type userAuthenticator interface {
	Login(ctx context.Context, email, password string) (*user.User, error)
}

type sessionCreator interface {
	Create(ctx context.Context, userID string) (token string, expiresAt *time.Time, err error)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthHandler struct {
	users    userAuthenticator
	sessions sessionCreator
	secure   bool
}

func NewAuthHandler(users userAuthenticator, sessions sessionCreator, secureCookies bool) *AuthHandler {
	return &AuthHandler{
		users:    users,
		sessions: sessions,
		secure:   secureCookies,
	}
}

// Login godoc
// @Summary Login to system
// @Description Validate login credentials and authorise access to the system.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login details"
// @Success 200 {object} user.User "Login successful"
// @Failure 400 {string} string "Invalid JSON body"
// @Failure 401 {string} string "Invalid email or password"
// @Failure 500 {string} string "Could not log in"
// @Router /login [post]
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

	// Generate session
	token, expiresAt, err := h.sessions.Create(r.Context(), loggedIn.ID)
	if err != nil || expiresAt == nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Expires:  *expiresAt,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(loggedIn)
}
