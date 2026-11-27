package api

import (
	"context"
	"digitalautonomy/internal/delegation"
	"digitalautonomy/internal/session"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type delegationCreator interface {
	Create(ctx context.Context, primaryUserID string, params delegation.CreateParams) (*delegation.Link, error)
}

type delegationSessionValidator interface {
	GetUserIDByToken(ctx context.Context, token string) (string, error)
}

type createDelegationLinkRequest struct {
	DelegateEmail string     `json:"delegate_email" binding:"required"`
	Permissions   []string   `json:"permissions" binding:"required"`
	ValidUntil    *time.Time `json:"valid_until"`
}

type DelegationHandler struct {
	sessions delegationSessionValidator
	links    delegationCreator
}

func NewDelegationHandler(sessions delegationSessionValidator, links delegationCreator) *DelegationHandler {
	return &DelegationHandler{sessions: sessions, links: links}
}

// Link godoc
// @Summary Create a delegation link
// @Description Creates a pending delegation link from the authenticated primary user to a delegate identified by email.
// @Tags delegation
// @Accept json
// @Produce json
// @Param request body createDelegationLinkRequest true "Delegate email, permissions, and optional expiry"
// @Success 201 {object} delegation.Link "Delegation link created"
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Missing or invalid session"
// @Failure 403 {string} string "User is not allowed to create delegation links"
// @Failure 404 {string} string "Delegate not found"
// @Failure 409 {string} string "Delegation link already exists"
// @Failure 500 {string} string "Could not create delegation link"
// @Router /delegation-links [post]
func (h *DelegationHandler) Link(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		http.Error(w, "missing session cookie", http.StatusUnauthorized)
		return
	}

	primaryUserID, err := h.sessions.GetUserIDByToken(r.Context(), cookie.Value)
	if errors.Is(err, session.ErrInvalidSession) ||
		errors.Is(err, session.ErrTokenExpired) ||
		errors.Is(err, session.ErrTokenRevoked) {
		http.Error(w, "unauthorised", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "could not validate session", http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var req createDelegationLinkRequest
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		http.Error(w, "request must contain a single JSON object", http.StatusBadRequest)
		return
	}

	created, err := h.links.Create(r.Context(), primaryUserID, delegation.CreateParams{
		DelegateEmail: req.DelegateEmail,
		Permissions:   req.Permissions,
		ValidUntil:    req.ValidUntil,
	})
	switch {
	case errors.Is(err, delegation.ErrInvalidInput):
		http.Error(w, "invalid delegation link details", http.StatusBadRequest)
		return
	case errors.Is(err, delegation.ErrUnauthorized):
		http.Error(w, "user is not allowed to create delegation links", http.StatusForbidden)
		return
	case errors.Is(err, delegation.ErrDelegateNotFound):
		http.Error(w, "delegate user not found", http.StatusNotFound)
		return
	case errors.Is(err, delegation.ErrDuplicateLink):
		http.Error(w, "delegation link already exists", http.StatusConflict)
		return
	case err != nil:
		http.Error(w, "could not create delegation link", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}
