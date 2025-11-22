package postgres

import (
	"avito-internship/internal/domain"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepo struct {
	db *pgxpool.Pool
}

func NewTeamRepo(db *pgxpool.Pool) *TeamRepo {
	return &TeamRepo{db: db}
}

func (r *TeamRepo) CreateTeam(ctx context.Context, team domain.Team) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO teams (team_name) VALUES ($1)`,
		team.Name,
	)
	return err
}

func (r *TeamRepo) GetTeam(ctx context.Context, name string) (*domain.Team, error) {
	var team domain.Team
	team.Name = name

	rows, err := r.db.Query(ctx,
		`SELECT user_id, username, is_active FROM teams WHERE team_name = $1`,
		name,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Username, &user.IsActive); err != nil {
			return nil, err
		}
		user.TeamName = name
		team.Members = append(team.Members, user)
	}

	return &team, nil
}
