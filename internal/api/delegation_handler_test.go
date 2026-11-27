package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"digitalautonomy/internal/delegation"
	"digitalautonomy/internal/session"
)

type delegationTestSessions struct {
	userID string
	err    error
	token  string
}

func (s *delegationTestSessions) GetUserIDByToken(_ context.Context, token string) (string, error) {
	s.token = token
	return s.userID, s.err
}

func (*delegationTestSessions) Revoke(context.Context, string) error { return nil }

type delegationTestCreator struct {
	primaryID  string
	params     delegation.CreateParams
	link       *delegation.Link
	err        error
	calls      int
	pending    []delegation.Link
	userLinks  []delegation.Link
	delegateID string
	linkID     string
}

func (d *delegationTestCreator) Create(_ context.Context, primaryID string, params delegation.CreateParams) (*delegation.Link, error) {
	d.calls++
	d.primaryID = primaryID
	d.params = params
	return d.link, d.err
}

func (d *delegationTestCreator) GetPendingForDelegate(_ context.Context, delegateID string) ([]delegation.Link, error) {
	d.delegateID = delegateID
	return d.pending, d.err
}

func (d *delegationTestCreator) GetForUser(_ context.Context, userID string) ([]delegation.Link, error) {
	d.delegateID = userID
	return d.userLinks, d.err
}

func (d *delegationTestCreator) Accept(_ context.Context, delegateID, linkID string) (*delegation.Link, error) {
	d.delegateID = delegateID
	d.linkID = linkID
	return d.link, d.err
}

