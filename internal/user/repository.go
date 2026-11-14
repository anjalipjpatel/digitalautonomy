package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailTaken = errors.New("email already registered.")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, params CreateUserParams) (*User, error) {
	var created User
	var role string

	err := r.db.QueryRow(ctx, queryCreateNewUser, pgx.NamedArgs{
		"email":           params.Email,
		"hashed_password": params.HashedPassword,
		"name":            params.Name,
		"role":            string(params.Role),
	}).
		Scan(&created.ID, &created.Email, &created.Name, &role, &created.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Email taken code
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("inserting user: %w", err)
	}

	created.Role = Role(role)
	return &created, nil
}
