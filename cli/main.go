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
	
	// Start a goroutine to simulate job progress
	go simulateJobProgress(engine)

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
	
	// Create a second pipeline (secure-build) based on a more complex example
	secureBuildPipeline := &core.Pipeline{
		ID:          "secure-build",
		Name:        "Secure Build Pipeline",
		Description: "A comprehensive secure CI/CD pipeline with security scanning",
		Stages: []core.Stage{
			{
				ID:   "pre-build",
				Name: "Pre-Build",
				Steps: []core.Step{
					{
						ID:      "setup",
						Name:    "Setup Environment",
						Type:    "script",
						Command: "echo 'Setting up environment'",
					},
					{
						ID:      "dependencies",
						Name:    "Install Dependencies",
						Type:    "script",
						Command: "echo 'Installing dependencies'",
					},
				},
			},
			{
				ID:   "security-checks",
				Name: "Security Checks",
				Steps: []core.Step{
					{
						ID:     "secret-scan",
						Name:   "Secret Scanner",
						Type:   "secret-scan",
						Plugin: "security",
						Config: map[string]interface{}{
							"scanTypes":        []string{"secret"},
							"severityThreshold": "HIGH",
							"failOnViolation":  true,
						},
					},
					{
						ID:     "code-scan",
						Name:   "Code Security Scanner",
						Type:   "vulnerability-scan",
						Plugin: "security",
						Config: map[string]interface{}{
							"scanTypes":        []string{"code"},
							"severityThreshold": "MEDIUM",
						},
					},
				},
			},
			{
				ID:   "build",
				Name: "Build",
				Steps: []core.Step{
					{
						ID:      "build-app",
						Name:    "Build Application",
						Type:    "script",
						Command: "echo 'Building application'",
					},
				},
			},
			{
				ID:   "test",
				Name: "Test",
				Steps: []core.Step{
					{
						ID:      "unit-tests",
						Name:    "Run Unit Tests",
						Type:    "script",
						Command: "echo 'Running unit tests'",
					},
					{
						ID:      "integration-tests",
						Name:    "Run Integration Tests",
						Type:    "script",
						Command: "echo 'Running integration tests'",
					},
				},
			},
			{
				ID:   "deploy",
				Name: "Deploy",
				Steps: []core.Step{
					{
						ID:      "deploy-app",
						Name:    "Deploy Application",
						Type:    "script",
						Command: "echo 'Deploying application'",
					},
				},
			},
		},
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}

	if err := engine.CreatePipeline(secureBuildPipeline); err != nil {
		log.Printf("Error creating secure build pipeline: %v", err)
	}
	
	// Create sample jobs for the pipelines
	createSampleJobs(engine)
}

