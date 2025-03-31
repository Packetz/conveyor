package api

import (
	"github.com/chip/conveyor/api/routes"
	"github.com/chip/conveyor/core"
	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up all API routes
func SetupRoutes(r *gin.Engine, engine *core.PipelineEngine) {
	// API group
	api := r.Group("/api")
	
	// Health endpoint
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	
	// Pipeline routes
	pipelineRoutes := api.Group("/pipelines")
	routes.RegisterPipelineRoutes(pipelineRoutes, engine)
	
	// Job routes
	jobRoutes := api.Group("/jobs")
	routes.RegisterJobRoutes(jobRoutes, engine)
	
	// Security routes
	securityRoutes := api.Group("/security")
	routes.RegisterSecurityRoutes(securityRoutes, engine)
} 