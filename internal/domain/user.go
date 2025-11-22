package domain

import "context"

type User struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type UserRepository interface {
	UpsertUsers(ctx context.Context, teamName string, users []User) error
	SetIsActive(ctx context.Context, userID string, active bool) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetActiveTeamMembers(ctx context.Context, teamName string, except string) ([]User, error)
}
