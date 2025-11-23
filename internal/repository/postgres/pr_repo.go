package postgres

import (
	"avito-internship/internal/domain"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"time"
)

type PRRepo struct {
	db     *pgxpool.Pool
	logger *logrus.Logger
}

func NewPRRepo(db *pgxpool.Pool, logger *logrus.Logger) *PRRepo {
	return &PRRepo{db, logger}
}

func (r *PRRepo) Create(ctx context.Context, pr domain.PullRequest) error {
	r.logger.WithFields(logrus.Fields{
		"pull_request_id": pr.ID,
		"author_id":       pr.AuthorID,
	}).Debug("Creating PR in database")

	// Insert PR
	_, err := r.db.Exec(ctx, `
			INSERT INTO pull_requests (
            pull_request_id, pull_request_name, author_id, status, created_at
        ) VALUES ($1, $2, $3, $4, $5)`,
		pr.ID,
		pr.Name,
		pr.AuthorID,
		pr.Status,
		pr.CreatedAt,
	)

	// Check unique
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				r.logger.WithFields(logrus.Fields{
					"pull_request_id": pr.ID,
				}).Warn("PR already exists")
				return domain.ErrPRExists
			}
		}
		r.logger.WithError(err).Error("Failed to create PR in database")
		return err
	}

	r.logger.WithFields(logrus.Fields{
		"pull_request_id": pr.ID,
	}).Debug("PR created in database")
	return nil
}

func (r *PRRepo) GetByUserID(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	r.logger.WithFields(logrus.Fields{
		"user_id": userID,
	}).Debug("Getting PRs by user ID")

	// Get user
	rows, err := r.db.Query(ctx, `
        SELECT pr.pull_request_id,
			   pr.pull_request_name,
			   pr.author_id,
			   pr.status
		FROM pull_requests pr
		JOIN pull_request_shorts prs
		  ON pr.pull_request_id = prs.pull_request_id
		WHERE prs.author_id = $1
    `, userID)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"user_id": userID,
		}).WithError(err).Error("Failed to query PRs by user ID")
		return nil, err
	}
	defer rows.Close()

	// Get PRs
	var pullRequests []domain.PullRequest

	for rows.Next() {
		var pr domain.PullRequest
		err = rows.Scan(
			&pr.ID,
			&pr.Name,
			&pr.AuthorID,
			&pr.Status,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan PR row")
			return nil, err
		}

		pullRequests = append(pullRequests, pr)
	}

	if rows.Err() != nil {
		r.logger.WithError(rows.Err()).Error("Error iterating PR rows")
		return nil, rows.Err()
	}

	r.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"prs_count": len(pullRequests),
	}).Debug("PRs by user ID retrieved successfully")
	return pullRequests, nil
}

