package routes

import (
	"project-x/handlers"
	"project-x/middleware"
	"project-x/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupHRProblemRoutes(r *gin.Engine, db *gorm.DB, notificationService *services.NotificationService) {
	// Create HR problem service and handler
	hrProblemService := services.NewHRProblemService(db, notificationService)
	hrProblemHandler := handlers.NewHRProblemHandler(db, hrProblemService)

	// HR Problem API routes
	hrProblemAPI := r.Group("/api/hr-problems")
	hrProblemAPI.Use(middleware.AuthMiddleware(db)) // All endpoints require authentication

	// Public endpoints (all authenticated users can access)
	{
		// Get problem categories and priorities
		hrProblemAPI.GET("/categories", hrProblemHandler.GetProblemCategories)

		// Report a new problem
		hrProblemAPI.POST("", hrProblemHandler.CreateProblem)

		// Get user's own problems
		hrProblemAPI.GET("/my-problems", hrProblemHandler.GetUserProblems)

		// Get specific problem (users can only see their own, HR/Admin can see all)
		hrProblemAPI.GET("/:id", hrProblemHandler.GetProblem)

		// Add comment to a problem (users can comment on their own problems)
		hrProblemAPI.POST("/:id/comments", hrProblemHandler.AddComment)
	}

	// HR and Admin only endpoints
	hrOnlyAPI := hrProblemAPI.Group("")
	hrOnlyAPI.Use(middleware.RequireHROrAdmin())
	{
		// Get all problems with filtering
		hrOnlyAPI.GET("", hrProblemHandler.GetAllProblems)

		// Get assigned problems for current HR user
		hrOnlyAPI.GET("/assigned", hrProblemHandler.GetAssignedProblems)

		// Update problem status
		hrOnlyAPI.PATCH("/:id/status", hrProblemHandler.UpdateProblemStatus)

		// Assign problem to HR user
		hrOnlyAPI.PATCH("/:id/assign", hrProblemHandler.AssignProblem)

		// Update HR internal notes
		hrOnlyAPI.PATCH("/:id/notes", hrProblemHandler.UpdateHRNotes)

		// Set final resolution
		hrOnlyAPI.PATCH("/:id/resolution", hrProblemHandler.SetResolution)

		// Get statistics
		hrOnlyAPI.GET("/statistics", hrProblemHandler.GetStatistics)
	}
}
