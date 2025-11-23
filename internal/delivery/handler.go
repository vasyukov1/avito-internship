package delivery

import (
	"avito-internship/internal/usecase"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	service *usecase.Server
	logger  *logrus.Logger
}

func NewHandler(service *usecase.Server, logger *logrus.Logger) *Handler {
	return &Handler{service, logger}
}
