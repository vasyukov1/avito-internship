package delivery

import (
	"github.com/gin-gonic/gin"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(h *Handler) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", h.Health)

	// Team routes
	h.RegisterTeamRoutes(r)

	// Swagger
	r.GET("/swagger/*any", gin.WrapH(httpSwagger.WrapHandler))

	return r
}
