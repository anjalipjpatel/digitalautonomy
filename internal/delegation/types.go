package delegation

import "time"

type Link struct {
	ID             string     `json:"id"`
	PrimaryUserID  string     `json:"primary_user_id"`
	DelegateUserID string     `json:"delegate_user_id"`
	ValidFrom      time.Time  `json:"valid_from"`
	ValidUntil     *time.Time `json:"valid_until,omitempty"`
	Status         string     `json:"status"`
	Permissions    []string   `json:"permissions"`
}

type CreateParams struct {
	PrimaryUserID  string
	DelegateUserID string
	DelegateEmail  string
	Permissions    []string
	ValidUntil     *time.Time
}
