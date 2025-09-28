package handlers

import (
	"net/http"
	"project-x/config"
	"project-x/models"
	"project-x/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskHandler struct {
	DB                  *gorm.DB
	TaskService         *services.TaskService
	NotificationService *services.NotificationService
	WorkSchedule        *config.WorkScheduleConfig
	AIOptimizer         *services.AITimeOptimizer
}

func NewTaskHandler(db *gorm.DB) *TaskHandler {
	taskService := services.NewTaskService(db)
	wsService := services.NewWebSocketService(db)
	notificationService := services.NewNotificationService(db, wsService)

	// Set the notification service in the task service
	taskService.SetNotificationService(notificationService)

	return &TaskHandler{
		DB:                  db,
		TaskService:         taskService,
		NotificationService: notificationService,
		WorkSchedule:        config.GetDefaultWorkSchedule(),
		AIOptimizer:         services.NewAITimeOptimizer(db),
	}
}

// CreateTask creates a new task
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var createTaskRequest struct {
		Title       string     `json:"title" binding:"required"`
		Description string     `json:"description" binding:"required"`
		ProjectID   *uint      `json:"project_id"`
		DueDate     *time.Time `json:"due_date"`
		AssignedTo  *uint      `json:"assigned_to"` // Optional: assign to specific user (Admin/Manager only)
		StartTime   *time.Time `json:"start_time"`
		EndTime     *time.Time `json:"end_time"`
	}

	if err := c.ShouldBindJSON(&createTaskRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user info from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	// Determine who the task should be assigned to
	var assignedUserID uint
	if createTaskRequest.AssignedTo != nil {
		// Check if current user has permission to assign tasks to others
		if userRole == models.RoleEmployee || userRole == models.RoleHead || userRole == models.RoleHR {
			c.JSON(http.StatusForbidden, gin.H{"error": "Employees, Heads, and HR cannot assign tasks to other users"})
			return
		}
		assignedUserID = *createTaskRequest.AssignedTo
	} else {
		// Assign to current user
		assignedUserID = userID.(uint)
	}

	task, err := h.TaskService.CreateTask(
		createTaskRequest.Title,
		createTaskRequest.Description,
		assignedUserID,
		createTaskRequest.ProjectID,
		createTaskRequest.StartTime,
		createTaskRequest.EndTime,
		createTaskRequest.DueDate,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send notification if task is assigned to someone else
	if assignedUserID != userID.(uint) {
		// Get user details for notification
		var assignedUser, currentUser models.User
		if err := h.DB.First(&assignedUser, assignedUserID).Error; err == nil {
			if err := h.DB.First(&currentUser, userID.(uint)).Error; err == nil {
				h.NotificationService.SendTaskAssignedNotification(task, &assignedUser, &currentUser)
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Task created successfully",
		"task": gin.H{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"user_id":     task.UserID,
			"project_id":  task.ProjectID,
			"assigned_at": task.AssignedAt,
			"due_date":    task.DueDate,
			"created_at":  task.CreatedAt,
		},
	})
}

// CreateTaskForUser allows Admin/Manager to create tasks for specific users
func (h *TaskHandler) CreateTaskForUser(c *gin.Context) {
	var createTaskRequest struct {
		Title       string     `json:"title" binding:"required"`
		Description string     `json:"description" binding:"required"`
		UserID      uint       `json:"user_id" binding:"required"`
		ProjectID   *uint      `json:"project_id"`
		DueDate     *time.Time `json:"due_date"`
		Priority    string     `json:"priority"` // high, medium, low
	}

	if err := c.ShouldBindJSON(&createTaskRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user info from context
	currentUserID, _ := c.Get("userID")

	// Validate priority
	if createTaskRequest.Priority != "" {
		validPriorities := []string{"high", "medium", "low"}
		validPriority := false
		for _, p := range validPriorities {
			if p == createTaskRequest.Priority {
				validPriority = true
				break
			}
		}
		if !validPriority {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid priority. Must be 'high', 'medium', or 'low'"})
			return
		}
	}

	task, err := h.TaskService.CreateTaskForUser(
		createTaskRequest.Title,
		createTaskRequest.Description,
		createTaskRequest.UserID,
		currentUserID.(uint),
		createTaskRequest.ProjectID,
		createTaskRequest.DueDate,
		createTaskRequest.Priority,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send notification for task assignment
	var assignedUser, currentUser models.User
	if err := h.DB.First(&assignedUser, createTaskRequest.UserID).Error; err == nil {
		if err := h.DB.First(&currentUser, currentUserID.(uint)).Error; err == nil {
			h.NotificationService.SendTaskAssignedNotification(task, &assignedUser, &currentUser)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Task assigned successfully",
		"task": gin.H{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"user_id":     task.UserID,
			"assigned_by": currentUserID,
			"project_id":  task.ProjectID,
			"assigned_at": task.AssignedAt,
			"due_date":    task.DueDate,
			"created_at":  task.CreatedAt,
		},
	})
}

// CreateCollaborativeTask creates a new collaborative task
func (h *TaskHandler) CreateCollaborativeTask(c *gin.Context) {
	var createTaskRequest struct {
		Title           string     `json:"title" binding:"required"`
		Description     string     `json:"description" binding:"required"`
		ProjectID       *uint      `json:"project_id"`
		StartTime       *time.Time `json:"start_time"`
		EndTime         *time.Time `json:"end_time"`
		DueDate         *time.Time `json:"due_date"`
		LeadUserID      *uint      `json:"lead_user_id"`     // Explicit lead selection
		ParticipantIDs  []uint     `json:"participant_ids"`  // Other team members
		MaxParticipants int        `json:"max_participants"` // Maximum participants (default 5)
		Priority        string     `json:"priority"`         // high, medium, low
		Complexity      string     `json:"complexity"`       // simple, medium, complex
		AssignedTo      *uint      `json:"assigned_to"`      // Legacy: Optional: assign to specific user (Admin/Manager only)
	}

	if err := c.ShouldBindJSON(&createTaskRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user info from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	// Determine who the task should be assigned to
	var assignedUserID uint
	if createTaskRequest.AssignedTo != nil {
		// Check if current user has permission to assign tasks to others
		if userRole == models.RoleEmployee || userRole == models.RoleHead || userRole == models.RoleHR {
			c.JSON(http.StatusForbidden, gin.H{"error": "Employees, Heads, and HR cannot assign collaborative tasks to other users"})
			return
		}
		assignedUserID = *createTaskRequest.AssignedTo
	} else {
		// Assign to current user
		assignedUserID = userID.(uint)
	}

	leadUserID := createTaskRequest.LeadUserID
	if leadUserID == nil {
		uid := userID.(uint)
		leadUserID = &uid
	}
	task, err := h.TaskService.CreateCollaborativeTask(
		createTaskRequest.Title,
		createTaskRequest.Description,
		*leadUserID,
		createTaskRequest.ProjectID,
		createTaskRequest.StartTime,
		createTaskRequest.EndTime,
		createTaskRequest.DueDate,
		createTaskRequest.ParticipantIDs,
		createTaskRequest.MaxParticipants,
		createTaskRequest.Priority,
		createTaskRequest.Complexity,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send notification if task is assigned to someone else
	if assignedUserID != userID.(uint) {
		var assignedUser, currentUser models.User
		if err := h.DB.First(&assignedUser, assignedUserID).Error; err == nil {
			if err := h.DB.First(&currentUser, userID.(uint)).Error; err == nil {
				h.NotificationService.SendCollaborativeTaskAssignedNotification(task, &assignedUser, &currentUser)
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Collaborative task created successfully",
		"task": gin.H{
			"id":           task.ID,
			"title":        task.Title,
			"description":  task.Description,
			"status":       task.Status,
			"lead_user_id": task.LeadUserID,
			"project_id":   task.ProjectID,
			"assigned_at":  task.AssignedAt,
			"due_date":     task.DueDate,
			"created_at":   task.CreatedAt,
		},
	})
}

// CreateCollaborativeTaskForUser allows Admin/Manager to create collaborative tasks for specific users
func (h *TaskHandler) CreateCollaborativeTaskForUser(c *gin.Context) {
	var createTaskRequest struct {
		Title           string     `json:"title" binding:"required"`
		Description     string     `json:"description" binding:"required"`
		UserID          uint       `json:"user_id" binding:"required"` // Lead user
		ProjectID       *uint      `json:"project_id"`
		StartTime       *time.Time `json:"start_time"`
		EndTime         *time.Time `json:"end_time"`
		DueDate         *time.Time `json:"due_date"`
		ParticipantIDs  []uint     `json:"participant_ids"`  // Other team members
		MaxParticipants int        `json:"max_participants"` // Maximum participants (default 5)
		Priority        string     `json:"priority"`         // high, medium, low
		Complexity      string     `json:"complexity"`       // simple, medium, complex
	}

	if err := c.ShouldBindJSON(&createTaskRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user info from context
	currentUserID, _ := c.Get("userID")

	// Validate priority
	if createTaskRequest.Priority != "" {
		validPriorities := []string{"high", "medium", "low"}
		validPriority := false
		for _, p := range validPriorities {
			if p == createTaskRequest.Priority {
				validPriority = true
				break
			}
		}
		if !validPriority {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid priority. Must be 'high', 'medium', or 'low'"})
			return
		}
	}

	task, err := h.TaskService.CreateCollaborativeTaskForUser(
		createTaskRequest.Title,
		createTaskRequest.Description,
		createTaskRequest.UserID, // Lead user
		currentUserID.(uint),
		createTaskRequest.ProjectID,
		createTaskRequest.StartTime,
		createTaskRequest.EndTime,
		createTaskRequest.DueDate,
		createTaskRequest.ParticipantIDs,
		createTaskRequest.MaxParticipants,
		createTaskRequest.Priority,
		createTaskRequest.Complexity,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send notification for collaborative task assignment
	var assignedUser, currentUser models.User
	if err := h.DB.First(&assignedUser, createTaskRequest.UserID).Error; err == nil {
		if err := h.DB.First(&currentUser, currentUserID.(uint)).Error; err == nil {
			h.NotificationService.SendCollaborativeTaskAssignedNotification(task, &assignedUser, &currentUser)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Collaborative task assigned successfully",
		"task": gin.H{
			"id":           task.ID,
			"title":        task.Title,
			"description":  task.Description,
			"status":       task.Status,
			"lead_user_id": task.LeadUserID,
			"assigned_by":  currentUserID,
			"project_id":   task.ProjectID,
			"assigned_at":  task.AssignedAt,
			"due_date":     task.DueDate,
			"created_at":   task.CreatedAt,
		},
	})
}

// GetUserTasks returns all tasks for the current user
func (h *TaskHandler) GetUserTasks(c *gin.Context) {
	userID, _ := c.Get("userID")

	taskService := services.NewTaskService(h.DB)
	tasks, err := taskService.GetUserTasks(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	var taskList []gin.H
	for _, task := range tasks {
		taskList = append(taskList, gin.H{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"project_id":  task.ProjectID,
			"assigned_at": task.AssignedAt,
			"due_date":    task.DueDate,
			"created_at":  task.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"tasks": taskList})
}

// GetUserCollaborativeTasks returns all collaborative tasks for the current user
func (h *TaskHandler) GetUserCollaborativeTasks(c *gin.Context) {
	userID, _ := c.Get("userID")

	taskService := services.NewTaskService(h.DB)
	tasks, err := taskService.GetUserCollaborativeTasks(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch collaborative tasks"})
		return
	}

	var taskList []gin.H
	for _, task := range tasks {
		taskList = append(taskList, gin.H{
			"id":           task.ID,
			"title":        task.Title,
			"description":  task.Description,
			"status":       task.Status,
			"lead_user_id": task.LeadUserID,
			"project_id":   task.ProjectID,
			"assigned_at":  task.AssignedAt,
			"due_date":     task.DueDate,
			"created_at":   task.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"collaborative_tasks": taskList})
}

// GetProjectTasks returns all tasks for a specific project (Admin or project member)
func (h *TaskHandler) GetProjectTasks(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	taskService := services.NewTaskService(h.DB)
	tasks, err := taskService.GetProjectTasks(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project tasks"})
		return
	}

	var taskList []gin.H
	for _, task := range tasks {
		taskList = append(taskList, gin.H{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"user_id":     task.UserID,
			"assigned_at": task.AssignedAt,
			"due_date":    task.DueDate,
			"created_at":  task.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"tasks": taskList})
}

// GetProjectCollaborativeTasks returns all collaborative tasks for a specific project
func (h *TaskHandler) GetProjectCollaborativeTasks(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	taskService := services.NewTaskService(h.DB)
	tasks, err := taskService.GetProjectCollaborativeTasks(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project collaborative tasks"})
		return
	}

	var taskList []gin.H
	for _, task := range tasks {
		taskList = append(taskList, gin.H{
			"id":             task.ID,
			"title":          task.Title,
			"description":    task.Description,
			"status":         task.Status,
			"lead_user_id":   task.LeadUserID,
			"lead_user_name": task.LeadUser.Username,
			"project_id":     task.ProjectID,
			"assigned_at":    task.AssignedAt,
			"due_date":       task.DueDate,
			"created_at":     task.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"collaborative_tasks": taskList})
}

// UpdateTaskStatus updates task status
func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var updateRequest struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	validStatuses := []string{"pending", "in_progress", "completed", "cancelled"}
	validStatus := false
	for _, status := range validStatuses {
		if status == updateRequest.Status {
			validStatus = true
			break
		}
	}
	if !validStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	// Get current task to check old status and ownership
	var task models.Task
	if err := h.DB.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check if current user is the task assignee
	currentUserID, _ := c.Get("userID")

	// Only allow the task assignee to update status
	if task.UserID != currentUserID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the task assignee can update task status"})
		return
	}

	oldStatus := task.Status

	err = h.TaskService.UpdateTaskStatus(uint(taskID), models.TaskStatus(updateRequest.Status))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task status"})
		return
	}

	// Send notification for status change
	var currentUser models.User
	if err := h.DB.First(&currentUser, currentUserID.(uint)).Error; err == nil {
		// Get updated task
		var updatedTask models.Task
		if err := h.DB.First(&updatedTask, taskID).Error; err == nil {
			h.NotificationService.SendTaskStatusChangedNotification(&updatedTask, &currentUser, oldStatus, models.TaskStatus(updateRequest.Status))

			// Send completion notification if task is completed
			if updateRequest.Status == "completed" {
				h.NotificationService.SendTaskCompletedNotification(&updatedTask, &currentUser)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task status updated successfully"})
}

// UpdateCollaborativeTaskStatus updates collaborative task status
func (h *TaskHandler) UpdateCollaborativeTaskStatus(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var updateRequest struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	validStatuses := []string{"pending", "in_progress", "completed", "cancelled"}
	validStatus := false
	for _, status := range validStatuses {
		if status == updateRequest.Status {
			validStatus = true
			break
		}
	}
	if !validStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	// Get current collaborative task to check ownership
	var collaborativeTask models.CollaborativeTask
	if err := h.DB.First(&collaborativeTask, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collaborative task not found"})
		return
	}

	// Check if current user is the lead user
	currentUserID, _ := c.Get("userID")

	// Only allow the lead user to update status
	if collaborativeTask.LeadUserID != currentUserID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the lead user can update collaborative task status"})
		return
	}

	err = h.TaskService.UpdateCollaborativeTaskStatus(uint(taskID), models.TaskStatus(updateRequest.Status))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update collaborative task status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collaborative task status updated successfully"})
}

// GetTasksByStatus returns tasks filtered by status for current user
func (h *TaskHandler) GetTasksByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status parameter is required"})
		return
	}

	userID, _ := c.Get("userID")

	taskService := services.NewTaskService(h.DB)
	tasks, err := taskService.GetTasksByStatus(userID.(uint), models.TaskStatus(status))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	var taskList []gin.H
	for _, task := range tasks {
		taskList = append(taskList, gin.H{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"project_id":  task.ProjectID,
			"assigned_at": task.AssignedAt,
			"due_date":    task.DueDate,
			"created_at":  task.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"tasks": taskList})
}

// GetCollaborativeTasksByStatus returns collaborative tasks filtered by status for current user
func (h *TaskHandler) GetCollaborativeTasksByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status parameter is required"})
		return
	}

	userID, _ := c.Get("userID")

	taskService := services.NewTaskService(h.DB)
	tasks, err := taskService.GetCollaborativeTasksByStatus(userID.(uint), models.TaskStatus(status))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch collaborative tasks"})
		return
	}

	var taskList []gin.H
	for _, task := range tasks {
		taskList = append(taskList, gin.H{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"project_id":  task.ProjectID,
			"assigned_at": task.AssignedAt,
			"due_date":    task.DueDate,
			"created_at":  task.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"collaborative_tasks": taskList})
}

// DeleteTask deletes a task (only by owner)
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	userID, _ := c.Get("userID")

	taskService := services.NewTaskService(h.DB)
	err = taskService.DeleteTask(uint(taskID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// DeleteCollaborativeTask deletes a collaborative task (only by owner)
func (h *TaskHandler) DeleteCollaborativeTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	userID, _ := c.Get("userID")

	taskService := services.NewTaskService(h.DB)
	err = taskService.DeleteCollaborativeTask(uint(taskID), userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collaborative task deleted successfully"})
}

// GetTasksByDepartment returns all tasks for users in a specific department (Head/Admin only)
func (h *TaskHandler) GetTasksByDepartment(c *gin.Context) {
	department := c.Param("dept")
	if department == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Department parameter is required"})
		return
	}

	taskService := services.NewTaskService(h.DB)
	tasks, err := taskService.GetTasksByDepartment(department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch department tasks"})
		return
	}

	var taskList []gin.H
	for _, task := range tasks {
		taskList = append(taskList, gin.H{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"user_id":     task.UserID,
			"user_name":   task.User.Username,
			"project_id":  task.ProjectID,
			"assigned_at": task.AssignedAt,
			"due_date":    task.DueDate,
			"created_at":  task.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"tasks": taskList})
}

// GetCollaborativeTasksByDepartment returns all collaborative tasks for users in a specific department (Head/Admin only)
func (h *TaskHandler) GetCollaborativeTasksByDepartment(c *gin.Context) {
	department := c.Param("dept")
	if department == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Department parameter is required"})
		return
	}

	taskService := services.NewTaskService(h.DB)
	tasks, err := taskService.GetCollaborativeTasksByDepartment(department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch department collaborative tasks"})
		return
	}

	var taskList []gin.H
	for _, task := range tasks {
		taskList = append(taskList, gin.H{
			"id":             task.ID,
			"title":          task.Title,
			"description":    task.Description,
			"status":         task.Status,
			"lead_user_id":   task.LeadUserID,
			"lead_user_name": task.LeadUser.Username,
			"project_id":     task.ProjectID,
			"assigned_at":    task.AssignedAt,
			"due_date":       task.DueDate,
			"created_at":     task.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"collaborative_tasks": taskList})
}

// BulkUpdateTaskStatus allows Manager/Admin to update multiple task statuses at once
func (h *TaskHandler) BulkUpdateTaskStatus(c *gin.Context) {
	var bulkUpdateRequest struct {
		TaskIDs []uint            `json:"task_ids" binding:"required"`
		Status  models.TaskStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&bulkUpdateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	validStatuses := []models.TaskStatus{
		models.TaskStatusPending,
		models.TaskStatusInProgress,
		models.TaskStatusCompleted,
		models.TaskStatusCancelled,
	}
	validStatus := false
	for _, status := range validStatuses {
		if status == bulkUpdateRequest.Status {
			validStatus = true
			break
		}
	}
	if !validStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	taskService := services.NewTaskService(h.DB)
	updatedCount, err := taskService.BulkUpdateTaskStatus(bulkUpdateRequest.TaskIDs, bulkUpdateRequest.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Tasks updated successfully",
		"updated_count":   updatedCount,
		"total_requested": len(bulkUpdateRequest.TaskIDs),
	})
}

// GetTaskStatistics returns task statistics for managers and admins
func (h *TaskHandler) GetTaskStatistics(c *gin.Context) {
	taskService := services.NewTaskService(h.DB)
	stats, err := taskService.GetTaskStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statistics": stats})
}

// GetProjectReport returns detailed project report with user performance
func (h *TaskHandler) GetProjectReport(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	period := c.Query("period")
	if period == "" {
		period = "weekly" // Default to weekly
	}

	if period != "weekly" && period != "monthly" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period. Use 'weekly' or 'monthly'"})
		return
	}

	taskService := services.NewTaskService(h.DB)
	report, err := taskService.GetProjectReport(uint(projectID), period)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report})
}

// GetUserReport returns detailed user report with project performance
func (h *TaskHandler) GetUserReport(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	period := c.Query("period")
	if period == "" {
		period = "weekly" // Default to weekly
	}

	if period != "weekly" && period != "monthly" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period. Use 'weekly' or 'monthly'"})
		return
	}

	taskService := services.NewTaskService(h.DB)
	report, err := taskService.GetUserReport(uint(userID), period)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report})
}

// UpdateTask updates task details (title, description, due date, etc.)
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var updateRequest struct {
		Title       string     `json:"title"`
		Description string     `json:"description"`
		DueDate     *time.Time `json:"due_date"`
		EndTime     *time.Time `json:"end_time"`
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user info from context
	currentUserID, _ := c.Get("userID")
	currentUserRole, _ := c.Get("userRole")

	// Get current task to check ownership and store old values for notification
	var task models.Task
	if err := h.DB.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Store old values for notification comparison
	oldTitle := task.Title
	oldDescription := task.Description
	oldDueDate := task.DueDate

	// Check permissions: Only Admin/Manager can update task fields
	// Task assignees can only update status, not other fields
	// HR cannot update task details
	if currentUserRole != models.RoleAdmin && currentUserRole != models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only Admin/Manager can update task details. Task assignees can only update status."})
		return
	}

	// Update the task
	updatedTask, err := h.TaskService.UpdateTask(
		uint(taskID),
		currentUserID.(uint),
		updateRequest.Title,
		updateRequest.Description,
		updateRequest.DueDate,
		nil, // StartTime is immutable - cannot be updated
		updateRequest.EndTime,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send notification to task assignee about the update
	var currentUser, assignedUser models.User
	if err := h.DB.First(&currentUser, currentUserID.(uint)).Error; err == nil {
		if err := h.DB.First(&assignedUser, task.UserID).Error; err == nil {
			h.NotificationService.SendTaskUpdatedNotification(updatedTask, &assignedUser, &currentUser, oldTitle, oldDescription, oldDueDate)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task updated successfully",
		"task": gin.H{
			"id":          updatedTask.ID,
			"title":       updatedTask.Title,
			"description": updatedTask.Description,
			"status":      updatedTask.Status,
			"user_id":     updatedTask.UserID,
			"project_id":  updatedTask.ProjectID,
			"assigned_at": updatedTask.AssignedAt,
			"start_time":  updatedTask.StartTime,
			"end_time":    updatedTask.EndTime,
			"due_date":    updatedTask.DueDate,
			"updated_at":  updatedTask.UpdatedAt,
		},
	})
}

// UpdateCollaborativeTask updates collaborative task details
func (h *TaskHandler) UpdateCollaborativeTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var updateRequest struct {
		Title       string     `json:"title"`
		Description string     `json:"description"`
		DueDate     *time.Time `json:"due_date"`
		EndTime     *time.Time `json:"end_time"`
		Priority    string     `json:"priority"`   // high, medium, low
		Complexity  string     `json:"complexity"` // simple, medium, complex
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate priority if provided
	if updateRequest.Priority != "" {
		validPriorities := []string{"high", "medium", "low"}
		validPriority := false
		for _, p := range validPriorities {
			if p == updateRequest.Priority {
				validPriority = true
				break
			}
		}
		if !validPriority {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid priority. Must be 'high', 'medium', or 'low'"})
			return
		}
	}

	// Validate complexity if provided
	if updateRequest.Complexity != "" {
		validComplexities := []string{"simple", "medium", "complex"}
		validComplexity := false
		for _, c := range validComplexities {
			if c == updateRequest.Complexity {
				validComplexity = true
				break
			}
		}
		if !validComplexity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid complexity. Must be 'simple', 'medium', or 'complex'"})
			return
		}
	}

	// Get current user info from context
	currentUserID, _ := c.Get("userID")
	currentUserRole, _ := c.Get("userRole")

	// Get current collaborative task to check ownership and store old values for notification
	var collaborativeTask models.CollaborativeTask
	if err := h.DB.First(&collaborativeTask, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collaborative task not found"})
		return
	}

	// Store old values for notification comparison
	oldTitle := collaborativeTask.Title
	oldDescription := collaborativeTask.Description
	oldDueDate := collaborativeTask.DueDate

	// Check permissions: Only Admin/Manager can update collaborative task fields
	// Lead users can only update status, not other fields
	// HR cannot update collaborative task details
	if currentUserRole != models.RoleAdmin && currentUserRole != models.RoleManager {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only Admin/Manager can update collaborative task details. Lead users can only update status."})
		return
	}

	// Update the collaborative task
	updatedTask, err := h.TaskService.UpdateCollaborativeTask(
		uint(taskID),
		currentUserID.(uint),
		updateRequest.Title,
		updateRequest.Description,
		updateRequest.DueDate,
		nil, // StartTime is immutable - cannot be updated
		updateRequest.EndTime,
		updateRequest.Priority,
		updateRequest.Complexity,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send notification to lead user about the update
	var currentUser, leadUser models.User
	if err := h.DB.First(&currentUser, currentUserID.(uint)).Error; err == nil {
		if err := h.DB.First(&leadUser, collaborativeTask.LeadUserID).Error; err == nil {
			h.NotificationService.SendCollaborativeTaskUpdatedNotification(updatedTask, &leadUser, &currentUser, oldTitle, oldDescription, oldDueDate)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Collaborative task updated successfully",
		"task": gin.H{
			"id":           updatedTask.ID,
			"title":        updatedTask.Title,
			"description":  updatedTask.Description,
			"status":       updatedTask.Status,
			"lead_user_id": updatedTask.LeadUserID,
			"project_id":   updatedTask.ProjectID,
			"assigned_at":  updatedTask.AssignedAt,
			"start_time":   updatedTask.StartTime,
			"end_time":     updatedTask.EndTime,
			"due_date":     updatedTask.DueDate,
			"priority":     updatedTask.Priority,
			"complexity":   updatedTask.Complexity,
			"updated_at":   updatedTask.UpdatedAt,
		},
	})
}

// SmartTaskHandler handles all task operations with query parameters
func (h *TaskHandler) SmartTaskHandler(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	// Get query parameters
	action := c.Query("action") // "create", "update", "delete", "bulk_update", "query" (default)
	taskType := c.Query("type") // "regular", "collaborative", "all"
	view := c.Query("view")     // "my_tasks", "project", "department", "all"
	status := c.Query("status") // Filter by status
	projectID := c.Query("project_id")
	department := c.Query("department")
	assignedTo := c.Query("assigned_to")
	period := c.Query("period") // For reports: "weekly", "monthly"
	sort := c.Query("sort")     // "due_date:desc", "created_at:asc", etc.
	page := c.Query("page")
	limit := c.Query("limit")

	// Default action is "query" if not specified
	if action == "" {
		action = "query"
	}

	switch action {
	case "create":
		h.handleSmartTaskCreation(c, userID.(uint), userRole.(string))
	case "update":
		h.handleSmartTaskUpdate(c, userID.(uint), userRole.(string))
	case "delete":
		h.handleSmartTaskDelete(c, userID.(uint), userRole.(string))
	case "bulk_update":
		h.handleSmartBulkUpdate(c, userID.(uint), userRole.(string))
	case "query":
		h.handleSmartTaskQuery(c, userID.(uint), userRole.(string), map[string]string{
			"type":        taskType,
			"view":        view,
			"status":      status,
			"project_id":  projectID,
			"department":  department,
			"assigned_to": assignedTo,
			"period":      period,
			"sort":        sort,
			"page":        page,
			"limit":       limit,
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
	}
}

func (h *TaskHandler) handleSmartTaskCreation(c *gin.Context, userID uint, userRole string) {
	var createRequest struct {
		Type            string     `json:"type" binding:"required"` // "regular" or "collaborative"
		Title           string     `json:"title" binding:"required"`
		Description     string     `json:"description" binding:"required"`
		ProjectID       *uint      `json:"project_id"`
		DueDate         *time.Time `json:"due_date"`
		AssignedTo      *uint      `json:"assigned_to"`      // For regular tasks
		LeadUserID      *uint      `json:"lead_user_id"`     // For collaborative tasks
		ParticipantIDs  []uint     `json:"participant_ids"`  // For collaborative tasks
		MaxParticipants int        `json:"max_participants"` // For collaborative tasks
		Priority        string     `json:"priority"`         // high, medium, low
		Complexity      string     `json:"complexity"`       // simple, medium, complex
		StartTime       *time.Time `json:"start_time"`
		EndTime         *time.Time `json:"end_time"`
	}

	if err := c.ShouldBindJSON(&createRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate task type
	if createRequest.Type != "regular" && createRequest.Type != "collaborative" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task type. Must be 'regular' or 'collaborative'"})
		return
	}

	if createRequest.Type == "regular" {
		// Handle regular task creation
		var assignedUserID uint
		if createRequest.AssignedTo != nil {
			// Check permissions for assigning tasks to others
			if userRole == string(models.RoleEmployee) || userRole == string(models.RoleHead) || userRole == string(models.RoleHR) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Employees, Heads, and HR cannot assign tasks to other users"})
				return
			}
			assignedUserID = *createRequest.AssignedTo
		} else {
			assignedUserID = userID
		}

		task, err := h.TaskService.CreateTask(
			createRequest.Title,
			createRequest.Description,
			assignedUserID,
			createRequest.ProjectID,
			createRequest.StartTime,
			createRequest.EndTime,
			createRequest.DueDate,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Send notification if task is assigned to someone else
		if assignedUserID != userID {
			var assignedUser, currentUser models.User
			if err := h.DB.First(&assignedUser, assignedUserID).Error; err == nil {
				if err := h.DB.First(&currentUser, userID).Error; err == nil {
					h.NotificationService.SendTaskAssignedNotification(task, &assignedUser, &currentUser)
				}
			}
		}

		c.JSON(http.StatusCreated, gin.H{
			"action": "created",
			"type":   "regular",
			"task": gin.H{
				"id":          task.ID,
				"title":       task.Title,
				"description": task.Description,
				"status":      task.Status,
				"user_id":     task.UserID,
				"project_id":  task.ProjectID,
				"assigned_at": task.AssignedAt,
				"due_date":    task.DueDate,
				"created_at":  task.CreatedAt,
			},
		})
	} else {
		// Handle collaborative task creation
		leadUserID := createRequest.LeadUserID
		if leadUserID == nil {
			leadUserID = &userID
		}

		task, err := h.TaskService.CreateCollaborativeTask(
			createRequest.Title,
			createRequest.Description,
			*leadUserID,
			createRequest.ProjectID,
			createRequest.StartTime,
			createRequest.EndTime,
			createRequest.DueDate,
			createRequest.ParticipantIDs,
			createRequest.MaxParticipants,
			createRequest.Priority,
			createRequest.Complexity,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"action": "created",
			"type":   "collaborative",
			"task": gin.H{
				"id":           task.ID,
				"title":        task.Title,
				"description":  task.Description,
				"status":       task.Status,
				"lead_user_id": task.LeadUserID,
				"project_id":   task.ProjectID,
				"assigned_at":  task.AssignedAt,
				"due_date":     task.DueDate,
				"created_at":   task.CreatedAt,
			},
		})
	}
}

func (h *TaskHandler) handleSmartTaskQuery(c *gin.Context, userID uint, userRole string, params map[string]string) {
	taskService := services.NewTaskService(h.DB)
	result := make(gin.H)

	// Handle reports (Manager+, HR, or Admin only)
	if params["period"] != "" {
		if userRole == string(models.RoleEmployee) || userRole == string(models.RoleHead) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only Manager+, HR, or Admin can access reports"})
			return
		}

		if params["period"] != "weekly" && params["period"] != "monthly" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period. Use 'weekly' or 'monthly'"})
			return
		}

		// Handle project reports
		if params["project_id"] != "" {
			projectIDUint, err := strconv.ParseUint(params["project_id"], 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
				return
			}

			report, err := taskService.GetProjectReport(uint(projectIDUint), params["period"])
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"view_type": "project_report",
				"period":    params["period"],
				"report":    report,
			})
			return
		}

		// Handle user reports
		if params["assigned_to"] != "" {
			userIDUint, err := strconv.ParseUint(params["assigned_to"], 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
				return
			}

			report, err := taskService.GetUserReport(uint(userIDUint), params["period"])
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"view_type": "user_report",
				"period":    params["period"],
				"report":    report,
			})
			return
		}
	}

	// Handle department-based queries (Manager+, HR, or Admin only)
	if params["department"] != "" {
		if userRole == string(models.RoleEmployee) || userRole == string(models.RoleHead) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only Manager+, HR, or Admin can access department tasks"})
			return
		}

		if params["type"] == "regular" || params["type"] == "all" || params["type"] == "" {
			tasks, err := taskService.GetTasksByDepartment(params["department"])
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch department tasks"})
				return
			}

			var taskList []gin.H
			for _, task := range tasks {
				taskList = append(taskList, gin.H{
					"id":          task.ID,
					"title":       task.Title,
					"description": task.Description,
					"status":      task.Status,
					"user_id":     task.UserID,
					"user_name":   task.User.Username,
					"project_id":  task.ProjectID,
					"assigned_at": task.AssignedAt,
					"due_date":    task.DueDate,
					"created_at":  task.CreatedAt,
				})
			}
			result["regular_tasks"] = taskList
		}

		if params["type"] == "collaborative" || params["type"] == "all" || params["type"] == "" {
			tasks, err := taskService.GetCollaborativeTasksByDepartment(params["department"])
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch department collaborative tasks"})
				return
			}

			var taskList []gin.H
			for _, task := range tasks {
				taskList = append(taskList, gin.H{
					"id":             task.ID,
					"title":          task.Title,
					"description":    task.Description,
					"status":         task.Status,
					"lead_user_id":   task.LeadUserID,
					"lead_user_name": task.LeadUser.Username,
					"project_id":     task.ProjectID,
					"assigned_at":    task.AssignedAt,
					"due_date":       task.DueDate,
					"created_at":     task.CreatedAt,
				})
			}
			result["collaborative_tasks"] = taskList
		}

		result["view_type"] = "department"
		result["department"] = params["department"]
		c.JSON(http.StatusOK, result)
		return
	}

	// Handle project-based queries
	if params["project_id"] != "" {
		projectIDUint, err := strconv.ParseUint(params["project_id"], 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}

		if params["type"] == "regular" || params["type"] == "all" || params["type"] == "" {
			tasks, err := taskService.GetProjectTasks(uint(projectIDUint))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project tasks"})
				return
			}

			var taskList []gin.H
			for _, task := range tasks {
				taskList = append(taskList, gin.H{
					"id":          task.ID,
					"title":       task.Title,
					"description": task.Description,
					"status":      task.Status,
					"user_id":     task.UserID,
					"assigned_at": task.AssignedAt,
					"due_date":    task.DueDate,
					"created_at":  task.CreatedAt,
				})
			}
			result["regular_tasks"] = taskList
		}

		if params["type"] == "collaborative" || params["type"] == "all" || params["type"] == "" {
			tasks, err := taskService.GetProjectCollaborativeTasks(uint(projectIDUint))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project collaborative tasks"})
				return
			}

			var taskList []gin.H
			for _, task := range tasks {
				taskList = append(taskList, gin.H{
					"id":             task.ID,
					"title":          task.Title,
					"description":    task.Description,
					"status":         task.Status,
					"lead_user_id":   task.LeadUserID,
					"lead_user_name": task.LeadUser.Username,
					"project_id":     task.ProjectID,
					"assigned_at":    task.AssignedAt,
					"due_date":       task.DueDate,
					"created_at":     task.CreatedAt,
				})
			}
			result["collaborative_tasks"] = taskList
		}

		result["view_type"] = "project"
		result["project_id"] = params["project_id"]
		c.JSON(http.StatusOK, result)
		return
	}

	// Handle user's own tasks (default view)
	if params["type"] == "regular" || params["type"] == "all" || params["type"] == "" {
		var tasks []models.Task
		var err error

		if params["status"] != "" {
			tasks, err = taskService.GetTasksByStatus(userID, models.TaskStatus(params["status"]))
		} else {
			tasks, err = taskService.GetUserTasks(userID)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
			return
		}

		var taskList []gin.H
		for _, task := range tasks {
			taskList = append(taskList, gin.H{
				"id":          task.ID,
				"title":       task.Title,
				"description": task.Description,
				"status":      task.Status,
				"project_id":  task.ProjectID,
				"assigned_at": task.AssignedAt,
				"due_date":    task.DueDate,
				"created_at":  task.CreatedAt,
			})
		}
		result["regular_tasks"] = taskList
	}

	if params["type"] == "collaborative" || params["type"] == "all" || params["type"] == "" {
		var tasks []models.CollaborativeTask
		var err error

		if params["status"] != "" {
			tasks, err = taskService.GetCollaborativeTasksByStatus(userID, models.TaskStatus(params["status"]))
		} else {
			tasks, err = taskService.GetUserCollaborativeTasks(userID)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch collaborative tasks"})
			return
		}

		var taskList []gin.H
		for _, task := range tasks {
			taskList = append(taskList, gin.H{
				"id":           task.ID,
				"title":        task.Title,
				"description":  task.Description,
				"status":       task.Status,
				"lead_user_id": task.LeadUserID,
				"project_id":   task.ProjectID,
				"assigned_at":  task.AssignedAt,
				"due_date":     task.DueDate,
				"created_at":   task.CreatedAt,
			})
		}
		result["collaborative_tasks"] = taskList
	}

	result["view_type"] = "my_tasks"
	if params["status"] != "" {
		result["status"] = params["status"]
	}
	c.JSON(http.StatusOK, result)
}