// createSampleJobs creates sample jobs for the pipelines
func createSampleJobs(engine *core.PipelineEngine) {
	// Create a completed job for pipeline-1
	completedJob := &core.Job{
		ID:         "job-1",
		PipelineID: "pipeline-1",
		Status:     "success",
		StartedAt:  time.Now().Add(-2 * time.Hour),
		EndedAt:    time.Now().Add(-1 * time.Hour),
		Steps: []core.StepStatus{
			{
				ID:        "step-1",
				Name:      "Build Application",
				Status:    "success",
				StartedAt: time.Now().Add(-2 * time.Hour),
				EndedAt:   time.Now().Add(-1*time.Hour - 30*time.Minute),
				ExitCode:  0,
				Output:    "Build successful!",
			},
			{
				ID:        "step-2",
				Name:      "Run Tests",
				Status:    "success",
				StartedAt: time.Now().Add(-1*time.Hour - 30*time.Minute),
				EndedAt:   time.Now().Add(-1*time.Hour - 15*time.Minute),
				ExitCode:  0,
				Output:    "All tests passed!",
			},
			{
				ID:        "step-3",
				Name:      "Vulnerability Scan",
				Status:    "success",
				StartedAt: time.Now().Add(-1*time.Hour - 15*time.Minute),
				EndedAt:   time.Now().Add(-1*time.Hour - 10*time.Minute),
				ExitCode:  0,
				Output:    "No vulnerabilities found!",
			},
			{
				ID:        "step-4",
				Name:      "Secret Scan",
				Status:    "success",
				StartedAt: time.Now().Add(-1*time.Hour - 10*time.Minute),
				EndedAt:   time.Now().Add(-1 * time.Hour),
				ExitCode:  0,
				Output:    "No secrets found!",
			},
		},
		Logs: []core.LogEntry{
			{
				Timestamp: time.Now().Add(-2 * time.Hour),
				Level:     "info",
				Message:   "Job started",
			},
			{
				Timestamp: time.Now().Add(-1*time.Hour - 30*time.Minute),
				Level:     "info",
				Message:   "Build step completed successfully",
				StepID:    "step-1",
			},
			{
				Timestamp: time.Now().Add(-1*time.Hour - 15*time.Minute),
				Level:     "info",
				Message:   "Test step completed successfully",
				StepID:    "step-2",
			},
			{
				Timestamp: time.Now().Add(-1 * time.Hour),
				Level:     "info",
				Message:   "Job completed successfully",
			},
		},
	}
	engine.AddJob(completedJob)
	
	// Create jobs for secure-build pipeline
	// A failed job
	failedJob := &core.Job{
		ID:         "job-2",
		PipelineID: "secure-build",
		Status:     "failed",
		StartedAt:  time.Now().Add(-12 * time.Hour),
		EndedAt:    time.Now().Add(-11 * time.Hour),
		Steps: []core.StepStatus{
			{
				ID:        "setup",
				Name:      "Setup Environment",
				Status:    "success",
				StartedAt: time.Now().Add(-12 * time.Hour),
				EndedAt:   time.Now().Add(-12*time.Hour + 10*time.Minute),
				ExitCode:  0,
				Output:    "Environment setup complete",
			},
			{
				ID:        "dependencies",
				Name:      "Install Dependencies",
				Status:    "success",
				StartedAt: time.Now().Add(-12*time.Hour + 10*time.Minute),
				EndedAt:   time.Now().Add(-12*time.Hour + 20*time.Minute),
				ExitCode:  0,
				Output:    "Dependencies installed",
			},
			{
				ID:        "secret-scan",
				Name:      "Secret Scanner",
				Status:    "failed",
				StartedAt: time.Now().Add(-12*time.Hour + 20*time.Minute),
				EndedAt:   time.Now().Add(-12*time.Hour + 30*time.Minute),
				ExitCode:  1,
				Output:    "Found 2 potential secrets in the codebase!",
			},
		},
		Logs: []core.LogEntry{
			{
				Timestamp: time.Now().Add(-12 * time.Hour),
				Level:     "info",
				Message:   "Job started",
			},
			{
				Timestamp: time.Now().Add(-12*time.Hour + 20*time.Minute),
				Level:     "info",
				Message:   "Starting security checks",
				StepID:    "secret-scan",
			},
			{
				Timestamp: time.Now().Add(-12*time.Hour + 25*time.Minute),
				Level:     "error",
				Message:   "Found AWS access key in config files",
				StepID:    "secret-scan",
			},
			{
				Timestamp: time.Now().Add(-12*time.Hour + 30*time.Minute),
				Level:     "error",
				Message:   "Secret scan failed: potential secrets found",
				StepID:    "secret-scan",
			},
		},
	}
	engine.AddJob(failedJob)
	
	// A successful job
	successJob := &core.Job{
		ID:         "job-3",
		PipelineID: "secure-build",
		Status:     "success",
		StartedAt:  time.Now().Add(-6 * time.Hour),
		EndedAt:    time.Now().Add(-5 * time.Hour),
		Steps: []core.StepStatus{
			{
				ID:        "setup",
				Name:      "Setup Environment",
				Status:    "success",
				StartedAt: time.Now().Add(-6 * time.Hour),
				EndedAt:   time.Now().Add(-6*time.Hour + 5*time.Minute),
				ExitCode:  0,
				Output:    "Environment setup complete",
			},
			{
				ID:        "dependencies",
				Name:      "Install Dependencies",
				Status:    "success",
				StartedAt: time.Now().Add(-6*time.Hour + 5*time.Minute),
				EndedAt:   time.Now().Add(-6*time.Hour + 15*time.Minute),
				ExitCode:  0,
				Output:    "Dependencies installed",
			},
			{
				ID:        "secret-scan",
				Name:      "Secret Scanner",
				Status:    "success",
				StartedAt: time.Now().Add(-6*time.Hour + 15*time.Minute),
				EndedAt:   time.Now().Add(-6*time.Hour + 25*time.Minute),
				ExitCode:  0,
				Output:    "No secrets found",
			},
			{
				ID:        "code-scan",
				Name:      "Code Security Scanner",
				Status:    "success",
				StartedAt: time.Now().Add(-6*time.Hour + 25*time.Minute),
				EndedAt:   time.Now().Add(-6*time.Hour + 35*time.Minute),
				ExitCode:  0,
				Output:    "No vulnerabilities found",
			},
			{
				ID:        "build-app",
				Name:      "Build Application",
				Status:    "success",
				StartedAt: time.Now().Add(-6*time.Hour + 35*time.Minute),
				EndedAt:   time.Now().Add(-6*time.Hour + 45*time.Minute),
				ExitCode:  0,
				Output:    "Build successful",
			},
		},
		Logs: []core.LogEntry{
			{
				Timestamp: time.Now().Add(-6 * time.Hour),
				Level:     "info",
				Message:   "Job started",
			},
			{
				Timestamp: time.Now().Add(-6*time.Hour + 15*time.Minute),
				Level:     "info",
				Message:   "Starting security checks",
			},
			{
				Timestamp: time.Now().Add(-6*time.Hour + 35*time.Minute),
				Level:     "info",
				Message:   "Security checks passed",
			},
			{
				Timestamp: time.Now().Add(-5 * time.Hour),
				Level:     "info",
				Message:   "Job completed successfully",
			},
		},
	}
	engine.AddJob(successJob)
	
	// A running job
	runningJob := &core.Job{
		ID:         "job-4",
		PipelineID: "secure-build",
		Status:     "running",
		StartedAt:  time.Now().Add(-15 * time.Minute),
		Steps: []core.StepStatus{
			{
				ID:        "setup",
				Name:      "Setup Environment",
				Status:    "success",
				StartedAt: time.Now().Add(-15 * time.Minute),
				EndedAt:   time.Now().Add(-12 * time.Minute),
				ExitCode:  0,
				Output:    "Environment setup complete",
			},
			{
				ID:        "dependencies",
				Name:      "Install Dependencies",
				Status:    "success",
				StartedAt: time.Now().Add(-12 * time.Minute),
				EndedAt:   time.Now().Add(-8 * time.Minute),
				ExitCode:  0,
				Output:    "Dependencies installed",
			},
			{
				ID:        "secret-scan",
				Name:      "Secret Scanner",
				Status:    "success",
				StartedAt: time.Now().Add(-8 * time.Minute),
				EndedAt:   time.Now().Add(-5 * time.Minute),
				ExitCode:  0,
				Output:    "No secrets found",
			},
			{
				ID:        "code-scan",
				Name:      "Code Security Scanner",
				Status:    "running",
				StartedAt: time.Now().Add(-5 * time.Minute),
				Output:    "Scanning code for vulnerabilities...",
			},
		},
		Logs: []core.LogEntry{
			{
				Timestamp: time.Now().Add(-15 * time.Minute),
				Level:     "info",
				Message:   "Job started",
			},
			{
				Timestamp: time.Now().Add(-12 * time.Minute),
				Level:     "info",
				Message:   "Environment setup complete",
				StepID:    "setup",
			},
			{
				Timestamp: time.Now().Add(-8 * time.Minute),
				Level:     "info",
				Message:   "Dependencies installed successfully",
				StepID:    "dependencies",
			},
			{
				Timestamp: time.Now().Add(-7 * time.Minute),
				Level:     "info", 
				Message:   "Starting secret scan",
				StepID:    "secret-scan",
			},
			{
				Timestamp: time.Now().Add(-5 * time.Minute),
				Level:     "info",
				Message:   "Secret scan complete, no secrets found",
				StepID:    "secret-scan",
			},
			{
				Timestamp: time.Now().Add(-5 * time.Minute),
				Level:     "info",
				Message:   "Starting code security scan",
				StepID:    "code-scan",
			},
			{
				Timestamp: time.Now().Add(-2 * time.Minute),
				Level:     "info",
				Message:   "Scanning dependencies for vulnerabilities...",
				StepID:    "code-scan",
			},
		},
	}
	engine.AddJob(runningJob)
}

