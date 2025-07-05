package services

import (
	"errors"
	"fmt"
	"project-x/models"
	"time"

	"gorm.io/gorm"
)

type HRProblemService struct {
	db                  *gorm.DB
	notificationService *NotificationService
}

func NewHRProblemService(db *gorm.DB, notificationService *NotificationService) *HRProblemService {
	return &HRProblemService{
		db:                  db,
		notificationService: notificationService,
	}
}

// CreateProblem creates a new HR problem report
func (s *HRProblemService) CreateProblem(
	title, description string,
	category models.ProblemCategory,
	priority models.ProblemPriority,
	reporterID uint,
	isAnonymous bool,
	contactMethod, phoneNumber, preferredTime string,
	witnessInfo, location string,
	incidentDate *time.Time,
	previousReports bool,
) (*models.HRProblem, error) {
	// Validate required fields
	if title == "" || description == "" {
		return nil, errors.New("title and description are required")
	}

	// Validate category
	categories := models.GetProblemCategories()
	if _, exists := categories[category]; !exists {
		return nil, errors.New("invalid problem category")
	}

	// Validate priority
	priorities := models.GetProblemPriorities()
	if _, exists := priorities[priority]; !exists {
		return nil, errors.New("invalid problem priority")
	}

	// Set urgent flag for high priority issues
	isUrgent := priority == models.ProblemPriorityUrgent || priority == models.ProblemPritorityCritical

	problem := &models.HRProblem{
		Title:           title,
		Description:     description,
		Category:        category,
		Priority:        priority,
		Status:          models.ProblemStatusPending,
		ReporterID:      reporterID,
		IsAnonymous:     isAnonymous,
		IsUrgent:        isUrgent,
		ReportedAt:      time.Now(),
		ContactMethod:   contactMethod,
		PhoneNumber:     phoneNumber,
		PreferredTime:   preferredTime,
		WitnessInfo:     witnessInfo,
		Location:        location,
		IncidentDate:    incidentDate,
		PreviousReports: previousReports,
	}

	if err := s.db.Create(problem).Error; err != nil {
		return nil, err
	}

	// Load reporter info
	s.db.Preload("Reporter").First(problem, problem.ID)

	// Create initial update record
	s.createStatusUpdate(problem.ID, reporterID, "", models.ProblemStatusPending, "Problem reported", true)

	// Send notification to all HR users
	s.notifyHRUsers(problem)

	return problem, nil
}

// GetProblemByID retrieves a problem by ID
func (s *HRProblemService) GetProblemByID(problemID uint, userID uint, userRole models.Role) (*models.HRProblem, error) {
	var problem models.HRProblem
	query := s.db.Preload("Reporter").Preload("AssignedHR").Preload("Comments.User").Preload("Updates.UpdatedByUser")

	// Non-HR users can only see their own reports
	if userRole != models.RoleHR && userRole != models.RoleAdmin {
		query = query.Where("reporter_id = ?", userID)
	}

	if err := query.First(&problem, problemID).Error; err != nil {
		return nil, err
	}

	// Filter HR-only comments for non-HR users
	if userRole != models.RoleHR && userRole != models.RoleAdmin {
		var filteredComments []models.HRProblemComment
		for _, comment := range problem.Comments {
			if !comment.IsHROnly {
				filteredComments = append(filteredComments, comment)
			}
		}
		problem.Comments = filteredComments
	}

	return &problem, nil
}

// GetUserProblems retrieves all problems reported by a user
func (s *HRProblemService) GetUserProblems(userID uint, page, limit int) ([]models.HRProblem, error) {
	var problems []models.HRProblem
	offset := (page - 1) * limit

	err := s.db.Where("reporter_id = ?", userID).
		Preload("AssignedHR").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&problems).Error

	return problems, err
}

// GetAllProblems retrieves all problems (HR/Admin only)
func (s *HRProblemService) GetAllProblems(page, limit int, status string, category string, priority string) ([]models.HRProblem, error) {
	var problems []models.HRProblem
	offset := (page - 1) * limit

	query := s.db.Preload("Reporter").Preload("AssignedHR")

	// Apply filters
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}

	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&problems).Error

	return problems, err
}

