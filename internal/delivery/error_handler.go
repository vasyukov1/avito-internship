package delivery

import (
	"avito-internship/internal/domain"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (h *Handler) handleError(c *gin.Context, err error) {
	fields := logrus.Fields{
		"method":    c.Request.Method,
		"path":      c.Request.URL.Path,
		"client_ip": c.ClientIP(),
	}

	switch {
	case errors.Is(err, domain.ErrTeamExists):
		h.logger.WithFields(fields).Warn("Team already exists")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "TEAM_EXISTS", "message": "team name already exists"},
		})

	case errors.Is(err, domain.ErrPRExists):
		h.logger.WithFields(fields).Warn("PR already exists")
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "PR_EXISTS", "message": "PR id already exists"},
		})

	case errors.Is(err, domain.ErrPRMerged):
		h.logger.WithFields(fields).Warn("Attempt to modify merged PR")
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "PR_MERGED", "message": "cannot reassign reviewer for merged PR"},
		})

	case errors.Is(err, domain.ErrNotAssigned):
		h.logger.WithFields(fields).Warn("Reviewer not assigned")
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "NOT_ASSIGNED", "message": "reviewer is not assigned to this PR"},
		})

	case errors.Is(err, domain.ErrNoCandidate):
		h.logger.WithFields(fields).Warn("No replacement candidate found")
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{"code": "NO_CANDIDATE", "message": "no active replacement candidate in team"},
		})

	case errors.Is(err, domain.ErrNotFound):
		h.logger.WithFields(fields).Warn("Resource not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "NOT_FOUND", "message": "resource not found"},
		})

	default:
		h.logger.WithFields(fields).WithError(err).Error("Internal server error")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
		})
	}
}
