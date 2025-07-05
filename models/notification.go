package models

import (
	"time"

	"gorm.io/gorm"
)

type NotificationType string

const (
	NotificationTypeTaskAssigned   NotificationType = "task_assigned"
	NotificationTypeTaskUpdated    NotificationType = "task_updated"
	NotificationTypeTaskCompleted  NotificationType = "task_completed"
	NotificationTypeTaskCommented  NotificationType = "task_commented"
	NotificationTypeTaskDueSoon    NotificationType = "task_due_soon"
	NotificationTypeProjectCreated NotificationType = "project_created"
	NotificationTypeUserJoined     NotificationType = "user_joined"
	NotificationTypeFileUploaded   NotificationType = "file_uploaded"
	// HR-related notifications
	NotificationTypeHRProblem         NotificationType = "hr_problem"
	NotificationTypeHRProblemUpdate   NotificationType = "hr_problem_update"
	NotificationTypeHRProblemAssigned NotificationType = "hr_problem_assigned"
)

type Notification struct {
	gorm.Model
	UserID           uint             `gorm:"not null;index"`
	Type             NotificationType `gorm:"not null;index;type:varchar(50)"`
	Title            string           `gorm:"not null;type:varchar(500) COLLATE \"default\""`
	Message          string           `gorm:"not null;type:text"`
	Data             string           `gorm:"type:json"` // Additional data as JSON
	IsRead           bool             `gorm:"default:false;index"`
	ReadAt           *time.Time
	RelatedTaskID    *uint `gorm:"index"`
	RelatedProjectID *uint `gorm:"index"`
	FromUserID       *uint `gorm:"index"`

	// Relationships
	User           User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	RelatedTask    *Task    `gorm:"foreignKey:RelatedTaskID;constraint:OnDelete:SET NULL"`
	RelatedProject *Project `gorm:"foreignKey:RelatedProjectID;constraint:OnDelete:SET NULL"`
	FromUser       *User    `gorm:"foreignKey:FromUserID;constraint:OnDelete:SET NULL"`
}

type UserNotificationPreference struct {
	gorm.Model
	UserID             uint `gorm:"uniqueIndex"`
	TaskAssigned       bool `gorm:"default:true"`
	TaskUpdated        bool `gorm:"default:true"`
	TaskCompleted      bool `gorm:"default:true"`
	TaskCommented      bool `gorm:"default:true"`
	TaskDueSoon        bool `gorm:"default:true"`
	ProjectCreated     bool `gorm:"default:true"`
	UserJoined         bool `gorm:"default:true"`
	FileUploaded       bool `gorm:"default:true"`
	EmailNotifications bool `gorm:"default:false"`
	PushNotifications  bool `gorm:"default:true"`
	InAppNotifications bool `gorm:"default:true"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}