// simulateJobProgress updates the running job to simulate progress
func simulateJobProgress(engine *core.PipelineEngine) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	codeScanComplete := false
	buildStepStarted := false
	buildStepComplete := false
	
	for range ticker.C {
		// Get the running job
		job, err := engine.GetJob("secure-build", "job-4")
		if err != nil {
			log.Printf("Error getting job: %v", err)
			continue
		}
		
		// If job is no longer running, stop simulating
		if job.Status != "running" {
			log.Println("Job is no longer running, stopping simulation")
			return
		}
		
		// Stage 1: Complete code scan
		if !codeScanComplete {
			// Create a modified copy of the job
			updatedJob := *job
			
			// Complete the code scan step
			for i, step := range updatedJob.Steps {
				if step.ID == "code-scan" {
					updatedJob.Steps[i].Status = "success"
					updatedJob.Steps[i].EndedAt = time.Now()
					updatedJob.Steps[i].ExitCode = 0
					updatedJob.Steps[i].Output = "No vulnerabilities found"
					
					// Add log entry
					updatedJob.Logs = append(updatedJob.Logs, core.LogEntry{
						Timestamp: time.Now(),
						Level:     "info",
						Message:   "Code scan completed successfully",
						StepID:    "code-scan",
					})
					
					codeScanComplete = true
					break
				}
			}
			
			// Update the job in the engine
			err = engine.UpdateJob(&updatedJob)
			if err != nil {
				log.Printf("Error updating job: %v", err)
				continue
			}
			
			// Emit step completed event through the engine
			engine.EmitStepCompletedEvent("secure-build", "job-4", "code-scan", "success")
			continue
		}
		
		// Stage 2: Start build step
		if codeScanComplete && !buildStepStarted {
			// Create a modified copy of the job
			updatedJob := *job
			
			// Add build step
			updatedJob.Steps = append(updatedJob.Steps, core.StepStatus{
				ID:        "build-app",
				Name:      "Build Application",
				Status:    "running",
				StartedAt: time.Now(),
				Output:    "Building application...",
			})
			
			// Add log entry
			updatedJob.Logs = append(updatedJob.Logs, core.LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "Starting build step",
				StepID:    "build-app",
			})
			
			buildStepStarted = true
			
			// Update the job in the engine
			err = engine.UpdateJob(&updatedJob)
			if err != nil {
				log.Printf("Error updating job: %v", err)
				continue
			}
			
			// Emit step started event through the engine
			engine.EmitStepStartedEvent("secure-build", "job-4", "build-app")
			continue
		}
		
		// Stage 3: Complete build step and job
		if buildStepStarted && !buildStepComplete {
			// Create a modified copy of the job
			updatedJob := *job
			
			// Complete the build step
			for i, step := range updatedJob.Steps {
				if step.ID == "build-app" {
					updatedJob.Steps[i].Status = "success"
					updatedJob.Steps[i].EndedAt = time.Now()
					updatedJob.Steps[i].ExitCode = 0
					updatedJob.Steps[i].Output = "Build completed successfully"
					
					// Add log entry
					updatedJob.Logs = append(updatedJob.Logs, core.LogEntry{
						Timestamp: time.Now(),
						Level:     "info",
						Message:   "Build completed successfully",
						StepID:    "build-app",
					})
					
					buildStepComplete = true
					break
				}
			}
			
			// Complete the job
			updatedJob.Status = "success"
			updatedJob.EndedAt = time.Now()
			
			// Add final log entry
			updatedJob.Logs = append(updatedJob.Logs, core.LogEntry{
				Timestamp: time.Now(),
				Level:     "info",
				Message:   "Job completed successfully",
			})
			
			// Update the job in the engine
			err = engine.UpdateJob(&updatedJob)
			if err != nil {
				log.Printf("Error updating job: %v", err)
				continue
			}
			
			// Emit step completed event through the engine
			engine.EmitStepCompletedEvent("secure-build", "job-4", "build-app", "success")
			
			// Emit job completed event through the engine
			engine.EmitJobCompletedEvent("secure-build", "job-4", "success")
			return // Simulation complete
		}
	}
} 