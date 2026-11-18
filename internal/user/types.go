package user

import "time"

type Role string

const (
	RoleVulnerableOwner Role = "vulnerable_owner"
	RoleDelegate        Role = "delegate"
	RoleAdmin           Role = "admin"
)

type User struct {
	ID        string
	Email     string
	Name      string
	Role      Role
	CreatedAt time.Time
}

type userWithPassword struct {
	User
	hashedPassword string
}

type CreateUserParams struct {
	Email          string
	HashedPassword string
	Name           string
	Role           Role
}

type LoginParams struct {
	Email    string
	Password string
}
