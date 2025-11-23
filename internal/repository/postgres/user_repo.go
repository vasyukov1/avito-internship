package postgres

import (
	"avito-internship/internal/domain"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) UpsertUsers(ctx context.Context, teamName string, users []domain.User) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO teams (team_name)
         VALUES ($1)
         ON CONFLICT (team_name) DO NOTHING`,
		teamName,
	)
	if err != nil {
		return err
	}

	for _, user := range users {
		_, err = tx.Exec(ctx,
			`INSERT INTO users (user_id, username, team_name, is_active)
             VALUES ($1, $2, $3, $4)
             ON CONFLICT (user_id)
             DO UPDATE SET username = EXCLUDED.username,
                           team_name = EXCLUDED.team_name,
                           is_active = EXCLUDED.is_active`,
			user.ID, user.Username, teamName, user.IsActive,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *UserRepo) SetIsActive(ctx context.Context, userID string, active bool) (*domain.User, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE users
         SET is_active = $2
         WHERE user_id = $1
         RETURNING user_id, username, team_name, is_active`,
		userID, active,
	)

	var user domain.User
	if err := row.Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT user_id, username, team_name, is_active
         FROM users
         WHERE user_id = $1`,
		id,
	)

	var user domain.User
	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetActiveTeamMembers(ctx context.Context, teamName string, except string) ([]domain.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT user_id, username, team_name, is_active
         FROM users
         WHERE team_name = $1
           AND is_active = TRUE
           AND user_id <> $2`,
		teamName, except,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, err
		}
		list = append(list, user)
	}

	return list, nil
}
