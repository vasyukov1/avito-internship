package delivery

import (
	"github.com/gin-gonic/gin"
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
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
