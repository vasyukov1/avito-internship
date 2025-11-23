package delivery

import (
	"avito-internship/internal/infrastructure"
)

type Handler struct {
	service *infrastructure.Server
}

func NewHandler(service *infrastructure.Server) *Handler {
	return &Handler{service}
}
