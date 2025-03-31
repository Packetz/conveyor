package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/chip/conveyor/api/routes"
	"github.com/chip/conveyor/core"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Server represents the API server
type Server struct {
	router         *gin.Engine
	httpServer     *http.Server
	pipelineEngine *core.PipelineEngine
	upgrader       websocket.Upgrader
}

// NewServer creates a new API server
func NewServer(pipelineEngine *core.PipelineEngine) *Server {
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

	server := &Server{
		router:         router,
		pipelineEngine: pipelineEngine,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

	// Register routes
	server.registerRoutes()

	return server
}

// Start starts the API server
func (s *Server) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	log.Printf("Starting API server on %s", addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the API server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// registerRoutes registers all API routes
func (s *Server) registerRoutes() {
	// Health check route
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Web UI
	s.router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/ui")
	})

	// API routes
	api := s.router.Group("/api")
	
	// API health endpoint
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
	
	// Pipeline routes
	pipelineRoutes := api.Group("/pipelines")
	routes.RegisterPipelineRoutes(pipelineRoutes, s.pipelineEngine)
	
	// Job routes
	jobRoutes := api.Group("/jobs")
	routes.RegisterJobRoutes(jobRoutes, s.pipelineEngine)
	
	// Plugin routes
	pluginRoutes := api.Group("/plugins")
	routes.RegisterPluginRoutes(pluginRoutes)
	
	// Security routes
	securityRoutes := api.Group("/security")
	routes.RegisterSecurityRoutes(securityRoutes, s.pipelineEngine)
	
	// WebSocket route for real-time updates
	s.router.GET("/ws", s.handleWebSocket)

	// Static files for UI
	s.router.Static("/ui", "./ui/dist")
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(c *gin.Context) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	// Create a channel for events
	eventCh := make(chan core.Event, 100)

	// Register the event listener
	s.pipelineEngine.RegisterEventListener(c.ClientIP(), eventCh)
	defer s.pipelineEngine.UnregisterEventListener(c.ClientIP())

	// Write events to the WebSocket
	go func() {
		for event := range eventCh {
			err := conn.WriteJSON(event)
			if err != nil {
				log.Printf("Error writing to WebSocket: %v", err)
				return
			}
		}
	}()

	// Read messages from the WebSocket (just ping-pong for now)
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading from WebSocket: %v", err)
			return
		}

		// Echo the message back for now
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Printf("Error writing to WebSocket: %v", err)
			return
		}
	}
} 