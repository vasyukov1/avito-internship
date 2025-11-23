package repository

import (
	"avito-internship/internal/domain"
	"avito-internship/internal/metrics"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"time"
)

type TeamRepo struct {
	db     *pgxpool.Pool
	logger *logrus.Logger
}

func NewTeamRepo(db *pgxpool.Pool, logger *logrus.Logger) *TeamRepo {
	return &TeamRepo{db, logger}
}

func (r *TeamRepo) CreateTeam(ctx context.Context, team domain.Team) error {
	start := time.Now()
	defer func() {
		metrics.DBQueriesTotal.WithLabelValues("create_team").Inc()
		metrics.DBQueryDuration.WithLabelValues("create_team").Observe(time.Since(start).Seconds())
	}()

	r.logger.WithFields(logrus.Fields{
		"team_name": team.Name,
	}).Debug("Creating team in database")

	// Check team existing
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT * 
			FROM teams 
			WHERE team_name = $1
		)`, team.Name,
	).Scan(&exists)
	if err != nil {
		r.logger.WithError(err).Error("Failed to check team existence")
		return err
	}
	if exists {
		r.logger.WithFields(logrus.Fields{
			"team_name": team.Name,
		}).Warn("Team already exists")
		return domain.ErrTeamExists
	}

	// Insert team
	_, err = r.db.Exec(ctx, `
		INSERT INTO teams (team_name) VALUES ($1)
		`, team.Name,
	)
	if err != nil {
		r.logger.WithError(err).Error("Failed to insert team into database")
		return err
	}

	r.logger.WithFields(logrus.Fields{
		"team_name": team.Name,
	}).Debug("Team created in database")

	return err
}

func (r *TeamRepo) GetTeam(ctx context.Context, name string) (*domain.Team, error) {
	start := time.Now()
	defer func() {
		metrics.DBQueriesTotal.WithLabelValues("get_team").Inc()
		metrics.DBQueryDuration.WithLabelValues("get_team").Observe(time.Since(start).Seconds())
	}()

	r.logger.WithFields(logrus.Fields{
		"team_name": name,
	}).Debug("Getting team from database")

	// Check team existing
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT * 
			FROM teams 
			WHERE team_name = $1
		)`, name,
	).Scan(&exists)
	if err != nil {
		r.logger.WithError(err).Error("Failed to check team existence")
		return nil, err
	}
	if !exists {
		r.logger.WithFields(logrus.Fields{
			"team_name": name,
		}).Warn("Team not found")
		return nil, domain.ErrNotFound
	}

	// Get team info
	var team domain.Team
	team.Name = name

	rows, err := r.db.Query(ctx, `
		SELECT user_id, username, is_active 
		FROM users 
		WHERE team_name = $1
		`, name,
	)
	if err != nil {
		r.logger.WithError(err).Error("Failed to query team members")
		return nil, err
	}
	defer rows.Close()

	// Get team users
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Username, &user.IsActive); err != nil {
			r.logger.WithError(err).Error("Failed to scan team member")
			return nil, err
		}
		user.TeamName = name
		team.Members = append(team.Members, user)
	}

	r.logger.WithFields(logrus.Fields{
		"team_name": name,
		"members":   len(team.Members),
	}).Debug("Team retrieved from database")
	return &team, nil
}
