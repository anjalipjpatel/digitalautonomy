package user

import (
	"context"
	passwordhash "digitalautonomy/internal/hash"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/jackc/pgx/v5"
)

var ErrInvalidInput = errors.New("invalid user input")
var ErrInvalidCredentials = errors.New("invalid credentials")

type Service struct {
	users *Repository
}

func NewService(users *Repository) *Service {
	return &Service{
		users: users,
	}
}

func (s *Service) Create(ctx context.Context, email, name, password, role string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)

	// Check for empty fields
	if name == "" || password == "" || email == "" {
		return nil, ErrInvalidInput
	}

	// Verify Email Format
	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		return nil, ErrInvalidInput
	}

	// Validate role type
	var userRole Role
	switch role {
	case string(RoleVulnerableOwner), string(RoleDelegate), string(RoleAdmin):
		userRole = Role(role)
	default:
		return nil, ErrInvalidInput
	}

	// hash password
	hashedPassword, err := passwordhash.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	return s.users.Create(
		ctx,
		CreateUserParams{
			Email:          email,
			HashedPassword: string(hashedPassword),
			Name:           name,
			Role:           userRole,
		},
	)
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	found, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	valid, err := passwordhash.Verify(password, found.hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}
	if !valid {
		return nil, ErrInvalidCredentials
	}

	return &found.User, nil
}

func (s *Service) GetByID(ctx context.Context, userID string) (*User, error) {
	found, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return found, nil
}

func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	found, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &found.User, nil
}