func (h *TaskHandler) handleSmartTaskUpdate(c *gin.Context, userID uint, userRole string) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	taskType := c.Query("type") // "regular" or "collaborative"
	if taskType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task type is required. Use ?type=regular or ?type=collaborative"})
		return
	}

	if taskType == "regular" {
		var updateRequest struct {
			Title       string     `json:"title"`
			Description string     `json:"description"`
			DueDate     *time.Time `json:"due_date"`
			StartTime   *time.Time `json:"start_time"`
			EndTime     *time.Time `json:"end_time"`
		}

		if err := c.ShouldBindJSON(&updateRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get current task to check ownership and store old values for notification
		var task models.Task
		if err := h.DB.First(&task, taskID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}

		// Store old values for notification comparison
		oldTitle := task.Title
		oldDescription := task.Description
		oldDueDate := task.DueDate

		// Check permissions
		if task.UserID != userID {
			if userRole != string(models.RoleAdmin) && userRole != string(models.RoleManager) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Only the task assignee or Admin/Manager can update task details"})
				return
			}
		}

		// Update the task
		updatedTask, err := h.TaskService.UpdateTask(
			uint(taskID),
			userID,
			updateRequest.Title,
			updateRequest.Description,
			updateRequest.DueDate,
			updateRequest.StartTime,
			updateRequest.EndTime,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Send notification to task assignee about the update
		var currentUser, assignedUser models.User
		if err := h.DB.First(&currentUser, userID).Error; err == nil {
			if err := h.DB.First(&assignedUser, task.UserID).Error; err == nil {
				h.NotificationService.SendTaskUpdatedNotification(updatedTask, &assignedUser, &currentUser, oldTitle, oldDescription, oldDueDate)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"action": "updated",
			"type":   "regular",
			"task": gin.H{
				"id":          updatedTask.ID,
				"title":       updatedTask.Title,
				"description": updatedTask.Description,
				"status":      updatedTask.Status,
				"user_id":     updatedTask.UserID,
				"project_id":  updatedTask.ProjectID,
				"assigned_at": updatedTask.AssignedAt,
				"start_time":  updatedTask.StartTime,
				"end_time":    updatedTask.EndTime,
				"due_date":    updatedTask.DueDate,
				"updated_at":  updatedTask.UpdatedAt,
			},
		})
	} else {
		var updateRequest struct {
			Title       string     `json:"title"`
			Description string     `json:"description"`
			DueDate     *time.Time `json:"due_date"`
			StartTime   *time.Time `json:"start_time"`
			EndTime     *time.Time `json:"end_time"`
			Priority    string     `json:"priority"`   // high, medium, low
			Complexity  string     `json:"complexity"` // simple, medium, complex
		}

		if err := c.ShouldBindJSON(&updateRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get current collaborative task to check ownership and store old values for notification
		var collaborativeTask models.CollaborativeTask
		if err := h.DB.First(&collaborativeTask, taskID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collaborative task not found"})
			return
		}

		// Store old values for notification comparison
		oldTitle := collaborativeTask.Title
		oldDescription := collaborativeTask.Description
		oldDueDate := collaborativeTask.DueDate

		// Check permissions
		if collaborativeTask.LeadUserID != userID {
			if userRole != string(models.RoleAdmin) && userRole != string(models.RoleManager) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Only the lead user or Admin/Manager can update collaborative task details"})
				return
			}
		}

		// Update the collaborative task
		updatedTask, err := h.TaskService.UpdateCollaborativeTask(
			uint(taskID),
			userID,
			updateRequest.Title,
			updateRequest.Description,
			updateRequest.DueDate,
			updateRequest.StartTime,
			updateRequest.EndTime,
			updateRequest.Priority,
			updateRequest.Complexity,
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Send notification to lead user about the update
		var currentUser, leadUser models.User
		if err := h.DB.First(&currentUser, userID).Error; err == nil {
			if err := h.DB.First(&leadUser, collaborativeTask.LeadUserID).Error; err == nil {
				h.NotificationService.SendCollaborativeTaskUpdatedNotification(updatedTask, &leadUser, &currentUser, oldTitle, oldDescription, oldDueDate)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"action": "updated",
			"type":   "collaborative",
			"task": gin.H{
				"id":           updatedTask.ID,
				"title":        updatedTask.Title,
				"description":  updatedTask.Description,
				"status":       updatedTask.Status,
				"lead_user_id": updatedTask.LeadUserID,
				"project_id":   updatedTask.ProjectID,
				"assigned_at":  updatedTask.AssignedAt,
				"start_time":   updatedTask.StartTime,
				"end_time":     updatedTask.EndTime,
				"due_date":     updatedTask.DueDate,
				"priority":     updatedTask.Priority,
				"complexity":   updatedTask.Complexity,
				"updated_at":   updatedTask.UpdatedAt,
			},
		})
	}
}

func (h *TaskHandler) handleSmartTaskDelete(c *gin.Context, userID uint, _ string) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	taskType := c.Query("type") // "regular" or "collaborative"
	if taskType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task type is required. Use ?type=regular or ?type=collaborative"})
		return
	}

	var err2 error
	if taskType == "collaborative" {
		err2 = h.TaskService.DeleteCollaborativeTask(uint(taskID), userID)
	} else {
		err2 = h.TaskService.DeleteTask(uint(taskID), userID)
	}

	if err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err2.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"action": "deleted",
		"type":   taskType,
		"id":     taskID,
	})
}

