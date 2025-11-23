package delivery

import (
	"avito-internship/internal/domain"
	"avito-internship/internal/metrics"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
		h.logger.WithFields(logrus.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"error":  err.Error(),
		}).Warn("Bad request for SetIsActive")

		metrics.ErrorsTotal.WithLabelValues("bad_request", "set_is_active").Inc()

		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "invalid request body",
			},
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":   req.UserID,
		"is_active": req.IsActive,
	}).Info("Setting user active status")

	// Update user activity
	user, err := h.service.Storage().User().SetIsActive(
		c.Request.Context(),
		req.UserID,
		req.IsActive,
	)
	if err != nil {
		metrics.ErrorsTotal.WithLabelValues("update_failed", "set_is_active").Inc()
		h.logger.WithError(err).Error("Failed to set user active status")
		h.handleError(c, err)
		return
	}

	if req.IsActive {
		metrics.UsersActiveTotal.WithLabelValues(user.TeamName).Inc()
	} else {
		metrics.UsersActiveTotal.WithLabelValues(user.TeamName).Dec()
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":   req.UserID,
		"is_active": req.IsActive,
	}).Info("User active status updated successfully")

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
		h.logger.WithFields(logrus.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
		}).Warn("user_id query parameter is missing")

		metrics.ErrorsTotal.WithLabelValues("bad_request", "get_review").Inc()

		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "user_id is required",
			},
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
	}).Debug("Getting PRs for user")

	// Get user by id
	rows, err := h.service.Storage().PullRequest().GetByUserID(c.Request.Context(), userID)
	if err != nil {
		metrics.ErrorsTotal.WithLabelValues("query_failed", "get_review").Inc()
		h.logger.WithError(err).Error("Failed to get PRs by user ID")
		h.handleError(c, err)
		return
	}

	// Get user's PRs
	pullRequests := make([]PullRequestResponse, len(rows))
	for i, pr := range rows {
		pullRequests[i] = ToResponse(pr)
	}

	metrics.UsersActiveTotal.WithLabelValues("review_assignment").Set(float64(len(pullRequests)))

	h.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"prs_count": len(pullRequests),
	}).Debug("PRs retrieved successfully")

	c.JSON(http.StatusOK, PRbyUserResponse{UserID: userID, PullRequests: pullRequests})
}
