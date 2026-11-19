package session

import (
	"context"
	"digitalautonomy/internal/user"
	"errors"
	"time"
)

var ErrTokenRevoed = errors.New("token revoked")
var ErrInvalidToken = errors.New("invalid token")

type Service struct {
	sessions *Repository
}

func NewService(sessions *Repository) *Service {
	return &Service{
		sessions: sessions,
	}
}

// Create
func (s *Service) Create(ctx context.Context, user user.User) (*tokenWithTime, error) {
	// Call repository to generate token and insert
	return s.sessions.Create(ctx, user.ID)
}

// Validate
func (s *Service) Validate(ctx context.Context, token string) (bool, error) {
	session, err := s.sessions.GetByToken(ctx, token)
	if err != nil {
		// Repo-level errors - bubble straight up
		return false, err
	}

	if session.revoked {
		return false, ErrTokenRevoed
	}

	if session.token.expiresAt.Before(time.Now()) {
		return false, ErrInvalidToken
	}

	return true, nil
}

// create session during login - auth_handler.go recieve session creator, create session adn set HTTPonly cookie
// finish user_handler.go session-validator field
// GetByID to user service/repositry to return only public profile
// /Me validate cookie, load profile and return 200 ok, 401 for expreid/invalid, 500 for unexpected db erros
// main.go: constructions ession repo/service, pass dependeies, and register GET /me
