package handlers

import (
	"net/http"
	"project-x/config"
	"project-x/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AITimeHandler struct {
	DB           *gorm.DB
	AIOptimizer  *services.AITimeOptimizer
	WorkSchedule *config.WorkScheduleConfig
}

func NewAITimeHandler(db *gorm.DB) *AITimeHandler {
	return &AITimeHandler{
		DB:           db,
		AIOptimizer:  services.NewAITimeOptimizer(db),
		WorkSchedule: config.GetDefaultWorkSchedule(),
	}
}

// GetTimeAnalysis returns comprehensive AI time analysis for all tasks (Manager+, HR, Admin only)
func (h *AITimeHandler) GetTimeAnalysis(c *gin.Context) {
	userRole, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	// Check permissions - only Manager+, HR, or Admin can access full analysis
	if userRole != "admin" && userRole != "manager" && userRole != "hr" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Only Manager+, HR, or Admin can access full time analysis"})
		return
	}

	analyses, err := h.AIOptimizer.AnalyzeTaskTimeRisks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze time risks", "details": err.Error()})
		return
	}

	// Add Arabic working schedule context
	now := time.Now()
	response := gin.H{
		"analyses": analyses,
		"working_schedule": gin.H{
			"working_days":    []string{"Saturday", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday"},
			"working_hours":   "9:00 AM - 4:00 PM",
			"weekly_capacity": h.WorkSchedule.WeeklyHours,
			"lunch_break":     "12:00 PM - 1:00 PM",
			"timezone":        h.WorkSchedule.TimeZone,
			"current_time":    now,
			"is_working_hour": h.WorkSchedule.IsWorkingHour(now),
			"is_working_day":  h.WorkSchedule.IsWorkingDay(now),
		},
		"summary": gin.H{
			"total_tasks":       len(analyses),
			"critical_tasks":    h.countTasksByRisk(analyses, "critical"),
			"high_risk_tasks":   h.countTasksByRisk(analyses, "high"),
			"medium_risk_tasks": h.countTasksByRisk(analyses, "medium"),
			"low_risk_tasks":    h.countTasksByRisk(analyses, "low"),
		},
		"generated_at": now,
	}

	c.JSON(http.StatusOK, response)
}

// GetMyTimeAnalysis returns personal time analysis for current user
func (h *AITimeHandler) GetMyTimeAnalysis(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	// Get user's workload analysis
	workloadAnalysis, err := h.AIOptimizer.AnalyzeUserWorkload(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze personal workload", "details": err.Error()})
		return
	}

	// Get user's tasks analysis
	allAnalyses, err := h.AIOptimizer.AnalyzeTaskTimeRisks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze task risks", "details": err.Error()})
		return
	}

	// Filter analyses for current user
	var userAnalyses []services.TimeAnalysis
	for _, analysis := range allAnalyses {
		// Get task to check if it belongs to current user
		var task struct {
			UserID uint
		}
		if err := h.DB.Model(&struct {
			ID     uint
			UserID uint
		}{}).Where("id = ?", analysis.TaskID).First(&task).Error; err == nil {
			if task.UserID == userID.(uint) {
				userAnalyses = append(userAnalyses, analysis)
			}
		}
	}

	now := time.Now()
	response := gin.H{
		"workload_analysis": workloadAnalysis,
		"task_analyses":     userAnalyses,
		"working_schedule": gin.H{
			"working_days":    []string{"Saturday", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday"},
			"working_hours":   "9:00 AM - 4:00 PM",
			"weekly_capacity": h.WorkSchedule.WeeklyHours,
			"lunch_break":     "12:00 PM - 1:00 PM",
			"timezone":        h.WorkSchedule.TimeZone,
			"current_time":    now,
			"is_working_hour": h.WorkSchedule.IsWorkingHour(now),
			"is_working_day":  h.WorkSchedule.IsWorkingDay(now),
		},
		"personal_summary": gin.H{
			"total_tasks":     len(userAnalyses),
			"critical_tasks":  h.countTasksByRisk(userAnalyses, "critical"),
			"high_risk_tasks": h.countTasksByRisk(userAnalyses, "high"),
			"capacity_usage":  workloadAnalysis.CapacityUtilization,
			"overload_risk":   workloadAnalysis.OverloadRisk,
		},
		"generated_at": now,
	}

	c.JSON(http.StatusOK, response)
}

