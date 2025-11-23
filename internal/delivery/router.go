package delivery

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(h *Handler) *gin.Engine {
	r := gin.New()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(LoggingMiddleware(h.logger))
	r.Use(MetricsMiddleware())

	// Health check
	r.GET("/health", h.Health)

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Team routes
	h.RegisterTeamRoutes(r)

	// User routes
	h.RegisterUserRoutes(r)

	// PR routes
	h.RegisterPRRoutes(r)

	// Swagger
	r.GET("/swagger/*any", gin.WrapH(httpSwagger.WrapHandler))

	return r
}
