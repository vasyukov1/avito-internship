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

// MergePRRequest represents request body for merging a pull request
// @Description Запрос на мердж
type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
}

// ReassignReviewerRequest represents request for reassigning a reviewer
// @Description Запрос на переназначение ревьювера
type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	OldReviewerID string `json:"old_reviewer_id" binding:"required"`
}

// ReassignReviewerResponse represents response for reassigning a reviewer
// @Description Ответ с переназначением ревьювера
type ReassignReviewerResponse struct {
	PullRequest domain.PullRequest `json:"pr"`
	ReplacedBy  string             `json:"replaced_by"`
}

func (h *Handler) RegisterPRRoutes(r *gin.Engine) {
	pr := r.Group("/pullRequest")
	{
		pr.POST("/create", h.CreatePR)
		pr.POST("/merge", h.MergePR)
		pr.POST("/reassign", h.ReassignReviewer)
	}
}

// CreatePR godoc
// @Summary Создать PR и автоматически назначить до 2 ревьюверов
// @Description Создает новый пул-реквест и автоматически назначает до двух активных ревьюверов из команды автора. Автор не может быть ревьювером своего PR.
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param request body CreatePRRequest true "Данные для создания пул-реквеста"
// @Success 201 {object} CreatePRResponse "PR успешно создан"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 404 {object} ErrorResponse "Автор или команда не найдены"
// @Failure 409 {object} ErrorResponse "PR с таким ID уже существует"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /pullRequest/create [post]
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

// MergePR godoc
// @Summary Пометить PR как MERGED
// @Description Идемпотентная операция: пометить существующий PR как MERGED. Повторный вызов для уже MERGED PR возвращает успех. pull_request_id передаётся в теле запроса.
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param request body MergePRRequest true "ID пул-реквеста для пометки как MERGED"
// @Success 200 {object} CreatePRResponse "PR в состоянии MERGED"
// @Failure 400 {object} ErrorResponse "Некорректные данные запроса"
// @Failure 404 {object} ErrorResponse "PR не найден"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /pullRequest/merge [post]
// @Example request {"pull_request_id": "pr-1001"}
// @Example response 200 {"pull_request": {"pull_request_id": "pr-1001", "pull_request_name": "Add search", "author_id": "u1", "status": "MERGED", "assigned_reviewers": ["u2", "u3"], "createdAt": "2025-01-15T10:30:00Z", "mergedAt": "2025-10-24T12:34:56Z"}}
func (h *Handler) MergePR(c *gin.Context) {
	var req MergePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "BAD_REQUEST", "message": err.Error()}})
		return
	}

	pr, err := h.service.MergePullRequest(c.Request.Context(), req.PullRequestID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, CreatePRResponse{PullRequest: *pr})
}

// ReassignReviewer godoc
// @Summary Переназначить ревьювера на другого члена команды
// @Description Заменяет одного ревьювера в PR на другого активного пользователя из команды автора. Для закрытых или MERGED PR переназначение запрещено.
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param request body ReassignReviewerRequest true "Данные для переназначения ревьювера"
// @Success 200 {object} ReassignReviewerResponse "Ревьювер успешно переназначен"
// @Failure 400 {object} ErrorResponse "Некорректные данные"
// @Failure 404 {object} ErrorResponse "PR или пользователь не найден"
// @Failure 409 {object} ErrorResponse "Бизнес-правило нарушено (например, ревьювер отсутствует у PR)"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /pullRequest/reassign [post]
// @Example request {"pull_request_id": "pr-1001", "old_reviewer_id": "u2"}
// @Example response 200 {"pr": {"pull_request_id": "pr-1001","pull_request_name":"Add search","author_id":"u1","status":"OPEN","assigned_reviewers":["u3","u4"],"createdAt":"2025-01-15T10:30:00Z","mergedAt":null}, "replaced_by": "u7"}
func (h *Handler) ReassignReviewer(c *gin.Context) {
	var req ReassignReviewerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "BAD_REQUEST", "message": err.Error()}})
		return
	}

	pr, newReviewer, err := h.service.ReassignReviewer(
		c.Request.Context(),
		req.PullRequestID,
		req.OldReviewerID,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ReassignReviewerResponse{PullRequest: *pr, ReplacedBy: newReviewer})
}
