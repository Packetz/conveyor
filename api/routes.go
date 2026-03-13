package api

import (
	"github.com/chip/conveyor/api/routes"
	"github.com/chip/conveyor/core"
	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up all API routes
func SetupRoutes(r *gin.Engine, engine *core.PipelineEngine, pipelineLoader interface {
	LoadFromBytes([]byte, string) (*core.Pipeline, []string, error)
}) {
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

	// Pipeline import route (needs loader)
	if pipelineLoader != nil {
		routes.RegisterPipelineImportRoute(pipelineRoutes, pipelineLoader)
	}

	// Job routes
	jobRoutes := api.Group("/jobs")
	routes.RegisterJobRoutes(jobRoutes, engine)

	// Plugin routes
	pluginRoutes := api.Group("/plugins")
	routes.RegisterPluginRoutes(pluginRoutes)

	// Security routes
	securityRoutes := api.Group("/security")
	routes.RegisterSecurityRoutes(securityRoutes, engine)

	// System stats routes
	api.GET("/system/stats", func(c *gin.Context) {
		routes.GetSystemStats(c)
	})
}
