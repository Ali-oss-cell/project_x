package services

import (
	"encoding/json"
	"fmt"
	"log"
	"project-x/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

type NotificationService struct {
	db        *gorm.DB
	wsService *WebSocketService
	// Role-based notification control
	roleNotificationRules map[models.NotificationType][]models.Role
}

type NotificationData struct {
	TaskID      *uint   `json:"task_id,omitempty"`
	ProjectID   *uint   `json:"project_id,omitempty"`
	FromUserID  *uint   `json:"from_user_id,omitempty"`
	TaskTitle   string  `json:"task_title,omitempty"`
	ProjectName string  `json:"project_name,omitempty"`
	FromUser    string  `json:"from_user,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
	FileURL     string  `json:"file_url,omitempty"`
	FileName    string  `json:"file_name,omitempty"`
}

func NewNotificationService(db *gorm.DB, wsService *WebSocketService) *NotificationService {
	// Default role-based notification rules
	roleRules := map[models.NotificationType][]models.Role{
		models.NotificationTypeTaskAssigned:   {models.RoleHead, models.RoleEmployee},                                       // Heads and Employees (Heads can assign to Employees)
		models.NotificationTypeTaskUpdated:    {models.RoleHead, models.RoleEmployee, models.RoleManager},                   // Heads, Employees, and Managers
		models.NotificationTypeTaskCompleted:  {models.RoleHead, models.RoleEmployee, models.RoleManager, models.RoleAdmin}, // All roles
		models.NotificationTypeTaskCommented:  {models.RoleHead, models.RoleEmployee, models.RoleManager},                   // Heads, Employees, and Managers
		models.NotificationTypeTaskDueSoon:    {models.RoleHead, models.RoleEmployee},                                       // Only Heads and Employees
		models.NotificationTypeProjectCreated: {models.RoleManager, models.RoleAdmin},                                       // Only Managers and Admins
		models.NotificationTypeUserJoined:     {models.RoleManager, models.RoleAdmin},                                       // Only Managers and Admins
		models.NotificationTypeFileUploaded:   {models.RoleHead, models.RoleEmployee, models.RoleManager},                   // Heads, Employees, and Managers
	}

	return &NotificationService{
		db:                    db,
		wsService:             wsService,
		roleNotificationRules: roleRules,
	}
}

// SendTaskAssignedNotification sends notification when task is assigned
func (ns *NotificationService) SendTaskAssignedNotification(task *models.Task, assignedToUser *models.User, assignedByUser *models.User) error {
	// Only send task assignment notifications to Heads and Employees
	// Admins and Managers (who assign tasks) should not receive these notifications
	if assignedToUser.Role == models.RoleAdmin || assignedToUser.Role == models.RoleManager {
		log.Printf("Skipping task assignment notification for %s (role: %s) - only Heads and Employees receive task assignment notifications",
			assignedToUser.Username, assignedToUser.Role)
		return nil
	}

	// Check if user wants this notification
	if !ns.shouldSendNotification(assignedToUser.ID, models.NotificationTypeTaskAssigned) {
		return nil
	}

	// Create notification data
	data := NotificationData{
		TaskID:     &task.ID,
		FromUserID: &assignedByUser.ID,
		TaskTitle:  task.Title,
		FromUser:   assignedByUser.Username,
	}

	dataJSON, _ := json.Marshal(data)

	notification := models.Notification{
		UserID:        assignedToUser.ID,
		Type:          models.NotificationTypeTaskAssigned,
		Title:         "New Task Assigned",
		Message:       fmt.Sprintf("You have been assigned a new task: '%s' by %s", task.Title, assignedByUser.Username),
		Data:          string(dataJSON),
		RelatedTaskID: &task.ID,
		FromUserID:    &assignedByUser.ID,
	}

	// Save to database
	if err := ns.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to save notification: %v", err)
	}

	// Send real-time notification
	ns.sendRealtimeNotification(assignedToUser.ID, notification)

	log.Printf("Task assigned notification sent to user %s (role: %s)", assignedToUser.Username, assignedToUser.Role)
	return nil
}

// SendTaskAssignedNotificationWithRoleControl sends notification when task is assigned with custom role control
func (ns *NotificationService) SendTaskAssignedNotificationWithRoleControl(task *models.Task, assignedToUser *models.User, assignedByUser *models.User, allowedRoles []models.Role) error {
	// Check if user's role is in the allowed roles list
	roleAllowed := false
	for _, role := range allowedRoles {
		if assignedToUser.Role == role {
			roleAllowed = true
			break
		}
	}

	if !roleAllowed {
		log.Printf("Skipping task assignment notification for %s (role: %s) - role not in allowed list: %v",
			assignedToUser.Username, assignedToUser.Role, allowedRoles)
		return nil
	}

	// Check if user wants this notification
	if !ns.shouldSendNotification(assignedToUser.ID, models.NotificationTypeTaskAssigned) {
		return nil
	}

	// Create notification data
	data := NotificationData{
		TaskID:     &task.ID,
		FromUserID: &assignedByUser.ID,
		TaskTitle:  task.Title,
		FromUser:   assignedByUser.Username,
	}

	dataJSON, _ := json.Marshal(data)

	notification := models.Notification{
		UserID:        assignedToUser.ID,
		Type:          models.NotificationTypeTaskAssigned,
		Title:         "New Task Assigned",
		Message:       fmt.Sprintf("You have been assigned a new task: '%s' by %s", task.Title, assignedByUser.Username),
		Data:          string(dataJSON),
		RelatedTaskID: &task.ID,
		FromUserID:    &assignedByUser.ID,
	}

	// Save to database
	if err := ns.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to save notification: %v", err)
	}

	// Send real-time notification
	ns.sendRealtimeNotification(assignedToUser.ID, notification)

	log.Printf("Task assigned notification sent to user %s (role: %s) with custom role control", assignedToUser.Username, assignedToUser.Role)
	return nil
}

// SendCollaborativeTaskAssignedNotification sends notification when collaborative task is assigned
func (ns *NotificationService) SendCollaborativeTaskAssignedNotification(task *models.CollaborativeTask, assignedToUser *models.User, assignedByUser *models.User) error {
	// Only send task assignment notifications to Heads and Employees
	// Admins and Managers (who assign tasks) should not receive these notifications
	if assignedToUser.Role == models.RoleAdmin || assignedToUser.Role == models.RoleManager {
		log.Printf("Skipping collaborative task assignment notification for %s (role: %s) - only Heads and Employees receive task assignment notifications",
			assignedToUser.Username, assignedToUser.Role)
		return nil
	}

	// Check if user wants this notification
	if !ns.shouldSendNotification(assignedToUser.ID, models.NotificationTypeTaskAssigned) {
		return nil
	}

	// Create notification data
	data := NotificationData{
		TaskID:     &task.ID,
		FromUserID: &assignedByUser.ID,
		TaskTitle:  task.Title,
		FromUser:   assignedByUser.Username,
	}

	dataJSON, _ := json.Marshal(data)

	notification := models.Notification{
		UserID:        assignedToUser.ID,
		Type:          models.NotificationTypeTaskAssigned,
		Title:         "New Collaborative Task Assigned",
		Message:       fmt.Sprintf("You have been assigned a new collaborative task: '%s' by %s", task.Title, assignedByUser.Username),
		Data:          string(dataJSON),
		RelatedTaskID: &task.ID,
		FromUserID:    &assignedByUser.ID,
	}

	// Save to database
	if err := ns.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to save notification: %v", err)
	}

	// Send real-time notification
	ns.sendRealtimeNotification(assignedToUser.ID, notification)

	log.Printf("Collaborative task assigned notification sent to user %s (role: %s)", assignedToUser.Username, assignedToUser.Role)
	return nil
}

// isRoleAllowedForNotification checks if a user's role is allowed to receive a specific notification type
func (ns *NotificationService) isRoleAllowedForNotification(userRole models.Role, notificationType models.NotificationType) bool {
	allowedRoles, exists := ns.roleNotificationRules[notificationType]
	if !exists {
		// If no rules defined for this notification type, allow all roles
		return true
	}

	for _, role := range allowedRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

// UpdateRoleNotificationRules allows you to update the role-based notification rules
func (ns *NotificationService) UpdateRoleNotificationRules(notificationType models.NotificationType, allowedRoles []models.Role) {
	ns.roleNotificationRules[notificationType] = allowedRoles
	log.Printf("Updated notification rules for %s: allowed roles = %v", notificationType, allowedRoles)
}

// GetRoleNotificationRules returns the current role-based notification rules
func (ns *NotificationService) GetRoleNotificationRules() map[models.NotificationType][]models.Role {
	return ns.roleNotificationRules
}

// SendTaskStatusChangedNotification sends notification when task status changes
func (ns *NotificationService) SendTaskStatusChangedNotification(task *models.Task, updatedByUser *models.User, oldStatus, newStatus models.TaskStatus) error {
	// Get task assignee
	var assignee models.User
	if err := ns.db.First(&assignee, task.UserID).Error; err != nil {
		return fmt.Errorf("failed to get task assignee: %v", err)
	}

	// Check if assignee wants this notification
	if !ns.shouldSendNotification(assignee.ID, models.NotificationTypeTaskUpdated) {
		return nil
	}

	// Create notification data
	data := NotificationData{
		TaskID:     &task.ID,
		FromUserID: &updatedByUser.ID,
		TaskTitle:  task.Title,
		FromUser:   updatedByUser.Username,
	}

	dataJSON, _ := json.Marshal(data)

	notification := models.Notification{
		UserID:        assignee.ID,
		Type:          models.NotificationTypeTaskUpdated,
		Title:         "Task Status Updated",
		Message:       fmt.Sprintf("Task '%s' status changed from %s to %s by %s", task.Title, oldStatus, newStatus, updatedByUser.Username),
		Data:          string(dataJSON),
		RelatedTaskID: &task.ID,
		FromUserID:    &updatedByUser.ID,
	}

	// Save to database
	if err := ns.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to save notification: %v", err)
	}

	// Send real-time notification
	ns.sendRealtimeNotification(assignee.ID, notification)

	log.Printf("Task status change notification sent to user %s", assignee.Username)
	return nil
}

// SendTaskCompletedNotification sends notification when task is completed
func (ns *NotificationService) SendTaskCompletedNotification(task *models.Task, completedByUser *models.User) error {
	// Get project members to notify
	var projectMembers []models.User
	if task.ProjectID != nil {
		// Get all project members
		err := ns.db.Table("users").
			Joins("JOIN user_projects ON users.id = user_projects.user_id").
			Where("user_projects.project_id = ?", *task.ProjectID).
			Find(&projectMembers).Error
		if err != nil {
			return fmt.Errorf("failed to get project members: %v", err)
		}
	} else {
		// If no project, notify the task creator (if different from assignee)
		if task.UserID != completedByUser.ID {
			var assignee models.User
			if err := ns.db.First(&assignee, task.UserID).Error; err == nil {
				projectMembers = append(projectMembers, assignee)
			}
		}
	}

	// Send notifications to all relevant users
	for _, member := range projectMembers {
		if member.ID == completedByUser.ID {
			continue // Don't notify the person who completed it
		}

		if !ns.shouldSendNotification(member.ID, models.NotificationTypeTaskCompleted) {
			continue
		}

		data := NotificationData{
			TaskID:     &task.ID,
			FromUserID: &completedByUser.ID,
			TaskTitle:  task.Title,
			FromUser:   completedByUser.Username,
		}

		dataJSON, _ := json.Marshal(data)

		notification := models.Notification{
			UserID:        member.ID,
			Type:          models.NotificationTypeTaskCompleted,
			Title:         "Task Completed",
			Message:       fmt.Sprintf("Task '%s' has been completed by %s", task.Title, completedByUser.Username),
			Data:          string(dataJSON),
			RelatedTaskID: &task.ID,
			FromUserID:    &completedByUser.ID,
		}

		// Save to database
		if err := ns.db.Create(&notification).Error; err != nil {
			log.Printf("Failed to save completion notification for user %s: %v", member.Username, err)
			continue
		}

		// Send real-time notification
		ns.sendRealtimeNotification(member.ID, notification)
	}

	log.Printf("Task completion notifications sent for task: %s", task.Title)
	return nil
}

// SendTaskDueSoonNotification sends notification when task is due soon
func (ns *NotificationService) SendTaskDueSoonNotification(task *models.Task) error {
	if task.DueDate == nil {
		return nil // No due date, no notification
	}

	// Check if due within 24 hours
	dueTime := *task.DueDate
	now := time.Now()
	if dueTime.Sub(now) > 24*time.Hour {
		return nil // Not due soon
	}

	// Get task assignee
	var assignee models.User
	if err := ns.db.First(&assignee, task.UserID).Error; err != nil {
		return fmt.Errorf("failed to get task assignee: %v", err)
	}

	// Check if assignee wants this notification
	if !ns.shouldSendNotification(assignee.ID, models.NotificationTypeTaskDueSoon) {
		return nil
	}

	// Check if we already sent a due soon notification for this task
	var existingNotification models.Notification
	err := ns.db.Where("user_id = ? AND type = ? AND related_task_id = ? AND created_at > ?",
		assignee.ID, models.NotificationTypeTaskDueSoon, task.ID, now.Add(-24*time.Hour)).First(&existingNotification).Error
	if err == nil {
		return nil // Already sent notification
	}

	dueDateStr := task.DueDate.Format("2006-01-02 15:04")
	data := NotificationData{
		TaskID:    &task.ID,
		TaskTitle: task.Title,
		DueDate:   &dueDateStr,
	}

	dataJSON, _ := json.Marshal(data)

	notification := models.Notification{
		UserID:        assignee.ID,
		Type:          models.NotificationTypeTaskDueSoon,
		Title:         "Task Due Soon",
		Message:       fmt.Sprintf("Task '%s' is due on %s", task.Title, task.DueDate.Format("Jan 2, 2006 at 3:04 PM")),
		Data:          string(dataJSON),
		RelatedTaskID: &task.ID,
	}

	// Save to database
	if err := ns.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to save notification: %v", err)
	}

	// Send real-time notification
	ns.sendRealtimeNotification(assignee.ID, notification)

	log.Printf("Task due soon notification sent to user %s for task: %s", assignee.Username, task.Title)
	return nil
}

// SendFileUploadedNotification sends notification when file is uploaded
func (ns *NotificationService) SendFileUploadedNotification(task *models.Task, uploadedByUser *models.User, fileName, fileURL string) error {
	// Get project members to notify
	var projectMembers []models.User
	if task.ProjectID != nil {
		err := ns.db.Table("users").
			Joins("JOIN user_projects ON users.id = user_projects.user_id").
			Where("user_projects.project_id = ?", *task.ProjectID).
			Find(&projectMembers).Error
		if err != nil {
			return fmt.Errorf("failed to get project members: %v", err)
		}
	} else {
		// If no project, notify the task assignee
		var assignee models.User
		if err := ns.db.First(&assignee, task.UserID).Error; err == nil {
			projectMembers = append(projectMembers, assignee)
		}
	}

	// Send notifications to all relevant users
	for _, member := range projectMembers {
		if member.ID == uploadedByUser.ID {
			continue // Don't notify the person who uploaded
		}

		if !ns.shouldSendNotification(member.ID, models.NotificationTypeFileUploaded) {
			continue
		}

		data := NotificationData{
			TaskID:     &task.ID,
			FromUserID: &uploadedByUser.ID,
			TaskTitle:  task.Title,
			FromUser:   uploadedByUser.Username,
			FileName:   fileName,
			FileURL:    fileURL,
		}

		dataJSON, _ := json.Marshal(data)

		notification := models.Notification{
			UserID:        member.ID,
			Type:          models.NotificationTypeFileUploaded,
			Title:         "File Uploaded",
			Message:       fmt.Sprintf("%s uploaded '%s' to task '%s'", uploadedByUser.Username, fileName, task.Title),
			Data:          string(dataJSON),
			RelatedTaskID: &task.ID,
			FromUserID:    &uploadedByUser.ID,
		}

		// Save to database
		if err := ns.db.Create(&notification).Error; err != nil {
			log.Printf("Failed to save file upload notification for user %s: %v", member.Username, err)
			continue
		}

		// Send real-time notification
		ns.sendRealtimeNotification(member.ID, notification)
	}

	log.Printf("File upload notifications sent for file: %s", fileName)
	return nil
}

// shouldSendNotification checks if user wants to receive this type of notification
func (ns *NotificationService) shouldSendNotification(userID uint, notificationType models.NotificationType) bool {
	// First, get the user to check their role
	var user models.User
	if err := ns.db.First(&user, userID).Error; err != nil {
		log.Printf("Failed to get user %d for notification check: %v", userID, err)
		return false
	}

	// Check role-based notification rules
	if !ns.isRoleAllowedForNotification(user.Role, notificationType) {
		log.Printf("User %s (role: %s) not allowed to receive %s notifications", user.Username, user.Role, notificationType)
		return false
	}

	var preference models.UserNotificationPreference
	err := ns.db.Where("user_id = ?", userID).First(&preference).Error

	if err != nil {
		// No preference found, create default one
		preference = models.UserNotificationPreference{
			UserID:             userID,
			TaskAssigned:       true,
			TaskUpdated:        true,
			TaskCompleted:      true,
			TaskCommented:      true,
			TaskDueSoon:        true,
			ProjectCreated:     true,
			UserJoined:         true,
			FileUploaded:       true,
			EmailNotifications: false,
			PushNotifications:  true,
			InAppNotifications: true,
		}
		ns.db.Create(&preference)
		return true
	}

	// Check specific notification type
	switch notificationType {
	case models.NotificationTypeTaskAssigned:
		return preference.TaskAssigned
	case models.NotificationTypeTaskUpdated:
		return preference.TaskUpdated
	case models.NotificationTypeTaskCompleted:
		return preference.TaskCompleted
	case models.NotificationTypeTaskCommented:
		return preference.TaskCommented
	case models.NotificationTypeTaskDueSoon:
		return preference.TaskDueSoon
	case models.NotificationTypeProjectCreated:
		return preference.ProjectCreated
	case models.NotificationTypeUserJoined:
		return preference.UserJoined
	case models.NotificationTypeFileUploaded:
		return preference.FileUploaded
	default:
		return true
	}
}

// sendRealtimeNotification sends notification via WebSocket
func (ns *NotificationService) sendRealtimeNotification(userID uint, notification models.Notification) {
	wsNotification := Notification{
		Type:      string(notification.Type),
		Title:     notification.Title,
		Message:   notification.Message,
		UserID:    notification.UserID,
		Timestamp: notification.CreatedAt,
	}

	// Parse additional data if available
	if notification.Data != "" {
		var data NotificationData
		if err := json.Unmarshal([]byte(notification.Data), &data); err == nil {
			wsNotification.Data = data
		}
	}

	ns.wsService.SendNotification(userID, wsNotification)
}

// GetUserNotifications gets notifications for a user
func (ns *NotificationService) GetUserNotifications(userID uint, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := ns.db.Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).
		Preload("FromUser").
		Preload("RelatedTask").
		Preload("RelatedProject").
		Find(&notifications).Error

	return notifications, err
}

// MarkNotificationAsRead marks a notification as read
func (ns *NotificationService) MarkNotificationAsRead(notificationID uint, userID uint) error {
	now := time.Now()
	return ns.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

// GetUnreadCount gets count of unread notifications for a user
func (ns *NotificationService) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	err := ns.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error

	return count, err
}

// GetDB returns the database instance
func (ns *NotificationService) GetDB() *gorm.DB {
	return ns.db
}

// SendTaskUpdatedNotification sends notification when task details are updated
func (ns *NotificationService) SendTaskUpdatedNotification(task *models.Task, targetUser *models.User, updatedByUser *models.User, oldTitle, oldDescription string, oldDueDate *time.Time) error {
	// Check if user wants this notification
	if !ns.shouldSendNotification(targetUser.ID, models.NotificationTypeTaskUpdated) {
		return nil
	}

	// Create notification data
	data := NotificationData{
		TaskID:     &task.ID,
		FromUserID: &updatedByUser.ID,
		TaskTitle:  task.Title,
		FromUser:   updatedByUser.Username,
	}

	dataJSON, _ := json.Marshal(data)

	// Create message based on what changed
	message := fmt.Sprintf("Task '%s' has been updated by %s", task.Title, updatedByUser.Username)

	// Add specific change details
	var changes []string
	if oldTitle != task.Title {
		changes = append(changes, "title")
	}
	if oldDescription != task.Description {
		changes = append(changes, "description")
	}
	if oldDueDate == nil && task.DueDate != nil {
		changes = append(changes, "due date added")
	} else if oldDueDate != nil && task.DueDate == nil {
		changes = append(changes, "due date removed")
	} else if oldDueDate != nil && task.DueDate != nil && !oldDueDate.Equal(*task.DueDate) {
		changes = append(changes, "due date changed")
	}

	if len(changes) > 0 {
		message += fmt.Sprintf(" (Changes: %s)", strings.Join(changes, ", "))
	}

	notification := models.Notification{
		UserID:        targetUser.ID,
		Type:          models.NotificationTypeTaskUpdated,
		Title:         "Task Updated",
		Message:       message,
		Data:          string(dataJSON),
		RelatedTaskID: &task.ID,
		FromUserID:    &updatedByUser.ID,
	}

	// Save to database
	if err := ns.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to save notification: %v", err)
	}

	// Send real-time notification
	ns.sendRealtimeNotification(targetUser.ID, notification)

	log.Printf("Task update notification sent to user %s (role: %s)", targetUser.Username, targetUser.Role)
	return nil
}

// SendCollaborativeTaskUpdatedNotification sends notification when collaborative task details are updated
func (ns *NotificationService) SendCollaborativeTaskUpdatedNotification(task *models.CollaborativeTask, targetUser *models.User, updatedByUser *models.User, oldTitle, oldDescription string, oldDueDate *time.Time) error {
	// Check if user wants this notification
	if !ns.shouldSendNotification(targetUser.ID, models.NotificationTypeTaskUpdated) {
		return nil
	}

	// Create notification data
	data := NotificationData{
		TaskID:     &task.ID,
		FromUserID: &updatedByUser.ID,
		TaskTitle:  task.Title,
		FromUser:   updatedByUser.Username,
	}

	dataJSON, _ := json.Marshal(data)

	// Create message based on what changed
	message := fmt.Sprintf("Collaborative task '%s' has been updated by %s", task.Title, updatedByUser.Username)

	// Add specific change details
	var changes []string
	if oldTitle != task.Title {
		changes = append(changes, "title")
	}
	if oldDescription != task.Description {
		changes = append(changes, "description")
	}
	if oldDueDate == nil && task.DueDate != nil {
		changes = append(changes, "due date added")
	} else if oldDueDate != nil && task.DueDate == nil {
		changes = append(changes, "due date removed")
	} else if oldDueDate != nil && task.DueDate != nil && !oldDueDate.Equal(*task.DueDate) {
		changes = append(changes, "due date changed")
	}

	if len(changes) > 0 {
		message += fmt.Sprintf(" (Changes: %s)", strings.Join(changes, ", "))
	}

	notification := models.Notification{
		UserID:        targetUser.ID,
		Type:          models.NotificationTypeTaskUpdated,
		Title:         "Collaborative Task Updated",
		Message:       message,
		Data:          string(dataJSON),
		RelatedTaskID: &task.ID,
		FromUserID:    &updatedByUser.ID,
	}

	// Save to database
	if err := ns.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to save notification: %v", err)
	}

	// Send real-time notification
	ns.sendRealtimeNotification(targetUser.ID, notification)

	log.Printf("Collaborative task update notification sent to user %s (role: %s)", targetUser.Username, targetUser.Role)
	return nil
}

// CreateNotification creates a generic notification
func (ns *NotificationService) CreateNotification(userID uint, title, message, notificationType string, data map[string]interface{}) error {
	// Convert data to JSON string
	var dataJSON string
	if data != nil {
		dataBytes, _ := json.Marshal(data)
		dataJSON = string(dataBytes)
	}

	notification := models.Notification{
		UserID:  userID,
		Type:    models.NotificationType(notificationType),
		Title:   title,
		Message: message,
		Data:    dataJSON,
	}

	// Save to database
	if err := ns.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to save notification: %v", err)
	}

	// Send real-time notification
	ns.sendRealtimeNotification(userID, notification)

	log.Printf("Generic notification sent to user %d: %s", userID, title)
	return nil
}
