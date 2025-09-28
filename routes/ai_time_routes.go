package routes

import (
	"project-x/handlers"
	"project-x/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupAITimeRoutes sets up all AI time optimization routes with Arabic working hours context
func SetupAITimeRoutes(r *gin.Engine, db *gorm.DB) {
	aiTimeHandler := handlers.NewAITimeHandler(db)

	// AI Time Optimization routes group
	aiTimeGroup := r.Group("/api/ai/time")
	aiTimeGroup.Use(middleware.AuthMiddleware(db)) // All routes require authentication

	// 🎯 **تحليل الوقت الشامل** - Full Time Analysis (Manager+, HR, Admin only)
	aiTimeGroup.GET("/analysis", aiTimeHandler.GetTimeAnalysis)

	// 👤 **تحليل الوقت الشخصي** - Personal Time Analysis (All users)
	aiTimeGroup.GET("/my-analysis", aiTimeHandler.GetMyTimeAnalysis)

	// 📊 **تقرير المشروع** - Project Time Report (Manager+, HR, Admin only)
	aiTimeGroup.GET("/project/:projectId/report", aiTimeHandler.GetProjectTimeReport)

	// 👥 **تحليل عبء العمل للمستخدم** - User Workload Analysis (Manager+, HR, Admin only)
	aiTimeGroup.GET("/user/:userId/workload", aiTimeHandler.GetUserWorkloadAnalysis)

	// 🚨 **التنبيهات الحرجة** - Critical Time Alerts (Manager+, HR, Admin only)
	aiTimeGroup.GET("/alerts", aiTimeHandler.GetCriticalTimeAlerts)

	// 📈 **لوحة القيادة** - Executive Dashboard (Manager+, HR, Admin only)
	aiTimeGroup.GET("/dashboard", aiTimeHandler.GetTimeOptimizationDashboard)

	// 💡 **التوصيات المدعومة بالذكاء الاصطناعي** - AI-Powered Recommendations (Manager+, HR, Admin only)
	aiTimeGroup.GET("/recommendations", aiTimeHandler.GetTimeOptimizationRecommendations)
}