// GetProblemsByAssignedHR retrieves problems assigned to a specific HR user
func (s *HRProblemService) GetProblemsByAssignedHR(hrUserID uint, page, limit int) ([]models.HRProblem, error) {
	var problems []models.HRProblem
	offset := (page - 1) * limit

	err := s.db.Where("assigned_hr_id = ?", hrUserID).
		Preload("Reporter").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&problems).Error

	return problems, err
}

// UpdateProblemStatus updates the status of a problem (HR/Admin only)
func (s *HRProblemService) UpdateProblemStatus(problemID uint, newStatus models.ProblemStatus, updatedBy uint, note string) error {
	var problem models.HRProblem
	if err := s.db.First(&problem, problemID).Error; err != nil {
		return err
	}

	oldStatus := problem.Status
	problem.Status = newStatus

	// Set resolved date if status is resolved or closed
	if newStatus == models.ProblemStatusResolved || newStatus == models.ProblemStatusClosed {
		now := time.Now()
		problem.ResolvedAt = &now
	}

	if err := s.db.Save(&problem).Error; err != nil {
		return err
	}

	// Create status update record
	s.createStatusUpdate(problemID, updatedBy, oldStatus, newStatus, note, false)

	// Notify reporter about status change
	s.notifyReporter(&problem, fmt.Sprintf("Your problem report status has been updated to: %s", newStatus))

	return nil
}

// AssignProblemToHR assigns a problem to a specific HR user
func (s *HRProblemService) AssignProblemToHR(problemID uint, hrUserID uint, assignedBy uint) error {
	var problem models.HRProblem
	if err := s.db.First(&problem, problemID).Error; err != nil {
		return err
	}

	// Verify the assignee is HR or Admin
	var hrUser models.User
	if err := s.db.First(&hrUser, hrUserID).Error; err != nil {
		return errors.New("HR user not found")
	}
	if hrUser.Role != models.RoleHR && hrUser.Role != models.RoleAdmin {
		return errors.New("can only assign to HR or Admin users")
	}

	problem.AssignedHRID = &hrUserID
	if err := s.db.Save(&problem).Error; err != nil {
		return err
	}

	// Create update record
	s.createStatusUpdate(problemID, assignedBy, "", "", fmt.Sprintf("Problem assigned to %s", hrUser.Username), false)

	// Notify assigned HR user
	s.notifyAssignedHR(&problem, &hrUser)

	return nil
}

// AddComment adds a comment to a problem
func (s *HRProblemService) AddComment(problemID uint, userID uint, comment string, isHROnly bool) error {
	// Verify problem exists
	var problem models.HRProblem
	if err := s.db.First(&problem, problemID).Error; err != nil {
		return err
	}

	hrComment := &models.HRProblemComment{
		ProblemID: problemID,
		UserID:    userID,
		Comment:   comment,
		IsHROnly:  isHROnly,
	}

	if err := s.db.Create(hrComment).Error; err != nil {
		return err
	}

	// Load user info
	s.db.Preload("User").First(hrComment, hrComment.ID)

	// Notify relevant parties about new comment
	if !isHROnly {
		s.notifyReporter(&problem, fmt.Sprintf("New comment added to your problem report by %s", hrComment.User.Username))
	}

	return nil
}

// UpdateHRNotes updates HR internal notes (HR/Admin only)
func (s *HRProblemService) UpdateHRNotes(problemID uint, notes string, updatedBy uint) error {
	var problem models.HRProblem
	if err := s.db.First(&problem, problemID).Error; err != nil {
		return err
	}

	problem.HRNotes = notes
	if err := s.db.Save(&problem).Error; err != nil {
		return err
	}

	// Create update record
	s.createStatusUpdate(problemID, updatedBy, "", "", "HR notes updated", false)

	return nil
}

// SetResolution sets the final resolution for a problem (HR/Admin only)
func (s *HRProblemService) SetResolution(problemID uint, resolution string, updatedBy uint) error {
	var problem models.HRProblem
	if err := s.db.First(&problem, problemID).Error; err != nil {
		return err
	}

	problem.Resolution = resolution
	problem.Status = models.ProblemStatusResolved
	now := time.Now()
	problem.ResolvedAt = &now

	if err := s.db.Save(&problem).Error; err != nil {
		return err
	}

	// Create update record
	s.createStatusUpdate(problemID, updatedBy, "", models.ProblemStatusResolved, "Problem resolved with final resolution", false)

	// Notify reporter about resolution
	s.notifyReporter(&problem, "Your problem report has been resolved")

	return nil
}

