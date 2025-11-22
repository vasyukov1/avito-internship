package delivery

import (
	"avito-internship/internal/domain"
	"github.com/gin-gonic/gin"
	"net/http"
)

// setActiveRequest represents request for setting user activity status
// @Description Запрос на установку флага активности пользователя
type setActiveRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive bool   `json:"is_active" binding:"required"`
}

// UserResponse represents response for user operations
// @Description Ответ с информацией о пользователе
type UserResponse struct {
	User *domain.User `json:"user"`
}

func (h *Handler) RegisterUserRoutes(r *gin.Engine) {
	users := r.Group("/users")
	{
		users.POST("/setIsActive", h.SetIsActive)
	}
}

// SetIsActive godoc
// @Summary Установить флаг активности пользователя
// @Description Обновляет статус активности пользователя. Если пользователь не найден, возвращает ошибку.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body setActiveRequest true "Данные для обновления активности пользователя"
// @Success 200 {object} UserResponse "Обновлённый пользователь"
// @Failure 400 {object} ErrorResponse "Неверный запрос"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /users/setIsActive [post]
func (h *Handler) SetIsActive(c *gin.Context) {
	var req setActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "invalid request body",
			},
		})
		return
	}

	user, err := h.service.Storage().User().SetIsActive(c.Request.Context(), req.UserID, req.IsActive)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
