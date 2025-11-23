package repository

import (
	"avito-internship/internal/domain"
	"avito-internship/internal/repository/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type Storage interface {
	Team() domain.TeamRepository
	User() domain.UserRepository
	PullRequest() domain.PullRequestRepository
}

type pgStorage struct {
	team        domain.TeamRepository
	user        domain.UserRepository
	pullRequest domain.PullRequestRepository
}

func NewPgStorage(db *pgxpool.Pool, logger *logrus.Logger) Storage {
	return &pgStorage{
		team:        postgres.NewTeamRepo(db, logger),
		user:        postgres.NewUserRepo(db, logger),
		pullRequest: postgres.NewPRRepo(db, logger),
	}
}

func (s *pgStorage) Team() domain.TeamRepository               { return s.team }
func (s *pgStorage) User() domain.UserRepository               { return s.user }
func (s *pgStorage) PullRequest() domain.PullRequestRepository { return s.pullRequest }
