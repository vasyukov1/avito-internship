package infrastructure

import "avito-internship/internal/domain"

type Server struct {
	storage domain.Storage
}

func NewServer(storage domain.Storage) *Server {
	return &Server{storage: storage}
}
