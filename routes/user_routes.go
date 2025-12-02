package routes

import (
	"project-x/handlers"
	"project-x/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupUserRoutes(r *gin.Engine, db *gorm.DB) {
	userHandler := handlers.NewUserHandler(db)

	userGroup := r.Group("/api/users")
	userGroup.Use(middleware.AuthMiddleware(db))
	{
		// Admin only routes
		userGroup.POST("", middleware.RequireAdmin(), userHandler.CreateUser)
		userGroup.PATCH("/:id/role", middleware.RequireAdmin(), userHandler.UpdateUserRole)
		userGroup.DELETE("/:id", middleware.RequireAdmin(), userHandler.DeleteUser)

		// HR or Admin routes - HR can view user information for HR purposes
		userGroup.GET("", middleware.RequireHROrHigher(), userHandler.ListUsers)
		userGroup.GET("/role/:role", middleware.RequireHROrAdmin(), userHandler.GetUsersByRole)
		userGroup.GET("/department/:department", middleware.RequireHROrAdmin(), userHandler.GetUsersByDepartment)
		userGroup.PATCH("/:id/department", middleware.RequireHROrAdmin(), userHandler.UpdateUserDepartment)

		// Routes accessible by admin or the user themselves
		userGroup.GET("/:id", middleware.RequireSelfOrAdmin(), userHandler.GetUser)
		userGroup.GET("/:id/stats", middleware.RequireSelfOrAdmin(), userHandler.GetUserStats)
		userGroup.PATCH("/:id/password", middleware.RequireSelfOrAdmin(), userHandler.UpdateUserPassword)
		userGroup.PATCH("/:id/skills", middleware.RequireSelfOrAdmin(), userHandler.UpdateUserSkills)
	}
}
