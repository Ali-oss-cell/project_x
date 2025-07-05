package routes

import (
	"project-x/handlers"
	"project-x/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupTaskRoutes(r *gin.Engine, db *gorm.DB) {
	taskHandler := handlers.NewTaskHandler(db)

	taskGroup := r.Group("/api/tasks")
	taskGroup.Use(middleware.AuthMiddleware(db))
	{
		// Smart unified endpoint - handles everything with query parameters
		taskGroup.POST("", middleware.RequireHeadOrHigher(), taskHandler.SmartTaskHandler)         // Create tasks (?action=create)
		taskGroup.GET("", taskHandler.SmartTaskHandler)                                            // Query tasks (?action=query or default)
		taskGroup.PUT("/:id", taskHandler.SmartTaskHandler)                                        // Update tasks (?action=update&type=regular/collaborative)
		taskGroup.DELETE("/:id", middleware.RequireHeadOrHigher(), taskHandler.SmartTaskHandler)   // Delete tasks (?action=delete&type=regular/collaborative)
		taskGroup.POST("/bulk", middleware.RequireManagerOrHigher(), taskHandler.SmartTaskHandler) // Bulk operations (?action=bulk_update)

		// Status updates (kept separate for clarity)
		taskGroup.PATCH("/:id/status", taskHandler.UpdateTaskStatus)                            // Update regular task status
		taskGroup.PATCH("/:id/collaborative/status", taskHandler.UpdateCollaborativeTaskStatus) // Update collaborative task status

		// Statistics endpoint (Manager+ only)
		taskGroup.GET("/statistics", middleware.RequireManagerOrHigher(), taskHandler.GetTaskStatistics)

		// Legacy endpoints (for backward compatibility - can be removed later)
		// These maintain the old API structure while you migrate frontend
		taskGroup.POST("/legacy", middleware.RequireHeadOrHigher(), taskHandler.CreateTask)
		taskGroup.POST("/legacy/collaborative", middleware.RequireHeadOrHigher(), taskHandler.CreateCollaborativeTask)
		taskGroup.GET("/legacy/my", taskHandler.GetUserTasks)
		taskGroup.GET("/legacy/collaborative", taskHandler.GetUserCollaborativeTasks)
		taskGroup.GET("/legacy/status/:status", taskHandler.GetTasksByStatus)
		taskGroup.GET("/legacy/collaborative/status/:status", taskHandler.GetCollaborativeTasksByStatus)
		taskGroup.GET("/legacy/project/:projectId", taskHandler.GetProjectTasks)
		taskGroup.GET("/legacy/project/:projectId/collaborative", taskHandler.GetProjectCollaborativeTasks)
		taskGroup.GET("/legacy/department/:dept", middleware.RequireManagerOrHigher(), taskHandler.GetTasksByDepartment)
		taskGroup.GET("/legacy/department/:dept/collaborative", middleware.RequireManagerOrHigher(), taskHandler.GetCollaborativeTasksByDepartment)
		taskGroup.POST("/legacy/assign", middleware.RequireManagerOrHigher(), taskHandler.CreateTaskForUser)
		taskGroup.POST("/legacy/collaborative/assign", middleware.RequireManagerOrHigher(), taskHandler.CreateCollaborativeTaskForUser)
		taskGroup.POST("/legacy/bulk-update", middleware.RequireManagerOrHigher(), taskHandler.BulkUpdateTaskStatus)
		taskGroup.GET("/legacy/reports/project/:projectId", middleware.RequireManagerOrHigher(), taskHandler.GetProjectReport)
		taskGroup.GET("/legacy/reports/user/:userId", middleware.RequireManagerOrHigher(), taskHandler.GetUserReport)
	}
}
