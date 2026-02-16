package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	DefaultPort    = "5555"
	AgentVersion   = "1.0.0"
	ShutdownTimeout = 5 * time.Second
)

func main() {
	// Parse command line flags
	port := flag.String("port", DefaultPort, "Port to run the agent on")
	install := flag.Bool("install", false, "Install agent to start on login")
	uninstall := flag.Bool("uninstall", false, "Remove agent from auto-start")
	flag.Parse()

	// Handle installation flags
	if *install {
		if err := installAgent(); err != nil {
			log.Fatalf("❌ Failed to install agent: %v", err)
		}
		return
	}

	if *uninstall {
		if err := uninstallAgent(); err != nil {
			log.Fatalf("❌ Failed to uninstall agent: %v", err)
		}
		return
	}

	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	// Create router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware())

	// Register routes
	router.GET("/health", healthHandler)
	router.POST("/execute", executeHandler)

	// Start server
	addr := fmt.Sprintf("127.0.0.1:%s", *port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Blink Agent v%s starting on http://%s", AgentVersion, addr)
		log.Printf("✓ Ready to accept requests from web application")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down agent...")
	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Agent shutdown error: %v", err)
	}
	log.Println("Agent stopped")
}

func corsMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all origins for now, can be restricted later
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
	return cors.New(config)
}

func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		duration := time.Since(start)
		log.Printf("%s %s - %d (%v)", c.Request.Method, path, c.Writer.Status(), duration)
	}
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": AgentVersion,
		"message": "Agent is running",
	})
}
