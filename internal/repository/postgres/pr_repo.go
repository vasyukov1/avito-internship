package postgres

import (
	"avito-internship/internal/domain"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type PRRepo struct {
	db *pgxpool.Pool
}

func NewPRRepo(db *pgxpool.Pool) *PRRepo {
	return &PRRepo{db: db}
}

func (r *PRRepo) Create(ctx context.Context, pr domain.PullRequest) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO pull_requests (
            pull_request_id, pull_request_name, author_id, status, created_at
        ) VALUES ($1, $2, $3, $4, $5)`,
		pr.ID,
		pr.Name,
		pr.AuthorID,
		pr.Status,
		pr.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return domain.ErrPRExists
			}
		}
	}

	return err
}

func (r *PRRepo) GetByUserID(ctx context.Context, userID string) ([]domain.PullRequest, error) {
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
		return nil, err
	}
	defer rows.Close()

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
			return nil, err
		}

		pullRequests = append(pullRequests, pr)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return pullRequests, nil
}

func (r *PRRepo) GetByID(ctx context.Context, id string) (*domain.PullRequest, error) {
	row := r.db.QueryRow(ctx, `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`, id)

	var pr domain.PullRequest
	err := row.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT author_id 
		FROM pull_request_shorts
		WHERE pull_request_id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewer string
		if err = rows.Scan(&reviewer); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewer)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	pr.Reviewers = reviewers

	return &pr, nil
}

func (r *PRRepo) AssignReviewers(ctx context.Context, prID string, reviewers []string) error {
	batch := &pgx.Batch{}
	for _, uid := range reviewers {
		batch.Queue(
			`INSERT INTO pull_request_shorts (pull_request_id, author_id)
             VALUES ($1, $2)
             ON CONFLICT DO NOTHING`,
			prID, uid,
		)
	}
	br := r.db.SendBatch(ctx, batch)
	return br.Close()
}

func (r *PRRepo) Merge(ctx context.Context, prID string) (*domain.PullRequest, error) {
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
		selectErr := r.db.QueryRow(ctx, `
			SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
			FROM pull_requests WHERE pull_request_id = $1
		`, prID).Scan(
			&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt,
		)

		if selectErr != nil {
			return nil, domain.ErrNotFound
		}

		if pr.Status == domain.PRMerged {
			return &pr, nil
		}
	}

	if err != nil {
		return nil, err
	}

	return &pr, nil
}

func (r *PRRepo) Reassign(ctx context.Context, prID, oldReviewerID, newReviewerID string) (*domain.PullRequest, error) {
	// Begin transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	// Delete old reviewer
	_, err = tx.Exec(ctx, `
		DELETE FROM pull_request_shorts WHERE pull_request_id=$1 AND author_id=$2
		`, prID, oldReviewerID)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	// Insert new reviewer
	_, err = tx.Exec(ctx, `
		INSERT INTO pull_request_shorts (pull_request_id, author_id) VALUES ($1,$2) ON CONFLICT DO NOTHING
		`, prID, newReviewerID)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	// Get actual PR
	pr, err := r.getPRTx(ctx, tx, prID)
	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

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

	rows, err := tx.Query(ctx, `SELECT author_id FROM pull_request_shorts WHERE pull_request_id = $1`, prID)
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
