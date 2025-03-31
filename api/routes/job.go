package routes

import (
	"net/http"
	"time"

	"github.com/chip/conveyor/core"
	"github.com/gin-gonic/gin"
)

// JobPayload represents a job creation payload
type JobPayload struct {
	PipelineID string                 `json:"pipelineId" binding:"required"`
	Params     map[string]interface{} `json:"params,omitempty"`
}

// JobResponse represents a job response
type JobResponse struct {
	ID         string                 `json:"id"`
	PipelineID string                 `json:"pipelineId"`
	Status     string                 `json:"status"`
	StartedAt  time.Time              `json:"startedAt"`
	EndedAt    time.Time              `json:"endedAt,omitempty"`
	Steps      []core.StepStatus      `json:"steps,omitempty"`
	Logs       []core.LogEntry        `json:"logs,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RegisterJobRoutes registers job routes
func RegisterJobRoutes(router *gin.RouterGroup, engine *core.PipelineEngine) {
	router.POST("", createJob(engine))
	router.GET("/:id", getJob(engine))
	router.POST("/:id/retry", retryJob(engine))
	router.POST("/:id/cancel", cancelJob(engine))
}

// createJob creates a new job
func createJob(engine *core.PipelineEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload JobPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// In a real implementation, we would validate the pipeline ID and create a job
		// For now, just return a placeholder
		c.JSON(http.StatusAccepted, gin.H{
			"id":         "job-" + time.Now().Format("20060102150405"),
			"pipelineId": payload.PipelineID,
			"status":     "pending",
			"startedAt":  time.Now(),
		})
	}
}

// getJob retrieves a job by ID
func getJob(engine *core.PipelineEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		pipelineID := c.DefaultQuery("pipelineId", "")
		
		// In a real implementation, we would validate the IDs and get the job
		// For now, just return a placeholder
		c.JSON(http.StatusOK, gin.H{
			"id":         id,
			"pipelineId": pipelineID,
			"status":     "running",
			"startedAt":  time.Now().Add(-5 * time.Minute),
			"steps": []gin.H{
				{
					"id":        "step-1",
					"name":      "Build",
					"status":    "success",
					"startedAt": time.Now().Add(-5 * time.Minute),
					"endedAt":   time.Now().Add(-4 * time.Minute),
				},
				{
					"id":        "step-2",
					"name":      "Test",
					"status":    "running",
					"startedAt": time.Now().Add(-3 * time.Minute),
				},
			},
		})
	}
}

// retryJob retries a job
func retryJob(engine *core.PipelineEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		pipelineID := c.DefaultQuery("pipelineId", "")
		
		// In a real implementation, we would validate the IDs and retry the job
		// For now, just return a placeholder
		c.JSON(http.StatusAccepted, gin.H{
			"id":         "job-" + time.Now().Format("20060102150405"),
			"pipelineId": pipelineID,
			"status":     "pending",
			"startedAt":  time.Now(),
			"metadata": gin.H{
				"retryOf": id,
			},
		})
	}
}

// cancelJob cancels a job
func cancelJob(engine *core.PipelineEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		pipelineID := c.DefaultQuery("pipelineId", "")
		
		// In a real implementation, we would validate the IDs and cancel the job
		// For now, just return a placeholder
		c.JSON(http.StatusOK, gin.H{
			"id":         id,
			"pipelineId": pipelineID,
			"status":     "cancelled",
			"startedAt":  time.Now().Add(-5 * time.Minute),
			"endedAt":    time.Now(),
		})
	}
} 