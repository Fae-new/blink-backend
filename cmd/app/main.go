package main

import (
	"log"

	"postman-runner/internal/config"
	"postman-runner/internal/db"
	"postman-runner/internal/handlers"
	"postman-runner/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	database, err := db.NewConnection(cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("✅ Connected to database successfully")

	// Initialize Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery()) // Panic recovery
	// CORS middleware - allow localhost:5173 for development
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}))
	router.Use(middleware.Logger()) // Request logging

	// Initialize rate limiter
	limiter := middleware.NewIPRateLimiter(rate.Limit(cfg.RateLimitRPS), cfg.RateLimitBurst)

	// Initialize handlers
	collectionHandler := handlers.NewCollectionHandler(database, cfg)
	executionHandler := handlers.NewExecutionHandler(database, cfg)
	itemHandler := handlers.NewItemHandler(database, cfg)
	environmentHandler := handlers.NewEnvironmentHandler(database, cfg)

	// Health check endpoint (no rate limit)
	router.GET("/health", handlers.HealthCheck)

	// Serve agent downloads
	router.Static("/downloads", "./downloads")

	// API routes
	api := router.Group("/api/v1")
	{
		// Collections
		api.POST("/collections/upload", collectionHandler.UploadCollection)
		api.GET("/collections", collectionHandler.ListCollections)
		api.GET("/collections/:id/tree", collectionHandler.GetCollectionTree)

		// Items
		api.POST("/collections/:id/items", itemHandler.CreateItem)
		api.GET("/items/:id", collectionHandler.GetItem)
		api.PUT("/items/:id", itemHandler.UpdateItem)
		api.DELETE("/items/:id", itemHandler.DeleteItem)

		// Execution (with rate limiting)
		api.POST("/items/:id/execute", middleware.RateLimitMiddleware(limiter), executionHandler.ExecuteRequest)

		// Environments
		api.POST("/environments", environmentHandler.CreateEnvironment)
		api.GET("/environments", environmentHandler.ListEnvironments)
		api.GET("/environments/:id", environmentHandler.GetEnvironment)
		api.PUT("/environments/:id", environmentHandler.UpdateEnvironment)
		api.PATCH("/environments/:id/variables", environmentHandler.BatchUpdateEnvironmentVariables)
		api.DELETE("/environments/:id", environmentHandler.DeleteEnvironment)
	}

	// Start server
	address := ":" + cfg.Port
	log.Printf("🚀 Server starting on %s", address)
	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
