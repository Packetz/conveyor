package routes

import (
	"io"
	"net/http"
	"time"

	"github.com/chip/conveyor/core"
	"github.com/gin-gonic/gin"
)

// RegisterPipelineRoutes registers all pipeline-related routes
func RegisterPipelineRoutes(router *gin.RouterGroup, engine *core.PipelineEngine) {
	// Get all pipelines
	router.GET("", func(c *gin.Context) {
		pipelines := engine.ListPipelines()
		c.JSON(http.StatusOK, pipelines)
	})

	// Get a single pipeline
	router.GET("/:id", func(c *gin.Context) {
		id := c.Param("id")
		pipeline, err := engine.GetPipeline(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, pipeline)
	})

	// Create a new pipeline
	router.POST("", func(c *gin.Context) {
		var pipeline core.Pipeline
		if err := c.ShouldBindJSON(&pipeline); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		now := time.Now()
		pipeline.CreatedAt = now
		pipeline.UpdatedAt = now
		
		err := engine.CreatePipeline(&pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusCreated, pipeline)
	})

	// Update a pipeline
	router.PUT("/:id", func(c *gin.Context) {
		id := c.Param("id")
		
		var pipeline core.Pipeline
		if err := c.ShouldBindJSON(&pipeline); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Ensure the ID matches
		if pipeline.ID != id {
			c.JSON(http.StatusBadRequest, gin.H{"error": "pipeline ID in URL does not match payload"})
			return
		}
		
		// Get the existing pipeline
		existing, err := engine.GetPipeline(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		
		// Delete the old pipeline
		err = engine.DeletePipeline(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		// Preserve creation time
		pipeline.CreatedAt = existing.CreatedAt
		pipeline.UpdatedAt = time.Now()
		
		// Create the updated pipeline
		err = engine.CreatePipeline(&pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, pipeline)
	})

	// Delete a pipeline
	router.DELETE("/:id", func(c *gin.Context) {
		id := c.Param("id")
		err := engine.DeletePipeline(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	// Execute a pipeline
	router.POST("/:id/execute", func(c *gin.Context) {
		id := c.Param("id")
		err := engine.ExecutePipeline(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusAccepted, gin.H{"status": "executing"})
	})

	// Get pipeline jobs
	router.GET("/:id/jobs", func(c *gin.Context) {
		id := c.Param("id")
		jobs, err := engine.ListJobs(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, jobs)
	})

	// Get a specific job
	router.GET("/:id/jobs/:jobId", func(c *gin.Context) {
		pipelineID := c.Param("id")
		jobID := c.Param("jobId")
		
		job, err := engine.GetJob(pipelineID, jobID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, job)
	})

	// Retry a job
	router.POST("/:id/jobs/:jobId/retry", func(c *gin.Context) {
		pipelineID := c.Param("id")
		jobID := c.Param("jobId")

		err := engine.RetryJob(pipelineID, jobID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"status": "retrying"})
	})
}

// RegisterPipelineImportRoute registers the YAML pipeline import route.
func RegisterPipelineImportRoute(router *gin.RouterGroup, pipelineLoader interface {
	LoadFromBytes([]byte, string) (*core.Pipeline, []string, error)
}) {
	router.POST("/import", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'name' is required"})
			return
		}

		body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20)) // 1MB limit
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}

		pipeline, warnings, err := pipelineLoader.LoadFromBytes(body, name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "warnings": warnings})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"pipeline": pipeline,
			"warnings": warnings,
		})
	})
} 