package delegation

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"digitalautonomy/internal/user"

	"github.com/jackc/pgx/v5"
)

var (
	ErrInvalidInput     = errors.New("invalid delegation link input")
	ErrUnauthorized     = errors.New("user is not allowed to create delegation links")
	ErrDelegateNotFound = errors.New("delegate user not found")
	ErrLinkNotPending   = errors.New("delegation link is not pending or has expired")
)

type linkRepository interface {
	Create(ctx context.Context, params CreateParams) (*Link, error)
	GetPendingForDelegate(ctx context.Context, delegateUserID string) ([]Link, error)
	Accept(ctx context.Context, delegateUserID, linkID string) (*Link, error)
}

type userLookup interface {
	GetByID(ctx context.Context, id string) (*user.User, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
}

type Service struct {
	links linkRepository
	users userLookup
}

func NewService(links linkRepository, users userLookup) *Service {
	return &Service{links: links, users: users}
}

func (s *Service) Create(ctx context.Context, primaryUserID string, params CreateParams) (*Link, error) {
	email := strings.ToLower(strings.TrimSpace(params.DelegateEmail))
	parsedEmail, err := mail.ParseAddress(email)
	if email == "" || err != nil || parsedEmail.Address != email || len(params.Permissions) == 0 {
		return nil, ErrInvalidInput
	}
	params.DelegateEmail = email

	if params.ValidUntil != nil && !params.ValidUntil.After(time.Now()) {
		return nil, ErrInvalidInput
	}

	permissions := make([]string, 0, len(params.Permissions))
	seen := make(map[string]struct{}, len(params.Permissions))
	for _, permission := range params.Permissions {
		permission = strings.TrimSpace(permission)
		if permission == "" {
			return nil, ErrInvalidInput
		}
		if _, exists := seen[permission]; exists {
			return nil, ErrInvalidInput
		}
		seen[permission] = struct{}{}
		permissions = append(permissions, permission)
	}

	primary, err := s.users.GetByID(ctx, primaryUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUnauthorized
		}
		return nil, fmt.Errorf("looking up primary user: %w", err)
	}
	if primary.Role != user.RoleVulnerableOwner {
		return nil, ErrUnauthorized
	}

	delegate, err := s.users.GetByEmail(ctx, params.DelegateEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDelegateNotFound
		}
		return nil, fmt.Errorf("looking up delegate by email: %w", err)
	}
	if delegate.ID == primaryUserID {
		return nil, ErrInvalidInput
	}
	if delegate.Role != user.RoleDelegate {
		return nil, ErrDelegateNotFound
	}

	params.PrimaryUserID = primaryUserID
	params.DelegateUserID = delegate.ID
	params.Permissions = permissions
	return s.links.Create(ctx, params)
}

func (s *Service) GetPendingForDelegate(ctx context.Context, delegateUserID string) ([]Link, error) {
	delegate, err := s.users.GetByID(ctx, delegateUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUnauthorized
		}
		return nil, fmt.Errorf("looking up delegate user: %w", err)
	}
	if delegate.Role != user.RoleDelegate {
		return nil, ErrUnauthorized
	}

	return s.links.GetPendingForDelegate(ctx, delegateUserID)
}

func (s *Service) Accept(ctx context.Context, delegateUserID, linkID string) (*Link, error) {
	if strings.TrimSpace(linkID) == "" {
		return nil, ErrInvalidInput
	}

	delegate, err := s.users.GetByID(ctx, delegateUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUnauthorized
		}
		return nil, fmt.Errorf("looking up delegate user: %w", err)
	}
	if delegate.Role != user.RoleDelegate {
		return nil, ErrUnauthorized
	}

	link, err := s.links.Accept(ctx, delegateUserID, linkID)
	if errors.Is(err, ErrLinkNotFound) {
		return nil, ErrLinkNotPending
	}
	if err != nil {
		return nil, err
	}
	return link, nil
}
