package main

import (
	"fmt"
	"log"
	"os"
	"project-x/config"
	"project-x/handlers"
	"project-x/models"
	"project-x/routes"
	"project-x/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize Gin router
	r := gin.Default()

	// Setup database connection
	db, err := setupDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate database tables
	if err := db.AutoMigrate(
		&models.User{},
		&models.Task{},
		&models.CollaborativeTask{},
		&models.CollaborativeTaskParticipant{},
		&models.Project{},
		&models.UserProject{},
		&models.Notification{},
		&models.UserNotificationPreference{},
		// Simple chat models
		&models.ChatRoom{},
		&models.ChatParticipant{},
		&models.ChatMessage{},
		// HR problem reporting models
		&models.HRProblem{},
		&models.HRProblemComment{},
		&models.HRProblemUpdate{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("✅ Database tables migrated successfully")

	// Initialize routes
	setupRoutes(r, db)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server starting on port %s", port)
	r.Run(":" + port)
}

func setupDatabase() (*gorm.DB, error) {
	config, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Updated DSN with proper UTF-8 encoding for Arabic language support
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC client_encoding=UTF8",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Set additional connection parameters for proper UTF-8 support
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Execute commands to ensure proper UTF-8 handling
	if _, err := sqlDB.Exec("SET client_encoding TO 'UTF8'"); err != nil {
		log.Printf("Warning: Could not set client encoding: %v", err)
	}

	if _, err := sqlDB.Exec("SET lc_collate TO 'en_US.UTF-8'"); err != nil {
		log.Printf("Warning: Could not set collation: %v", err)
	}

	return db, nil
}

func setupRoutes(r *gin.Engine, db *gorm.DB) {
	// Health check endpoint for Railway
	r.GET("/health", func(c *gin.Context) {
		// Check database connection
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": "Database connection failed"})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(500, gin.H{"status": "error", "message": "Database ping failed"})
			return
		}

		c.JSON(200, gin.H{"status": "healthy", "message": "Service is running"})
	})

	// Initialize WebSocket service
	wsService := services.NewWebSocketService(db)

	// Initialize Notification service
	notificationService := services.NewNotificationService(db, wsService)

	// Initialize Notification handler
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	routes.SetupAuthRoutes(r, db)
	routes.SetupUserRoutes(r, db)
	routes.SetupTaskRoutes(r, db)
	routes.SetupProjectRoutes(r, db)
	routes.SetupCollaborativeTaskRoutes(r, db)
	routes.SetupWebSocketRoutes(r, db, wsService)
	routes.SetupNotificationRoutes(r, notificationHandler, db)
	routes.SetupChatRoutes(r, db, wsService, notificationService)
	routes.SetupHRProblemRoutes(r, db, notificationService)
}
