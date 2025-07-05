package handlers

import (
	"net/http"
	"project-x/models"
	"project-x/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HRProblemHandler struct {
	DB               *gorm.DB
	HRProblemService *services.HRProblemService
}

func NewHRProblemHandler(db *gorm.DB, hrProblemService *services.HRProblemService) *HRProblemHandler {
	return &HRProblemHandler{
		DB:               db,
		HRProblemService: hrProblemService,
	}
}

// GetProblemCategories returns all available problem categories
func (h *HRProblemHandler) GetProblemCategories(c *gin.Context) {
	categories := models.GetProblemCategories()
	priorities := models.GetProblemPriorities()

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
		"priorities": priorities,
	})
}

// CreateProblem allows users to report a new problem
func (h *HRProblemHandler) CreateProblem(c *gin.Context) {
	var request struct {
		Title           string                 `json:"title" binding:"required"`
		Description     string                 `json:"description" binding:"required"`
		Category        models.ProblemCategory `json:"category" binding:"required"`
		Priority        models.ProblemPriority `json:"priority"`
		IsAnonymous     bool                   `json:"is_anonymous"`
		ContactMethod   string                 `json:"contact_method"`
		PhoneNumber     string                 `json:"phone_number"`
		PreferredTime   string                 `json:"preferred_time"`
		WitnessInfo     string                 `json:"witness_info"`
		Location        string                 `json:"location"`
		IncidentDate    *time.Time             `json:"incident_date"`
		PreviousReports bool                   `json:"previous_reports"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	// Set default priority if not provided
	if request.Priority == "" {
		request.Priority = models.ProblemPriorityMedium
	}

	// Set default contact method if not provided
	if request.ContactMethod == "" {
		request.ContactMethod = "email"
	}

	problem, err := h.HRProblemService.CreateProblem(
		request.Title,
		request.Description,
		request.Category,
		request.Priority,
		userID.(uint),
		request.IsAnonymous,
		request.ContactMethod,
		request.PhoneNumber,
		request.PreferredTime,
		request.WitnessInfo,
		request.Location,
		request.IncidentDate,
		request.PreviousReports,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Problem reported successfully",
		"problem": gin.H{
			"id":          problem.ID,
			"title":       problem.Title,
			"category":    problem.Category,
			"priority":    problem.Priority,
			"status":      problem.Status,
			"reported_at": problem.ReportedAt,
			"is_urgent":   problem.IsUrgent,
		},
	})
}

// GetProblem retrieves a specific problem by ID
func (h *HRProblemHandler) GetProblem(c *gin.Context) {
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
		return
	}

	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	problem, err := h.HRProblemService.GetProblemByID(uint(problemID), userID.(uint), userRole.(models.Role))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Problem not found"})
		return
	}

	response := gin.H{
		"id":               problem.ID,
		"title":            problem.Title,
		"description":      problem.Description,
		"category":         problem.Category,
		"priority":         problem.Priority,
		"status":           problem.Status,
		"is_anonymous":     problem.IsAnonymous,
		"is_urgent":        problem.IsUrgent,
		"reported_at":      problem.ReportedAt,
		"resolved_at":      problem.ResolvedAt,
		"contact_method":   problem.ContactMethod,
		"phone_number":     problem.PhoneNumber,
		"preferred_time":   problem.PreferredTime,
		"witness_info":     problem.WitnessInfo,
		"location":         problem.Location,
		"incident_date":    problem.IncidentDate,
		"previous_reports": problem.PreviousReports,
		"resolution":       problem.Resolution,
		"created_at":       problem.CreatedAt,
		"updated_at":       problem.UpdatedAt,
	}

	// Add reporter info if not anonymous and user has permission
	if !problem.IsAnonymous && (userRole.(models.Role) == models.RoleHR || userRole.(models.Role) == models.RoleAdmin) {
		response["reporter"] = gin.H{
			"id":       problem.Reporter.ID,
			"username": problem.Reporter.Username,
			"role":     problem.Reporter.Role,
		}
	}

	// Add assigned HR info if available
	if problem.AssignedHR != nil {
		response["assigned_hr"] = gin.H{
			"id":       problem.AssignedHR.ID,
			"username": problem.AssignedHR.Username,
		}
	}

	// Add HR notes if user is HR/Admin
	if userRole.(models.Role) == models.RoleHR || userRole.(models.Role) == models.RoleAdmin {
		response["hr_notes"] = problem.HRNotes
		response["follow_up_date"] = problem.FollowUpDate
	}

	// Add comments
	var comments []gin.H
	for _, comment := range problem.Comments {
		comments = append(comments, gin.H{
			"id":         comment.ID,
			"comment":    comment.Comment,
			"user":       comment.User.Username,
			"is_hr_only": comment.IsHROnly,
			"created_at": comment.CreatedAt,
		})
	}
	response["comments"] = comments

	// Add updates
	var updates []gin.H
	for _, update := range problem.Updates {
		updates = append(updates, gin.H{
			"id":           update.ID,
			"old_status":   update.OldStatus,
			"new_status":   update.NewStatus,
			"update_note":  update.UpdateNote,
			"updated_by":   update.UpdatedByUser.Username,
			"is_automatic": update.IsAutomatic,
			"created_at":   update.CreatedAt,
		})
	}
	response["updates"] = updates

	c.JSON(http.StatusOK, gin.H{"problem": response})
}

// GetUserProblems retrieves all problems reported by the current user
func (h *HRProblemHandler) GetUserProblems(c *gin.Context) {
	userID, _ := c.Get("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	problems, err := h.HRProblemService.GetUserProblems(userID.(uint), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch problems"})
		return
	}

	var problemList []gin.H
	for _, problem := range problems {
		problemData := gin.H{
			"id":          problem.ID,
			"title":       problem.Title,
			"category":    problem.Category,
			"priority":    problem.Priority,
			"status":      problem.Status,
			"is_urgent":   problem.IsUrgent,
			"reported_at": problem.ReportedAt,
			"resolved_at": problem.ResolvedAt,
		}

		if problem.AssignedHR != nil {
			problemData["assigned_hr"] = problem.AssignedHR.Username
		}

		problemList = append(problemList, problemData)
	}

	c.JSON(http.StatusOK, gin.H{
		"problems": problemList,
		"page":     page,
		"limit":    limit,
	})
}

// GetAllProblems retrieves all problems (HR/Admin only)
func (h *HRProblemHandler) GetAllProblems(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	category := c.Query("category")
	priority := c.Query("priority")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	problems, err := h.HRProblemService.GetAllProblems(page, limit, status, category, priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch problems"})
		return
	}

	var problemList []gin.H
	for _, problem := range problems {
		problemData := gin.H{
			"id":           problem.ID,
			"title":        problem.Title,
			"category":     problem.Category,
			"priority":     problem.Priority,
			"status":       problem.Status,
			"is_urgent":    problem.IsUrgent,
			"is_anonymous": problem.IsAnonymous,
			"reported_at":  problem.ReportedAt,
			"resolved_at":  problem.ResolvedAt,
		}

		if !problem.IsAnonymous {
			problemData["reporter"] = gin.H{
				"id":       problem.Reporter.ID,
				"username": problem.Reporter.Username,
				"role":     problem.Reporter.Role,
			}
		}

		if problem.AssignedHR != nil {
			problemData["assigned_hr"] = gin.H{
				"id":       problem.AssignedHR.ID,
				"username": problem.AssignedHR.Username,
			}
		}

		problemList = append(problemList, problemData)
	}

	c.JSON(http.StatusOK, gin.H{
		"problems": problemList,
		"page":     page,
		"limit":    limit,
		"filters": gin.H{
			"status":   status,
			"category": category,
			"priority": priority,
		},
	})
}

// GetAssignedProblems retrieves problems assigned to the current HR user
func (h *HRProblemHandler) GetAssignedProblems(c *gin.Context) {
	userID, _ := c.Get("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	problems, err := h.HRProblemService.GetProblemsByAssignedHR(userID.(uint), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assigned problems"})
		return
	}

	var problemList []gin.H
	for _, problem := range problems {
		problemData := gin.H{
			"id":           problem.ID,
			"title":        problem.Title,
			"category":     problem.Category,
			"priority":     problem.Priority,
			"status":       problem.Status,
			"is_urgent":    problem.IsUrgent,
			"is_anonymous": problem.IsAnonymous,
			"reported_at":  problem.ReportedAt,
			"resolved_at":  problem.ResolvedAt,
		}

		if !problem.IsAnonymous {
			problemData["reporter"] = gin.H{
				"id":       problem.Reporter.ID,
				"username": problem.Reporter.Username,
				"role":     problem.Reporter.Role,
			}
		}

		problemList = append(problemList, problemData)
	}

	c.JSON(http.StatusOK, gin.H{
		"problems": problemList,
		"page":     page,
		"limit":    limit,
	})
}

// UpdateProblemStatus updates the status of a problem (HR/Admin only)
func (h *HRProblemHandler) UpdateProblemStatus(c *gin.Context) {
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
		return
	}

	var request struct {
		Status models.ProblemStatus `json:"status" binding:"required"`
		Note   string               `json:"note"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	err = h.HRProblemService.UpdateProblemStatus(uint(problemID), request.Status, userID.(uint), request.Note)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Problem status updated successfully"})
}

// AssignProblem assigns a problem to a specific HR user (HR/Admin only)
func (h *HRProblemHandler) AssignProblem(c *gin.Context) {
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
		return
	}

	var request struct {
		HRUserID uint `json:"hr_user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	err = h.HRProblemService.AssignProblemToHR(uint(problemID), request.HRUserID, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Problem assigned successfully"})
}

// AddComment adds a comment to a problem
func (h *HRProblemHandler) AddComment(c *gin.Context) {
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
		return
	}

	var request struct {
		Comment  string `json:"comment" binding:"required"`
		IsHROnly bool   `json:"is_hr_only"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	// Only HR/Admin can create HR-only comments
	if request.IsHROnly && userRole.(models.Role) != models.RoleHR && userRole.(models.Role) != models.RoleAdmin {
		request.IsHROnly = false
	}

	err = h.HRProblemService.AddComment(uint(problemID), userID.(uint), request.Comment, request.IsHROnly)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment added successfully"})
}

// UpdateHRNotes updates HR internal notes (HR/Admin only)
func (h *HRProblemHandler) UpdateHRNotes(c *gin.Context) {
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
		return
	}

	var request struct {
		Notes string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	err = h.HRProblemService.UpdateHRNotes(uint(problemID), request.Notes, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "HR notes updated successfully"})
}

// SetResolution sets the final resolution for a problem (HR/Admin only)
func (h *HRProblemHandler) SetResolution(c *gin.Context) {
	problemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
		return
	}

	var request struct {
		Resolution string `json:"resolution" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	err = h.HRProblemService.SetResolution(uint(problemID), request.Resolution, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Problem resolution set successfully"})
}

// GetStatistics returns HR problem statistics (HR/Admin only)
func (h *HRProblemHandler) GetStatistics(c *gin.Context) {
	stats, err := h.HRProblemService.GetProblemStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statistics": stats})
}