func (h *TaskHandler) handleSmartBulkUpdate(c *gin.Context, currentUserID uint, userRole string) {
	var bulkUpdateRequest struct {
		TaskIDs     []uint     `json:"task_ids" binding:"required"`
		TaskType    string     `json:"task_type" binding:"required"` // "regular" or "collaborative"
		Title       *string    `json:"title"`                        // Optional: update title
		Description *string    `json:"description"`                  // Optional: update description
		Status      *string    `json:"status"`                       // Optional: update status
		DueDate     *time.Time `json:"due_date"`                     // Optional: update due date
		EndTime     *time.Time `json:"end_time"`                     // Optional: update end time
		Priority    *string    `json:"priority"`                     // Optional: update priority (for collaborative tasks)
		Complexity  *string    `json:"complexity"`                   // Optional: update complexity (for collaborative tasks)
		AssignedTo  *uint      `json:"assigned_to"`                  // Optional: reassign regular tasks
		LeadUserID  *uint      `json:"lead_user_id"`                 // Optional: change lead for collaborative tasks
		ProjectID   *uint      `json:"project_id"`                   // Optional: change project
	}

	if err := c.ShouldBindJSON(&bulkUpdateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate task type
	if bulkUpdateRequest.TaskType != "regular" && bulkUpdateRequest.TaskType != "collaborative" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task type. Must be 'regular' or 'collaborative'"})
		return
	}

	// Validate status if provided
	if bulkUpdateRequest.Status != nil {
		validStatuses := []string{"pending", "in_progress", "completed", "cancelled"}
		validStatus := false
		for _, status := range validStatuses {
			if status == *bulkUpdateRequest.Status {
				validStatus = true
				break
			}
		}
		if !validStatus {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
			return
		}
	}

	// Validate priority if provided
	if bulkUpdateRequest.Priority != nil {
		validPriorities := []string{"high", "medium", "low"}
		validPriority := false
		for _, p := range validPriorities {
			if p == *bulkUpdateRequest.Priority {
				validPriority = true
				break
			}
		}
		if !validPriority {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid priority. Must be 'high', 'medium', or 'low'"})
			return
		}
	}

	// Validate complexity if provided
	if bulkUpdateRequest.Complexity != nil {
		validComplexities := []string{"simple", "medium", "complex"}
		validComplexity := false
		for _, c := range validComplexities {
			if c == *bulkUpdateRequest.Complexity {
				validComplexity = true
				break
			}
		}
		if !validComplexity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid complexity. Must be 'simple', 'medium', or 'complex'"})
			return
		}
	}

	// Check permissions for bulk operations
	if userRole == string(models.RoleEmployee) || userRole == string(models.RoleHead) || userRole == string(models.RoleHR) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only Manager+ can perform bulk operations"})
		return
	}

	// If only status is being updated, use the existing bulk status update
	if bulkUpdateRequest.Status != nil &&
		bulkUpdateRequest.Title == nil &&
		bulkUpdateRequest.Description == nil &&
		bulkUpdateRequest.DueDate == nil &&
		bulkUpdateRequest.EndTime == nil &&
		bulkUpdateRequest.Priority == nil &&
		bulkUpdateRequest.Complexity == nil &&
		bulkUpdateRequest.AssignedTo == nil &&
		bulkUpdateRequest.LeadUserID == nil &&
		bulkUpdateRequest.ProjectID == nil {

		updatedCount, err := h.TaskService.BulkUpdateTaskStatus(bulkUpdateRequest.TaskIDs, models.TaskStatus(*bulkUpdateRequest.Status))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"action":        "bulk_status_updated",
			"type":          bulkUpdateRequest.TaskType,
			"updated_count": updatedCount,
			"total":         len(bulkUpdateRequest.TaskIDs),
		})
		return
	}

	// For comprehensive bulk updates, update each task individually with proper permission checks
	updatedCount := 0
	failedTasks := []uint{}
	permissionDeniedTasks := []uint{}

	for _, taskID := range bulkUpdateRequest.TaskIDs {
		if bulkUpdateRequest.TaskType == "regular" {
			// Check permissions for regular task
			var task models.Task
			if err := h.DB.First(&task, taskID).Error; err != nil {
				failedTasks = append(failedTasks, taskID)
				continue
			}

			// Store old values for notification comparison
			oldTitle := task.Title
			oldDescription := task.Description
			oldDueDate := task.DueDate

			// Check if current user can update this task
			if task.UserID != currentUserID {
				// Only Admin/Manager can update tasks assigned to others
				if userRole != string(models.RoleAdmin) && userRole != string(models.RoleManager) {
					permissionDeniedTasks = append(permissionDeniedTasks, taskID)
					continue
				}
			}

			// Update regular task
			title := ""
			description := ""
			if bulkUpdateRequest.Title != nil {
				title = *bulkUpdateRequest.Title
			}
			if bulkUpdateRequest.Description != nil {
				description = *bulkUpdateRequest.Description
			}

			updatedTask, err := h.TaskService.UpdateTask(
				taskID,
				currentUserID,
				title,
				description,
				bulkUpdateRequest.DueDate,
				nil, // StartTime is immutable - cannot be updated
				bulkUpdateRequest.EndTime,
			)
			if err != nil {
				failedTasks = append(failedTasks, taskID)
				continue
			}

			// Send notification for task detail updates if any details were changed
			if bulkUpdateRequest.Title != nil || bulkUpdateRequest.Description != nil || bulkUpdateRequest.DueDate != nil {
				var currentUser, assignedUser models.User
				if err := h.DB.First(&currentUser, currentUserID).Error; err == nil {
					if err := h.DB.First(&assignedUser, task.UserID).Error; err == nil {
						h.NotificationService.SendTaskUpdatedNotification(updatedTask, &assignedUser, &currentUser, oldTitle, oldDescription, oldDueDate)
					}
				}
			}

			// Update status separately if provided
			if bulkUpdateRequest.Status != nil {
				// Only the task assignee can update status (Admin/Manager cannot override this)
				if task.UserID != currentUserID {
					permissionDeniedTasks = append(permissionDeniedTasks, taskID)
					continue
				}
				if err := h.TaskService.UpdateTaskStatus(taskID, models.TaskStatus(*bulkUpdateRequest.Status)); err != nil {
					failedTasks = append(failedTasks, taskID)
					continue
				}
			}

			// Update assignment if provided (Admin/Manager only)
			if bulkUpdateRequest.AssignedTo != nil {
				if userRole != string(models.RoleAdmin) && userRole != string(models.RoleManager) {
					permissionDeniedTasks = append(permissionDeniedTasks, taskID)
					continue
				}
				if err := h.DB.Model(&models.Task{}).Where("id = ?", taskID).Update("user_id", *bulkUpdateRequest.AssignedTo).Error; err != nil {
					failedTasks = append(failedTasks, taskID)
					continue
				}
			}

			// Update project if provided (Admin/Manager only)
			if bulkUpdateRequest.ProjectID != nil {
				if userRole != string(models.RoleAdmin) && userRole != string(models.RoleManager) {
					permissionDeniedTasks = append(permissionDeniedTasks, taskID)
					continue
				}
				if err := h.DB.Model(&models.Task{}).Where("id = ?", taskID).Update("project_id", *bulkUpdateRequest.ProjectID).Error; err != nil {
					failedTasks = append(failedTasks, taskID)
					continue
				}
			}
		} else {
			// Check permissions for collaborative task
			var collaborativeTask models.CollaborativeTask
			if err := h.DB.First(&collaborativeTask, taskID).Error; err != nil {
				failedTasks = append(failedTasks, taskID)
				continue
			}

			// Store old values for notification comparison
			oldTitle := collaborativeTask.Title
			oldDescription := collaborativeTask.Description
			oldDueDate := collaborativeTask.DueDate

			// Check if current user can update this collaborative task
			if collaborativeTask.LeadUserID != currentUserID {
				// Only Admin/Manager can update collaborative tasks led by others
				if userRole != string(models.RoleAdmin) && userRole != string(models.RoleManager) {
					permissionDeniedTasks = append(permissionDeniedTasks, taskID)
					continue
				}
			}

			// Update collaborative task
			title := ""
			description := ""
			priority := ""
			complexity := ""
			if bulkUpdateRequest.Title != nil {
				title = *bulkUpdateRequest.Title
			}
			if bulkUpdateRequest.Description != nil {
				description = *bulkUpdateRequest.Description
			}
			if bulkUpdateRequest.Priority != nil {
				priority = *bulkUpdateRequest.Priority
			}
			if bulkUpdateRequest.Complexity != nil {
				complexity = *bulkUpdateRequest.Complexity
			}

			updatedTask, err := h.TaskService.UpdateCollaborativeTask(
				taskID,
				currentUserID,
				title,
				description,
				bulkUpdateRequest.DueDate,
				nil, // StartTime is immutable - cannot be updated
				bulkUpdateRequest.EndTime,
				priority,
				complexity,
			)
			if err != nil {
				failedTasks = append(failedTasks, taskID)
				continue
			}

			// Send notification for collaborative task detail updates if any details were changed
			if bulkUpdateRequest.Title != nil || bulkUpdateRequest.Description != nil || bulkUpdateRequest.DueDate != nil || bulkUpdateRequest.Priority != nil || bulkUpdateRequest.Complexity != nil {
				var currentUser, leadUser models.User
				if err := h.DB.First(&currentUser, currentUserID).Error; err == nil {
					if err := h.DB.First(&leadUser, collaborativeTask.LeadUserID).Error; err == nil {
						h.NotificationService.SendCollaborativeTaskUpdatedNotification(updatedTask, &leadUser, &currentUser, oldTitle, oldDescription, oldDueDate)
					}
				}
			}

			// Update status separately if provided
			if bulkUpdateRequest.Status != nil {
				// Only the lead user can update status (Admin/Manager cannot override this)
				if collaborativeTask.LeadUserID != currentUserID {
					permissionDeniedTasks = append(permissionDeniedTasks, taskID)
					continue
				}
				if err := h.TaskService.UpdateCollaborativeTaskStatus(taskID, models.TaskStatus(*bulkUpdateRequest.Status)); err != nil {
					failedTasks = append(failedTasks, taskID)
					continue
				}
			}

			// Update lead user if provided (Admin/Manager only)
			if bulkUpdateRequest.LeadUserID != nil {
				if userRole != string(models.RoleAdmin) && userRole != string(models.RoleManager) {
					permissionDeniedTasks = append(permissionDeniedTasks, taskID)
					continue
				}
				if err := h.DB.Model(&models.CollaborativeTask{}).Where("id = ?", taskID).Update("lead_user_id", *bulkUpdateRequest.LeadUserID).Error; err != nil {
					failedTasks = append(failedTasks, taskID)
					continue
				}
			}

			// Update project if provided (Admin/Manager only)
			if bulkUpdateRequest.ProjectID != nil {
				if userRole != string(models.RoleAdmin) && userRole != string(models.RoleManager) {
					permissionDeniedTasks = append(permissionDeniedTasks, taskID)
					continue
				}
				if err := h.DB.Model(&models.CollaborativeTask{}).Where("id = ?", taskID).Update("project_id", *bulkUpdateRequest.ProjectID).Error; err != nil {
					failedTasks = append(failedTasks, taskID)
					continue
				}
			}
		}

		updatedCount++
	}

	// Send notifications for reassignments
	if bulkUpdateRequest.AssignedTo != nil && bulkUpdateRequest.TaskType == "regular" {
		var assignedUser, currentUser models.User
		if err := h.DB.First(&assignedUser, *bulkUpdateRequest.AssignedTo).Error; err == nil {
			if err := h.DB.First(&currentUser, currentUserID).Error; err == nil {
				// Send notifications for each reassigned task
				for _, taskID := range bulkUpdateRequest.TaskIDs {
					var task models.Task
					if err := h.DB.First(&task, taskID).Error; err == nil {
						h.NotificationService.SendTaskAssignedNotification(&task, &assignedUser, &currentUser)
					}
				}
			}
		}
	}

	if bulkUpdateRequest.LeadUserID != nil && bulkUpdateRequest.TaskType == "collaborative" {
		var leadUser, currentUser models.User
		if err := h.DB.First(&leadUser, *bulkUpdateRequest.LeadUserID).Error; err == nil {
			if err := h.DB.First(&currentUser, currentUserID).Error; err == nil {
				// Send notifications for each reassigned collaborative task
				for _, taskID := range bulkUpdateRequest.TaskIDs {
					var task models.CollaborativeTask
					if err := h.DB.First(&task, taskID).Error; err == nil {
						h.NotificationService.SendCollaborativeTaskAssignedNotification(&task, &leadUser, &currentUser)
					}
				}
			}
		}
	}

	response := gin.H{
		"action":        "bulk_updated",
		"type":          bulkUpdateRequest.TaskType,
		"updated_count": updatedCount,
		"total":         len(bulkUpdateRequest.TaskIDs),
		"updated_fields": gin.H{
			"title":        bulkUpdateRequest.Title != nil,
			"description":  bulkUpdateRequest.Description != nil,
			"status":       bulkUpdateRequest.Status != nil,
			"due_date":     bulkUpdateRequest.DueDate != nil,
			"end_time":     bulkUpdateRequest.EndTime != nil,
			"priority":     bulkUpdateRequest.Priority != nil,
			"complexity":   bulkUpdateRequest.Complexity != nil,
			"assigned_to":  bulkUpdateRequest.AssignedTo != nil,
			"lead_user_id": bulkUpdateRequest.LeadUserID != nil,
			"project_id":   bulkUpdateRequest.ProjectID != nil,
		},
	}

	if len(failedTasks) > 0 {
		response["failed_tasks"] = failedTasks
	}

	if len(permissionDeniedTasks) > 0 {
		response["permission_denied_tasks"] = permissionDeniedTasks
	}

	c.JSON(http.StatusOK, response)
}