func TestDelegationHandlerLink(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		cookie     string
		sessionID  string
		sessionErr error
		createErr  error
		wantStatus int
		wantCalls  int
	}{
		{
			name:       "requires session",
			body:       `{"delegate_email":"delegate@example.com","permissions":["view_balance"]}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "rejects invalid session",
			body:       `{"delegate_email":"delegate@example.com","permissions":["view_balance"]}`,
			cookie:     "invalid",
			sessionErr: session.ErrInvalidSession,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "rejects malformed JSON",
			body:       `{"delegate_user_id":`,
			cookie:     "valid",
			sessionID:  "owner",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects unknown fields",
			body:       `{"primary_user_id":"another-user","delegate_email":"delegate@example.com","permissions":["view_balance"]}`,
			cookie:     "valid",
			sessionID:  "owner",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "rejects invalid link input",
			body:       `{"delegate_email":"delegate@example.com","permissions":[]}`,
			cookie:     "valid",
			sessionID:  "owner",
			createErr:  delegation.ErrInvalidInput,
			wantStatus: http.StatusBadRequest,
			wantCalls:  1,
		},
		{
			name:       "forbids users without owner role",
			body:       `{"delegate_email":"delegate@example.com","permissions":["view_balance"]}`,
			cookie:     "valid",
			sessionID:  "delegate",
			createErr:  delegation.ErrUnauthorized,
			wantStatus: http.StatusForbidden,
			wantCalls:  1,
		},
		{
			name:       "creates link",
			body:       `{"delegate_email":"delegate@example.com","permissions":["view_balance"],"valid_until":"2030-01-01T00:00:00Z"}`,
			cookie:     "valid",
			sessionID:  "owner",
			wantStatus: http.StatusCreated,
			wantCalls:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created := &delegation.Link{
				ID:             "link-id",
				PrimaryUserID:  "owner",
				DelegateUserID: "delegate",
				Status:         "pending",
				Permissions:    []string{"view_balance"},
			}
			sessions := &delegationTestSessions{userID: tt.sessionID, err: tt.sessionErr}
			links := &delegationTestCreator{link: created, err: tt.createErr}
			handler := NewDelegationHandler(sessions, links)

			request := httptest.NewRequest(http.MethodPost, "/delegation-links", strings.NewReader(tt.body))
			if tt.cookie != "" {
				request.AddCookie(&http.Cookie{Name: "session", Value: tt.cookie})
			}
			recorder := httptest.NewRecorder()

			handler.Link(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%q", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			if links.calls != tt.wantCalls {
				t.Fatalf("service called %d times, want %d", links.calls, tt.wantCalls)
			}
			if tt.cookie != "" && sessions.token != tt.cookie {
				t.Errorf("validated token = %q, want %q", sessions.token, tt.cookie)
			}
			if tt.wantStatus == http.StatusCreated {
				if links.primaryID != "owner" {
					t.Errorf("primary user ID = %q, want owner from authenticated session", links.primaryID)
				}
				if links.params.DelegateEmail != "delegate@example.com" ||
					len(links.params.Permissions) != 1 ||
					links.params.Permissions[0] != "view_balance" ||
					links.params.ValidUntil == nil {
					t.Errorf("service params = %#v, want decoded request", links.params)
				}
				if got := recorder.Header().Get("Content-Type"); got != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", got)
				}
				var got delegation.Link
				if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
					t.Fatalf("response is not valid JSON: %v", err)
				}
				if got.ID != created.ID || got.Status != created.Status {
					t.Errorf("response link = %#v, want %#v", got, created)
				}
			}
		})
	}

	t.Run("maps duplicate to conflict", func(t *testing.T) {
		sessions := &delegationTestSessions{userID: "owner"}
		links := &delegationTestCreator{err: delegation.ErrDuplicateLink}
		handler := NewDelegationHandler(sessions, links)
		request := httptest.NewRequest(http.MethodPost, "/delegation-links",
			strings.NewReader(`{"delegate_email":"delegate@example.com","permissions":["view_balance"]}`))
		request.AddCookie(&http.Cookie{Name: "session", Value: "valid"})
		recorder := httptest.NewRecorder()

		handler.Link(recorder, request)

		if recorder.Code != http.StatusConflict {
			t.Errorf("status = %d, want %d", recorder.Code, http.StatusConflict)
		}
	})

	t.Run("lists pending links for authenticated delegate", func(t *testing.T) {
		pending := []delegation.Link{{ID: "link-id", Status: "pending"}}
		sessions := &delegationTestSessions{userID: "delegate"}
		links := &delegationTestCreator{pending: pending}
		handler := NewDelegationHandler(sessions, links)
		request := httptest.NewRequest(http.MethodGet, "/delegation-links/pending", nil)
		request.AddCookie(&http.Cookie{Name: "session", Value: "valid"})
		recorder := httptest.NewRecorder()

		handler.Pending(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%q", recorder.Code, http.StatusOK, recorder.Body.String())
		}
		if links.delegateID != "delegate" {
			t.Errorf("delegate ID = %q, want authenticated user ID", links.delegateID)
		}
		var got []delegation.Link
		if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
			t.Fatalf("response is not valid JSON: %v", err)
		}
		if len(got) != 1 || got[0].ID != pending[0].ID {
			t.Errorf("response links = %#v, want %#v", got, pending)
		}
	})

	t.Run("lists all links for authenticated user", func(t *testing.T) {
		allLinks := []delegation.Link{{ID: "link-id", Status: "active"}}
		sessions := &delegationTestSessions{userID: "owner"}
		links := &delegationTestCreator{userLinks: allLinks}
		handler := NewDelegationHandler(sessions, links)
		request := httptest.NewRequest(http.MethodGet, "/delegation-links", nil)
		request.AddCookie(&http.Cookie{Name: "session", Value: "valid"})
		recorder := httptest.NewRecorder()

		handler.List(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%q", recorder.Code, http.StatusOK, recorder.Body.String())
		}
		if links.delegateID != "owner" {
			t.Errorf("listed user ID = %q, want authenticated user ID", links.delegateID)
		}
		var got []delegation.Link
		if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
			t.Fatalf("response is not valid JSON: %v", err)
		}
		if len(got) != 1 || got[0].ID != allLinks[0].ID {
			t.Errorf("response links = %#v, want %#v", got, allLinks)
		}
	})

	t.Run("accepts pending link for authenticated delegate", func(t *testing.T) {
		accepted := &delegation.Link{ID: "link-id", Status: "active"}
		sessions := &delegationTestSessions{userID: "delegate"}
		links := &delegationTestCreator{link: accepted}
		handler := NewDelegationHandler(sessions, links)
		request := httptest.NewRequest(http.MethodPost, "/delegation-links/link-id/accept", nil)
		request.SetPathValue("id", "link-id")
		request.AddCookie(&http.Cookie{Name: "session", Value: "valid"})
		recorder := httptest.NewRecorder()

		handler.Accept(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%q", recorder.Code, http.StatusOK, recorder.Body.String())
		}
		if links.delegateID != "delegate" || links.linkID != "link-id" {
			t.Errorf("accept args = (%q, %q), want (delegate, link-id)", links.delegateID, links.linkID)
		}
	})
}
