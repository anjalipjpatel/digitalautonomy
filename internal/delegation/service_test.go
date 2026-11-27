package delegation

import (
	"context"
	"errors"
	"testing"
	"time"

	"digitalautonomy/internal/user"

	"github.com/jackc/pgx/v5"
)

type serviceTestUsers struct {
	users map[string]*user.User
	err   error
}

func (s serviceTestUsers) GetByID(_ context.Context, id string) (*user.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	found, ok := s.users[id]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return found, nil
}

func (s serviceTestUsers) GetByEmail(_ context.Context, email string) (*user.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	for _, found := range s.users {
		if found.Email == email {
			return found, nil
		}
	}
	return nil, pgx.ErrNoRows
}

type serviceTestLinks struct {
	params CreateParams
	link   *Link
	err    error
	calls  int
}

func (s *serviceTestLinks) Create(_ context.Context, params CreateParams) (*Link, error) {
	s.calls++
	s.params = params
	return s.link, s.err
}

func TestServiceCreate(t *testing.T) {
	owner := &user.User{ID: "owner-id", Email: "owner@example.com", Role: user.RoleVulnerableOwner}
	delegate := &user.User{ID: "delegate-id", Email: "delegate@example.com", Role: user.RoleDelegate}
	otherOwner := &user.User{ID: "other-owner-id", Email: "other-owner@example.com", Role: user.RoleVulnerableOwner}
	until := time.Now().Add(time.Hour)
	expected := &Link{ID: "link-id", Status: "pending"}

	tests := []struct {
		name      string
		primaryID string
		params    CreateParams
		users     map[string]*user.User
		userErr   error
		repoErr   error
		wantErr   error
		wantCalls int
		wantPerms []string
	}{
		{
			name:      "creates pending link with normalized permissions",
			primaryID: owner.ID,
			params: CreateParams{
				DelegateEmail: " DELEGATE@example.com ",
				Permissions:   []string{" view_balance ", "make_payment"},
				ValidUntil:    &until,
			},
			users:     map[string]*user.User{owner.ID: owner, delegate.ID: delegate},
			wantCalls: 1,
			wantPerms: []string{"view_balance", "make_payment"},
		},
		{
			name:      "rejects empty permissions",
			primaryID: owner.ID,
			params:    CreateParams{DelegateEmail: delegate.Email},
			users:     map[string]*user.User{owner.ID: owner, delegate.ID: delegate},
			wantErr:   ErrInvalidInput,
		},
		{
			name:      "rejects duplicate permissions",
			primaryID: owner.ID,
			params: CreateParams{
				DelegateEmail: delegate.Email,
				Permissions:   []string{"view_balance", " view_balance "},
			},
			users:   map[string]*user.User{owner.ID: owner, delegate.ID: delegate},
			wantErr: ErrInvalidInput,
		},
		{
			name:      "rejects expired valid until",
			primaryID: owner.ID,
			params: CreateParams{
				DelegateEmail: delegate.Email,
				Permissions:   []string{"view_balance"},
				ValidUntil:    timePtr(time.Now().Add(-time.Minute)),
			},
			users:   map[string]*user.User{owner.ID: owner, delegate.ID: delegate},
			wantErr: ErrInvalidInput,
		},
		{
			name:      "rejects self delegation",
			primaryID: owner.ID,
			params: CreateParams{
				DelegateEmail: owner.Email,
				Permissions:   []string{"view_balance"},
			},
			users:   map[string]*user.User{owner.ID: owner},
			wantErr: ErrInvalidInput,
		},
		{
			name:      "only vulnerable owners can create links",
			primaryID: delegate.ID,
			params: CreateParams{
				DelegateEmail: owner.Email,
				Permissions:   []string{"view_balance"},
			},
			users:   map[string]*user.User{owner.ID: owner, delegate.ID: delegate},
			wantErr: ErrUnauthorized,
		},
		{
			name:      "target must be a delegate",
			primaryID: owner.ID,
			params: CreateParams{
				DelegateEmail: otherOwner.Email,
				Permissions:   []string{"view_balance"},
			},
			users:   map[string]*user.User{owner.ID: owner, otherOwner.ID: otherOwner},
			wantErr: ErrDelegateNotFound,
		},
		{
			name:      "unknown delegate",
			primaryID: owner.ID,
			params: CreateParams{
				DelegateEmail: delegate.Email,
				Permissions:   []string{"view_balance"},
			},
			users:   map[string]*user.User{owner.ID: owner},
			wantErr: ErrDelegateNotFound,
		},
		{
			name:      "user lookup failure",
			primaryID: owner.ID,
			params: CreateParams{
				DelegateEmail: delegate.Email,
				Permissions:   []string{"view_balance"},
			},
			userErr: errors.New("database unavailable"),
			wantErr: errors.New("database unavailable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			links := &serviceTestLinks{link: expected, err: tt.repoErr}
			service := NewService(links, serviceTestUsers{users: tt.users, err: tt.userErr})

			got, err := service.Create(context.Background(), tt.primaryID, tt.params)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("Create() error = nil, want an error")
				}
				if errors.Is(tt.wantErr, ErrInvalidInput) ||
					errors.Is(tt.wantErr, ErrUnauthorized) ||
					errors.Is(tt.wantErr, ErrDelegateNotFound) {
					if !errors.Is(err, tt.wantErr) {
						t.Fatalf("Create() error = %v, want %v", err, tt.wantErr)
					}
				}
			} else if err != nil {
				t.Fatalf("Create() error = %v", err)
			}

			if links.calls != tt.wantCalls {
				t.Errorf("repository Create() called %d times, want %d", links.calls, tt.wantCalls)
			}
			if tt.wantCalls > 0 {
				if got != expected {
					t.Errorf("Create() = %#v, want %#v", got, expected)
				}
				if links.params.PrimaryUserID != tt.primaryID {
					t.Errorf("primary user ID = %q, want %q", links.params.PrimaryUserID, tt.primaryID)
				}
				if links.params.DelegateUserID != delegate.ID {
					t.Errorf("delegate user ID = %q, want %q", links.params.DelegateUserID, delegate.ID)
				}
				if len(links.params.Permissions) != len(tt.wantPerms) {
					t.Fatalf("permissions = %#v, want %#v", links.params.Permissions, tt.wantPerms)
				}
				for i := range tt.wantPerms {
					if links.params.Permissions[i] != tt.wantPerms[i] {
						t.Errorf("permissions = %#v, want %#v", links.params.Permissions, tt.wantPerms)
						break
					}
				}
			}
		})
	}
}

func timePtr(value time.Time) *time.Time {
	return &value
}
