package routes

import (
	"project-x/handlers"
	"project-x/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupAdminRoutes sets up all admin-only routes
func SetupAdminRoutes(r *gin.Engine, db *gorm.DB) {
	adminHandler := handlers.NewAdminHandler(db)

	// Admin routes group - all require Admin role
	adminGroup := r.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(db))
	adminGroup.Use(middleware.RequireAdmin())
	{
		// Main dashboard endpoint
		adminGroup.GET("/dashboard", adminHandler.GetAdminDashboard)

		// Items requiring attention
		adminGroup.GET("/attention-items", adminHandler.GetAttentionItems)

		// Project health summary
		adminGroup.GET("/project-health", adminHandler.GetProjectHealthSummary)

		// Inactive users
		adminGroup.GET("/inactive-users", adminHandler.GetInactiveUsers)

		// Recent activity
		adminGroup.GET("/recent-activity", adminHandler.GetRecentActivity)

		// System health
		adminGroup.GET("/system-health", adminHandler.GetSystemHealth)

		// Checklist endpoints
		adminGroup.GET("/checklist/status", adminHandler.GetChecklistStatus)
		adminGroup.POST("/checklist/update", adminHandler.UpdateChecklistItem)
		adminGroup.GET("/checklist/history", adminHandler.GetChecklistHistory)
		adminGroup.GET("/checklist/stats", adminHandler.GetChecklistStats)
	}
}
