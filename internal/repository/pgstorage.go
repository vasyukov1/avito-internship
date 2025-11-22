package repository

import (
	"avito-internship/internal/domain"
	"avito-internship/internal/repository/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	Team() domain.TeamRepository
	User() domain.UserRepository
}

type pgStorage struct {
	team domain.TeamRepository
	user domain.UserRepository
}

func NewPgStorage(db *pgxpool.Pool) Storage {
	return &pgStorage{
		team: postgres.NewTeamRepo(db),
		user: postgres.NewUserRepo(db),
	}
}

func (s *pgStorage) Team() domain.TeamRepository { return s.team }
func (s *pgStorage) User() domain.UserRepository { return s.user }