// GetProjectTimeReport returns AI-powered project time analysis
func (h *AITimeHandler) GetProjectTimeReport(c *gin.Context) {
	userRole, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	// Check permissions
	if userRole != "admin" && userRole != "manager" && userRole != "hr" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Only Manager+, HR, or Admin can access project reports"})
		return
	}

	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	report, err := h.AIOptimizer.AnalyzeProjectTimeRisks(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze project time risks", "details": err.Error()})
		return
	}

	now := time.Now()
	response := gin.H{
		"report": report,
		"working_schedule": gin.H{
			"working_days":    []string{"Saturday", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday"},
			"working_hours":   "9:00 AM - 4:00 PM",
			"weekly_capacity": h.WorkSchedule.WeeklyHours,
			"lunch_break":     "12:00 PM - 1:00 PM",
			"timezone":        h.WorkSchedule.TimeZone,
			"current_time":    now,
			"is_working_hour": h.WorkSchedule.IsWorkingHour(now),
			"is_working_day":  h.WorkSchedule.IsWorkingDay(now),
		},
		"arabic_context": gin.H{
			"working_days_remaining": report.WorkingDaysRemaining,
			"total_working_hours":    report.TotalWorkingHours,
			"estimated_work_weeks":   report.TotalWorkingHours / h.WorkSchedule.WeeklyHours,
		},
		"generated_at": now,
	}

	c.JSON(http.StatusOK, response)
}

// GetUserWorkloadAnalysis returns workload analysis for specific user
func (h *AITimeHandler) GetUserWorkloadAnalysis(c *gin.Context) {
	userRole, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	// Check permissions
	if userRole != "admin" && userRole != "manager" && userRole != "hr" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Only Manager+, HR, or Admin can access user workload analysis"})
		return
	}

	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	analysis, err := h.AIOptimizer.AnalyzeUserWorkload(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze user workload", "details": err.Error()})
		return
	}

	now := time.Now()
	response := gin.H{
		"analysis": analysis,
		"working_schedule": gin.H{
			"working_days":    []string{"Saturday", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday"},
			"working_hours":   "9:00 AM - 4:00 PM",
			"weekly_capacity": h.WorkSchedule.WeeklyHours,
			"lunch_break":     "12:00 PM - 1:00 PM",
			"timezone":        h.WorkSchedule.TimeZone,
			"current_time":    now,
			"is_working_hour": h.WorkSchedule.IsWorkingHour(now),
			"is_working_day":  h.WorkSchedule.IsWorkingDay(now),
		},
		"capacity_insights": gin.H{
			"weekly_capacity_hours":    analysis.WeeklyCapacity,
			"current_workload_hours":   analysis.CurrentWorkload,
			"available_capacity_hours": analysis.WeeklyCapacity - analysis.CurrentWorkload,
			"utilization_percentage":   analysis.CapacityUtilization,
			"overload_risk":            analysis.OverloadRisk,
		},
		"generated_at": now,
	}

	c.JSON(http.StatusOK, response)
}

// GetCriticalTimeAlerts returns urgent time-related alerts
func (h *AITimeHandler) GetCriticalTimeAlerts(c *gin.Context) {
	userRole, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	// Check permissions
	if userRole != "admin" && userRole != "manager" && userRole != "hr" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Only Manager+, HR, or Admin can access critical alerts"})
		return
	}

	alerts, err := h.AIOptimizer.GetCriticalTimeAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get critical alerts", "details": err.Error()})
		return
	}

	now := time.Now()
	response := gin.H{
		"alerts": alerts,
		"working_schedule": gin.H{
			"working_days":    []string{"Saturday", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday"},
			"working_hours":   "9:00 AM - 4:00 PM",
			"weekly_capacity": h.WorkSchedule.WeeklyHours,
			"lunch_break":     "12:00 PM - 1:00 PM",
			"timezone":        h.WorkSchedule.TimeZone,
			"current_time":    now,
			"is_working_hour": h.WorkSchedule.IsWorkingHour(now),
			"is_working_day":  h.WorkSchedule.IsWorkingDay(now),
		},
		"alert_summary": gin.H{
			"total_alerts":    len(alerts),
			"critical_alerts": h.countAlertsBySeverity(alerts, "critical"),
			"high_alerts":     h.countAlertsBySeverity(alerts, "high"),
		},
		"generated_at": now,
	}

	c.JSON(http.StatusOK, response)
}

