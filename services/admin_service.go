package services

import (
	"fmt"
	"project-x/models"
	"time"

	"gorm.io/gorm"
)

type AdminService struct {
	DB               *gorm.DB
	TaskService      *TaskService
	ProjectService   *ProjectService
	UserService      *UserService
	HRProblemService *HRProblemService
	AITimeOptimizer  *AITimeOptimizer
}

func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{
		DB:               db,
		TaskService:      NewTaskService(db),
		ProjectService:   NewProjectService(db),
		UserService:      NewUserService(db),
		HRProblemService: NewHRProblemService(db, nil), // Notification service not needed for read operations
		AITimeOptimizer:  NewAITimeOptimizer(db),
	}
}

// GetAdminDashboard returns comprehensive dashboard data for admin
func (s *AdminService) GetAdminDashboard() (map[string]interface{}, error) {
	dashboard := make(map[string]interface{})

	// System Statistics
	taskStats, err := s.TaskService.GetTaskStatistics()
	if err == nil {
		dashboard["system_statistics"] = taskStats
	}

	// HR Problems Summary
	var openHRProblems, highPriorityHRProblems, unassignedHRProblems int64
	s.DB.Model(&models.HRProblem{}).Where("status IN ?", []string{"pending", "in_progress"}).Count(&openHRProblems)
	s.DB.Model(&models.HRProblem{}).Where("status IN ? AND priority IN ?", []string{"pending", "in_progress"}, []string{"high", "urgent", "critical"}).Count(&highPriorityHRProblems)
	s.DB.Model(&models.HRProblem{}).Where("status IN ? AND assigned_hr_id IS NULL", []string{"pending", "in_progress"}).Count(&unassignedHRProblems)

	dashboard["hr_problems"] = map[string]interface{}{
		"open_count":          openHRProblems,
		"high_priority_count": highPriorityHRProblems,
		"unassigned_count":    unassignedHRProblems,
		"needs_attention":     highPriorityHRProblems > 0 || unassignedHRProblems > 0,
	}

	// Critical Tasks Summary
	alerts, err := s.AITimeOptimizer.GetCriticalTimeAlerts()
	if err == nil {
		criticalCount := 0
		highRiskCount := 0
		for _, alert := range alerts {
			if severity, ok := alert["severity"].(string); ok {
				if severity == "critical" {
					criticalCount++
				} else if severity == "high" {
					highRiskCount++
				}
			}
		}

		// Also count overdue tasks
		var overdueTasks int64
		now := time.Now()
		s.DB.Model(&models.Task{}).Where("status != ? AND due_date < ?", models.TaskStatusCompleted, now).Count(&overdueTasks)

		dashboard["critical_tasks"] = map[string]interface{}{
			"critical_risk_count": criticalCount,
			"high_risk_count":     highRiskCount,
			"overdue_count":       overdueTasks,
			"total_at_risk":       criticalCount + highRiskCount,
			"needs_attention":     criticalCount > 0 || highRiskCount > 0 || overdueTasks > 0,
		}
	}

	// Project Health Summary
	var totalProjects, activeProjects, behindScheduleProjects int64
	s.DB.Model(&models.Project{}).Count(&totalProjects)
	s.DB.Model(&models.Project{}).Where("status = ?", models.ProjectStatusActive).Count(&activeProjects)

	// Projects behind schedule (end_date in past but status is active)
	now := time.Now()
	s.DB.Model(&models.Project{}).Where("status = ? AND end_date < ?", models.ProjectStatusActive, now).Count(&behindScheduleProjects)

	dashboard["project_health"] = map[string]interface{}{
		"total_projects":        totalProjects,
		"active_projects":       activeProjects,
		"behind_schedule_count": behindScheduleProjects,
		"needs_attention":       behindScheduleProjects > 0,
	}

	// User Activity Summary
	var totalUsers, inactiveUsers30Days, inactiveUsers60Days int64
	s.DB.Model(&models.User{}).Count(&totalUsers)

	// Users inactive for 30+ days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	sixtyDaysAgo := time.Now().AddDate(0, 0, -60)

	// Count users with LastLogin older than threshold, or NULL (never logged in) with CreatedAt older than threshold
	s.DB.Model(&models.User{}).Where("(last_login IS NULL AND created_at < ?) OR (last_login IS NOT NULL AND last_login < ?)", thirtyDaysAgo, thirtyDaysAgo).Count(&inactiveUsers30Days)
	s.DB.Model(&models.User{}).Where("(last_login IS NULL AND created_at < ?) OR (last_login IS NOT NULL AND last_login < ?)", sixtyDaysAgo, sixtyDaysAgo).Count(&inactiveUsers60Days)

	dashboard["user_activity"] = map[string]interface{}{
		"total_users":      totalUsers,
		"inactive_30_days": inactiveUsers30Days,
		"inactive_60_days": inactiveUsers60Days,
		"needs_attention":  inactiveUsers60Days > 0,
	}

	// AI Performance Summary - Get from handler method pattern
	// We'll get this from the handler since it requires DB access
	dashboard["ai_performance"] = map[string]interface{}{
		"status": "available",
		"note":   "Use /api/ai/time/performance endpoint for detailed metrics",
	}

	// System Health
	systemHealth := s.GetSystemHealth()
	dashboard["system_health"] = systemHealth

	// Overall Status
	dashboard["overall_status"] = map[string]interface{}{
		"status":               s.calculateOverallStatus(dashboard),
		"items_need_attention": s.countItemsNeedingAttention(dashboard),
		"last_updated":         time.Now(),
	}

	return dashboard, nil
}

