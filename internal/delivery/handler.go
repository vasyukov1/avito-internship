package delivery

import (
	"avito-internship/internal/domain"
	"avito-internship/internal/infrastructure"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Handler struct {
	service *infrastructure.Server
}

func NewHandler(service *infrastructure.Server) *Handler {
	return &Handler{service}
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrTeamExists):
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "TEAM_EXISTS", "message": "team_name already exists"},
		})

	case errors.Is(err, domain.ErrPRExists):
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "PR_EXISTS", "message": "PR id already exists"},
		})

	case errors.Is(err, domain.ErrPRMerged):
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "PR_MERGED", "message": "cannot reassign reviewer for merged PR"},
		})

	case errors.Is(err, domain.ErrNotAssigned):
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "NOT_ASSIGNED", "message": "reviewer is not assigned to this PR"},
		})

	case errors.Is(err, domain.ErrNoCandidate):
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "NO_CANDIDATE", "message": "no active replacement candidate in team"},
		})

	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "NOT_FOUND", "message": "resource not found"},
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
}
