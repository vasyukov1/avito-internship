package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgStorage struct {
	db *pgxpool.Pool
}

func NewPgStorage(db *pgxpool.Pool) *PgStorage {
	return &PgStorage{db}
}
