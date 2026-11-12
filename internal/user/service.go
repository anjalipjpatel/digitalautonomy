package user

import (
	"context"
	passwordhash "digitalautonomy/internal/hash"
	"errors"
	"fmt"
	"net/mail"
	"strings"
)

var ErrInvalidInput = errors.New("invalid user input")

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
