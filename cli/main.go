package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chip/conveyor/api"
	"github.com/chip/conveyor/core"
	"github.com/chip/conveyor/plugins/security"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set up the pipeline engine
	engine := core.NewPipelineEngine()

	// Register plugins
	securityPlugin := security.NewSecurityPlugin()
	engine.RegisterPlugin(securityPlugin)

	// Create the router
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Register API routes
	api.SetupRoutes(router, engine)

	// Create some sample data for testing
	createSampleData(engine)

	// Start watching for plugins and samples
	log.Println("Watching plugins and samples...")

	// Start the server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Run the server in a goroutine
	go func() {
		log.Println("Server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

// createSampleData creates some sample data for testing
func createSampleData(engine *core.PipelineEngine) {
	// Create a sample pipeline
	pipeline := &core.Pipeline{
		ID:          "pipeline-1",
		Name:        "Sample Pipeline",
		Description: "A sample pipeline for testing",
		Stages: []core.Stage{
			{
				ID:   "stage-1",
				Name: "Build",
				Steps: []core.Step{
					{
						ID:      "step-1",
						Name:    "Build Application",
						Type:    "build",
						Command: "npm run build",
					},
				},
			},
			{
				ID:   "stage-2",
				Name: "Test",
				Steps: []core.Step{
					{
						ID:      "step-2",
						Name:    "Run Tests",
						Type:    "test",
						Command: "npm run test",
					},
				},
			},
			{
				ID:   "stage-3",
				Name: "Security Scan",
				Steps: []core.Step{
					{
						ID:     "step-3",
						Name:   "Vulnerability Scan",
						Type:   "vulnerability-scan",
						Plugin: "security",
						Config: map[string]interface{}{
							"pipelineId": "pipeline-1",
							"jobId":      "job-1",
						},
					},
					{
						ID:     "step-4",
						Name:   "Secret Scan",
						Type:   "secret-scan",
						Plugin: "security",
						Config: map[string]interface{}{
							"pipelineId": "pipeline-1",
							"jobId":      "job-1",
						},
					},
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := engine.CreatePipeline(pipeline); err != nil {
		log.Printf("Error creating sample pipeline: %v", err)
	}
} 