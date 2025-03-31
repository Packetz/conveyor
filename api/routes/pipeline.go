package routes

import (
	"net/http"
	"time"

	"github.com/chip/conveyor/core"
	"github.com/gin-gonic/gin"
)

// RegisterPipelineRoutes registers all pipeline-related routes
func RegisterPipelineRoutes(router *gin.Engine, pipelineEngine *core.PipelineEngine) {
	pipelineGroup := router.Group("/api/pipelines")
	{
		// Get all pipelines
		pipelineGroup.GET("", func(c *gin.Context) {
			// This would load from a database in a real implementation
			// For now, we'll return a mock response
			pipelines := []map[string]interface{}{
				{
					"id":          "pipeline-1",
					"name":        "Build and Test",
					"description": "Builds and tests the application",
					"status":      "running",
					"createdAt":   time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
					"updatedAt":   time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				},
				{
					"id":          "pipeline-2",
					"name":        "Deploy to Production",
					"description": "Deploys the application to production",
					"status":      "success",
					"createdAt":   time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
					"updatedAt":   time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
				},
				{
					"id":          "pipeline-3",
					"name":        "Security Scan",
					"description": "Performs security scanning of the codebase",
					"status":      "failed",
					"createdAt":   time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
					"updatedAt":   time.Now().Add(-6 * time.Hour).Format(time.RFC3339),
				},
			}

			c.JSON(http.StatusOK, pipelines)
		})

		// Get a single pipeline
		pipelineGroup.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")

			// In a real implementation, this would query the database
			// For now, we'll return mock data
			pipeline := map[string]interface{}{
				"id":          id,
				"name":        "Build and Test",
				"description": "Builds and tests the application",
				"status":      "running",
				"createdAt":   time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
				"updatedAt":   time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				"stages": []map[string]interface{}{
					{
						"id":     "stage-1",
						"name":   "Build",
						"status": "success",
					},
					{
						"id":     "stage-2",
						"name":   "Test",
						"status": "running",
					},
					{
						"id":     "stage-3",
						"name":   "Deploy",
						"status": "pending",
					},
				},
			}

			c.JSON(http.StatusOK, pipeline)
		})

		// Create a new pipeline
		pipelineGroup.POST("", func(c *gin.Context) {
			var pipelineRequest map[string]interface{}
			if err := c.ShouldBindJSON(&pipelineRequest); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
				return
			}

			// In a real implementation, this would create a pipeline in the database
			// For now, we'll return a mock response
			pipeline := map[string]interface{}{
				"id":          "pipeline-new",
				"name":        pipelineRequest["name"],
				"description": pipelineRequest["description"],
				"status":      "created",
				"createdAt":   time.Now().Format(time.RFC3339),
				"updatedAt":   time.Now().Format(time.RFC3339),
			}

			c.JSON(http.StatusCreated, pipeline)
		})

		// Update a pipeline
		pipelineGroup.PUT("/:id", func(c *gin.Context) {
			id := c.Param("id")

			var pipelineRequest map[string]interface{}
			if err := c.ShouldBindJSON(&pipelineRequest); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
				return
			}

			// In a real implementation, this would update a pipeline in the database
			// For now, we'll return a mock response
			pipeline := map[string]interface{}{
				"id":          id,
				"name":        pipelineRequest["name"],
				"description": pipelineRequest["description"],
				"status":      "updated",
				"createdAt":   time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
				"updatedAt":   time.Now().Format(time.RFC3339),
			}

			c.JSON(http.StatusOK, pipeline)
		})

		// Delete a pipeline
		pipelineGroup.DELETE("/:id", func(c *gin.Context) {
			id := c.Param("id")

			// In a real implementation, this would delete a pipeline from the database
			// For now, we'll return a mock response
			c.JSON(http.StatusOK, gin.H{"message": "Pipeline deleted successfully", "id": id})
		})

		// Run a pipeline
		pipelineGroup.POST("/:id/run", func(c *gin.Context) {
			id := c.Param("id")

			// In a real implementation, this would start a pipeline execution
			// For now, we'll return a mock response
			job := map[string]interface{}{
				"id":         "job-123",
				"pipelineId": id,
				"status":     "running",
				"startedAt":  time.Now().Format(time.RFC3339),
			}

			c.JSON(http.StatusOK, job)
		})
	}
} 