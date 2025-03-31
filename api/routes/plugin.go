package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterPluginRoutes registers all plugin-related routes
func RegisterPluginRoutes(router *gin.Engine) {
	pluginGroup := router.Group("/api/plugins")
	{
		// Get all plugins
		pluginGroup.GET("", func(c *gin.Context) {
			// This would load from a plugin registry in a real implementation
			// For now, we'll return a mock response
			plugins := []map[string]interface{}{
				{
					"name":        "security-scanner",
					"version":     "2.0.0",
					"description": "Comprehensive security scanner for code, dependencies, and configurations",
					"author":      "Conveyor Team",
					"homepage":    "https://conveyor.example.com/plugins/security",
					"license":     "MIT",
					"type":        "security",
					"icon":        "shield-check",
					"installed":   true,
					"enabled":     true,
					"installedAt": time.Now().Add(-7 * 24 * time.Hour).Format(time.RFC3339),
				},
				{
					"name":        "docker-build",
					"version":     "1.2.0",
					"description": "Build Docker images with advanced caching and optimization",
					"author":      "Conveyor Team",
					"homepage":    "https://conveyor.example.com/plugins/docker-build",
					"license":     "MIT",
					"type":        "builder",
					"icon":        "docker",
					"installed":   true,
					"enabled":     true,
					"installedAt": time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339),
				},
				{
					"name":        "aws-deploy",
					"version":     "1.0.1",
					"description": "Deploy applications to AWS services",
					"author":      "Conveyor Team",
					"homepage":    "https://conveyor.example.com/plugins/aws-deploy",
					"license":     "MIT",
					"type":        "deployment",
					"icon":        "cloud-upload",
					"installed":   true,
					"enabled":     false,
					"installedAt": time.Now().Add(-2 * 24 * time.Hour).Format(time.RFC3339),
				},
				{
					"name":        "test-reporter",
					"version":     "1.1.0",
					"description": "Generate test reports with coverage metrics and visualizations",
					"author":      "Conveyor Team",
					"homepage":    "https://conveyor.example.com/plugins/test-reporter",
					"license":     "MIT",
					"type":        "reporting",
					"icon":        "chart-bar",
					"installed":   false,
					"enabled":     false,
				},
			}

			c.JSON(http.StatusOK, plugins)
		})

		// Get a single plugin
		pluginGroup.GET("/:name", func(c *gin.Context) {
			name := c.Param("name")

			// In a real implementation, this would query the plugin registry
			// For now, we'll return mock data based on the requested plugin name
			var plugin map[string]interface{}

			if name == "security-scanner" {
				plugin = map[string]interface{}{
					"name":        "security-scanner",
					"version":     "2.0.0",
					"description": "Comprehensive security scanner for code, dependencies, and configurations",
					"author":      "Conveyor Team",
					"homepage":    "https://conveyor.example.com/plugins/security",
					"license":     "MIT",
					"type":        "security",
					"icon":        "shield-check",
					"installed":   true,
					"enabled":     true,
					"installedAt": time.Now().Add(-7 * 24 * time.Hour).Format(time.RFC3339),
					"stepTypes": []string{
						"security-scan",
						"vulnerability-scan",
						"secret-scan",
						"sbom-generate",
					},
					"categories": []string{
						"security",
						"quality",
						"compliance",
					},
					"configSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"severityThreshold": map[string]interface{}{
								"type":        "string",
								"enum":        []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "INFO"},
								"default":     "HIGH",
								"description": "Minimum severity level that will cause a scan to fail",
							},
							"scanTypes": map[string]interface{}{
								"type":        "array",
								"items":       map[string]interface{}{"type": "string"},
								"default":     []string{"secret", "vulnerability", "code", "license"},
								"description": "Types of security scans to perform",
							},
						},
					},
				}
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
				return
			}

			c.JSON(http.StatusOK, plugin)
		})

		// Install a plugin
		pluginGroup.POST("/:name/install", func(c *gin.Context) {
			name := c.Param("name")

			// In a real implementation, this would install a plugin
			// For now, we'll return a mock response
			plugin := map[string]interface{}{
				"name":        name,
				"version":     "1.0.0",
				"description": "Plugin description",
				"author":      "Conveyor Team",
				"installed":   true,
				"enabled":     true,
				"installedAt": time.Now().Format(time.RFC3339),
			}

			c.JSON(http.StatusOK, plugin)
		})

		// Uninstall a plugin
		pluginGroup.POST("/:name/uninstall", func(c *gin.Context) {
			name := c.Param("name")

			// In a real implementation, this would uninstall a plugin
			// For now, we'll return a mock response
			c.JSON(http.StatusOK, gin.H{"name": name, "status": "uninstalled"})
		})

		// Enable a plugin
		pluginGroup.POST("/:name/enable", func(c *gin.Context) {
			name := c.Param("name")

			// In a real implementation, this would enable a plugin
			// For now, we'll return a mock response
			c.JSON(http.StatusOK, gin.H{"name": name, "enabled": true})
		})

		// Disable a plugin
		pluginGroup.POST("/:name/disable", func(c *gin.Context) {
			name := c.Param("name")

			// In a real implementation, this would disable a plugin
			// For now, we'll return a mock response
			c.JSON(http.StatusOK, gin.H{"name": name, "enabled": false})
		})

		// Get plugin settings
		pluginGroup.GET("/:name/settings", func(c *gin.Context) {
			name := c.Param("name")

			// In a real implementation, this would fetch plugin settings
			// For now, we'll return mock data
			settings := map[string]interface{}{
				"name": name,
				"config": map[string]interface{}{
					"severityThreshold": "HIGH",
					"ignorePatterns":    []string{"node_modules/", "vendor/", ".git/"},
					"scanTypes":         []string{"secret", "vulnerability", "code", "license"},
					"failOnViolation":   true,
					"generateSBOM":      true,
				},
			}

			c.JSON(http.StatusOK, settings)
		})

		// Update plugin settings
		pluginGroup.PUT("/:name/settings", func(c *gin.Context) {
			name := c.Param("name")

			var settingsRequest map[string]interface{}
			if err := c.ShouldBindJSON(&settingsRequest); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
				return
			}

			// In a real implementation, this would update plugin settings
			// For now, we'll return a mock response
			settings := map[string]interface{}{
				"name":    name,
				"config":  settingsRequest,
				"updated": true,
			}

			c.JSON(http.StatusOK, settings)
		})
	}
} 