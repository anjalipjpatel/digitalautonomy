package user

import (
	"context"
	"errors"
	"testing"
	"time"

	passwordhash "digitalautonomy/internal/hash"

	"github.com/jackc/pgx/v5"
)

type serviceTestDB struct {
	args []any
	row  pgx.Row
}

func (db *serviceTestDB) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	db.args = args
	return db.row
}

type serviceTestRow func(dest ...any) error

func (row serviceTestRow) Scan(dest ...any) error {
	return row(dest...)
}

func TestServiceCreate(t *testing.T) {
	createdAt := time.Date(2026, time.November, 14, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		email    string
		userName string
		password string
		role     string
		wantErr  bool
	}{
		{
			name:     "empty email",
			email:    " ",
			userName: "Person",
			password: "secret",
			role:     string(RoleAdmin),
			wantErr:  true,
		},
		{
			name:     "empty name",
			email:    "person@example.com",
			userName: " ",
			password: "secret",
			role:     string(RoleAdmin),
			wantErr:  true,
		},
		{
			name:     "empty password",
			email:    "person@example.com",
			userName: "Person",
			role:     string(RoleAdmin),
			wantErr:  true,
		},
		{
			name:     "malformed email",
			email:    "not-an-email",
			userName: "Person",
			password: "secret",
			role:     string(RoleAdmin),
			wantErr:  true,
		},
		{
			name:     "email with display name",
			email:    "Person <person@example.com>",
			userName: "Person",
			password: "secret",
			role:     string(RoleAdmin),
			wantErr:  true,
		},
		{
			name:     "unsupported role",
			email:    "person@example.com",
			userName: "Person",
			password: "secret",
			role:     "owner",
			wantErr:  true,
		},
		{
			name:     "valid input",
			email:    "  PERSON@example.com ",
			userName: "  Test Person  ",
			password: "secret",
			role:     string(RoleDelegate),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := &serviceTestDB{
				row: serviceTestRow(func(dest ...any) error {
					*dest[0].(*string) = "user-123"
					*dest[1].(*string) = "person@example.com"
					*dest[2].(*string) = "Test Person"
					*dest[3].(*string) = string(RoleDelegate)
					*dest[4].(*time.Time) = createdAt
					return nil
				}),
			}
			service := NewService(NewRepository(db))

			got, err := service.Create(context.Background(), tc.email, tc.userName, tc.password, tc.role)

			// Handle error states
			if tc.wantErr {
				if !errors.Is(err, ErrInvalidInput) {
					t.Errorf("Create() error = %v, want %v", err, ErrInvalidInput)
				}
				if got != nil {
					t.Errorf("Create() user = %#v, want nil", got)
				}
				if len(db.args) != 0 {
					t.Errorf("database called with arguments %#v for invalid input", db.args)
				}
				return
			}

			// Handle valid state
			if err != nil {
				t.Fatalf("Create() error = %v, want nil", err)
			}
			want := &User{
				ID:        "user-123",
				Email:     "person@example.com",
				Name:      "Test Person",
				Role:      RoleDelegate,
				CreatedAt: createdAt,
			}
			if got.ID != want.ID || got.Email != want.Email || got.Name != want.Name ||
				got.Role != want.Role || !got.CreatedAt.Equal(want.CreatedAt) {
				t.Errorf("Create() user = %#v, want %#v", got, want)
			}

			if len(db.args) != 1 {
				t.Fatalf("database received %d argument groups, want 1", len(db.args))
			}
			args, ok := db.args[0].(pgx.NamedArgs)
			if !ok {
				t.Fatalf("database argument type = %T, want pgx.NamedArgs", db.args[0])
			}
			if args["email"] != "person@example.com" {
				t.Errorf("email argument = %v, want %q", args["email"], "person@example.com")
			}
			if args["name"] != "Test Person" {
				t.Errorf("name argument = %v, want %q", args["name"], "Test Person")
			}
			if args["role"] != string(RoleDelegate) {
				t.Errorf("role argument = %v, want %q", args["role"], RoleDelegate)
			}
			encodedPassword, ok := args["hashed_password"].(string)
			if !ok {
				t.Fatalf("hashed_password argument type = %T, want string", args["hashed_password"])
			}
			if encodedPassword == tc.password || encodedPassword == "" {
				t.Fatalf("hashed_password = %q, want a non-empty password hash", encodedPassword)
			}
			valid, err := passwordhash.Verify(tc.password, encodedPassword)
			if err != nil {
				t.Fatalf("verifying password hash: %v", err)
			}
			if !valid {
				t.Error("hashed password does not verify against the original password")
			}
		})
	}
}