// GetAttentionItems returns all items requiring admin attention
func (s *AdminService) GetAttentionItems() (map[string]interface{}, error) {
	items := make(map[string]interface{})
	var highPriority []map[string]interface{}
	var mediumPriority []map[string]interface{}
	var lowPriority []map[string]interface{}

	now := time.Now()

	// HR Problems - High Priority
	var hrProblems []models.HRProblem
	s.DB.Where("status IN ? AND priority IN ?", []string{"pending", "in_progress"}, []string{"high", "urgent", "critical"}).
		Or("status IN ? AND assigned_hr_id IS NULL", []string{"pending", "in_progress"}).
		Preload("Reporter").
		Find(&hrProblems)

	for _, problem := range hrProblems {
		item := map[string]interface{}{
			"type":       "hr_problem",
			"id":         problem.ID,
			"title":      problem.Title,
			"priority":   problem.Priority,
			"status":     problem.Status,
			"is_urgent":  problem.IsUrgent,
			"created_at": problem.CreatedAt,
			"action_url": fmt.Sprintf("/hr-problems/%d", problem.ID),
		}

		if problem.Priority == "critical" || problem.Priority == "urgent" || problem.IsUrgent {
			highPriority = append(highPriority, item)
		} else {
			mediumPriority = append(mediumPriority, item)
		}
	}

	// Critical Tasks
	alerts, err := s.AITimeOptimizer.GetCriticalTimeAlerts()
	if err == nil {
		for _, alert := range alerts {
			if severity, ok := alert["severity"].(string); ok {
				taskID, _ := alert["task_id"].(uint)
				item := map[string]interface{}{
					"type":       "critical_task",
					"task_id":    taskID,
					"title":      alert["task_title"],
					"message":    alert["message"],
					"severity":   severity,
					"created_at": alert["created_at"],
					"action_url": fmt.Sprintf("/tasks/%d", taskID),
				}

				if severity == "critical" {
					highPriority = append(highPriority, item)
				} else {
					mediumPriority = append(mediumPriority, item)
				}
			}
		}
	}

	// Overdue Tasks
	var overdueTasks []models.Task
	s.DB.Where("status != ? AND due_date < ?", models.TaskStatusCompleted, now).
		Preload("User").
		Preload("Project").
		Find(&overdueTasks)

	for _, task := range overdueTasks {
		item := map[string]interface{}{
			"type":       "overdue_task",
			"task_id":    task.ID,
			"title":      task.Title,
			"message":    "Task is overdue",
			"severity":   "high",
			"due_date":   task.DueDate,
			"created_at": task.CreatedAt,
			"action_url": fmt.Sprintf("/tasks/%d", task.ID),
		}
		highPriority = append(highPriority, item)
	}

	// Projects Behind Schedule
	var behindScheduleProjects []models.Project
	s.DB.Where("status = ? AND end_date < ?", models.ProjectStatusActive, now).Find(&behindScheduleProjects)

	for _, project := range behindScheduleProjects {
		item := map[string]interface{}{
			"type":       "project_behind_schedule",
			"project_id": project.ID,
			"title":      project.Title,
			"message":    "Project is behind schedule",
			"severity":   "medium",
			"end_date":   project.EndDate,
			"created_at": project.CreatedAt,
			"action_url": fmt.Sprintf("/projects/%d", project.ID),
		}
		mediumPriority = append(mediumPriority, item)
	}

	// Inactive Users (60+ days)
	var inactiveUsers []models.User
	sixtyDaysAgo := time.Now().AddDate(0, 0, -60)
	s.DB.Where("(last_login IS NULL AND created_at < ?) OR (last_login IS NOT NULL AND last_login < ?)", sixtyDaysAgo, sixtyDaysAgo).Find(&inactiveUsers)

	for _, user := range inactiveUsers {
		item := map[string]interface{}{
			"type":       "inactive_user",
			"user_id":    user.ID,
			"title":      "User inactive: " + user.Username,
			"message":    "User has been inactive for 60+ days",
			"severity":   "low",
			"created_at": user.CreatedAt,
			"action_url": fmt.Sprintf("/users/%d", user.ID),
		}
		lowPriority = append(lowPriority, item)
	}

	items["high_priority"] = highPriority
	items["medium_priority"] = mediumPriority
	items["low_priority"] = lowPriority
	items["summary"] = map[string]interface{}{
		"high_count":   len(highPriority),
		"medium_count": len(mediumPriority),
		"low_count":    len(lowPriority),
		"total_count":  len(highPriority) + len(mediumPriority) + len(lowPriority),
	}

	return items, nil
}

