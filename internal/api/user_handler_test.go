package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"digitalautonomy/internal/session"
	"digitalautonomy/internal/user"
)

type meTestUsers struct {
	profile *user.User
	err     error
	gotID   string
}

func (users *meTestUsers) Create(context.Context, string, string, string, string) (*user.User, error) {
	return nil, nil
}

func (users *meTestUsers) GetByID(_ context.Context, id string) (*user.User, error) {
	users.gotID = id
	return users.profile, users.err
}

type meTestSessions struct {
	userID string
	err    error
	token  string
}

func (sessions *meTestSessions) GetUserIDByToken(_ context.Context, token string) (string, error) {
	sessions.token = token
	return sessions.userID, sessions.err
}

func TestUserHandlerMe(t *testing.T) {
	expectedProfile := &user.User{
		ID:        "user-123",
		Email:     "person@example.com",
		Name:      "Test Person",
		Role:      user.RoleDelegate,
		CreatedAt: time.Date(2026, time.November, 19, 12, 0, 0, 0, time.UTC),
	}
	databaseErr := errors.New("database unavailable")

	tests := []struct {
		name             string
		cookie           string
		sessionUserID    string
		sessionErr       error
		profile          *user.User
		profileErr       error
		wantStatus       int
		wantSessionToken string
		wantLookupID     string
		wantProfile      bool
	}{
		{
			name:       "missing cookie",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "empty cookie",
			cookie:     "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:             "unknown session",
			cookie:           "bad-token",
			sessionErr:       session.ErrInvalidSession,
			wantStatus:       http.StatusUnauthorized,
			wantSessionToken: "bad-token",
		},
		{
			name:             "expired session",
			cookie:           "expired-token",
			sessionErr:       session.ErrTokenExpired,
			wantStatus:       http.StatusUnauthorized,
			wantSessionToken: "expired-token",
		},
		{
			name:             "revoked session",
			cookie:           "revoked-token",
			sessionErr:       session.ErrTokenRevoked,
			wantStatus:       http.StatusUnauthorized,
			wantSessionToken: "revoked-token",
		},
		{
			name:             "session database failure",
			cookie:           "valid-looking-token",
			sessionErr:       databaseErr,
			wantStatus:       http.StatusInternalServerError,
			wantSessionToken: "valid-looking-token",
		},
		{
			name:             "user lookup failure",
			cookie:           "valid-token",
			sessionUserID:    "user-123",
			profileErr:       databaseErr,
			wantStatus:       http.StatusInternalServerError,
			wantSessionToken: "valid-token",
			wantLookupID:     "user-123",
		},
		{
			name:             "valid session",
			cookie:           "valid-token",
			sessionUserID:    "user-123",
			profile:          expectedProfile,
			wantStatus:       http.StatusOK,
			wantSessionToken: "valid-token",
			wantLookupID:     "user-123",
			wantProfile:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := &meTestUsers{profile: tt.profile, err: tt.profileErr}
			sessions := &meTestSessions{userID: tt.sessionUserID, err: tt.sessionErr}
			handler := NewUserHandler(users, sessions)

			request := httptest.NewRequest(http.MethodGet, "/me", nil)
			if tt.name != "missing cookie" {
				request.AddCookie(&http.Cookie{Name: "session", Value: tt.cookie})
			}
			recorder := httptest.NewRecorder()

			handler.Me(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%q", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			if sessions.token != tt.wantSessionToken {
				t.Errorf("session validator token = %q, want %q", sessions.token, tt.wantSessionToken)
			}
			if users.gotID != tt.wantLookupID {
				t.Errorf("user lookup ID = %q, want %q", users.gotID, tt.wantLookupID)
			}

			if tt.wantProfile {
				if got := recorder.Header().Get("Content-Type"); got != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", got)
				}
				var got user.User
				if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
					t.Fatalf("response is not valid user JSON: %v", err)
				}
				if got != *expectedProfile {
					t.Errorf("response profile = %#v, want %#v", got, *expectedProfile)
				}
			}
		})
	}
}
