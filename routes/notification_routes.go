package routes

import (
	"project-x/handlers"
	"project-x/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupNotificationRoutes(router *gin.Engine, notificationHandler *handlers.NotificationHandler, db *gorm.DB) {
	// Notification routes group
	notificationRoutes := router.Group("/api/notifications")
	{
		// Apply auth middleware to all notification routes
		notificationRoutes.Use(middleware.AuthMiddleware(db))

		// Get user's notifications
		notificationRoutes.GET("", notificationHandler.GetUserNotifications)

		// Get unread count
		notificationRoutes.GET("/unread-count", notificationHandler.GetUnreadCount)

		// Get notification preferences
		notificationRoutes.GET("/preferences", notificationHandler.GetNotificationPreferences)

		// Update notification preferences
		notificationRoutes.PUT("/preferences", notificationHandler.UpdateNotificationPreferences)

		// Mark notification as read
		notificationRoutes.PUT("/:id/read", notificationHandler.MarkNotificationAsRead)

		// Mark all notifications as read
		notificationRoutes.PUT("/read-all", notificationHandler.MarkAllNotificationsAsRead)

		// Delete a notification
		notificationRoutes.DELETE("/:id", notificationHandler.DeleteNotification)

		// Get notification rules (Admin only)
		notificationRoutes.GET("/rules", notificationHandler.GetNotificationRules)

		// Update notification rules (Admin only)
		notificationRoutes.PUT("/rules", notificationHandler.UpdateNotificationRules)
	}
}
