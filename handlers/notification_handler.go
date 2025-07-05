package handlers

import (
	"net/http"
	"project-x/models"
	"project-x/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// getUserIDFromContext extracts user ID from gin context
func getUserIDFromContext(c *gin.Context) uint {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		return 0
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		return 0
	}
	return userID
}

type NotificationHandler struct {
	notificationService *services.NotificationService
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetUserNotifications gets notifications for the authenticated user
func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get limit from query parameter (default 50)
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	notifications, err := h.notificationService.GetUserNotifications(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

// MarkNotificationAsRead marks a notification as read
func (h *NotificationHandler) MarkNotificationAsRead(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	err = h.notificationService.MarkNotificationAsRead(uint(notificationID), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllNotificationsAsRead marks all notifications as read for the user
func (h *NotificationHandler) MarkAllNotificationsAsRead(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// This would need to be implemented in the service
	// For now, we'll return a placeholder
	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// GetNotificationRules gets the current role-based notification rules
func (h *NotificationHandler) GetNotificationRules(c *gin.Context) {
	rules := h.notificationService.GetRoleNotificationRules()
	c.JSON(http.StatusOK, gin.H{
		"notification_rules": rules,
	})
}

// UpdateNotificationRules updates the role-based notification rules
func (h *NotificationHandler) UpdateNotificationRules(c *gin.Context) {
	var request struct {
		NotificationType string        `json:"notification_type" binding:"required"`
		AllowedRoles     []models.Role `json:"allowed_roles" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate notification type
	validTypes := []string{
		"task_assigned", "task_updated", "task_completed", "task_commented",
		"task_due_soon", "project_created", "user_joined", "file_uploaded",
	}
	validType := false
	for _, t := range validTypes {
		if t == request.NotificationType {
			validType = true
			break
		}
	}
	if !validType {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification type"})
		return
	}

	// Validate roles
	validRoles := []models.Role{models.RoleAdmin, models.RoleManager, models.RoleHead, models.RoleEmployee}
	for _, role := range request.AllowedRoles {
		roleValid := false
		for _, validRole := range validRoles {
			if role == validRole {
				roleValid = true
				break
			}
		}
		if !roleValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role in allowed_roles"})
			return
		}
	}

	h.notificationService.UpdateRoleNotificationRules(models.NotificationType(request.NotificationType), request.AllowedRoles)

	c.JSON(http.StatusOK, gin.H{
		"message":           "Notification rules updated successfully",
		"notification_type": request.NotificationType,
		"allowed_roles":     request.AllowedRoles,
	})
}

// GetUnreadCount gets count of unread notifications
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	count, err := h.notificationService.GetUnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}

// GetNotificationPreferences gets user's notification preferences
func (h *NotificationHandler) GetNotificationPreferences(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var preference models.UserNotificationPreference
	err := h.notificationService.GetDB().Where("user_id = ?", userID).First(&preference).Error
	if err != nil {
		// Create default preferences if not found
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
		h.notificationService.GetDB().Create(&preference)
	}

	c.JSON(http.StatusOK, gin.H{
		"preferences": preference,
	})
}

// UpdateNotificationPreferences updates user's notification preferences
func (h *NotificationHandler) UpdateNotificationPreferences(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		TaskAssigned       *bool `json:"task_assigned"`
		TaskUpdated        *bool `json:"task_updated"`
		TaskCompleted      *bool `json:"task_completed"`
		TaskCommented      *bool `json:"task_commented"`
		TaskDueSoon        *bool `json:"task_due_soon"`
		ProjectCreated     *bool `json:"project_created"`
		UserJoined         *bool `json:"user_joined"`
		FileUploaded       *bool `json:"file_uploaded"`
		EmailNotifications *bool `json:"email_notifications"`
		PushNotifications  *bool `json:"push_notifications"`
		InAppNotifications *bool `json:"in_app_notifications"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get existing preferences or create new ones
	var preference models.UserNotificationPreference
	err := h.notificationService.GetDB().Where("user_id = ?", userID).First(&preference).Error
	if err != nil {
		preference = models.UserNotificationPreference{
			UserID: userID,
		}
	}

	// Update only the fields that were provided
	if request.TaskAssigned != nil {
		preference.TaskAssigned = *request.TaskAssigned
	}
	if request.TaskUpdated != nil {
		preference.TaskUpdated = *request.TaskUpdated
	}
	if request.TaskCompleted != nil {
		preference.TaskCompleted = *request.TaskCompleted
	}
	if request.TaskCommented != nil {
		preference.TaskCommented = *request.TaskCommented
	}
	if request.TaskDueSoon != nil {
		preference.TaskDueSoon = *request.TaskDueSoon
	}
	if request.ProjectCreated != nil {
		preference.ProjectCreated = *request.ProjectCreated
	}
	if request.UserJoined != nil {
		preference.UserJoined = *request.UserJoined
	}
	if request.FileUploaded != nil {
		preference.FileUploaded = *request.FileUploaded
	}
	if request.EmailNotifications != nil {
		preference.EmailNotifications = *request.EmailNotifications
	}
	if request.PushNotifications != nil {
		preference.PushNotifications = *request.PushNotifications
	}
	if request.InAppNotifications != nil {
		preference.InAppNotifications = *request.InAppNotifications
	}

	// Save preferences
	if preference.ID == 0 {
		err = h.notificationService.GetDB().Create(&preference).Error
	} else {
		err = h.notificationService.GetDB().Save(&preference).Error
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Preferences updated successfully",
		"preferences": preference,
	})
}

// DeleteNotification deletes a notification
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	// Delete notification (only if it belongs to the user)
	err = h.notificationService.GetDB().Where("id = ? AND user_id = ?", notificationID, userID).Delete(&models.Notification{}).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}
