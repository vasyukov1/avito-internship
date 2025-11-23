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

func (s *Server) MergePullRequest(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.storage.PullRequest().Merge(ctx, prID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return pr, nil
}

func (s *Server) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	// Get PR
	pr, err := s.storage.PullRequest().GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, "", domain.ErrNotFound
		}
		return nil, "", err
	}

	// Cannot change merged PR
	if pr.Status == domain.PRMerged {
		return nil, "", domain.ErrPRMerged
	}

	// Check oldReviewer existing
	isAssigned := false
	for _, r := range pr.Reviewers {
		if r == oldReviewerID {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		return nil, "", domain.ErrNotAssigned
	}

	// Get oldReviewer's team
	oldUser, err := s.storage.User().GetByID(ctx, oldReviewerID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, "", domain.ErrNotFound
		}
		return nil, "", err
	}

	// Get active teammates
	candidates, err := s.storage.User().GetActiveTeamMembers(ctx, oldUser.TeamName, oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	// Remove old reviewers
	assignedSet := map[string]struct{}{}
	for _, a := range pr.Reviewers {
		assignedSet[a] = struct{}{}
	}
	assignedSet[pr.AuthorID] = struct{}{}

	var newReviewerID string
	for _, m := range candidates {
		if m.ID == oldReviewerID {
			continue
		}
		if _, ok := assignedSet[m.ID]; ok {
			continue
		}
		newReviewerID = m.ID
		break
	}

	if newReviewerID == "" {
		return nil, "", domain.ErrNoCandidate
	}

	// Transaction updating
	updatedPR, err := s.storage.PullRequest().Reassign(ctx, prID, oldReviewerID, newReviewerID)
	if err != nil {
		return nil, "", err
	}

	return updatedPR, newReviewerID, nil
}
