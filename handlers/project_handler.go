package handlers

import (
	"fmt"
	"net/http"
	"project-x/models"
	"project-x/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	DB *gorm.DB
}

func NewProjectHandler(db *gorm.DB) *ProjectHandler {
	return &ProjectHandler{DB: db}
}

// CreateProject creates a new project (Manager/Admin only)
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var createProjectRequest struct {
		Title       string     `json:"title" binding:"required"`
		Description string     `json:"description" binding:"required"`
		StartDate   time.Time  `json:"start_date" binding:"required"`
		EndDate     *time.Time `json:"end_date"`
	}

	if err := c.ShouldBindJSON(&createProjectRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user ID from context
	userID, _ := c.Get("userID")

	projectService := services.NewProjectService(h.DB)
	project, err := projectService.CreateProject(
		createProjectRequest.Title,
		createProjectRequest.Description,
		userID.(uint),
		createProjectRequest.StartDate,
		createProjectRequest.EndDate,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Project created successfully",
		"project": gin.H{
			"id":          project.ID,
			"title":       project.Title,
			"description": project.Description,
			"status":      project.Status,
			"created_by":  project.CreatedBy,
			"start_date":  project.StartDate,
			"end_date":    project.EndDate,
			"created_at":  project.CreatedAt,
		},
	})
}

// GetUserProjects returns all projects the current user is part of
func (h *ProjectHandler) GetUserProjects(c *gin.Context) {
	userID, _ := c.Get("userID")

	projectService := services.NewProjectService(h.DB)
	projects, err := projectService.GetUserProjects(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	var projectList []gin.H
	for _, project := range projects {
		projectList = append(projectList, gin.H{
			"id":          project.ID,
			"title":       project.Title,
			"description": project.Description,
			"status":      project.Status,
			"created_by":  project.CreatedBy,
			"start_date":  project.StartDate,
			"end_date":    project.EndDate,
			"created_at":  project.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"projects": projectList})
}

// GetProjectDetails returns detailed information about a specific project
func (h *ProjectHandler) GetProjectDetails(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	projectService := services.NewProjectService(h.DB)
	project, err := projectService.GetProjectWithDetails(uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project": gin.H{
			"id":          project.ID,
			"title":       project.Title,
			"description": project.Description,
			"status":      project.Status,
			"created_by":  project.CreatedBy,
			"start_date":  project.StartDate,
			"end_date":    project.EndDate,
			"created_at":  project.CreatedAt,
			"creator": gin.H{
				"id":       project.Creator.ID,
				"username": project.Creator.Username,
				"role":     project.Creator.Role,
			},
			"member_count": len(project.Users),
			"task_count":   len(project.Tasks) + len(project.CollaborativeTasks),
		},
	})
}

// GetProjectMembers returns all members of a project
func (h *ProjectHandler) GetProjectMembers(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	projectService := services.NewProjectService(h.DB)
	members, userProjects, err := projectService.GetProjectMembers(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project members"})
		return
	}

	// Create a map of userID to UserProject for quick lookup
	userProjectMap := make(map[uint]models.UserProject)
	for _, up := range userProjects {
		userProjectMap[up.UserID] = up
	}

	var memberList []gin.H
	for _, member := range members {
		userProject := userProjectMap[member.ID]
		memberList = append(memberList, gin.H{
			"id":           member.ID,
			"username":     member.Username,
			"role":         member.Role,
			"department":   member.Department,
			"skills":       member.Skills,
			"project_role": userProject.Role,    // Project management role
			"job_role":     userProject.JobRole, // Project-specific job role
		})
	}

	c.JSON(http.StatusOK, gin.H{"members": memberList})
}

