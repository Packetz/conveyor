package routes

import (
	"net/http"
	"time"

	"github.com/chip/conveyor/core"
	"github.com/gin-gonic/gin"
)

// SecurityRequest represents a request to run a security scan
type SecurityRequest struct {
	PipelineID       string   `json:"pipelineId"`
	TargetDir        string   `json:"targetDir"`
	ScanTypes        []string `json:"scanTypes"`
	SeverityThreshold string  `json:"severityThreshold"`
	FailOnViolation  bool     `json:"failOnViolation"`
	GenerateSBOM     bool     `json:"generateSBOM"`
	CustomRules      []map[string]interface{} `json:"customRules"`
}

// RegisterSecurityRoutes registers all security-related routes
func RegisterSecurityRoutes(router *gin.RouterGroup, pipelineEngine *core.PipelineEngine) {
	// Get security configuration
	router.GET("/config", func(c *gin.Context) {
		// In a real implementation, we would get this from the security plugin
		// For now, we'll return a mock response
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
	})

	// Update security configuration
	router.PUT("/config", func(c *gin.Context) {
		var config map[string]interface{}
		if err := c.ShouldBindJSON(&config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// In a real implementation, we would update the security plugin configuration
		// For now, we'll just return the config
		c.JSON(http.StatusOK, config)
	})

	// Get all security scans
	router.GET("/scans", func(c *gin.Context) {
		// In a real implementation, we would get this from storage
		// For now, we'll return a mock response
		c.JSON(http.StatusOK, []gin.H{
			{
				"id":            "scan-1",
				"timestamp":     time.Now().Add(-24 * time.Hour),
				"type":          "vulnerability",
				"pipelineId":    "pipeline-1",
				"jobId":         "job-1",
				"status":        "completed",
				"findingsCount": 3,
				"highCount":     1,
				"mediumCount":   2,
				"lowCount":      0,
			},
			{
				"id":            "scan-2",
				"timestamp":     time.Now().Add(-12 * time.Hour),
				"type":          "secret",
				"pipelineId":    "pipeline-1",
				"jobId":         "job-2",
				"status":        "completed",
				"findingsCount": 1,
			},
		})
	})

	// Create a new security scan
	router.POST("/scans", func(c *gin.Context) {
		var scanRequest struct {
			Type       string `json:"type" binding:"required"`
			PipelineID string `json:"pipelineId" binding:"required"`
			JobID      string `json:"jobId" binding:"required"`
		}

		if err := c.ShouldBindJSON(&scanRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// In a real implementation, we would run a security scan
		// For now, we'll just return a mock response
		c.JSON(http.StatusAccepted, gin.H{
			"id":         "scan-" + time.Now().Format("20060102150405"),
			"type":       scanRequest.Type,
			"pipelineId": scanRequest.PipelineID,
			"jobId":      scanRequest.JobID,
			"status":     "pending",
			"timestamp":  time.Now(),
		})
	})

	// Get a specific security scan
	router.GET("/scans/:id", func(c *gin.Context) {
		id := c.Param("id")

		// In a real implementation, we would get this from storage
		// For now, we'll return a mock response based on the ID
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
	})

	// Get scan history for a pipeline
	router.GET("/history/:pipelineId", func(c *gin.Context) {
		pipelineID := c.Param("pipelineId")
		
		// This would load from a database in a real implementation
		// For now, we'll return a mock response
		history := []map[string]interface{}{
			{
				"id": "scan-123",
				"pipelineId": pipelineID,
				"timestamp": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
				"status": "success",
				"findings": 5,
				"duration": "4.2s",
			},
			{
				"id": "scan-124",
				"pipelineId": pipelineID,
				"timestamp": time.Now().Add(-12 * time.Hour).Format(time.RFC3339),
				"status": "failed",
				"findings": 12,
				"duration": "3.8s",
			},
			{
				"id": "scan-125",
				"pipelineId": pipelineID,
				"timestamp": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				"status": "success",
				"findings": 3,
				"duration": "4.5s",
			},
		}
		
		c.JSON(http.StatusOK, history)
	})

	// Get a specific scan result
	router.GET("/scan/:scanId", func(c *gin.Context) {
		scanID := c.Param("scanId")
		
		// In a real implementation, this would load the scan result from a database
		// For now, we'll simulate finding a report file
		
		// Simulated file reading error
		if scanID == "invalid" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Scan report not found"})
			return
		}
		
		// For demo purposes, return mock data
		mockData := generateMockScanResult(scanID)
		c.JSON(http.StatusOK, mockData)
	})

	// Get the latest scan for a pipeline
	router.GET("/latest/:pipelineId", func(c *gin.Context) {
		pipelineID := c.Param("pipelineId")
		
		// In a real implementation, this would query the most recent scan
		// For now, we'll return mock data
		mockData := generateMockScanResult("latest-" + pipelineID)
		c.JSON(http.StatusOK, mockData)
	})
}

// generateMockScanResult creates mock scan data for demonstration purposes
func generateMockScanResult(scanID string) map[string]interface{} {
	findingsBySeverity := map[string]int{
		"CRITICAL": 1,
		"HIGH":     3,
		"MEDIUM":   2,
		"LOW":      0,
		"INFO":     0,
	}

	findings := []map[string]interface{}{
		{
			"ruleId":      "SECRET-001",
			"severity":    "CRITICAL",
			"description": "AWS Access Key ID detected",
			"location":    "config/settings.js",
			"lineNumber":  42,
			"context":     "const awsKey = 'AKIA[REDACTED]';",
			"remediation": "Remove AWS Access Key from code and use environment variables or AWS IAM roles",
		},
		{
			"ruleId":      "CODE-002",
			"severity":    "HIGH",
			"description": "Potential SQL injection vulnerability",
			"location":    "src/controllers/users.js",
			"lineNumber":  87,
			"context":     "const query = `SELECT * FROM users WHERE id = ${userId}`;",
			"remediation": "Use parameterized queries or prepared statements for database operations",
		},
		{
			"ruleId":      "VULN-001",
			"severity":    "HIGH",
			"description": "Axios before 0.21.2 allows server-side request forgery (CVE-2021-3749)",
			"location":    "package.json",
			"lineNumber":  15,
			"context":     "\"axios\": \"^0.21.1\"",
			"remediation": "Update axios to version 0.21.2 or later",
		},
		{
			"ruleId":      "CODE-003",
			"severity":    "MEDIUM",
			"description": "Hardcoded IP address",
			"location":    "src/services/api.js",
			"lineNumber":  12,
			"context":     "const API_HOST = '192.168.1.100';",
			"remediation": "Move IP addresses to configuration files or environment variables",
		},
		{
			"ruleId":      "CODE-001",
			"severity":    "HIGH",
			"description": "Use of insecure random number generator",
			"location":    "src/utils/crypto.js",
			"lineNumber":  8,
			"context":     "const token = Math.random().toString(36).substring(2);",
			"remediation": "Use a cryptographically secure random number generator",
		},
		{
			"ruleId":      "LICENSE-001",
			"severity":    "MEDIUM",
			"description": "Detected GPL-3.0 license which may conflict with project requirements",
			"location":    "node_modules/some-gpl-lib/LICENSE",
			"lineNumber":  1,
			"context":     "GNU General Public License v3.0",
			"remediation": "Review license compatibility with legal team",
		},
	}

	// Calculate total findings
	totalFindings := 0
	for _, count := range findingsBySeverity {
		totalFindings += count
	}

	components := []map[string]interface{}{
		{
			"name":    "axios",
			"version": "0.21.1",
			"type":    "npm",
			"license": "MIT",
			"source":  "https://www.npmjs.com/package/axios",
			"vulnerabilities": []map[string]interface{}{
				{
					"id":          "CVE-2021-3749",
					"severity":    "HIGH",
					"cvss":        8.1,
					"description": "Axios before 0.21.2 allows server-side request forgery",
					"fixedIn":     "0.21.2",
				},
			},
		},
		{
			"name":    "lodash",
			"version": "4.17.20",
			"type":    "npm",
			"license": "MIT",
			"source":  "https://www.npmjs.com/package/lodash",
		},
	}

	return map[string]interface{}{
		"id":        scanID,
		"timestamp": time.Now().Format(time.RFC3339),
		"environment": "development",
		"duration":  "5.2s",
		"findings":  findings,
		"summary": map[string]interface{}{
			"totalFiles":         120,
			"filesScanned":       98,
			"filesSkipped":       22,
			"totalFindings":      totalFindings,
			"findingsBySeverity": findingsBySeverity,
			"passedCheck":        false,
		},
		"sbom": map[string]interface{}{
			"components": components,
			"format":     "cyclonedx",
			"version":    "1.0",
		},
	}
} 