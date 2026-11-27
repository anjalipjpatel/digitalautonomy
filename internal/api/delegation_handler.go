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
	GetPendingForDelegate(ctx context.Context, delegateUserID string) ([]delegation.Link, error)
	Accept(ctx context.Context, delegateUserID, linkID string) (*delegation.Link, error)
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
	primaryUserID, ok := h.authenticatedUserID(w, r)
	if !ok {
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

// Pending godoc
// @Summary List pending delegation links
// @Description Lists unexpired pending links addressed to the authenticated delegate.
// @Tags delegation
// @Produce json
// @Success 200 {array} delegation.Link "Pending delegation links"
// @Failure 401 {string} string "Missing or invalid session"
// @Failure 403 {string} string "User is not a delegate"
// @Failure 500 {string} string "Could not load pending links"
// @Router /delegation-links/pending [get]
func (h *DelegationHandler) Pending(w http.ResponseWriter, r *http.Request) {
	delegateUserID, ok := h.authenticatedUserID(w, r)
	if !ok {
		return
	}

	links, err := h.links.GetPendingForDelegate(r.Context(), delegateUserID)
	switch {
	case errors.Is(err, delegation.ErrUnauthorized):
		http.Error(w, "only delegates can view pending delegation links", http.StatusForbidden)
		return
	case err != nil:
		http.Error(w, "could not load pending delegation links", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(links)
}

// Accept godoc
// @Summary Accept a delegation link
// @Description Accepts an unexpired pending delegation link belonging to the authenticated delegate.
// @Tags delegation
// @Produce json
// @Param id path string true "Delegation link ID"
// @Success 200 {object} delegation.Link "Delegation link accepted"
// @Failure 400 {string} string "Invalid link ID"
// @Failure 401 {string} string "Missing or invalid session"
// @Failure 403 {string} string "User is not a delegate"
// @Failure 404 {string} string "Pending link not found"
// @Failure 409 {string} string "Link is no longer pending or has expired"
// @Failure 500 {string} string "Could not accept delegation link"
// @Router /delegation-links/{id}/accept [post]
func (h *DelegationHandler) Accept(w http.ResponseWriter, r *http.Request) {
	delegateUserID, ok := h.authenticatedUserID(w, r)
	if !ok {
		return
	}

	link, err := h.links.Accept(r.Context(), delegateUserID, r.PathValue("id"))
	switch {
	case errors.Is(err, delegation.ErrInvalidInput):
		http.Error(w, "invalid delegation link ID", http.StatusBadRequest)
		return
	case errors.Is(err, delegation.ErrUnauthorized):
		http.Error(w, "only delegates can accept delegation links", http.StatusForbidden)
		return
	case errors.Is(err, delegation.ErrLinkNotPending):
		http.Error(w, "delegation link is no longer pending or has expired", http.StatusConflict)
		return
	case err != nil:
		http.Error(w, "could not accept delegation link", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(link)
}

func (h *DelegationHandler) authenticatedUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		http.Error(w, "missing session cookie", http.StatusUnauthorized)
		return "", false
	}

	userID, err := h.sessions.GetUserIDByToken(r.Context(), cookie.Value)
	if errors.Is(err, session.ErrInvalidSession) ||
		errors.Is(err, session.ErrTokenExpired) ||
		errors.Is(err, session.ErrTokenRevoked) {
		http.Error(w, "unauthorised", http.StatusUnauthorized)
		return "", false
	}
	if err != nil {
		http.Error(w, "could not validate session", http.StatusInternalServerError)
		return "", false
	}
	return userID, true
}