// GetTasksWithArabicTimeContext returns tasks with Arabic working hours context
func (h *TaskHandler) GetTasksWithArabicTimeContext(c *gin.Context) {
	userID, _ := c.Get("userID")

	// Get user's tasks
	tasks, err := h.TaskService.GetUserTasks(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	// Enhance tasks with Arabic working hours context
	var enhancedTasks []gin.H
	now := time.Now()

	for _, task := range tasks {
		taskInfo := gin.H{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"project_id":  task.ProjectID,
			"assigned_at": task.AssignedAt,
			"due_date":    task.DueDate,
			"created_at":  task.CreatedAt,
		}

		// Add Arabic working hours context
		if task.DueDate != nil {
			workingHoursUntilDeadline := h.WorkSchedule.GetWorkingHoursInPeriod(now, *task.DueDate)
			workingDaysUntilDeadline := h.WorkSchedule.GetWorkingDaysUntil(now, *task.DueDate)
			deadlineRisk := h.WorkSchedule.CalculateDeadlineRisk(*task.DueDate, 8.0) // Assume 8 hours default

			taskInfo["arabic_time_context"] = gin.H{
				"working_hours_until_deadline": workingHoursUntilDeadline,
				"working_days_until_deadline":  workingDaysUntilDeadline,
				"deadline_risk":                deadlineRisk,
				"is_overdue":                   task.DueDate.Before(now),
				"optimal_start_time":           h.WorkSchedule.GetOptimalTaskSchedule(8.0, *task.DueDate),
			}
		}

		// Add current working status
		taskInfo["current_working_status"] = gin.H{
			"is_working_hour": h.WorkSchedule.IsWorkingHour(now),
			"is_working_day":  h.WorkSchedule.IsWorkingDay(now),
			"current_time":    now,
		}

		enhancedTasks = append(enhancedTasks, taskInfo)
	}

	response := gin.H{
		"tasks": enhancedTasks,
		"working_schedule": gin.H{
			"working_days":    []string{"Saturday", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday"},
			"working_hours":   "9:00 AM - 4:00 PM",
			"weekly_capacity": h.WorkSchedule.WeeklyHours,
			"lunch_break":     "12:00 PM - 1:00 PM",
			"timezone":        h.WorkSchedule.TimeZone,
		},
		"summary": gin.H{
			"total_tasks":      len(enhancedTasks),
			"current_time":     now,
			"is_working_hour":  h.WorkSchedule.IsWorkingHour(now),
			"is_working_day":   h.WorkSchedule.IsWorkingDay(now),
			"next_working_day": h.WorkSchedule.GetNextWorkingDay(now),
		},
	}

	c.JSON(http.StatusOK, response)
}

// CreateTaskWithArabicScheduling creates a task with Arabic working hours optimization
func (h *TaskHandler) CreateTaskWithArabicScheduling(c *gin.Context) {
	var createTaskRequest struct {
		Title               string     `json:"title" binding:"required"`
		Description         string     `json:"description" binding:"required"`
		ProjectID           *uint      `json:"project_id"`
		DueDate             *time.Time `json:"due_date"`
		AssignedTo          *uint      `json:"assigned_to"`
		EstimatedHours      *float64   `json:"estimated_hours"`       // Optional: user can provide estimate
		UseAIOptimization   bool       `json:"use_ai_optimization"`   // Whether to use AI for scheduling
		RespectWorkingHours bool       `json:"respect_working_hours"` // Default true
		Priority            string     `json:"priority"`              // high, medium, low
	}

	if err := c.ShouldBindJSON(&createTaskRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user info from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	// Determine who the task should be assigned to
	var assignedUserID uint
	if createTaskRequest.AssignedTo != nil {
		// Check if current user has permission to assign tasks to others
		if userRole == models.RoleEmployee || userRole == models.RoleHead || userRole == models.RoleHR {
			c.JSON(http.StatusForbidden, gin.H{"error": "Employees, Heads, and HR cannot assign tasks to other users"})
			return
		}
		assignedUserID = *createTaskRequest.AssignedTo
	} else {
		assignedUserID = userID.(uint)
	}

	// Calculate optimal timing based on Arabic working schedule
	now := time.Now()
	var optimalStartTime, optimalEndTime *time.Time
	var aiRecommendations []string
	var estimatedHours float64 = 8.0 // Default

	if createTaskRequest.EstimatedHours != nil {
		estimatedHours = *createTaskRequest.EstimatedHours
	}

	// Determine optimal start time
	if createTaskRequest.RespectWorkingHours {
		optimalStart := h.WorkSchedule.GetOptimalTaskSchedule(estimatedHours, *createTaskRequest.DueDate)
		optimalStartTime = &optimalStart

		// Calculate end time based on working hours
		workingDaysNeeded := int(estimatedHours / h.WorkSchedule.WorkingHours)
		if estimatedHours > float64(workingDaysNeeded)*h.WorkSchedule.WorkingHours {
			workingDaysNeeded++ // Round up
		}

		endTime := optimalStart
		for i := 0; i < workingDaysNeeded; i++ {
			endTime = h.WorkSchedule.GetNextWorkingDay(endTime)
		}
		optimalEndTime = &endTime

		// Generate recommendations
		aiRecommendations = []string{
			"ابدأ العمل في بداية اليوم لأفضل تركيز",
			"خصص وقت للمراجعة قبل نهاية اليوم",
			"تجنب العمل على المهام المعقدة بعد 3 عصراً",
		}

		if createTaskRequest.DueDate != nil {
			workingHoursUntilDeadline := h.WorkSchedule.GetWorkingHoursInPeriod(now, *createTaskRequest.DueDate)
			deadlineRisk := h.WorkSchedule.CalculateDeadlineRisk(*createTaskRequest.DueDate, estimatedHours)

			if deadlineRisk == "high" || deadlineRisk == "critical" {
				aiRecommendations = append(aiRecommendations,
					"المهمة تحتاج اهتمام عاجل - ابدأ فوراً",
					"فكر في طلب مساعدة إضافية",
					"قلل من المهام الأخرى للتركيز على هذه المهمة",
				)
			}

			if workingHoursUntilDeadline < estimatedHours {
				aiRecommendations = append(aiRecommendations,
					"الوقت المتاح أقل من المطلوب - أعد تقييم الموعد النهائي",
					"فكر في تقسيم المهمة لأجزاء أصغر",
				)
			}
		}
	}

	// Create the task
	task, err := h.TaskService.CreateTask(
		createTaskRequest.Title,
		createTaskRequest.Description,
		assignedUserID,
		createTaskRequest.ProjectID,
		optimalStartTime,
		optimalEndTime,
		createTaskRequest.DueDate,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send notification if task is assigned to someone else
	if assignedUserID != userID.(uint) {
		var assignedUser, currentUser models.User
		if err := h.DB.First(&assignedUser, assignedUserID).Error; err == nil {
			if err := h.DB.First(&currentUser, userID.(uint)).Error; err == nil {
				h.NotificationService.SendTaskAssignedNotification(task, &assignedUser, &currentUser)
			}
		}
	}

	// Prepare response with Arabic working hours context
	response := gin.H{
		"message": "Task created successfully with Arabic working hours optimization",
		"task": gin.H{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"user_id":     task.UserID,
			"project_id":  task.ProjectID,
			"assigned_at": task.AssignedAt,
			"due_date":    task.DueDate,
			"start_time":  optimalStartTime,
			"end_time":    optimalEndTime,
			"created_at":  task.CreatedAt,
		},
		"arabic_optimization": gin.H{
			"estimated_hours":     estimatedHours,
			"optimal_start_time":  optimalStartTime,
			"optimal_end_time":    optimalEndTime,
			"ai_recommendations":  aiRecommendations,
			"working_days_needed": int(estimatedHours / h.WorkSchedule.WorkingHours),
		},
		"working_schedule": gin.H{
			"working_days":    []string{"Saturday", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday"},
			"working_hours":   "9:00 AM - 4:00 PM",
			"weekly_capacity": h.WorkSchedule.WeeklyHours,
			"lunch_break":     "12:00 PM - 1:00 PM",
			"timezone":        h.WorkSchedule.TimeZone,
		},
	}

	if createTaskRequest.DueDate != nil {
		workingHoursUntilDeadline := h.WorkSchedule.GetWorkingHoursInPeriod(now, *createTaskRequest.DueDate)
		workingDaysUntilDeadline := h.WorkSchedule.GetWorkingDaysUntil(now, *createTaskRequest.DueDate)
		deadlineRisk := h.WorkSchedule.CalculateDeadlineRisk(*createTaskRequest.DueDate, estimatedHours)

		response["deadline_analysis"] = gin.H{
			"working_hours_until_deadline": workingHoursUntilDeadline,
			"working_days_until_deadline":  workingDaysUntilDeadline,
			"deadline_risk":                deadlineRisk,
			"feasible":                     workingHoursUntilDeadline >= estimatedHours,
		}
	}

	c.JSON(http.StatusCreated, response)
}

// GetTaskTimeAnalysis returns AI-powered time analysis for a specific task
func (h *TaskHandler) GetTaskTimeAnalysis(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	// Get the task
	var task models.Task
	if err := h.DB.Preload("User").Preload("Project").First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Check permissions - user can only analyze their own tasks unless they're Manager+/HR/Admin
	if task.UserID != userID.(uint) {
		if userRole == models.RoleEmployee || userRole == models.RoleHead {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only analyze your own tasks"})
			return
		}
	}

	// Get AI analysis if available
	var aiAnalysis *services.TimeAnalysis
	if h.AIOptimizer != nil {
		// Try to get AI analysis for this specific task
		allAnalyses, err := h.AIOptimizer.AnalyzeTaskTimeRisks()
		if err == nil {
			for _, analysis := range allAnalyses {
				if analysis.TaskID == uint(taskID) {
					aiAnalysis = &analysis
					break
				}
			}
		}
	}

	// Calculate basic working hours context
	now := time.Now()
	var workingHoursContext gin.H

	if task.DueDate != nil {
		workingHoursUntilDeadline := h.WorkSchedule.GetWorkingHoursInPeriod(now, *task.DueDate)
		workingDaysUntilDeadline := h.WorkSchedule.GetWorkingDaysUntil(now, *task.DueDate)
		deadlineRisk := h.WorkSchedule.CalculateDeadlineRisk(*task.DueDate, 8.0) // Default 8 hours

		workingHoursContext = gin.H{
			"working_hours_until_deadline": workingHoursUntilDeadline,
			"working_days_until_deadline":  workingDaysUntilDeadline,
			"deadline_risk":                deadlineRisk,
			"is_overdue":                   task.DueDate.Before(now),
			"optimal_start_time":           h.WorkSchedule.GetOptimalTaskSchedule(8.0, *task.DueDate),
		}
	}

	response := gin.H{
		"task": gin.H{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"status":      task.Status,
			"user_id":     task.UserID,
			"username":    task.User.Username,
			"project_id":  task.ProjectID,
			"due_date":    task.DueDate,
			"created_at":  task.CreatedAt,
		},
		"working_hours_context": workingHoursContext,
		"current_status": gin.H{
			"is_working_hour": h.WorkSchedule.IsWorkingHour(now),
			"is_working_day":  h.WorkSchedule.IsWorkingDay(now),
			"current_time":    now,
		},
		"working_schedule": gin.H{
			"working_days":    []string{"Saturday", "Sunday", "Monday", "Tuesday", "Wednesday", "Thursday"},
			"working_hours":   "9:00 AM - 4:00 PM",
			"weekly_capacity": h.WorkSchedule.WeeklyHours,
			"lunch_break":     "12:00 PM - 1:00 PM",
			"timezone":        h.WorkSchedule.TimeZone,
		},
	}

	// Add AI analysis if available
	if aiAnalysis != nil {
		response["ai_analysis"] = gin.H{
			"estimated_duration_hours":     aiAnalysis.EstimatedDuration,
			"deadline_risk":                aiAnalysis.DeadlineRisk,
			"risk_factors":                 aiAnalysis.RiskFactors,
			"recommendations":              aiAnalysis.Recommendations,
			"optimal_start_date":           aiAnalysis.OptimalStartDate,
			"predicted_completion":         aiAnalysis.PredictedCompletion,
			"working_hours_until_deadline": aiAnalysis.WorkingHoursUntilDeadline,
			"is_within_working_hours":      aiAnalysis.IsWithinWorkingHours,
		}
	} else {
		response["ai_analysis"] = gin.H{
			"status": "AI analysis not available",
			"reason": "AI service may be disabled or task data insufficient",
		}
	}

	c.JSON(http.StatusOK, response)
}
