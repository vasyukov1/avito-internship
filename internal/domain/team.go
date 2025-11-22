package domain

import "context"

type Team struct {
	Name    string `json:"team_name"`
	Members []User `json:"members"`
}

type TeamRepository interface {
	CreateTeam(ctx context.Context, team Team) error
	GetTeam(ctx context.Context, name string) (*Team, error)
}
