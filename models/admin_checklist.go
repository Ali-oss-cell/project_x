package models

import (
	"time"

	"gorm.io/gorm"
)

// AdminDailyChecklist represents the daily checklist completion status for admin
type AdminDailyChecklist struct {
	gorm.Model
	Date                  time.Time  `gorm:"uniqueIndex:idx_date;not null;index;type:date"` // Date of the checklist (YYYY-MM-DD)
	HRProblemsReviewed    bool       `gorm:"default:false"`
	CriticalTasksReviewed bool       `gorm:"default:false"`
	ProjectHealthReviewed bool       `gorm:"default:false"`
	UserActivityReviewed  bool       `gorm:"default:false"`
	SystemAlertsReviewed  bool       `gorm:"default:false"`
	AIPerformanceReviewed bool       `gorm:"default:false"`
	BackupStatusReviewed  bool       `gorm:"default:false"`
	SecurityLogsReviewed  bool       `gorm:"default:false"`
	CompletedAt           *time.Time `gorm:"index"`     // When all items were completed
	CompletedBy           *uint      `gorm:"index"`     // Admin user who completed it
	Notes                 string     `gorm:"type:text"` // Optional notes for the day

	// Relationships
	CompletedByUser *User `gorm:"foreignKey:CompletedBy;constraint:OnDelete:SET NULL"`
}

// GetCompletionPercentage calculates the percentage of completed items
func (c *AdminDailyChecklist) GetCompletionPercentage() float64 {
	totalItems := 8.0
	completedItems := 0.0

	if c.HRProblemsReviewed {
		completedItems++
	}
	if c.CriticalTasksReviewed {
		completedItems++
	}
	if c.ProjectHealthReviewed {
		completedItems++
	}
	if c.UserActivityReviewed {
		completedItems++
	}
	if c.SystemAlertsReviewed {
		completedItems++
	}
	if c.AIPerformanceReviewed {
		completedItems++
	}
	if c.BackupStatusReviewed {
		completedItems++
	}
	if c.SecurityLogsReviewed {
		completedItems++
	}

	if totalItems == 0 {
		return 0.0
	}

	return (completedItems / totalItems) * 100.0
}

// IsCompleted checks if all checklist items are completed
func (c *AdminDailyChecklist) IsCompleted() bool {
	return c.HRProblemsReviewed &&
		c.CriticalTasksReviewed &&
		c.ProjectHealthReviewed &&
		c.UserActivityReviewed &&
		c.SystemAlertsReviewed &&
		c.AIPerformanceReviewed &&
		c.BackupStatusReviewed &&
		c.SecurityLogsReviewed
}
