package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type serviceStubRepository struct {
	gotSession *session
	getErr     error
}

func (repo *serviceStubRepository) Create(context.Context, string) (string, *time.Time, error) {
	return "", nil, nil
}

func (repo *serviceStubRepository) GetByToken(context.Context, string) (*session, error) {
	return repo.gotSession, repo.getErr
}

func TestGetUserIDByToken(t *testing.T) {
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)
	databaseErr := errors.New("database unavailable")

	tests := []struct {
		name          string
		session       *session
		repositoryErr error
		wantUserID    string
		wantErr       error
	}{
		{
			name: "valid session",
			session: &session{
				userID:    "user-123",
				expiresAt: &future,
			},
			wantUserID: "user-123",
		},
		{
			name:          "unknown token",
			repositoryErr: pgx.ErrNoRows,
			wantErr:       ErrInvalidSession,
		},
		{
			name: "expired session",
			session: &session{
				userID:    "user-123",
				expiresAt: &past,
			},
			wantErr: ErrTokenExpired,
		},
		{
			name: "missing expiry",
			session: &session{
				userID: "user-123",
			},
			wantErr: ErrTokenExpired,
		},
		{
			name:    "nil session",
			wantErr: ErrInvalidSession,
		},
		{
			name: "revoked session",
			session: &session{
				userID:    "user-123",
				expiresAt: &future,
				revoked:   true,
			},
			wantErr: ErrTokenRevoked,
		},
		{
			name:          "repository failure",
			repositoryErr: databaseErr,
			wantErr:       databaseErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(&serviceStubRepository{
				gotSession: tt.session,
				getErr:     tt.repositoryErr,
			})

			gotID, err := service.GetUserIDByToken(context.Background(), "test-token")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("GetUserIDByToken() error = %v, want matching %v", err, tt.wantErr)
			}
			if gotID != tt.wantUserID {
				t.Errorf("GetUserIDByToken() userID = %q, want %q", gotID, tt.wantUserID)
			}
		})
	}
}
