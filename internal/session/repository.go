package session

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:embed queries/create_session.sql
var queryCreateSession string

//go:embed queries/get_by_token.sql
var queryGetByToken string

type sessionDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type Repository struct {
	db sessionDB
}

// todo
type session struct {
	token   tokenWithTime
	userID  string
	revoked bool
}

type tokenWithTime struct {
	token     string
	expiresAt *time.Time
}

func NewRepository(db sessionDB) *Repository {
	return &Repository{
		db: db,
	}
}

// Create
func (r *Repository) Create(ctx context.Context, userID string) (*tokenWithTime, error) {
	// Generate token
	rawToken := make([]byte, 32)
	if _, err := rand.Read(rawToken); err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}

	expiresAt := time.Now().UTC().Add(time.Hour) // Expiration = 1 hour after creation.
	token := tokenWithTime{
		token:     base64.RawURLEncoding.EncodeToString(rawToken),
		expiresAt: &(expiresAt),
	}

	// Insert to database
	_, err := r.db.Exec(ctx, queryCreateSession, pgx.NamedArgs{
		"token":      token,
		"user_id":    userID,
		"expires_at": expiresAt,
	})

	if err != nil {
		return nil, fmt.Errorf("inserting session: %w", err)
	}

	return &token, nil
}

// GetByToken
func (r *Repository) GetByToken(ctx context.Context, token string) (*session, error) {
	var found session

	err := r.db.QueryRow(ctx, queryGetByToken, pgx.NamedArgs{
		"token": token,
	}).Scan(
		&found.userID,
		&found.token.expiresAt,
		&found.revoked,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("no matching session found: %w", err)
		}
		return nil, fmt.Errorf("looking up session by token: %w", err)
	}

	found.token.token = token
	return &found, nil
}
