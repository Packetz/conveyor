package api

import (
	"net/http"
	"time"

	"github.com/chip/conveyor/core"
	"github.com/gin-gonic/gin"
)

// API is the main handler for the API
type API struct {
	engine    *core.PipelineEngine
	pipelines map[string]*core.Pipeline
	plugins   map[string]core.Plugin
}

// NewAPI creates a new API
func NewAPI(engine *core.PipelineEngine) *API {
	return &API{
		engine:    engine,
		pipelines: make(map[string]*core.Pipeline),
		plugins:   make(map[string]core.Plugin),
	}
}

// RegisterRoutes registers the API routes
func (a *API) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	
	// Health endpoint
	api.GET("/health", a.GetHealth)
	
	// Pipeline endpoints
	pipelines := api.Group("/pipelines")
	{
		pipelines.GET("", a.ListPipelines)
		pipelines.POST("", a.CreatePipeline)
		pipelines.GET("/:id", a.GetPipeline)
		pipelines.PUT("/:id", a.UpdatePipeline)
		pipelines.DELETE("/:id", a.DeletePipeline)
		pipelines.POST("/:id/execute", a.ExecutePipeline)
		pipelines.GET("/:id/jobs", a.ListPipelineJobs)
		pipelines.GET("/:id/jobs/:jobID", a.GetPipelineJob)
		pipelines.POST("/:id/jobs/:jobID/retry", a.RetryPipelineJob)
	}
	
	// Plugin endpoints
	plugins := api.Group("/plugins")
	{
		plugins.GET("", a.ListPlugins)
		plugins.GET("/:name", a.GetPlugin)
	}
	
	// Security endpoints
	security := api.Group("/security")
	{
		security.GET("/config", a.GetSecurityConfig)
		security.PUT("/config", a.UpdateSecurityConfig)
		security.GET("/scans", a.ListSecurityScans)
		security.POST("/scans", a.CreateSecurityScan)
		security.GET("/scans/:id", a.GetSecurityScan)
	}
}

// GetHealth returns the health status of the API
func (a *API) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// CreatePipeline creates a new pipeline
func (a *API) CreatePipeline(c *gin.Context) {
	var pipeline core.Pipeline
	if err := c.ShouldBindJSON(&pipeline); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err := a.engine.CreatePipeline(&pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, pipeline)
}

// GetPipeline retrieves a pipeline by ID
func (a *API) GetPipeline(c *gin.Context) {
	id := c.Param("id")
	pipeline, err := a.engine.GetPipeline(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, pipeline)
}

