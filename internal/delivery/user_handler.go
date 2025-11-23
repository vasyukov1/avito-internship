package delivery

import (
	"avito-internship/internal/domain"
	"github.com/gin-gonic/gin"
	"net/http"
)

// SetActiveRequest represents request for setting user activity status
// @Description Запрос на установку флага активности пользователя
type SetActiveRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive bool   `json:"is_active" binding:"required"`
}

// UserResponse represents response for user operations
// @Description Ответ с информацией о пользователе
type UserResponse struct {
	User *domain.User `json:"user"`
}

// UserIDRequest represents request for user ID
// @Description Запрос с идентификатором пользователя
type UserIDRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// PullRequestResponse represents short pull request information
// @Description Краткая информация о пул-реквесте
type PullRequestResponse struct {
	ID       string                   `json:"pull_request_id"`
	Name     string                   `json:"pull_request_name"`
	AuthorID string                   `json:"author_id"`
	Status   domain.PullRequestStatus `json:"status"`
}

// PRbyUserResponse represents response with user's pull requests
// @Description Ответ со списком PR, где пользователь назначен ревьювером
type PRbyUserResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestResponse `json:"pull_requests"`
}

func (h *Handler) RegisterUserRoutes(r *gin.Engine) {
	users := r.Group("/users")
	{
		users.POST("/setIsActive", h.SetIsActive)
		users.GET("/getReview", h.GetReview)
	}
}

func ToResponse(pr domain.PullRequest) PullRequestResponse {
	return PullRequestResponse{
		ID:       pr.ID,
		Name:     pr.Name,
		AuthorID: pr.AuthorID,
		Status:   pr.Status,
	}
}

// SetIsActive godoc
// @Summary Установить флаг активности пользователя
// @Description Обновляет статус активности пользователя. Если пользователь не найден, возвращает ошибку.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body SetActiveRequest true "Данные для обновления активности пользователя"
// @Success 200 {object} UserResponse "Обновлённый пользователь"
// @Failure 400 {object} ErrorResponse "Некорректные данные запроса"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /users/setIsActive [post]
func (h *Handler) SetIsActive(c *gin.Context) {
	// Get user request
	var req SetActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "invalid request body",
			},
		})
		return
	}

	// Update user activity
	user, err := h.service.Storage().User().SetIsActive(
		c.Request.Context(),
		req.UserID,
		req.IsActive,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// GetReview godoc
// @Summary Получить PR'ы, где пользователь назначен ревьювером
// @Description Возвращает список пул-реквестов, в которых пользователь назначен ревьювером
// @Tags Users
// @Accept json
// @Produce json
// @Param user_id query string true "Идентификатор пользователя" example("u2")
// @Success 200 {object} PRbyUserResponse "Список PR'ов пользователя"
// @Failure 400 {object} ErrorResponse "Некорректные данные запроса"
// @Failure 404 {object} ErrorResponse "Пользователь не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /users/getReview [get]
func (h *Handler) GetReview(c *gin.Context) {
	// Get user id
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "user_id is required",
			},
		})
		return
	}

	// Get user by id
	rows, err := h.service.Storage().PullRequest().GetByUserID(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Get user's PRs
	pullRequests := make([]PullRequestResponse, len(rows))
	for i, pr := range rows {
		pullRequests[i] = ToResponse(pr)
	}

	c.JSON(http.StatusOK, PRbyUserResponse{UserID: userID, PullRequests: pullRequests})
}
