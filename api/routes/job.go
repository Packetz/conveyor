package routes

import (
	"net/http"
	"time"

	"github.com/chip/conveyor/core"
	"github.com/gin-gonic/gin"
)

// RegisterJobRoutes registers all job-related routes
func RegisterJobRoutes(router *gin.Engine, pipelineEngine *core.PipelineEngine) {
	jobGroup := router.Group("/api/jobs")
	{
		// Get all jobs
		jobGroup.GET("", func(c *gin.Context) {
			// Filter by pipeline ID if provided
			pipelineID := c.Query("pipelineId")

			// This would load from a database in a real implementation
			// For now, we'll return a mock response
			jobs := []map[string]interface{}{
				{
					"id":         "job-123",
					"pipelineId": "pipeline-1",
					"status":     "running",
					"startedAt":  time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
					"duration":   "1h 30m",
				},
				{
					"id":         "job-124",
					"pipelineId": "pipeline-1",
					"status":     "success",
					"startedAt":  time.Now().Add(-5 * time.Hour).Format(time.RFC3339),
					"duration":   "45m",
				},
				{
					"id":         "job-125",
					"pipelineId": "pipeline-2",
					"status":     "failed",
					"startedAt":  time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
					"duration":   "20m",
				},
			}

			// Filter by pipeline ID if provided
			if pipelineID != "" {
				filteredJobs := make([]map[string]interface{}, 0)
				for _, job := range jobs {
					if job["pipelineId"] == pipelineID {
						filteredJobs = append(filteredJobs, job)
					}
				}
				jobs = filteredJobs
			}

			c.JSON(http.StatusOK, jobs)
		})

		// Get a single job
		jobGroup.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")

			// In a real implementation, this would query the database
			// For now, we'll return mock data
			job := map[string]interface{}{
				"id":         id,
				"pipelineId": "pipeline-1",
				"status":     "running",
				"startedAt":  time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				"duration":   "1h 30m",
				"steps": []map[string]interface{}{
					{
						"id":     "step-1",
						"name":   "checkout",
						"status": "success",
						"output": "Checked out repository at commit abc123",
					},
					{
						"id":     "step-2",
						"name":   "build",
						"status": "success",
						"output": "Build completed successfully",
					},
					{
						"id":     "step-3",
						"name":   "test",
						"status": "running",
						"output": "Running tests...",
					},
				},
			}

			c.JSON(http.StatusOK, job)
		})

		// Retry a job
		jobGroup.POST("/:id/retry", func(c *gin.Context) {
			id := c.Param("id")

			// In a real implementation, this would restart a job
			// For now, we'll return a mock response
			newJob := map[string]interface{}{
				"id":         "job-" + time.Now().Format("20060102150405"),
				"pipelineId": "pipeline-1",
				"status":     "running",
				"startedAt":  time.Now().Format(time.RFC3339),
				"retryOf":    id,
			}

			c.JSON(http.StatusOK, newJob)
		})

		// Cancel a job
		jobGroup.POST("/:id/cancel", func(c *gin.Context) {
			id := c.Param("id")

			// In a real implementation, this would cancel a running job
			// For now, we'll return a mock response
			job := map[string]interface{}{
				"id":         id,
				"pipelineId": "pipeline-1",
				"status":     "cancelled",
				"startedAt":  time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				"endedAt":    time.Now().Format(time.RFC3339),
				"duration":   "2h 0m",
			}

			c.JSON(http.StatusOK, job)
		})

		// Get job logs
		jobGroup.GET("/:id/logs", func(c *gin.Context) {
			id := c.Param("id")

			// In a real implementation, this would fetch logs from storage
			// For now, we'll return mock data
			logs := []map[string]interface{}{
				{
					"time":    time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
					"level":   "INFO",
					"message": "Job started",
				},
				{
					"time":    time.Now().Add(-1*time.Hour - 45*time.Minute).Format(time.RFC3339),
					"level":   "INFO",
					"message": "Checking out repository",
				},
				{
					"time":    time.Now().Add(-1*time.Hour - 44*time.Minute).Format(time.RFC3339),
					"level":   "INFO",
					"message": "Repository checkout complete",
				},
				{
					"time":    time.Now().Add(-1*time.Hour - 40*time.Minute).Format(time.RFC3339),
					"level":   "INFO",
					"message": "Starting build step",
				},
				{
					"time":    time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
					"level":   "INFO",
					"message": "Build completed successfully",
				},
				{
					"time":    time.Now().Add(-55 * time.Minute).Format(time.RFC3339),
					"level":   "INFO",
					"message": "Starting test step",
				},
				{
					"time":    time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
					"level":   "WARNING",
					"message": "Test failure in module X",
				},
				{
					"time":    time.Now().Add(-15 * time.Minute).Format(time.RFC3339),
					"level":   "INFO",
					"message": "Tests completed with warnings",
				},
			}

			c.JSON(http.StatusOK, logs)
		})
	}
} 