package handlers

import (
	"net/http"
	"project-x/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	DB           *gorm.DB
	AdminService *services.AdminService
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{
		DB:           db,
		AdminService: services.NewAdminService(db),
	}
}

// GetAdminDashboard returns comprehensive dashboard data for admin
func (h *AdminHandler) GetAdminDashboard(c *gin.Context) {
	dashboard, err := h.AdminService.GetAdminDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dashboard": dashboard})
}

// GetAttentionItems returns all items requiring admin attention
func (h *AdminHandler) GetAttentionItems(c *gin.Context) {
	items, err := h.AdminService.GetAttentionItems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch attention items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GetProjectHealthSummary returns health status of all projects
func (h *AdminHandler) GetProjectHealthSummary(c *gin.Context) {
	summary, err := h.AdminService.GetProjectHealthSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project health summary"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"project_health": summary})
}

// GetInactiveUsers returns users who haven't logged in recently
func (h *AdminHandler) GetInactiveUsers(c *gin.Context) {
	daysStr := c.Query("days")
	days := 30 // Default to 30 days

	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	users, err := h.AdminService.GetInactiveUsers(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch inactive users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"inactive_users": users,
		"days":           days,
		"count":          len(users),
	})
}

// GetRecentActivity returns recent system activity
func (h *AdminHandler) GetRecentActivity(c *gin.Context) {
	limitStr := c.Query("limit")
	limit := 20 // Default to 20 items

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	activities, err := h.AdminService.GetRecentActivity(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
		"limit":      limit,
		"count":      len(activities),
	})
}

// GetSystemHealth returns detailed system health information
func (h *AdminHandler) GetSystemHealth(c *gin.Context) {
	health := h.AdminService.GetSystemHealth()
	c.JSON(http.StatusOK, gin.H{"system_health": health})
}

// GetChecklistStatus returns today's checklist status
func (h *AdminHandler) GetChecklistStatus(c *gin.Context) {
	checklist, err := h.AdminService.GetTodayChecklistStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch checklist status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"checklist":             checklist,
		"completion_percentage": checklist.GetCompletionPercentage(),
		"is_completed":          checklist.IsCompleted(),
	})
}

// UpdateChecklistItem updates a specific checklist item
func (h *AdminHandler) UpdateChecklistItem(c *gin.Context) {
	var request struct {
		Item      string `json:"item" binding:"required"` // hr_problems, critical_tasks, etc.
		Completed bool   `json:"completed"`               // true to mark as complete
		Notes     string `json:"notes"`                   // Optional notes
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	checklist, err := h.AdminService.UpdateChecklistItem(
		request.Item,
		request.Completed,
		userID.(uint),
		request.Notes,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"checklist":             checklist,
		"completion_percentage": checklist.GetCompletionPercentage(),
		"is_completed":          checklist.IsCompleted(),
		"message":               "Checklist item updated successfully",
	})
}

// GetChecklistHistory returns checklist history for a date range
func (h *AdminHandler) GetChecklistHistory(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	limitStr := c.Query("limit")

	var startDate, endDate time.Time
	var err error

	// Default to last 30 days if not provided
	if startDateStr == "" {
		startDate = time.Now().AddDate(0, 0, -30)
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
			return
		}
		// Set to midnight
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	}

	if endDateStr == "" {
		endDate = time.Now()
	} else {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
			return
		}
		// Set to end of day
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())
	}

	limit := 30 // Default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	checklists, err := h.AdminService.GetChecklistHistory(startDate, endDate, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch checklist history"})
		return
	}

	// Format response with completion percentages
	formattedChecklists := make([]map[string]interface{}, len(checklists))
	for i, checklist := range checklists {
		formattedChecklists[i] = map[string]interface{}{
			"id":                      checklist.ID,
			"date":                    checklist.Date,
			"hr_problems_reviewed":    checklist.HRProblemsReviewed,
			"critical_tasks_reviewed": checklist.CriticalTasksReviewed,
			"project_health_reviewed": checklist.ProjectHealthReviewed,
			"user_activity_reviewed":  checklist.UserActivityReviewed,
			"system_alerts_reviewed":  checklist.SystemAlertsReviewed,
			"ai_performance_reviewed": checklist.AIPerformanceReviewed,
			"backup_status_reviewed":  checklist.BackupStatusReviewed,
			"security_logs_reviewed":  checklist.SecurityLogsReviewed,
			"completion_percentage":   checklist.GetCompletionPercentage(),
			"is_completed":            checklist.IsCompleted(),
			"completed_at":            checklist.CompletedAt,
			"completed_by":            checklist.CompletedBy,
			"notes":                   checklist.Notes,
			"created_at":              checklist.CreatedAt,
			"updated_at":              checklist.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"checklists": formattedChecklists,
		"count":      len(checklists),
		"start_date": startDate,
		"end_date":   endDate,
	})
}

// GetChecklistStats returns statistics about checklist completion
func (h *AdminHandler) GetChecklistStats(c *gin.Context) {
	daysStr := c.Query("days")
	days := 30 // Default to 30 days

	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 && parsedDays <= 365 {
			days = parsedDays
		}
	}

	stats, err := h.AdminService.GetChecklistStats(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch checklist statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}
