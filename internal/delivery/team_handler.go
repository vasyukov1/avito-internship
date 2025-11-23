package delivery

import (
	"avito-internship/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

// TeamMember represents request and response for an user in a team
// @Description Запрос и ответ о пользователе в команде
type TeamMember struct {
	ID       string `json:"user_id" binding:"required"`
	Username string `json:"username" binding:"required"`
	IsActive bool   `json:"is_active" binding:"required"`
}

// TeamRequest represents request for creating a team
// @Description Запрос на создание команды с участниками
type TeamRequest struct {
	Name    string       `json:"team_name" binding:"required"`
	Members []TeamMember `json:"members" binding:"required"`
}

// TeamResponse represents response for team operations
// @Description Ответ с информацией о команде
type TeamResponse struct {
	Name    string       `json:"team_name"`
	Members []TeamMember `json:"members"`
}

// ErrorResponse represents error response
// @Description Ответ с ошибкой
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code" binding:"required"`
		Message string `json:"message" binding:"required"`
	} `json:"error"`
}

func (h *Handler) RegisterTeamRoutes(r *gin.Engine) {
	teams := r.Group("/team")
	{
		teams.POST("/add", h.CreateTeam)
		teams.GET("/get", h.GetTeam)
	}
}

func (tm *TeamMember) ToDomain(teamName string) domain.User {
	return domain.User{
		ID:       tm.ID,
		Username: tm.Username,
		TeamName: teamName,
		IsActive: tm.IsActive,
	}
}

// CreateTeam godoc
// @Summary Создать команду с участниками
// @Description Создает новую команду и обновляет/создает пользователей. Если команда уже существует, возвращает ошибку.
// @Tags Teams
// @Accept json
// @Produce json
// @Param request body TeamRequest true "Данные команды"
// @Success 201 {object} TeamResponse "Команда создана"
// @Failure 400 {object} ErrorResponse "Команда уже существует"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /team/add [post]
func (h *Handler) CreateTeam(c *gin.Context) {
	var req TeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"error":  err.Error(),
		}).Warn("Bad request for CreateTeam")

		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"team_name": req.Name,
		"members":   len(req.Members),
	}).Info("Creating team")

	ctx := c.Request.Context()

	// Create team in database
	team := domain.Team{Name: req.Name}
	if err := h.service.Storage().Team().CreateTeam(ctx, team); err != nil {
		h.logger.WithError(err).Error("Failed to create team in database")
		h.handleError(c, err)
		return
	}

	// Convert members from request to domain
	members := make([]domain.User, len(req.Members))
	for i, userReq := range req.Members {
		members[i] = userReq.ToDomain(req.Name)
	}

	// Save members in database
	if len(members) > 0 {
		if err := h.service.Storage().User().UpsertUsers(ctx, req.Name, members); err != nil {
			h.logger.WithError(err).Error("Failed to upsert users in database")
			h.handleError(c, err)
			return
		}
	}

	h.logger.WithFields(logrus.Fields{
		"team_name": req.Name,
	}).Info("Team created successfully")

	c.JSON(http.StatusCreated, gin.H{"team": TeamResponse{Name: req.Name, Members: req.Members}})
}

// GetTeam godoc
// @Summary Получить команду с участниками
// @Description Возвращает информацию о команде и ее участниках по имени команды
// @Tags Teams
// @Accept json
// @Produce json
// @Param team_name query string true "Уникальное имя команды" example("payments")
// @Success 200 {object} domain.Team "Объект команды"
// @Failure 400 {object} ErrorResponse "Некорректные данные запроса"
// @Failure 404 {object} ErrorResponse "Команда не найдена"
// @Failure 500 {object} ErrorResponse "Внутренняя ошибка сервера"
// @Router /team/get [get]
func (h *Handler) GetTeam(c *gin.Context) {
	// Get team name
	teamName := c.Query("team_name")
	if teamName == "" {
		h.logger.WithFields(logrus.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
		}).Warn("team_name query parameter is missing")

		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "team_name is required",
			},
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"team_name": teamName,
	}).Debug("Getting team")

	// Get team info
	ctx := c.Request.Context()
	team, err := h.service.Storage().Team().GetTeam(ctx, teamName)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get team from database")
		h.handleError(c, err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"team_name": teamName,
		"members":   len(team.Members),
	}).Debug("Team retrieved successfully")

	c.JSON(http.StatusOK, team)
}
