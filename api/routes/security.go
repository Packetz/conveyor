package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/chip/conveyor/core"
	"github.com/chip/conveyor/plugins/security"
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
	CustomRules      []security.Rule `json:"customRules"`
}

// RegisterSecurityRoutes registers all security-related routes
func RegisterSecurityRoutes(router *gin.Engine, pipelineEngine *core.PipelineEngine) {
	securityGroup := router.Group("/api/security")
	{
		// Get security scan configuration
		securityGroup.GET("/config", func(c *gin.Context) {
			// This would load from a configuration store in a real implementation
			config := map[string]interface{}{
				"scanTypes": []string{"secret", "vulnerability", "code", "license"},
				"severityThreshold": "HIGH",
				"ignorePatterns": []string{"node_modules/", "vendor/", ".git/"},
				"failOnViolation": true,
				"generateSBOM": true,
				"sbomFormat": "cyclonedx",
				"outputDir": "security-reports",
			}
			
			c.JSON(http.StatusOK, config)
		})

		// Run a security scan
		securityGroup.POST("/scan", func(c *gin.Context) {
			var req SecurityRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
				return
			}

			// Set defaults if not provided
			if req.TargetDir == "" {
				req.TargetDir = "."
			}
			if len(req.ScanTypes) == 0 {
				req.ScanTypes = []string{"secret", "vulnerability", "code"}
			}
			if req.SeverityThreshold == "" {
				req.SeverityThreshold = "HIGH"
			}

			// Create a step for the security scan
			step := core.Step{
				ID:   fmt.Sprintf("security-scan-%d", time.Now().Unix()),
				Name: "security-scan",
				Type: "plugin",
				Plugin: "security-scanner",
				Config: map[string]interface{}{
					"targetDir":         req.TargetDir,
					"scanTypes":         req.ScanTypes,
					"severityThreshold": req.SeverityThreshold,
					"failOnViolation":   req.FailOnViolation,
					"generateSBOM":      req.GenerateSBOM,
				},
			}

			// Add custom rules if provided
			if len(req.CustomRules) > 0 {
				step.Config["customRules"] = req.CustomRules
			}

			// Create a security plugin instance
			plugin, err := security.NewSecurityPlugin("")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize security scanner", "details": err.Error()})
				return
			}

			// Run the scan
			ctx := c.Request.Context()
			result, err := plugin.Execute(ctx, step)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Security scan failed", "details": err.Error()})
				return
			}

			// Return the result
			c.JSON(http.StatusOK, result)
		})

		// Get scan history for a pipeline
		securityGroup.GET("/history/:pipelineId", func(c *gin.Context) {
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
		securityGroup.GET("/scan/:scanId", func(c *gin.Context) {
			scanID := c.Param("scanId")
			
			// In a real implementation, this would load the scan result from a database
			// For now, we'll simulate finding a report file
			reportPath := filepath.Join("security-reports", fmt.Sprintf("%s.json", scanID))
			
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
		securityGroup.GET("/latest/:pipelineId", func(c *gin.Context) {
			pipelineID := c.Param("pipelineId")
			
			// In a real implementation, this would query the most recent scan
			// For now, we'll return mock data
			mockData := generateMockScanResult("latest-" + pipelineID)
			c.JSON(http.StatusOK, mockData)
		})
	}
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