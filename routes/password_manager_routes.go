package routes

import (
	"project-x/handlers"
	"project-x/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupPasswordManagerRoutes sets up all password manager routes
func SetupPasswordManagerRoutes(r *gin.Engine, db *gorm.DB) {
	passwordManagerHandler, err := handlers.NewPasswordManagerHandler(db)
	if err != nil {
		// Log error but don't fail - encryption service will fail on first use if key is missing
		// This allows the app to start but password manager endpoints will fail gracefully
		return
	}

	// Admin routes - all require Admin role
	adminGroup := r.Group("/api/password-manager")
	adminGroup.Use(middleware.AuthMiddleware(db))
	adminGroup.Use(middleware.RequireAdmin())
	{
		// Credential management
		adminGroup.POST("/credentials", passwordManagerHandler.CreateCredential)
		adminGroup.GET("/credentials", passwordManagerHandler.GetAllCredentials)
		adminGroup.GET("/credentials/:id", passwordManagerHandler.GetCredentialDetails)
		adminGroup.PUT("/credentials/:id", passwordManagerHandler.UpdateCredential)
		adminGroup.DELETE("/credentials/:id", passwordManagerHandler.DeleteCredential)

		// Sharing management
		adminGroup.POST("/credentials/:id/share", passwordManagerHandler.ShareCredential)
		adminGroup.DELETE("/credentials/:id/share/:userId", passwordManagerHandler.UnshareCredential)

		// Access logs (Admin only)
		adminGroup.GET("/credentials/:id/logs", passwordManagerHandler.GetAccessLogs)
	}

	// User routes - accessible to all authenticated users
	userGroup := r.Group("/api/password-manager")
	userGroup.Use(middleware.AuthMiddleware(db))
	{
		// Get credentials shared with me
		userGroup.GET("/my-credentials", passwordManagerHandler.GetMyCredentials)
		userGroup.GET("/my-credentials/:id", passwordManagerHandler.GetCredentialDetails)
		userGroup.POST("/my-credentials/:id/copy", passwordManagerHandler.CopyPassword)
	}
}
