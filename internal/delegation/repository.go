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

//go:embed queries/get_pending_delegation_links.sql
var queryGetPendingDelegationLinks string

//go:embed queries/accept_delegation_link.sql
var queryAcceptDelegationLink string

var (
	ErrDuplicateLink = errors.New("delegation link already exists")
	ErrLinkNotFound  = errors.New("pending delegation link not found")
)

type repositoryDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
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

func (r *Repository) GetPendingForDelegate(ctx context.Context, delegateUserID string) ([]Link, error) {
	rows, err := r.db.Query(ctx, queryGetPendingDelegationLinks, pgx.NamedArgs{
		"delegate_user_id": delegateUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("querying pending delegation links: %w", err)
	}
	defer rows.Close()

	links := make([]Link, 0)
	for rows.Next() {
		link, err := scanLink(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning pending delegation link: %w", err)
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating pending delegation links: %w", err)
	}
	return links, nil
}

func (r *Repository) Accept(ctx context.Context, delegateUserID, linkID string) (*Link, error) {
	var link Link
	var status string

	err := r.db.QueryRow(ctx, queryAcceptDelegationLink, pgx.NamedArgs{
		"id":               linkID,
		"delegate_user_id": delegateUserID,
	}).Scan(
		&link.ID,
		&link.PrimaryUserID,
		&link.DelegateUserID,
		&link.ValidFrom,
		&link.ValidUntil,
		&status,
		&link.Permissions,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLinkNotFound
		}
		return nil, fmt.Errorf("accepting delegation link: %w", err)
	}
	link.Status = status
	return &link, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanLink(row rowScanner) (Link, error) {
	var link Link
	var status string
	err := row.Scan(
		&link.ID,
		&link.PrimaryUserID,
		&link.DelegateUserID,
		&link.ValidFrom,
		&link.ValidUntil,
		&status,
		&link.Permissions,
	)
	link.Status = status
	return link, err
}
