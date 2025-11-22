package domain

import (
	"context"
	"time"
)

type PullRequestStatus string

const (
	PROpen   PullRequestStatus = "OPEN"
	PRMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	ID        string            `json:"pull_request_id"`
	Name      string            `json:"pull_request_name"`
	AuthorID  string            `json:"author_id"`
	Status    PullRequestStatus `json:"status"`
	Reviewers []string          `json:"assigned_reviewers"`
	CreatedAt time.Time         `json:"createdAt"`
	MergedAt  *time.Time        `json:"mergedAt"`
}

type PullRequestShort struct {
	ID       string            `json:"pull_request_id"`
	Name     string            `json:"pull_request_name"`
	AuthorID string            `json:"author_id"`
	Status   PullRequestStatus `json:"status"`
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr PullRequest) error
	GetByID(ctx context.Context, id string) (*PullRequest, error)
	AssignReviewers(ctx context.Context, prID string, reviewers []string) error
	Merge(ctx context.Context, prID string) (*PullRequest, error)
	Reassign(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	GetReviewList(ctx context.Context, userID string) ([]PullRequestShort, error)
}