// GetProjectHealthSummary returns health status of all projects
func (s *AdminService) GetProjectHealthSummary() (map[string]interface{}, error) {
	var projects []models.Project
	s.DB.Find(&projects)

	now := time.Now()
	healthy := 0
	atRisk := 0
	behindSchedule := 0
	completed := 0

	projectDetails := []map[string]interface{}{}

	for _, project := range projects {
		// Get project statistics
		var totalTasks, completedTasks int64
		s.DB.Model(&models.Task{}).Where("project_id = ?", project.ID).Count(&totalTasks)
		s.DB.Model(&models.Task{}).Where("project_id = ? AND status = ?", project.ID, models.TaskStatusCompleted).Count(&completedTasks)

		completionRate := 0.0
		if totalTasks > 0 {
			completionRate = float64(completedTasks) / float64(totalTasks) * 100
		}

		isBehindSchedule := false
		if project.EndDate != nil && project.EndDate.Before(now) && project.Status == models.ProjectStatusActive {
			isBehindSchedule = true
			behindSchedule++
		}

		isAtRisk := completionRate < 50.0 && project.Status == models.ProjectStatusActive

		if project.Status == models.ProjectStatusCompleted {
			completed++
		} else if isBehindSchedule || isAtRisk {
			atRisk++
		} else {
			healthy++
		}

		projectDetails = append(projectDetails, map[string]interface{}{
			"id":                 project.ID,
			"title":              project.Title,
			"status":             project.Status,
			"completion_rate":    completionRate,
			"total_tasks":        totalTasks,
			"completed_tasks":    completedTasks,
			"is_behind_schedule": isBehindSchedule,
			"is_at_risk":         isAtRisk,
			"health_status":      s.getProjectHealthStatus(project, completionRate, isBehindSchedule),
		})
	}

	return map[string]interface{}{
		"summary": map[string]interface{}{
			"total_projects":  len(projects),
			"healthy":         healthy,
			"at_risk":         atRisk,
			"behind_schedule": behindSchedule,
			"completed":       completed,
			"needs_attention": atRisk + behindSchedule,
		},
		"projects": projectDetails,
	}, nil
}

