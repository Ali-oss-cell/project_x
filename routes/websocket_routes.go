package routes

import (
	"project-x/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupWebSocketRoutes(r *gin.Engine, db *gorm.DB, wsService *services.WebSocketService) {
	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		wsService.HandleWebSocket(c)
	})

	// WebSocket status endpoint
	r.GET("/api/ws/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"online_users": wsService.GetConnectionCount(),
			"status":       "running",
		})
	})
}
