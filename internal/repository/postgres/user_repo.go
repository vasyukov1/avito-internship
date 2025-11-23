package postgres

import (
	"avito-internship/internal/domain"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type UserRepo struct {
	db     *pgxpool.Pool
	logger *logrus.Logger
}

func NewUserRepo(db *pgxpool.Pool, logger *logrus.Logger) *UserRepo {
	return &UserRepo{db, logger}
}

func (r *UserRepo) UpsertUsers(ctx context.Context, teamName string, users []domain.User) error {
	r.logger.WithFields(logrus.Fields{
		"team_name": teamName,
		"users":     len(users),
	}).Debug("Upserting users in database")

	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.logger.WithError(err).Error("Failed to begin transaction for upserting users")
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
		r.logger.WithError(err).Error("Failed to insert team for upserting users")
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
			r.logger.WithError(err).Error("Failed to upsert user")
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.WithError(err).Error("Failed to commit transaction for upserting users")
		return err
	}

	r.logger.WithFields(logrus.Fields{
		"team_name": teamName,
		"users":     len(users),
	}).Debug("Users upserted successfully")

	return nil
}

func (r *UserRepo) SetIsActive(ctx context.Context, userID string, active bool) (*domain.User, error) {
	r.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"is_active": active,
	}).Debug("Setting user active status in database")

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
			r.logger.WithFields(logrus.Fields{
				"user_id": userID,
			}).Warn("User not found when setting active status")
			return nil, domain.ErrNotFound
		}
		r.logger.WithError(err).Error("Failed to set user active status")
		return nil, err
	}

	r.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"is_active": active,
	}).Debug("User active status set successfully")
	return &user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.logger.WithFields(logrus.Fields{
		"user_id": id,
	}).Debug("Getting user by ID from database")

	// Get user
	row := r.db.QueryRow(ctx, `
		SELECT user_id, username, team_name, is_active
         FROM users
         WHERE user_id = $1
         `, id,
	)

	// Make user domain
	var user domain.User
	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.WithFields(logrus.Fields{
				"user_id": id,
			}).Warn("User not found by ID")
			return nil, domain.ErrNotFound
		}
		r.logger.WithError(err).Error("Failed to get user by ID")
		return nil, err
	}

	r.logger.WithFields(logrus.Fields{
		"user_id": id,
	}).Debug("User retrieved by ID successfully")
	return &user, nil
}

func (r *UserRepo) GetActiveTeamMembers(ctx context.Context, teamName string, except string) ([]domain.User, error) {
	r.logger.WithFields(logrus.Fields{
		"team_name": teamName,
		"except":    except,
	}).Debug("Getting active team members from database")

	// Get active users
	rows, err := r.db.Query(ctx, `
		SELECT user_id, username, team_name, is_active
        FROM users
        WHERE team_name = $1
          AND is_active = TRUE
          AND user_id <> $2
          `, teamName, except,
	)
	if err != nil {
		r.logger.WithError(err).Error("Failed to query active team members")
		return nil, err
	}
	defer rows.Close()

	// Make user domains
	var list []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.TeamName,
			&user.IsActive); err != nil {
			r.logger.WithError(err).Error("Failed to scan active team member")
			return nil, err
		}
		list = append(list, user)
	}

	r.logger.WithFields(logrus.Fields{
		"team_name": teamName,
		"count":     len(list),
	}).Debug("Active team members retrieved successfully")
	return list, nil
}
