package postgres

import (
	"avito-internship/internal/domain"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
               pr.status,
               pr.created_at,
               pr.merged_at
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

//func (r *PRRepo) Merge(ctx context.Context, prID string) (*domain.PullRequest, error) {
//	now := time.Now()
//
//	_, err := r.db.Exec(ctx,
//		`UPDATE pull_requests
//         SET status = 'MERGED',
//             merged_at = COALESCE(merged_at, $2)
//         WHERE pull_request_id = $1`,
//		prID, now,
//	)
//	if err != nil {
//		return nil, err
//	}
//
//	return r.GetByUserID(ctx, prID)
//}

func (r *PRRepo) Reassign(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx,
		`SELECT status FROM pull_requests WHERE pull_request_id = $1`,
		prID,
	).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("not_found")
	}
	if err != nil {
		return err
	}
	if status == "MERGED" {
		return fmt.Errorf("pr_merged")
	}

	var exists bool
	err = tx.QueryRow(ctx,
		`SELECT true FROM pull_request_reviewers 
         WHERE pull_request_id = $1 AND reviewer_id = $2`,
		prID, oldReviewerID,
	).Scan(&exists)

	if err == pgx.ErrNoRows {
		return fmt.Errorf("not_assigned")
	}
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`DELETE FROM pull_request_reviewers
         WHERE pull_request_id = $1 AND reviewer_id = $2`,
		prID, oldReviewerID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
         VALUES ($1, $2)`,
		prID, newReviewerID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PRRepo) GetReviewList(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	rows, err := r.db.Query(ctx,
		`SELECT pr.pull_request_id,
                pr.pull_request_name,
                pr.author_id,
                pr.status
         FROM pull_requests pr
         JOIN pull_request_shorts s
           ON pr.pull_request_id = s.pull_request_id
         WHERE s.reviewer_id = $1
         ORDER BY pr.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.PullRequestShort

	for rows.Next() {
		var pr domain.PullRequestShort
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		result = append(result, pr)
	}

	return result, nil
}