// GetProblemStatistics returns statistics for HR dashboard
func (s *HRProblemService) GetProblemStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total problems
	var totalProblems int64
	s.db.Model(&models.HRProblem{}).Count(&totalProblems)
	stats["total_problems"] = totalProblems

	// Problems by status
	statusStats := make(map[string]int64)
	statuses := []models.ProblemStatus{
		models.ProblemStatusPending,
		models.ProblemStatusReviewing,
		models.ProblemStatusInProgress,
		models.ProblemStatusResolved,
		models.ProblemStatusClosed,
		models.ProblemStatusRejected,
	}
	for _, status := range statuses {
		var count int64
		s.db.Model(&models.HRProblem{}).Where("status = ?", status).Count(&count)
		statusStats[string(status)] = count
	}
	stats["by_status"] = statusStats

	// Problems by category
	categoryStats := make(map[string]int64)
	categories := models.GetProblemCategories()
	for category := range categories {
		var count int64
		s.db.Model(&models.HRProblem{}).Where("category = ?", category).Count(&count)
		categoryStats[string(category)] = count
	}
	stats["by_category"] = categoryStats

	// Problems by priority
	priorityStats := make(map[string]int64)
	priorities := models.GetProblemPriorities()
	for priority := range priorities {
		var count int64
		s.db.Model(&models.HRProblem{}).Where("priority = ?", priority).Count(&count)
		priorityStats[string(priority)] = count
	}
	stats["by_priority"] = priorityStats

	// Urgent problems
	var urgentProblems int64
	s.db.Model(&models.HRProblem{}).Where("is_urgent = ?", true).Where("status NOT IN ?", []string{"resolved", "closed"}).Count(&urgentProblems)
	stats["urgent_open"] = urgentProblems

	// Recent problems (last 7 days)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	var recentProblems int64
	s.db.Model(&models.HRProblem{}).Where("created_at >= ?", sevenDaysAgo).Count(&recentProblems)
	stats["recent_problems"] = recentProblems

	return stats, nil
}

// Helper function to create status update records
func (s *HRProblemService) createStatusUpdate(problemID uint, updatedBy uint, oldStatus models.ProblemStatus, newStatus models.ProblemStatus, note string, isAutomatic bool) {
	update := &models.HRProblemUpdate{
		ProblemID:   problemID,
		UpdatedBy:   updatedBy,
		OldStatus:   oldStatus,
		NewStatus:   newStatus,
		UpdateNote:  note,
		IsAutomatic: isAutomatic,
	}
	s.db.Create(update)
}

// Helper function to notify HR users about new problems
func (s *HRProblemService) notifyHRUsers(problem *models.HRProblem) {
	if s.notificationService == nil {
		return
	}

	// Get all HR users
	var hrUsers []models.User
	s.db.Where("role = ?", models.RoleHR).Find(&hrUsers)

	for _, hrUser := range hrUsers {
		title := "New HR Problem Report"
		message := fmt.Sprintf("New %s problem reported: %s", problem.Category, problem.Title)
		if problem.IsUrgent {
			title = "URGENT: " + title
		}

		s.notificationService.CreateNotification(
			hrUser.ID,
			title,
			message,
			"hr_problem",
			map[string]interface{}{
				"problem_id": problem.ID,
				"category":   problem.Category,
				"priority":   problem.Priority,
				"is_urgent":  problem.IsUrgent,
			},
		)
	}
}

// Helper function to notify reporter about updates
func (s *HRProblemService) notifyReporter(problem *models.HRProblem, message string) {
	if s.notificationService == nil || problem.IsAnonymous {
		return
	}

	s.notificationService.CreateNotification(
		problem.ReporterID,
		"HR Problem Report Update",
		message,
		"hr_problem_update",
		map[string]interface{}{
			"problem_id": problem.ID,
			"status":     problem.Status,
		},
	)
}

// Helper function to notify assigned HR user
func (s *HRProblemService) notifyAssignedHR(problem *models.HRProblem, hrUser *models.User) {
	if s.notificationService == nil {
		return
	}

	s.notificationService.CreateNotification(
		hrUser.ID,
		"HR Problem Assigned",
		fmt.Sprintf("You have been assigned to handle problem: %s", problem.Title),
		"hr_problem_assigned",
		map[string]interface{}{
			"problem_id": problem.ID,
			"category":   problem.Category,
			"priority":   problem.Priority,
		},
	)
}