// GetTimeOptimizationDashboard returns executive dashboard with key metrics
func (h *AITimeHandler) GetTimeOptimizationDashboard(c *gin.Context) {
	userRole, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	// Check permissions
	if userRole != "admin" && userRole != "manager" && userRole != "hr" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Only Manager+, HR, or Admin can access optimization dashboard"})
		return
	}

	// Get comprehensive analysis
	taskAnalyses, err := h.AIOptimizer.AnalyzeTaskTimeRisks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze tasks", "details": err.Error()})
		return
	}

	alerts, err := h.AIOptimizer.GetCriticalTimeAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get alerts", "details": err.Error()})
		return
	}

	// Calculate dashboard metrics
	now := time.Now()
	totalTasks := len(taskAnalyses)
	criticalTasks := h.countTasksByRisk(taskAnalyses, "critical")
	highRiskTasks := h.countTasksByRisk(taskAnalyses, "high")

	// Calculate total working hours across all tasks
	totalWorkingHours := 0.0
	for _, analysis := range taskAnalyses {
		totalWorkingHours += float64(analysis.EstimatedDuration)
	}

	// Get user count and calculate average workload
	var userCount int64
	h.DB.Model(&struct {
		ID uint
	}{}).Count(&userCount)

	avgWorkloadPerUser := 0.0
	if userCount > 0 {
		avgWorkloadPerUser = totalWorkingHours / float64(userCount)
	}

	// Calculate organization capacity utilization
	totalCapacity := float64(userCount) * h.WorkSchedule.WeeklyHours
	organizationUtilization := 0.0
	if totalCapacity > 0 {
		organizationUtilization = (totalWorkingHours / totalCapacity) * 100
	}

	response := gin.H{
		"dashboard": gin.H{
			"overview": gin.H{
				"total_tasks":               totalTasks,
				"critical_tasks":            criticalTasks,
				"high_risk_tasks":           highRiskTasks,
				"total_working_hours":       totalWorkingHours,
				"average_workload_per_user": avgWorkloadPerUser,
				"organization_utilization":  organizationUtilization,
			},
			"capacity_metrics": gin.H{
				"total_users":                 userCount,
				"weekly_capacity_per_user":    h.WorkSchedule.WeeklyHours,
				"total_organization_capacity": totalCapacity,
				"utilized_capacity":           totalWorkingHours,
				"available_capacity":          totalCapacity - totalWorkingHours,
			},
			"risk_distribution": gin.H{
				"critical": criticalTasks,
				"high":     highRiskTasks,
				"medium":   h.countTasksByRisk(taskAnalyses, "medium"),
				"low":      h.countTasksByRisk(taskAnalyses, "low"),
			},
			"alerts": gin.H{
				"total_alerts":    len(alerts),
				"critical_alerts": h.countAlertsBySeverity(alerts, "critical"),
				"high_alerts":     h.countAlertsBySeverity(alerts, "high"),
			},
		},
		"working_schedule": gin.H{
			"working_days":    []string{"Saturday", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday"},
			"working_hours":   "9:00 AM - 4:00 PM",
			"weekly_capacity": h.WorkSchedule.WeeklyHours,
			"lunch_break":     "12:00 PM - 1:00 PM",
			"timezone":        h.WorkSchedule.TimeZone,
			"current_time":    now,
			"is_working_hour": h.WorkSchedule.IsWorkingHour(now),
			"is_working_day":  h.WorkSchedule.IsWorkingDay(now),
		},
		"recommendations": gin.H{
			"optimization_tips": []string{
				"Focus on resolving " + strconv.Itoa(criticalTasks) + " critical tasks first",
				"Organization capacity utilization is " + strconv.FormatFloat(organizationUtilization, 'f', 1, 64) + "%",
				"Average workload per user is " + strconv.FormatFloat(avgWorkloadPerUser, 'f', 1, 64) + " hours",
				"Consider redistributing tasks if utilization exceeds 85%",
				"Schedule complex tasks during peak hours (10 AM - 12 PM)",
			},
		},
		"generated_at": now,
	}

	c.JSON(http.StatusOK, response)
}

