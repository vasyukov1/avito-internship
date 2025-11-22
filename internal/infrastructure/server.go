package infrastructure

import "avito-internship/internal/repository"

type Server struct {
	storage repository.Storage
}

func NewServer(storage repository.Storage) *Server {
	return &Server{storage: storage}
}

func (s *Server) Storage() repository.Storage {
	return s.storage
}
