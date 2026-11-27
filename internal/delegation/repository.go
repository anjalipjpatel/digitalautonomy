package delegation

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

//go:embed queries/create_delegation_link.sql
var queryCreateDelegationLink string

var ErrDuplicateLink = errors.New("delegation link already exists")

type repositoryDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct {
	db repositoryDB
}

func NewRepository(db repositoryDB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, params CreateParams) (*Link, error) {
	var created Link
	var status string

	err := r.db.QueryRow(ctx, queryCreateDelegationLink, pgx.NamedArgs{
		"primary_user_id":  params.PrimaryUserID,
		"delegate_user_id": params.DelegateUserID,
		"valid_until":      params.ValidUntil,
		"status":           "pending",
		"permissions":      params.Permissions,
	}).Scan(
		&created.ID,
		&created.PrimaryUserID,
		&created.DelegateUserID,
		&created.ValidFrom,
		&created.ValidUntil,
		&status,
		&created.Permissions,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateLink
		}
		return nil, fmt.Errorf("creating delegation link: %w", err)
	}

	created.Status = status
	return &created, nil
}