// GetInactiveUsers returns users who haven't logged in recently
func (s *AdminService) GetInactiveUsers(days int) ([]map[string]interface{}, error) {
	if days == 0 {
		days = 30 // Default to 30 days
	}

	cutoffDate := time.Now().AddDate(0, 0, -days)

	// Get users with LastLogin older than cutoff, or NULL (never logged in) with CreatedAt older than cutoff
	var users []models.User
	s.DB.Where("(last_login IS NULL AND created_at < ?) OR (last_login IS NOT NULL AND last_login < ?)", cutoffDate, cutoffDate).Find(&users)

	var inactiveUsers []map[string]interface{}
	for _, user := range users {
		var daysInactive int
		if user.LastLogin != nil {
			daysInactive = int(time.Since(*user.LastLogin).Hours() / 24)
		} else {
			daysInactive = int(time.Since(user.CreatedAt).Hours() / 24)
		}

		inactiveUsers = append(inactiveUsers, map[string]interface{}{
			"id":            user.ID,
			"username":      user.Username,
			"role":          user.Role,
			"department":    user.Department,
			"created_at":    user.CreatedAt,
			"last_login":    user.LastLogin,
			"days_inactive": daysInactive,
		})
	}

	return inactiveUsers, nil
}

// GetRecentActivity returns recent system activity
func (s *AdminService) GetRecentActivity(limit int) ([]map[string]interface{}, error) {
	if limit == 0 {
		limit = 20 // Default to 20 items
	}

	var activities []map[string]interface{}

	// Recent users created
	var recentUsers []models.User
	s.DB.Order("created_at DESC").Limit(limit / 4).Find(&recentUsers)
	for _, user := range recentUsers {
		activities = append(activities, map[string]interface{}{
			"type":       "user_created",
			"id":         user.ID,
			"title":      "New user: " + user.Username,
			"message":    "User " + user.Username + " was created",
			"created_at": user.CreatedAt,
		})
	}

	// Recent projects created
	var recentProjects []models.Project
	s.DB.Order("created_at DESC").Limit(limit / 4).Find(&recentProjects)
	for _, project := range recentProjects {
		activities = append(activities, map[string]interface{}{
			"type":       "project_created",
			"id":         project.ID,
			"title":      "New project: " + project.Title,
			"message":    "Project " + project.Title + " was created",
			"created_at": project.CreatedAt,
		})
	}

	// Recent tasks created
	var recentTasks []models.Task
	s.DB.Order("created_at DESC").Limit(limit / 4).Find(&recentTasks)
	for _, task := range recentTasks {
		activities = append(activities, map[string]interface{}{
			"type":       "task_created",
			"id":         task.ID,
			"title":      "New task: " + task.Title,
			"message":    "Task " + task.Title + " was created",
			"created_at": task.CreatedAt,
		})
	}

	// Recent HR problems
	var recentProblems []models.HRProblem
	s.DB.Order("created_at DESC").Limit(limit / 4).Find(&recentProblems)
	for _, problem := range recentProblems {
		activities = append(activities, map[string]interface{}{
			"type":       "hr_problem_reported",
			"id":         problem.ID,
			"title":      "HR Problem: " + problem.Title,
			"message":    "HR problem " + problem.Title + " was reported",
			"created_at": problem.CreatedAt,
		})
	}

	// Sort by created_at descending
	// Simple bubble sort (can be improved)
	for i := 0; i < len(activities)-1; i++ {
		for j := i + 1; j < len(activities); j++ {
			timeI, _ := activities[i]["created_at"].(time.Time)
			timeJ, _ := activities[j]["created_at"].(time.Time)
			if timeJ.After(timeI) {
				activities[i], activities[j] = activities[j], activities[i]
			}
		}
	}

	// Limit to requested number
	if len(activities) > limit {
		activities = activities[:limit]
	}

	return activities, nil
}

