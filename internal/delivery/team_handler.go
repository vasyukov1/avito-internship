package delivery

import (
	"avito-internship/internal/domain"
	"github.com/gin-gonic/gin"
	"net/http"
)

// UserRequest represents request for creating a user in a team
// @Description Запрос на создание пользователя в команде
type UserRequest struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// TeamRequest represents request for creating a team
// @Description Запрос на создание команды с участниками
type TeamRequest struct {
	Name    string        `json:"team_name" binding:"required"`
	Members []UserRequest `json:"members"`
}

// TeamResponse represents response for team operations
// @Description Ответ с информацией о команде
type TeamResponse struct {
	Team domain.Team `json:"team"`
}

// ErrorResponse represents error response
// @Description Ответ с ошибкой
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (h *Handler) RegisterTeamRoutes(r *gin.Engine) {
	r.POST("/team/add", h.CreateTeam)
	r.GET("/team/get", h.GetTeam)
}

func (ur *UserRequest) ToDomain(teamName string) domain.User {
	return domain.User{
		ID:       ur.ID,
		Username: ur.Username,
		TeamName: teamName,
		IsActive: ur.IsActive,
	}
}

// CreateTeam godoc
// @Summary Создать команду с участниками
// @Description Создает новую команду и обновляет/создает пользователей. Если команда уже существует, возвращает ошибку.
// @Tags Teams
// @Accept json
// @Produce json
// @Param request body TeamRequest true "Данные команды"
// @Success 201 {object} TeamResponse "Команда успешно создана"
// @Failure 400 {object} ErrorResponse "Неверный запрос или команда уже существует"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /team/add [post]
func (h *Handler) CreateTeam(c *gin.Context) {
	var req TeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "BAD_REQUEST", "message": err.Error()}})
		return
	}

	ctx := c.Request.Context()

	team := domain.Team{Name: req.Name}
	if err := h.service.Storage().Team().CreateTeam(ctx, team); err != nil {
		h.handleError(c, err)
		return
	}

	members := make([]domain.User, len(req.Members))
	for i, userReq := range req.Members {
		members[i] = userReq.ToDomain(req.Name)
	}

	if len(members) > 0 {
		if err := h.service.Storage().User().UpsertUsers(ctx, req.Name, members); err != nil {
			h.handleError(c, err)
			return
		}
	}

	team.Members = members
	c.JSON(http.StatusCreated, TeamResponse{Team: team})
}

// GetTeam godoc
// @Summary Получить команду с участниками
// @Description Возвращает информацию о команде и ее участниках по имени команды
// @Tags Teams
// @Accept json
// @Produce json
// @Param team_name query string true "Уникальное имя команды" example("payments")
// @Success 200 {object} domain.Team "Информация о команде"
// @Failure 400 {object} ErrorResponse "Не указано имя команды"
// @Failure 404 {object} ErrorResponse "Команда не найдена"
// @Router /team/get [get]
func (h *Handler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "BAD_REQUEST", "message": "team_name is required"}})
		return
	}

	ctx := c.Request.Context()
	team, err := h.service.Storage().Team().GetTeam(ctx, teamName)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, team)
}
