package delivery

import (
	"avito-internship/internal/domain"
	"github.com/gin-gonic/gin"
	"net/http"
)

// CreatePRRequest represents request for creating a pull request
// @Description Запрос на создание пул-реквеста
type CreatePRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	Name          string `json:"pull_request_name" binding:"required"`
	AuthorID      string `json:"author_id" binding:"required"`
}

// CreatePRResponse represents response for creating a pull request
// @Description Ответ с созданным пул-реквестом
type CreatePRResponse struct {
	PullRequest domain.PullRequest `json:"pull_request"`
}

func (h *Handler) RegisterPRRoutes(r *gin.Engine) {
	pr := r.Group("/pr")
	{
		pr.POST("/create", h.CreatePR)
	}
}

// CreatePR godoc
// @Summary Создать PR и автоматически назначить до 2 ревьюверов из команды автора
// @Description Создает новый пул-реквест и автоматически назначает до 2 активных ревьюверов из команды автора. Автор не может быть назначен ревьювером своего же PR.
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param request body CreatePRRequest true "Данные для создания пул-реквеста"
// @Success 201 {object} CreatePRResponse "PR успешно создан"
// @Failure 400 {object} ErrorResponse "Неверный запрос"
// @Failure 404 {object} ErrorResponse "Автор или команда не найдены"
// @Failure 409 {object} ErrorResponse "PR уже существует"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /pr/create [post]
// @Example request {"pull_request_id": "pr-1001", "pull_request_name": "Add search", "author_id": "u1"}
// @Example response 201 {"pull_request": {"pull_request_id": "pr-1001", "pull_request_name": "Add search", "author_id": "u1", "status": "OPEN", "assigned_reviewers": ["u2", "u3"], "createdAt": "2025-01-15T10:30:00Z", "mergedAt": null}}
func (h *Handler) CreatePR(c *gin.Context) {
	var req CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "BAD_REQUEST", "message": err.Error()}})
		return
	}

	pr, err := h.service.CreatePullRequest(c.Request.Context(), req.PullRequestID, req.Name, req.AuthorID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, CreatePRResponse{PullRequest: *pr})
}