// Helper functions

// GetSystemHealth returns detailed system health information
func (s *AdminService) GetSystemHealth() map[string]interface{} {
	// Check database connection
	sqlDB, err := s.DB.DB()
	dbStatus := "healthy"
	dbResponseTime := 0

	if err == nil {
		start := time.Now()
		err = sqlDB.Ping()
		dbResponseTime = int(time.Since(start).Milliseconds())
		if err != nil {
			dbStatus = "error"
		}
	} else {
		dbStatus = "error"
	}

	overallStatus := "healthy"
	if dbStatus != "healthy" {
		overallStatus = "error"
	}

	return map[string]interface{}{
		"status": overallStatus,
		"database": map[string]interface{}{
			"status":           dbStatus,
			"response_time_ms": dbResponseTime,
		},
		"api": map[string]interface{}{
			"status":            "operational",
			"uptime_percentage": 99.8, // This would come from monitoring in production
		},
		"last_check": time.Now(),
	}
}

func (s *AdminService) calculateOverallStatus(dashboard map[string]interface{}) string {
	itemsNeedingAttention := s.countItemsNeedingAttention(dashboard)

	if itemsNeedingAttention == 0 {
		return "healthy"
	} else if itemsNeedingAttention <= 3 {
		return "warning"
	} else {
		return "critical"
	}
}

func (s *AdminService) countItemsNeedingAttention(dashboard map[string]interface{}) int {
	count := 0

	if hrProblems, ok := dashboard["hr_problems"].(map[string]interface{}); ok {
		if needsAttention, ok := hrProblems["needs_attention"].(bool); ok && needsAttention {
			count++
		}
	}

	if criticalTasks, ok := dashboard["critical_tasks"].(map[string]interface{}); ok {
		if needsAttention, ok := criticalTasks["needs_attention"].(bool); ok && needsAttention {
			count++
		}
	}

	if projectHealth, ok := dashboard["project_health"].(map[string]interface{}); ok {
		if needsAttention, ok := projectHealth["needs_attention"].(bool); ok && needsAttention {
			count++
		}
	}

	if userActivity, ok := dashboard["user_activity"].(map[string]interface{}); ok {
		if needsAttention, ok := userActivity["needs_attention"].(bool); ok && needsAttention {
			count++
		}
	}

	return count
}

func (s *AdminService) getProjectHealthStatus(project models.Project, completionRate float64, isBehindSchedule bool) string {
	if project.Status == models.ProjectStatusCompleted {
		return "completed"
	}
	if isBehindSchedule {
		return "behind_schedule"
	}
	if completionRate < 30.0 {
		return "critical"
	}
	if completionRate < 50.0 {
		return "at_risk"
	}
	if completionRate < 70.0 {
		return "warning"
	}
	return "healthy"
}

// GetTodayChecklistStatus gets or creates today's checklist
func (s *AdminService) GetTodayChecklistStatus() (*models.AdminDailyChecklist, error) {
	today := time.Now()
	// Get date at midnight for consistent date comparison
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	var checklist models.AdminDailyChecklist
	err := s.DB.Where("date = ?", todayDate).First(&checklist).Error

	if err != nil {
		// Checklist doesn't exist for today, create it
		checklist = models.AdminDailyChecklist{
			Date: todayDate,
		}
		if err := s.DB.Create(&checklist).Error; err != nil {
			return nil, err
		}
	}

	return &checklist, nil
}

