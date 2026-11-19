package session

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrInvalidSession = errors.New("invalid session")
var ErrTokenRevoked = errors.New("token revoked")
var ErrTokenExpired = errors.New("token expired")

type Service struct {
	sessions sessionRepository
}

type sessionRepository interface {
	Create(ctx context.Context, userID string) (string, *time.Time, error)
	GetByToken(ctx context.Context, token string) (*session, error)
	RevokeByToken(ctx context.Context, token string) error
}

func NewService(sessions sessionRepository) *Service {
	return &Service{
		sessions: sessions,
	}
}

// Create
func (s *Service) Create(ctx context.Context, userID string) (string, *time.Time, error) {
	// Call repository to generate token and insert
	return s.sessions.Create(ctx, userID)
}

// GetUserIDByToken
func (s *Service) GetUserIDByToken(ctx context.Context, token string) (string, error) {
	session, err := s.sessions.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrInvalidSession
		}
		// Repo-level errors - bubble straight up
		return "", err
	}

	if session == nil {
		return "", ErrInvalidSession
	}

	if session.revoked {
		return "", ErrTokenRevoked
	}

	if session.expiresAt == nil || !session.expiresAt.After(time.Now()) {
		return "", ErrTokenExpired
	}

	return session.userID, nil
}

func (s *Service) Revoke(ctx context.Context, token string) error {
	err := s.sessions.RevokeByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}
	return nil
}
