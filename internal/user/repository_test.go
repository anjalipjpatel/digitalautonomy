package user

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Captures the repository query
type stubUserDB struct {
	query string
	args  []any
	row   pgx.Row
}

func (db *stubUserDB) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	db.query = query
	db.args = args
	return db.row
}

// Simulates values returned by scanning rows
type stubRow struct {
	scan func(dest ...any) error
}

func (row stubRow) Scan(dest ...any) error {
	return row.scan(dest...)
}

func TestRepositoryCreate(t *testing.T) {
	createdAt := time.Date(2026, time.November, 14, 10, 0, 0, 0, time.UTC)
	params := CreateUserParams{
		Email:          "person@example.com",
		HashedPassword: "hashed-password",
		Name:           "Test Person",
		Role:           RoleAdmin,
	}
	databaseErr := errors.New("database unavailable")

	tests := []struct {
		name        string
		scanErr     error
		wantUser    *User
		wantErr     error
		wantErrText string
	}{
		{
			name: "created",
			wantUser: &User{
				ID:        "user-123",
				Email:     params.Email,
				Name:      params.Name,
				Role:      params.Role,
				CreatedAt: createdAt,
			},
		},
		{
			name:    "email already exists",
			scanErr: &pgconn.PgError{Code: "23505"},
			wantErr: ErrEmailTaken,
		},
		{
			name:        "database error",
			scanErr:     databaseErr,
			wantErr:     databaseErr,
			wantErrText: "inserting user: database unavailable",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := &stubUserDB{
				row: stubRow{scan: func(dest ...any) error {
					if tc.scanErr != nil {
						return tc.scanErr
					}
					*dest[0].(*string) = "user-123"
					*dest[1].(*string) = params.Email
					*dest[2].(*string) = params.Name
					*dest[3].(*string) = string(params.Role)
					*dest[4].(*time.Time) = createdAt
					return nil
				}},
			}
			repo := NewRepository(db)

			got, err := repo.Create(context.Background(), params)

			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("Create() error = %v, want nil", err)
				}
				if !reflect.DeepEqual(got, tc.wantUser) {
					t.Errorf("Create() = %#v, want %#v", got, tc.wantUser)
				}
			} else {
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("Create() error = %v, want error matching %v", err, tc.wantErr)
				}
				if tc.wantErrText != "" && err.Error() != tc.wantErrText {
					t.Errorf("Create() error = %q, want %q", err, tc.wantErrText)
				}
				if got != nil {
					t.Errorf("Create() user = %#v, want nil on error", got)
				}
			}

			// Check SQL + Named Args
			if db.query != queryCreateNewUser {
				t.Errorf("QueryRow() query = %q, want %q", db.query, queryCreateNewUser)
			}
			if len(db.args) != 1 {
				t.Fatalf("QueryRow() received %d argument groups, want 1", len(db.args))
			}
			gotArgs, ok := db.args[0].(pgx.NamedArgs)
			if !ok {
				t.Fatalf("QueryRow() argument type = %T, want pgx.NamedArgs", db.args[0])
			}
			wantArgs := pgx.NamedArgs{
				"email":           params.Email,
				"hashed_password": params.HashedPassword,
				"name":            params.Name,
				"role":            string(params.Role),
			}
			if !reflect.DeepEqual(gotArgs, wantArgs) {
				t.Errorf("QueryRow() named arguments = %#v, want %#v", gotArgs, wantArgs)
			}
		})
	}
}
