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
	"github.com/chip/conveyor/core/executor"
	"github.com/chip/conveyor/core/loader"
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

	// Wire up the step executor and pipeline orchestrator
	shellExecutor := &executor.ShellExecutor{}
	orchestrator := executor.NewPipelineOrchestrator(engine, shellExecutor)
	engine.SetJobRunner(orchestrator)

	// Load pipelines from YAML directory
	pipelineLoader := loader.NewPipelineLoader(engine, "pipelines")
	result, err := pipelineLoader.LoadDirectory()
	if err != nil {
		log.Fatalf("Failed to scan pipeline directory: %v", err)
	}
	for file, warnings := range result.Warnings {
		for _, w := range warnings {
			log.Printf("WARN [%s]: %s", file, w)
		}
	}
	for file, loadErr := range result.Errors {
		log.Printf("ERROR [%s]: %s", file, loadErr)
	}
	log.Printf("Loaded %d pipelines from YAML", len(result.Loaded))

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
	api.SetupRoutes(router, engine, pipelineLoader)

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
