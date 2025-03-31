package api

import (
	"context"
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
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
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
				return true // Allow all origins for WebSocket connections
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
	// Health check
	s.router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Web UI
	s.router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/ui")
	})

	// API routes
	routes.RegisterPipelineRoutes(s.router, s.pipelineEngine)
	routes.RegisterJobRoutes(s.router, s.pipelineEngine)
	routes.RegisterPluginRoutes(s.router)
	routes.RegisterSecurityRoutes(s.router, s.pipelineEngine)
	
	// WebSocket route for real-time updates
	s.router.GET("/api/ws", s.handleWebSocket)

	// Static files for UI
	s.router.Static("/ui", "./ui/dist")
}

// handleWebSocket handles WebSocket connections for real-time updates
func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection to WebSocket"})
		return
	}
	defer conn.Close()

	// Create a unique client ID
	clientID := c.Query("clientId")
	if clientID == "" {
		clientID = "client-" + time.Now().Format("20060102-150405.000")
	}

	// Register client for event notifications
	events := make(chan core.Event)
	s.pipelineEngine.RegisterEventListener(clientID, events)
	defer s.pipelineEngine.UnregisterEventListener(clientID)

	// Create a context that's cancelled when the connection is closed
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Read messages from client in a separate goroutine
	go func() {
		defer cancel()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return // Connection closed
			}
		}
	}()

	// Send events to client
	for {
		select {
		case event := <-events:
			err := conn.WriteJSON(event)
			if err != nil {
				return // Connection closed
			}
		case <-ctx.Done():
			return // Context cancelled
		}
	}
} 