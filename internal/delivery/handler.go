package delivery

import (
	"avito-internship/internal/infrastructure"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	service *infrastructure.Server
	logger  *logrus.Logger
}

func NewHandler(service *infrastructure.Server, logger *logrus.Logger) *Handler {
	return &Handler{service, logger}
}
