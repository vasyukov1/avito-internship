package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

// Health godoc
// @Summary Проверка здоровья сервиса
// @Description Проверяет, что сервис работает корректно
// @Tags Health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Статус сервиса"
// @Router /health [get]
func (h *Handler) Health(c *gin.Context) {
	h.logger.WithFields(logrus.Fields{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
	}).Debug("Health check requested")

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
