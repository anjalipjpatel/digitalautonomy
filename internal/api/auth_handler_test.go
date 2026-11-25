package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"digitalautonomy/internal/user"
)

type authTestUsers struct {
	profile  *user.User
	err      error
	email    string
	password string
}

func (users *authTestUsers) Login(_ context.Context, email, password string) (*user.User, error) {
	users.email = email
	users.password = password
	return users.profile, users.err
}

type authTestSessions struct {
	token     string
	expiresAt *time.Time
	err       error
	userID    string
}

func (sessions *authTestSessions) Create(_ context.Context, userID string) (string, *time.Time, error) {
	sessions.userID = userID
	return sessions.token, sessions.expiresAt, sessions.err
}

func TestAuthHandlerLogin(t *testing.T) {
	expiresAt := time.Now().UTC().Add(time.Hour)
	expectedUser := &user.User{
		ID:    "user-123",
		Email: "person@example.com",
		Name:  "Test Person",
		Role:  user.RoleDelegate,
	}

	databaseErr := errors.New("database unavailable")

	tests := []struct {
		name            string
		body            string
		user            *user.User
		loginErr        error
		token           string
		sessionExpiry   *time.Time
		sessionErr      error
		secureCookie    bool
		wantStatus      int
		wantSessionUser string
		wantCookie      bool
	}{
		{
			name:       "invalid JSON",
			body:       "{",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid credentials",
			body:       `{"email":"person@example.com","password":"wrong"}`,
			loginErr:   user.ErrInvalidCredentials,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "login backend failure",
			body:       `{"email":"person@example.com","password":"secret"}`,
			loginErr:   databaseErr,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "nil user",
			body:       `{"email":"person@example.com","password":"secret"}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:            "session creation failure",
			body:            `{"email":"person@example.com","password":"secret"}`,
			user:            expectedUser,
			sessionErr:      databaseErr,
			wantStatus:      http.StatusInternalServerError,
			wantSessionUser: "user-123",
		},
		{
			name:            "missing expiry",
			body:            `{"email":"person@example.com","password":"secret"}`,
			user:            expectedUser,
			token:           "session-token",
			wantStatus:      http.StatusInternalServerError,
			wantSessionUser: "user-123",
		},
		{
			name:            "successful login sets secure cookie",
			body:            `{"email":"person@example.com","password":"secret"}`,
			user:            expectedUser,
			token:           "session-token",
			sessionExpiry:   &expiresAt,
			secureCookie:    true,
			wantStatus:      http.StatusOK,
			wantSessionUser: "user-123",
			wantCookie:      true,
		},
		{
			name:            "successful login sets local development cookie",
			body:            `{"email":"person@example.com","password":"secret"}`,
			user:            expectedUser,
			token:           "session-token",
			sessionExpiry:   &expiresAt,
			wantStatus:      http.StatusOK,
			wantSessionUser: "user-123",
			wantCookie:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := &authTestUsers{profile: tt.user, err: tt.loginErr}
			sessions := &authTestSessions{
				token:     tt.token,
				expiresAt: tt.sessionExpiry,
				err:       tt.sessionErr,
			}
			handler := NewAuthHandler(users, sessions, tt.secureCookie)
			request := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()

			handler.Login(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%q", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			if users.email != "" && users.email != "person@example.com" {
				t.Errorf("login email = %q, want person@example.com", users.email)
			}
			if users.password != "" && users.password != "wrong" && users.password != "secret" {
				t.Errorf("unexpected login password passed: %q", users.password)
			}
			if sessions.userID != tt.wantSessionUser {
				t.Errorf("session created for user %q, want %q", sessions.userID, tt.wantSessionUser)
			}

			cookies := recorder.Result().Cookies()
			if !tt.wantCookie {
				if len(cookies) != 0 {
					t.Errorf("got cookies %#v, want none", cookies)
				}
				return
			}
			if len(cookies) != 1 {
				t.Fatalf("got %d cookies, want 1", len(cookies))
			}
			cookie := cookies[0]
			if cookie.Name != "session" || cookie.Value != tt.token {
				t.Errorf("session cookie = %q=%q, want session=%q", cookie.Name, cookie.Value, tt.token)
			}
			if cookie.Path != "/" || !cookie.HttpOnly || cookie.Secure != tt.secureCookie ||
				cookie.SameSite != http.SameSiteLaxMode || !cookie.Expires.Equal(expiresAt.Truncate(time.Second)) {
				t.Errorf("session cookie attributes incorrect: %#v", cookie)
			}

			var responseUser user.User
			if err := json.Unmarshal(recorder.Body.Bytes(), &responseUser); err != nil {
				t.Fatalf("response is not valid user JSON: %v", err)
			}
			if responseUser.ID != expectedUser.ID || responseUser.Email != expectedUser.Email {
				t.Errorf("response user = %#v, want %#v", responseUser, expectedUser)
			}
			if strings.Contains(recorder.Body.String(), tt.token) {
				t.Error("raw session token should not be included in the JSON response")
			}
		})
	}
}
