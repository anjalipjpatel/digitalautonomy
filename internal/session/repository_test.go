package session

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type repositoryStubDB struct {
	query   string
	args    []any
	execErr error
	scan    func(dest ...any) error
}

func (db *repositoryStubDB) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	db.query = query
	db.args = args
	return repositoryStubRow{scan: db.scan}
}

func (db *repositoryStubDB) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	db.query = query
	db.args = args
	return pgconn.NewCommandTag("INSERT 0 1"), db.execErr
}

type repositoryStubRow struct {
	scan func(dest ...any) error
}

func (row repositoryStubRow) Scan(dest ...any) error {
	return row.scan(dest...)
}

func namedArgs(t *testing.T, args []any) pgx.NamedArgs {
	t.Helper()
	if len(args) != 1 {
		t.Fatalf("got %d argument groups, want 1", len(args))
	}
	got, ok := args[0].(pgx.NamedArgs)
	if !ok {
		t.Fatalf("argument group has type %T, want pgx.NamedArgs", args[0])
	}
	return got
}

func TestRepositoryCreate(t *testing.T) {
	dbErr := errors.New("database unavailable")
	tests := []struct {
		name    string
		execErr error
		wantErr bool
	}{
		{name: "inserts token hash"},
		{name: "database error", execErr: dbErr, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &repositoryStubDB{execErr: tt.execErr}
			repo := NewRepository(db)

			before := time.Now().UTC()
			token, expiresAt, err := repo.Create(context.Background(), "user-123")
			after := time.Now().UTC()

			if tt.wantErr {
				if err == nil || !errors.Is(err, dbErr) {
					t.Fatalf("Create() error = %v, want wrapped database error", err)
				}
				if token != "" || expiresAt != nil {
					t.Fatalf("Create() = (%q, %v), want empty token and nil expiry on error", token, expiresAt)
				}
				return
			}
			if err != nil {
				t.Fatalf("Create() error = %v, want nil", err)
			}
			if len(token) != 43 {
				t.Errorf("token length = %d, want 43", len(token))
			}
			if expiresAt == nil {
				t.Fatal("Create() expiry is nil")
			}
			if expiresAt.Before(before.Add(time.Hour)) || expiresAt.After(after.Add(time.Hour)) {
				t.Errorf("expiry %v is not approximately one hour from creation", expiresAt)
			}

			if db.query != queryCreateSession {
				t.Errorf("Exec() query = %q, want %q", db.query, queryCreateSession)
			}
			gotArgs := namedArgs(t, db.args)
			if gotArgs["user_id"] != "user-123" {
				t.Errorf("user_id = %v, want user-123", gotArgs["user_id"])
			}
			if gotArgs["expires_at"] != *expiresAt {
				t.Errorf("expires_at = %v, want %v", gotArgs["expires_at"], *expiresAt)
			}
			hash, ok := gotArgs["token_hash"].([]byte)
			if !ok {
				t.Fatalf("token_hash has type %T, want []byte", gotArgs["token_hash"])
			}
			wantHash := hashToken(token)
			if !reflect.DeepEqual(hash, wantHash[:]) {
				t.Errorf("stored token hash = %x, want %x", hash, wantHash)
			}
			if _, found := gotArgs["token"]; found {
				t.Error("raw token was passed to the database")
			}
		})
	}
}

func TestRepositoryGetByToken(t *testing.T) {
	expiresAt := time.Now().UTC().Add(time.Hour)
	dbErr := errors.New("database unavailable")
	tests := []struct {
		name    string
		scanErr error
		wantErr error
	}{
		{name: "found"},
		{name: "not found", scanErr: pgx.ErrNoRows, wantErr: pgx.ErrNoRows},
		{name: "database error", scanErr: dbErr, wantErr: dbErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &repositoryStubDB{
				scan: func(dest ...any) error {
					if tt.scanErr != nil {
						return tt.scanErr
					}
					*dest[0].(*string) = "user-123"
					*dest[1].(**time.Time) = &expiresAt
					*dest[2].(*bool) = false
					return nil
				},
			}
			repo := NewRepository(db)
			const token = "test-session-token"

			got, err := repo.GetByToken(context.Background(), token)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("GetByToken() error = %v, want matching %v", err, tt.wantErr)
				}
				if got != nil {
					t.Fatalf("GetByToken() = %#v, want nil on error", got)
				}
			} else {
				if err != nil {
					t.Fatalf("GetByToken() error = %v, want nil", err)
				}
				if got.userID != "user-123" || got.token != token || got.expiresAt == nil ||
					!got.expiresAt.Equal(expiresAt) || got.revoked {
					t.Errorf("GetByToken() = %#v, unexpected session values", got)
				}
			}

			if db.query != queryGetByToken {
				t.Errorf("QueryRow() query = %q, want %q", db.query, queryGetByToken)
			}
			gotArgs := namedArgs(t, db.args)
			gotHash, ok := gotArgs["token_hash"].([]byte)
			if !ok {
				t.Fatalf("token_hash has type %T, want []byte", gotArgs["token_hash"])
			}
			wantHash := hashToken(token)
			if !reflect.DeepEqual(gotHash, wantHash[:]) {
				t.Errorf("token hash = %x, want %x", gotHash, wantHash)
			}
		})
	}
}