// GetTimeOptimizationRecommendations returns AI-powered recommendations
func (h *AITimeHandler) GetTimeOptimizationRecommendations(c *gin.Context) {
	userRole, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	// Check permissions
	if userRole != "admin" && userRole != "manager" && userRole != "hr" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Only Manager+, HR, or Admin can access recommendations"})
		return
	}

	// Get task analyses
	taskAnalyses, err := h.AIOptimizer.AnalyzeTaskTimeRisks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze tasks", "details": err.Error()})
		return
	}

	// Collect all recommendations
	var allRecommendations []string
	var taskSpecificRecommendations []gin.H

	for _, analysis := range taskAnalyses {
		if analysis.DeadlineRisk == "critical" || analysis.DeadlineRisk == "high" {
			taskSpecificRecommendations = append(taskSpecificRecommendations, gin.H{
				"task_id":                 analysis.TaskID,
				"task_title":              analysis.TaskTitle,
				"risk_level":              analysis.DeadlineRisk,
				"recommendations":         analysis.Recommendations,
				"working_hours_remaining": analysis.WorkingHoursUntilDeadline,
			})
		}
		allRecommendations = append(allRecommendations, analysis.Recommendations...)
	}

	// Generate organization-level recommendations
	organizationRecommendations := []string{
		"Implement daily stand-ups during working hours (9 AM - 4 PM)",
		"Schedule weekly planning sessions on Saturday mornings",
		"Use Wednesday afternoons for task reviews and adjustments",
		"Avoid scheduling critical deadlines on Thursday afternoons",
		"Plan buffer time for tasks that span multiple working days",
		"Consider workload redistribution for users exceeding 85% capacity",
		"Implement time-blocking for complex tasks during peak hours",
	}

	now := time.Now()
	response := gin.H{
		"recommendations": gin.H{
			"organization_level":  organizationRecommendations,
			"task_specific":       taskSpecificRecommendations,
			"all_recommendations": allRecommendations,
		},
		"working_schedule": gin.H{
			"working_days":    []string{"Saturday", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday"},
			"working_hours":   "9:00 AM - 4:00 PM",
			"weekly_capacity": h.WorkSchedule.WeeklyHours,
			"lunch_break":     "12:00 PM - 1:00 PM",
			"timezone":        h.WorkSchedule.TimeZone,
			"current_time":    now,
			"is_working_hour": h.WorkSchedule.IsWorkingHour(now),
			"is_working_day":  h.WorkSchedule.IsWorkingDay(now),
		},
		"implementation_guide": gin.H{
			"immediate_actions": []string{
				"Address all critical risk tasks within current working day",
				"Redistribute high-risk tasks if user capacity exceeds 100%",
				"Schedule task reviews for next working day morning",
			},
			"weekly_actions": []string{
				"Conduct capacity planning every Saturday morning",
				"Review and adjust task deadlines mid-week (Tuesday/Wednesday)",
				"Prepare weekly reports every Thursday afternoon",
			},
			"cultural_considerations": []string{
				"Respect Friday as a complete day off",
				"Plan important deadlines before Thursday to avoid weekend pressure",
				"Consider Ramadan schedule adjustments when applicable",
				"Maintain work-life balance with strict 9 AM - 4 PM boundaries",
			},
		},
		"generated_at": now,
	}

	c.JSON(http.StatusOK, response)
}

// GetAIPerformanceMetrics returns AI prediction accuracy metrics
func (h *AITimeHandler) GetAIPerformanceMetrics(c *gin.Context) {
	userRole, exists := c.Get("userRole")
	if !exists || (userRole != "admin" && userRole != "manager" && userRole != "hr") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Only Manager+, HR, or Admin can access AI performance metrics"})
		return
	}

	var metrics struct {
		TotalPredictions     int64   `json:"total_predictions"`
		CompletedPredictions int64   `json:"completed_predictions"`
		AvgAccuracyScore     float64 `json:"avg_accuracy_score"`
		AccuracyRate         float64 `json:"accuracy_rate"`
		AvgDurationDiff      float64 `json:"avg_duration_difference_hours"`
	}

	h.DB.Raw(`
		SELECT 
			COUNT(*) as total_predictions,
			COUNT(CASE WHEN actual_duration IS NOT NULL THEN 1 END) as completed_predictions,
			COALESCE(AVG(accuracy_score), 0) as avg_accuracy_score,
			COALESCE(COUNT(CASE WHEN was_accurate = true THEN 1 END)::float / NULLIF(COUNT(CASE WHEN actual_duration IS NOT NULL THEN 1 END), 0) * 100, 0) as accuracy_rate,
			COALESCE(AVG(duration_difference), 0) as avg_duration_difference_hours
		FROM ai_analyses
	`).Scan(&metrics)

	c.JSON(http.StatusOK, gin.H{
		"metrics":      metrics,
		"generated_at": time.Now(),
		"message":      "AI performance metrics retrieved successfully",
	})
}

// Helper functions

func (h *AITimeHandler) countTasksByRisk(analyses []services.TimeAnalysis, risk string) int {
	count := 0
	for _, analysis := range analyses {
		if analysis.DeadlineRisk == risk {
			count++
		}
	}
	return count
}

func (h *AITimeHandler) countAlertsBySeverity(alerts []map[string]interface{}, severity string) int {
	count := 0
	for _, alert := range alerts {
		if alert["severity"] == severity {
			count++
		}
	}
	return count
}