// UpdateChecklistItem updates a specific checklist item
func (s *AdminService) UpdateChecklistItem(itemName string, completed bool, userID uint, notes string) (*models.AdminDailyChecklist, error) {
	checklist, err := s.GetTodayChecklistStatus()
	if err != nil {
		return nil, err
	}

	// Update the specific item
	switch itemName {
	case "hr_problems":
		checklist.HRProblemsReviewed = completed
	case "critical_tasks":
		checklist.CriticalTasksReviewed = completed
	case "project_health":
		checklist.ProjectHealthReviewed = completed
	case "user_activity":
		checklist.UserActivityReviewed = completed
	case "system_alerts":
		checklist.SystemAlertsReviewed = completed
	case "ai_performance":
		checklist.AIPerformanceReviewed = completed
	case "backup_status":
		checklist.BackupStatusReviewed = completed
	case "security_logs":
		checklist.SecurityLogsReviewed = completed
	default:
		return nil, fmt.Errorf("invalid checklist item: %s", itemName)
	}

	// Update notes if provided
	if notes != "" {
		if checklist.Notes != "" {
			checklist.Notes += "\n" + notes
		} else {
			checklist.Notes = notes
		}
	}

	// If all items are completed, set completed_at and completed_by
	if checklist.IsCompleted() {
		now := time.Now()
		checklist.CompletedAt = &now
		checklist.CompletedBy = &userID
	}

	if err := s.DB.Save(&checklist).Error; err != nil {
		return nil, err
	}

	return checklist, nil
}

// GetChecklistHistory gets checklist history for a date range
func (s *AdminService) GetChecklistHistory(startDate, endDate time.Time, limit int) ([]models.AdminDailyChecklist, error) {
	if limit == 0 {
		limit = 30 // Default to 30 days
	}

	var checklists []models.AdminDailyChecklist
	query := s.DB.Where("date >= ? AND date <= ?", startDate, endDate).
		Order("date DESC").
		Limit(limit).
		Preload("CompletedByUser")

	if err := query.Find(&checklists).Error; err != nil {
		return nil, err
	}

	return checklists, nil
}

// GetChecklistStats gets statistics about checklist completion
func (s *AdminService) GetChecklistStats(days int) (map[string]interface{}, error) {
	if days == 0 {
		days = 30 // Default to 30 days
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	var totalChecklists, completedChecklists int64
	var avgCompletionRate float64

	// Count total checklists in period
	s.DB.Model(&models.AdminDailyChecklist{}).
		Where("date >= ? AND date <= ?", startDate, endDate).
		Count(&totalChecklists)

	// Count completed checklists
	s.DB.Model(&models.AdminDailyChecklist{}).
		Where("date >= ? AND date <= ? AND completed_at IS NOT NULL", startDate, endDate).
		Count(&completedChecklists)

	// Get all checklists to calculate average completion rate
	var checklists []models.AdminDailyChecklist
	s.DB.Where("date >= ? AND date <= ?", startDate, endDate).Find(&checklists)

	if len(checklists) > 0 {
		totalRate := 0.0
		for _, checklist := range checklists {
			totalRate += checklist.GetCompletionPercentage()
		}
		avgCompletionRate = totalRate / float64(len(checklists))
	}

	completionRate := 0.0
	if totalChecklists > 0 {
		completionRate = float64(completedChecklists) / float64(totalChecklists) * 100.0
	}

	return map[string]interface{}{
		"period_days":          days,
		"total_checklists":     totalChecklists,
		"completed_checklists": completedChecklists,
		"completion_rate":      completionRate,
		"avg_completion_rate":  avgCompletionRate,
		"start_date":           startDate,
		"end_date":             endDate,
	}, nil
}
