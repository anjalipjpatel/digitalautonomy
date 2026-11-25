package user

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:embed queries/create_user.sql
var queryCreateNewUser string

//go:embed queries/login.sql
var queryLogin string

//go:embed queries/get_by_user_id.sql
var queryGetByUserID string

var ErrEmailTaken = errors.New("email already registered.")

type userDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct {
	db userDB
}

func NewRepository(db userDB) *Repository {
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

func (r *Repository) GetByEmail(ctx context.Context, email string) (*userWithPassword, error) {
	var found userWithPassword
	var role string

	err := r.db.QueryRow(ctx, queryLogin, pgx.NamedArgs{
		"email": email,
	}).Scan(
		&found.ID,
		&found.Email,
		&found.Name,
		&role,
		&found.CreatedAt,
		&found.hashedPassword,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("no matching user found: %w", err)
		}
		return nil, fmt.Errorf("looking up user by email: %w", err)
	}

	found.Role = Role(role)
	return &found, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	var found User
	var role string
	err := r.db.QueryRow(ctx, queryGetByUserID, pgx.NamedArgs{
		"id": id,
	}).Scan(
		&found.ID,
		&found.Email,
		&found.Name,
		&role,
		&found.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("no matching user found: %w", err)
		}
		return nil, fmt.Errorf("looking up user by email: %w", err)
	}

	found.Role = Role(role)
	return &found, nil
}