// UpdatePipeline updates a pipeline
func (a *API) UpdatePipeline(c *gin.Context) {
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
	existing, err := a.engine.GetPipeline(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	// Delete the old pipeline
	err = a.engine.DeletePipeline(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Preserve creation time
	pipeline.CreatedAt = existing.CreatedAt
	pipeline.UpdatedAt = time.Now()
	
	// Create the updated pipeline
	err = a.engine.CreatePipeline(&pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, pipeline)
}

// DeletePipeline deletes a pipeline
func (a *API) DeletePipeline(c *gin.Context) {
	id := c.Param("id")
	err := a.engine.DeletePipeline(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ListPipelines returns all pipelines
func (a *API) ListPipelines(c *gin.Context) {
	pipelines := a.engine.ListPipelines()
	c.JSON(http.StatusOK, pipelines)
}

// ExecutePipeline executes a pipeline
func (a *API) ExecutePipeline(c *gin.Context) {
	id := c.Param("id")
	err := a.engine.ExecutePipeline(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusAccepted, gin.H{"status": "executing"})
}

// ListPipelineJobs returns all jobs for a pipeline
func (a *API) ListPipelineJobs(c *gin.Context) {
	id := c.Param("id")
	jobs, err := a.engine.ListJobs(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, jobs)
}

// GetPipelineJob retrieves a job by ID
func (a *API) GetPipelineJob(c *gin.Context) {
	pipelineID := c.Param("id")
	jobID := c.Param("jobID")
	
	job, err := a.engine.GetJob(pipelineID, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, job)
}

// RetryPipelineJob retries a job
func (a *API) RetryPipelineJob(c *gin.Context) {
	pipelineID := c.Param("id")
	jobID := c.Param("jobID")
	
	err := a.engine.RetryJob(pipelineID, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusAccepted, gin.H{"status": "retrying"})
}

// ListPlugins returns all registered plugins
func (a *API) ListPlugins(c *gin.Context) {
	// In a real implementation, we would get this from the engine
	// For now, return a placeholder
	c.JSON(http.StatusOK, []gin.H{
		{
			"name":        "security",
			"version":     "1.0.0",
			"description": "Security scanning plugin",
			"author":      "Conveyor Team",
			"type":        "scanner",
			"stepTypes":   []string{"vulnerability-scan", "secret-scan", "license-scan"},
		},
	})
}

// GetPlugin retrieves a plugin by name
func (a *API) GetPlugin(c *gin.Context) {
	name := c.Param("name")
	// In a real implementation, we would get this from the engine
	// For now, return a placeholder based on the name
	if name == "security" {
		c.JSON(http.StatusOK, gin.H{
			"name":        "security",
			"version":     "1.0.0",
			"description": "Security scanning plugin",
			"author":      "Conveyor Team",
			"type":        "scanner",
			"stepTypes":   []string{"vulnerability-scan", "secret-scan", "license-scan"},
		})
		return
	}
	
	c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found"})
}

// GetSecurityConfig retrieves the security configuration
func (a *API) GetSecurityConfig(c *gin.Context) {
	// In a real implementation, we would get this from the security plugin
	// For now, return a placeholder
	c.JSON(http.StatusOK, gin.H{
		"vulnerabilityScan": gin.H{
			"enabled":     true,
			"threshold":   "medium",
			"excludeDeps": []string{"dev-dependencies"},
		},
		"secretScan": gin.H{
			"enabled":  true,
			"patterns": []string{"api_key", "password", "secret", "key"},
		},
		"licenseScan": gin.H{
			"enabled":     true,
			"allowedList": []string{"MIT", "Apache-2.0", "BSD-3-Clause"},
			"blockedList": []string{"GPL-3.0"},
		},
	})
}

// UpdateSecurityConfig updates the security configuration
func (a *API) UpdateSecurityConfig(c *gin.Context) {
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// In a real implementation, we would update the configuration in the security plugin
	// For now, just return success
	c.JSON(http.StatusOK, config)
}

// ListSecurityScans returns all security scans
func (a *API) ListSecurityScans(c *gin.Context) {
	// In a real implementation, we would get this from the security plugin
	// For now, return a placeholder
	c.JSON(http.StatusOK, []gin.H{
		{
			"id":           "scan-1",
			"timestamp":    time.Now().Add(-24 * time.Hour),
			"type":         "vulnerability",
			"pipelineId":   "pipeline-1",
			"jobId":        "job-1",
			"status":       "completed",
			"findingsCount": 3,
			"highCount":     1,
			"mediumCount":   2,
			"lowCount":      0,
		},
		{
			"id":           "scan-2",
			"timestamp":    time.Now().Add(-12 * time.Hour),
			"type":         "secret",
			"pipelineId":   "pipeline-1",
			"jobId":        "job-2",
			"status":       "completed",
			"findingsCount": 1,
		},
	})
}

// CreateSecurityScan creates a new security scan
func (a *API) CreateSecurityScan(c *gin.Context) {
	var scanRequest struct {
		Type       string `json:"type" binding:"required"`
		PipelineID string `json:"pipelineId" binding:"required"`
		JobID      string `json:"jobId" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&scanRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// In a real implementation, we would trigger a scan in the security plugin
	// For now, just return a placeholder
	c.JSON(http.StatusAccepted, gin.H{
		"id":          "scan-" + time.Now().Format("20060102150405"),
		"type":        scanRequest.Type,
		"pipelineId":  scanRequest.PipelineID,
		"jobId":       scanRequest.JobID,
		"status":      "pending",
		"timestamp":   time.Now(),
	})
}

// GetSecurityScan retrieves a security scan by ID
func (a *API) GetSecurityScan(c *gin.Context) {
	id := c.Param("id")
	// In a real implementation, we would get this from the security plugin
	// For now, return a placeholder based on the ID
	c.JSON(http.StatusOK, gin.H{
		"id":            id,
		"timestamp":     time.Now().Add(-6 * time.Hour),
		"type":          "vulnerability",
		"pipelineId":    "pipeline-1",
		"jobId":         "job-3",
		"status":        "completed",
		"findingsCount": 5,
		"highCount":     2,
		"mediumCount":   2,
		"lowCount":      1,
		"findings": []gin.H{
			{
				"id":          "CVE-2021-1234",
				"severity":    "high",
				"package":     "lodash",
				"version":     "4.17.20",
				"description": "Prototype pollution vulnerability in lodash",
				"fixVersion":  "4.17.21",
			},
			{
				"id":          "CVE-2021-5678",
				"severity":    "high",
				"package":     "express",
				"version":     "4.17.1",
				"description": "Memory leak in Express.js",
				"fixVersion":  "4.17.2",
			},
		},
	})
} 