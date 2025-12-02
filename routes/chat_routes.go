package routes

import (
	"project-x/handlers"
	"project-x/middleware"
	"project-x/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupChatRoutes(r *gin.Engine, db *gorm.DB, wsService *services.WebSocketService, notificationService *services.NotificationService) {
	// Create task service for AI chat
	taskService := services.NewTaskService(db)
	taskService.SetNotificationService(notificationService)

	// Create AI chat service
	aiChatService := services.NewAIChatService(db, taskService)

	// Create chat service and handler
	chatService := services.NewChatService(db, wsService, notificationService)
	chatService.SetAIChatService(aiChatService) // Connect AI chat service
	chatHandler := handlers.NewChatHandler(db, chatService)

	// Create AI chat handler
	aiChatHandler := handlers.NewAIChatHandler(db, aiChatService)

	// WebSocket endpoint for real-time chat (requires authentication)
	r.GET("/ws/chat", middleware.AuthMiddleware(db), func(c *gin.Context) {
		wsService.HandleWebSocket(c)
	})

	// Simple Team Chat API routes
	chatAPI := r.Group("/api/chat")
	chatAPI.Use(middleware.AuthMiddleware(db)) // All chat endpoints require authentication

	// Simple team chat endpoints
	{
		// Get or create the main team chat room (auto-joins user)
		chatAPI.GET("/team-chat", chatHandler.GetOrCreateTeamChat)

		// Send a message to team chat
		chatAPI.POST("/rooms/:roomId/messages", chatHandler.SendMessage)

		// Get messages from team chat (with pagination)
		chatAPI.GET("/rooms/:roomId/messages", chatHandler.GetMessages)

		// Get team chat members
		chatAPI.GET("/rooms/:roomId/members", chatHandler.GetRoomMembers)
	}

	// AI Chat routes
	{
		// Get or create private AI chat room
		chatAPI.GET("/ai/room", aiChatHandler.GetOrCreateAIChatRoom)

		// Process AI message (can be used for direct API calls)
		chatAPI.POST("/ai/rooms/:roomId/message", aiChatHandler.ProcessAIMessage)
	}

	// WebSocket status endpoint
	{
		// Get WebSocket connection status
		chatAPI.GET("/ws/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"online_users": wsService.GetConnectionCount(),
				"status":       "running",
				"users":        wsService.GetOnlineUsers(),
			})
		})
	}
}