func (r *PRRepo) GetByID(ctx context.Context, id string) (*domain.PullRequest, error) {
	r.logger.WithFields(logrus.Fields{
		"pull_request_id": id,
	}).Debug("Getting PR by ID")

	// Get PR
	row := r.db.QueryRow(ctx, `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`, id)

	// Make PR domain
	var pr domain.PullRequest
	err := row.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.WithFields(logrus.Fields{
				"pull_request_id": id,
			}).Warn("PR not found by ID")
			return nil, domain.ErrNotFound
		}
		r.logger.WithError(err).Error("Failed to scan PR row")
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT author_id 
		FROM pull_request_shorts
		WHERE pull_request_id = $1
	`, id)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"pull_request_id": id,
		}).WithError(err).Error("Failed to query PR reviewers")
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewer string
		if err = rows.Scan(&reviewer); err != nil {
			r.logger.WithError(err).Error("Failed to scan reviewer row")
			return nil, err
		}
		reviewers = append(reviewers, reviewer)
	}
	if rows.Err() != nil {
		r.logger.WithError(rows.Err()).Error("Error iterating reviewer rows")
		return nil, rows.Err()
	}
	pr.Reviewers = reviewers

	r.logger.WithFields(logrus.Fields{
		"pull_request_id": id,
		"reviewers_count": len(reviewers),
	}).Debug("PR by ID retrieved successfully")
	return &pr, nil
}

func (r *PRRepo) AssignReviewers(ctx context.Context, prID string, reviewers []string) error {
	r.logger.WithFields(logrus.Fields{
		"pull_request_id": prID,
		"reviewers":       reviewers,
	}).Debug("Assigning reviewers to PR")

	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.logger.WithError(err).Error("Failed to begin transaction for assigning reviewers")
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	stmt := `
		INSERT INTO pull_request_shorts (pull_request_id, author_id) 
		VALUES ($1, $2) 
		ON CONFLICT DO NOTHING
	`
	for _, uid := range reviewers {
		if _, err := tx.Exec(ctx, stmt, prID, uid); err != nil {
			r.logger.WithError(err).Error("Failed to assign reviewer")
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.WithError(err).Error("Failed to commit transaction for assigning reviewers")
		return err
	}

	r.logger.WithFields(logrus.Fields{
		"pull_request_id": prID,
		"reviewers":       reviewers,
	}).Debug("Reviewers assigned successfully")
	return nil
}

func (r *PRRepo) Merge(ctx context.Context, prID string) (*domain.PullRequest, error) {
	r.logger.WithFields(logrus.Fields{
		"pull_request_id": prID,
	}).Debug("Merging PR in database")

	now := time.Now()

	var pr domain.PullRequest
	err := r.db.QueryRow(ctx, `
        UPDATE pull_requests
        SET status = 'MERGED',
            merged_at = COALESCE(merged_at, $2)
        WHERE pull_request_id = $1
        RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at
    `, prID, now).Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&pr.Status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		r.logger.WithFields(logrus.Fields{
			"pull_request_id": prID,
		}).Debug("PR not found in first merge attempt, checking current status")

		selectErr := r.db.QueryRow(ctx, `
			SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
			FROM pull_requests WHERE pull_request_id = $1
		`, prID).Scan(
			&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt,
		)

		if selectErr != nil {
			r.logger.WithFields(logrus.Fields{
				"pull_request_id": prID,
			}).Warn("PR not found for merge")
			return nil, domain.ErrNotFound
		}

		if pr.Status == domain.PRMerged {
			r.logger.WithFields(logrus.Fields{
				"pull_request_id": prID,
			}).Debug("PR already merged")
			return &pr, nil
		}
	}

	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"pull_request_id": prID,
		}).WithError(err).Error("Failed to merge PR")
		return nil, err
	}

	r.logger.WithFields(logrus.Fields{
		"pull_request_id": prID,
		"status":          pr.Status,
	}).Debug("PR merged successfully")
	return &pr, nil
}

func (r *PRRepo) Reassign(ctx context.Context, prID, oldReviewerID, newReviewerID string) (*domain.PullRequest, error) {
	r.logger.WithFields(logrus.Fields{
		"pull_request_id": prID,
		"old_reviewer_id": oldReviewerID,
		"new_reviewer_id": newReviewerID,
	}).Debug("Reassigning reviewer")

	// Begin transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.logger.WithError(err).Error("Failed to begin transaction for reassigning reviewer")
		return nil, err
	}

	// Delete old reviewer
	_, err = tx.Exec(ctx, `
		DELETE FROM pull_request_shorts WHERE pull_request_id=$1 AND author_id=$2
		`, prID, oldReviewerID)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"pull_request_id": prID,
			"old_reviewer_id": oldReviewerID,
		}).WithError(err).Error("Failed to delete old reviewer")
		tx.Rollback(ctx)
		return nil, err
	}

	// Insert new reviewer
	_, err = tx.Exec(ctx, `
		INSERT INTO pull_request_shorts (pull_request_id, author_id) VALUES ($1,$2) ON CONFLICT DO NOTHING
		`, prID, newReviewerID)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"pull_request_id": prID,
			"new_reviewer_id": newReviewerID,
		}).WithError(err).Error("Failed to insert new reviewer")
		tx.Rollback(ctx)
		return nil, err
	}

	// Get actual PR
	pr, err := r.getPRTx(ctx, tx, prID)
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"pull_request_id": prID,
		}).WithError(err).Error("Failed to get PR after reassign")
		tx.Rollback(ctx)
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		r.logger.WithError(err).Error("Failed to commit transaction for reassigning reviewer")
		return nil, err
	}

	r.logger.WithFields(logrus.Fields{
		"pull_request_id": prID,
		"old_reviewer_id": oldReviewerID,
		"new_reviewer_id": newReviewerID,
	}).Debug("Reviewer reassigned successfully")
	return pr, nil
}

func (r *PRRepo) getPRTx(ctx context.Context, tx pgx.Tx, prID string) (*domain.PullRequest, error) {
	var pr domain.PullRequest
	err := tx.QueryRow(ctx, `
        SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
        FROM pull_requests
        WHERE pull_request_id = $1
    `, prID).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	rows, err := tx.Query(ctx, `
		SELECT author_id 
		FROM pull_request_shorts 
		WHERE pull_request_id = $1
	`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, uid)
	}
	pr.Reviewers = reviewers

	return &pr, nil
}
