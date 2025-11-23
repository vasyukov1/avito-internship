package infrastructure

import (
	"avito-internship/internal/domain"
	"avito-internship/internal/repository"
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type Server struct {
	storage repository.Storage
	rng     *rand.Rand
	logger  *logrus.Logger
}

func NewServer(storage repository.Storage, logger *logrus.Logger) *Server {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &Server{
		storage: storage,
		rng:     rng,
		logger:  logger,
	}
}

func (s *Server) Storage() repository.Storage {
	return s.storage
}

func (s *Server) CreatePullRequest(ctx context.Context, id string, name string, authorID string) (*domain.PullRequest, error) {
	s.logger.WithFields(logrus.Fields{
		"pull_request_id": id,
		"name":            name,
		"author_id":       authorID,
	}).Debug("Creating pull request")

	author, err := s.storage.User().GetByID(ctx, authorID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.logger.WithFields(logrus.Fields{
				"author_id": authorID,
			}).Warn("Author not found")
			return nil, domain.ErrNotFound
		}
		s.logger.WithError(err).Error("Failed to get author")
		return nil, err
	}

	members, err := s.storage.User().GetActiveTeamMembers(ctx, author.TeamName, authorID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get active team members")
		return nil, err
	}

	var reviewers []string
	if len(members) > 0 {
		perm := s.rng.Perm(len(members))
		for i := 0; i < len(perm) && i < 2; i++ {
			reviewers = append(reviewers, members[perm[i]].ID)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"pull_request_id": id,
		"reviewers":       reviewers,
		"team_name":       author.TeamName,
	}).Debug("Selected reviewers for PR")

	now := time.Now()
	pr := domain.PullRequest{
		ID:        id,
		Name:      name,
		AuthorID:  authorID,
		Status:    domain.PROpen,
		Reviewers: []string{},
		CreatedAt: now,
	}

	if err := s.storage.PullRequest().Create(ctx, pr); err != nil {
		s.logger.WithError(err).Error("Failed to create PR in storage")
		return nil, err
	}

	assigned, err := s.storage.PullRequest().AssignReviewers(ctx, pr.ID, author.TeamName, authorID, 2)
	if err != nil {
		s.logger.WithError(err).Error("Failed to assign reviewers")
		return nil, err
	}

	updatedPR, err := s.storage.PullRequest().GetByID(ctx, pr.ID)
	if err != nil {
		return nil, err
	}
	updatedPR.Reviewers = assigned

	s.logger.WithFields(logrus.Fields{
		"pull_request_id": id,
	}).Info("Pull request created successfully")

	return updatedPR, nil
}

func (s *Server) MergePullRequest(ctx context.Context, prID string) (*domain.PullRequest, error) {
	s.logger.WithFields(logrus.Fields{
		"pull_request_id": prID,
	}).Debug("Merging pull request")

	pr, err := s.storage.PullRequest().Merge(ctx, prID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.logger.WithFields(logrus.Fields{
				"pull_request_id": prID,
			}).Warn("Pull request not found for merge")
			return nil, domain.ErrNotFound
		}
		s.logger.WithError(err).Error("Failed to merge pull request")
		return nil, err
	}
	s.logger.WithFields(logrus.Fields{
		"pull_request_id": prID,
	}).Info("Pull request merged successfully")
	return pr, nil
}

func (s *Server) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	s.logger.WithFields(logrus.Fields{
		"pull_request_id": prID,
		"old_reviewer_id": oldReviewerID,
	}).Debug("Reassigning reviewer")

	// Get PR
	pr, err := s.storage.PullRequest().GetByID(ctx, prID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.logger.WithFields(logrus.Fields{
				"pull_request_id": prID,
			}).Warn("Pull request not found for reassign")
			return nil, "", domain.ErrNotFound
		}
		s.logger.WithError(err).Error("Failed to get pull request for reassign")
		return nil, "", err
	}

	// Cannot change merged PR
	if pr.Status == domain.PRMerged {
		s.logger.WithFields(logrus.Fields{
			"pull_request_id": prID,
		}).Warn("Attempt to reassign reviewer for merged PR")
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
		s.logger.WithFields(logrus.Fields{
			"pull_request_id": prID,
			"old_reviewer_id": oldReviewerID,
		}).Warn("Reviewer not assigned to PR")
		return nil, "", domain.ErrNotAssigned
	}

	// Get oldReviewer's team
	oldUser, err := s.storage.User().GetByID(ctx, oldReviewerID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.logger.WithFields(logrus.Fields{
				"old_reviewer_id": oldReviewerID,
			}).Warn("Old reviewer not found")
			return nil, "", domain.ErrNotFound
		}
		s.logger.WithError(err).Error("Failed to get old reviewer")
		return nil, "", err
	}

	// Get active teammates
	candidates, err := s.storage.User().GetActiveTeamMembers(ctx, oldUser.TeamName, oldReviewerID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get active team members for reassign")
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
		s.logger.WithFields(logrus.Fields{
			"pull_request_id": prID,
			"team_name":       oldUser.TeamName,
		}).Warn("No candidate found for reassign")
		return nil, "", domain.ErrNoCandidate
	}

	// Transaction updating
	updatedPR, err := s.storage.PullRequest().Reassign(ctx, prID, oldReviewerID, newReviewerID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to reassign reviewer in storage")
		return nil, "", err
	}

	s.logger.WithFields(logrus.Fields{
		"pull_request_id": prID,
		"old_reviewer":    oldReviewerID,
		"new_reviewer":    newReviewerID,
	}).Info("Reviewer reassigned successfully")

	return updatedPR, newReviewerID, nil
}