// AddUserToProject adds a user to a project (Manager/Admin only)
func (h *ProjectHandler) AddUserToProject(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var addUserRequest struct {
		UserID  uint   `json:"user_id" binding:"required"`
		Role    string `json:"role" binding:"required"` // Project management role: manager, head, employee, member
		JobRole string `json:"job_role"`                // Project-specific job role: "UX/UI Designer", "Backend Developer", etc.
	}

	if err := c.ShouldBindJSON(&addUserRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projectService := services.NewProjectService(h.DB)
	err = projectService.AddUserToProject(addUserRequest.UserID, uint(projectID), addUserRequest.Role, addUserRequest.JobRole)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added to project successfully"})
}

// RemoveUserFromProject removes a user from a project (Manager/Admin only)
func (h *ProjectHandler) RemoveUserFromProject(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	projectService := services.NewProjectService(h.DB)
	err = projectService.RemoveUserFromProject(uint(userID), uint(projectID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed from project successfully"})
}

// UpdateProjectStatus updates project status (Manager/Admin only)
func (h *ProjectHandler) UpdateProjectStatus(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
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
	validStatuses := []string{"active", "paused", "completed", "cancelled"}
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

	projectService := services.NewProjectService(h.DB)
	err = projectService.UpdateProjectStatus(uint(projectID), models.ProjectStatus(updateRequest.Status))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project status updated successfully"})
}

// DeleteProject deletes a project (Admin only)
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	projectService := services.NewProjectService(h.DB)
	err = projectService.DeleteProject(uint(projectID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

// GetProjectStatistics returns statistics for a project
func (h *ProjectHandler) GetProjectStatistics(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get project details
	projectService := services.NewProjectService(h.DB)
	project, err := projectService.GetProjectWithDetails(uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Count tasks by status
	var pendingTasks, inProgressTasks, completedTasks, cancelledTasks int
	for _, task := range project.Tasks {
		switch task.Status {
		case models.TaskStatusPending:
			pendingTasks++
		case models.TaskStatusInProgress:
			inProgressTasks++
		case models.TaskStatusCompleted:
			completedTasks++
		case models.TaskStatusCancelled:
			cancelledTasks++
		}
	}

	for _, task := range project.CollaborativeTasks {
		switch task.Status {
		case models.TaskStatusPending:
			pendingTasks++
		case models.TaskStatusInProgress:
			inProgressTasks++
		case models.TaskStatusCompleted:
			completedTasks++
		case models.TaskStatusCancelled:
			cancelledTasks++
		}
	}

	totalTasks := len(project.Tasks) + len(project.CollaborativeTasks)
	completionRate := 0.0
	if totalTasks > 0 {
		completionRate = float64(completedTasks) / float64(totalTasks) * 100
	}

	stats := gin.H{
		"project_id":        project.ID,
		"project_title":     project.Title,
		"status":            project.Status,
		"member_count":      len(project.Users),
		"total_tasks":       totalTasks,
		"pending_tasks":     pendingTasks,
		"in_progress_tasks": inProgressTasks,
		"completed_tasks":   completedTasks,
		"cancelled_tasks":   cancelledTasks,
		"completion_rate":   completionRate,
		"start_date":        project.StartDate,
		"end_date":          project.EndDate,
	}

	c.JSON(http.StatusOK, gin.H{"statistics": stats})
}

// GenerateProjectTasksWithAI generates tasks for a project using AI
func (h *ProjectHandler) GenerateProjectTasksWithAI(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get project details
	projectService := services.NewProjectService(h.DB)
	project, err := projectService.GetProjectWithDetails(uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Get project members with their job roles
	members, userProjects, err := projectService.GetProjectMembers(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project members"})
		return
	}

	if len(members) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project has no team members. Please add team members first."})
		return
	}

	// Create a map of userID to UserProject for quick lookup
	userProjectMap := make(map[uint]models.UserProject)
	for _, up := range userProjects {
		userProjectMap[up.UserID] = up
	}

	// Build team member info
	var teamMembers []services.TeamMemberInfo
	for _, member := range members {
		userProject := userProjectMap[member.ID]
		teamMembers = append(teamMembers, services.TeamMemberInfo{
			UserID:     member.ID,
			Username:   member.Username,
			JobRole:    userProject.JobRole,
			Role:       userProject.Role,
			Department: member.Department,
			Skills:     member.Skills,
		})
	}

	// Initialize AI task generator
	taskService := services.NewTaskService(h.DB)
	aiGenerator := services.NewAIProjectTaskGenerator(h.DB, taskService)

	// Prepare generation request
	req := services.ProjectTaskGenerationRequest{
		ProjectID:    project.ID,
		ProjectTitle: project.Title,
		Description:  project.Description,
		TeamMembers:  teamMembers,
		StartDate:    project.StartDate,
		EndDate:      project.EndDate,
	}

	// Generate tasks
	generationResponse, err := aiGenerator.GenerateProjectTasks(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate tasks: %v", err)})
		return
	}

	// Always return preview - user must confirm to create
	c.JSON(http.StatusOK, gin.H{
		"message":         "Tasks generated successfully. Review and edit before confirming.",
		"summary":         generationResponse.Summary,
		"generated_tasks": generationResponse.Tasks,
		"next_step":       "Edit tasks if needed, then call POST /api/projects/:id/confirm-tasks to create them",
		"instructions": gin.H{
			"review":  "Review all generated tasks below",
			"edit":    "Edit any task fields (start_time, end_time, assigned_to_user_id, priority, etc.)",
			"confirm": "Call POST /api/projects/:id/confirm-tasks with the edited tasks array to create them",
		},
	})
}

// ConfirmAndCreateProjectTasks creates tasks after review and editing
func (h *ProjectHandler) ConfirmAndCreateProjectTasks(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get project to verify it exists
	projectService := services.NewProjectService(h.DB)
	project, err := projectService.GetProjectWithDetails(uint(projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var confirmRequest struct {
		Tasks []services.GeneratedTask `json:"tasks" binding:"required"` // Edited tasks
	}

	if err := c.ShouldBindJSON(&confirmRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(confirmRequest.Tasks) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No tasks provided to create"})
		return
	}

	// Validate that all tasks have required fields
	for i, task := range confirmRequest.Tasks {
		if task.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Task %d: title is required", i+1)})
			return
		}
		if task.AssignedToID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Task %d: assigned_to_user_id is required", i+1)})
			return
		}
		// Verify assigned user exists and is in the project
		var userProject models.UserProject
		if err := h.DB.Where("project_id = ? AND user_id = ?", projectID, task.AssignedToID).First(&userProject).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Task %d: assigned user (ID: %d) is not a member of this project", i+1, task.AssignedToID)})
			return
		}
	}

	// Initialize AI task generator
	taskService := services.NewTaskService(h.DB)
	aiGenerator := services.NewAIProjectTaskGenerator(h.DB, taskService)

	// Create tasks from the edited/confirmed tasks
	createdTasks, err := aiGenerator.CreateTasksFromGeneration(project.ID, confirmRequest.Tasks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create tasks: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "Tasks created successfully",
		"created_tasks_count": len(createdTasks),
		"tasks":               createdTasks,
	})
}
