package infrastructure

import (
	"avito-internship/internal/domain"
	"avito-internship/internal/repository"
	"context"
	"errors"
	"time"
)

type Server struct {
	storage repository.Storage
}

func NewServer(storage repository.Storage) *Server {
	return &Server{storage: storage}
}

func (s *Server) Storage() repository.Storage {
	return s.storage
}

func (s *Server) CreatePullRequest(ctx context.Context, id string, name string, authorID string) (*domain.PullRequest, error) {
	author, err := s.storage.User().GetByID(ctx, authorID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	members, err := s.storage.User().GetActiveTeamMembers(ctx, author.TeamName, authorID)
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		return nil, domain.ErrNoCandidate
	}

	var reviewers []string
	for i := 0; i < len(members) && i < 2; i++ {
		reviewers = append(reviewers, members[i].ID)
	}

	now := time.Now()
	pr := domain.PullRequest{
		ID:        id,
		Name:      name,
		AuthorID:  authorID,
		Status:    domain.PROpen,
		Reviewers: reviewers,
		CreatedAt: now,
	}

	if err := s.storage.PullRequest().Create(ctx, pr); err != nil {
		return nil, err
	}

	if len(reviewers) > 0 {
		if err := s.storage.PullRequest().AssignReviewers(ctx, pr.ID, reviewers); err != nil {
			return nil, err
		}
	}

	return &pr, nil
}
