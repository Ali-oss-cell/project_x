package models

import (
	"time"

	"gorm.io/gorm"
)

// AIAnalysis stores AI predictions and tracks their accuracy for regular tasks
type AIAnalysis struct {
	gorm.Model
	TaskID   uint   `gorm:"not null;index"`
	TaskType string `gorm:"not null;type:varchar(50)"` // "regular" or "collaborative"

	// AI Predictions (saved when analysis is done)
	PredictedDuration         int       `gorm:"not null"` // hours
	PredictedCompletion       time.Time `gorm:"not null"`
	DeadlineRisk              string    `gorm:"not null;type:varchar(50)"` // low/medium/high/critical
	RiskFactors               string    `gorm:"type:json"`                 // JSON array of strings
	Recommendations           string    `gorm:"type:json"`                 // JSON array of strings
	OptimalStartDate          time.Time `gorm:"not null"`
	WorkingHoursUntilDeadline float64   `gorm:"default:0"`

	// Actual Results (updated when task completes)
	ActualDuration     *int       `gorm:"default:null"` // hours (null until completed)
	ActualCompletion   *time.Time `gorm:"default:null"`
	WasAccurate        *bool      `gorm:"default:null"` // true if prediction was close
	AccuracyScore      *float64   `gorm:"default:null"` // 0-100, how close prediction was
	DurationDifference *int       `gorm:"default:null"` // actual - predicted (hours)

	// Metadata
	AnalysisDate    time.Time `gorm:"not null;index"`
	ModelVersion    string    `gorm:"default:'gemini-2.0-flash';type:varchar(100)"`
	ConfidenceScore int       `gorm:"default:0"` // 0-100

	// Relationships
	Task Task `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`
}

// CollaborativeAIAnalysis stores AI predictions for collaborative tasks
type CollaborativeAIAnalysis struct {
	gorm.Model
	CollaborativeTaskID uint `gorm:"not null;index"`

	// AI Predictions
	PredictedDuration         int       `gorm:"not null"`
	PredictedCompletion       time.Time `gorm:"not null"`
	DeadlineRisk              string    `gorm:"not null;type:varchar(50)"`
	RiskFactors               string    `gorm:"type:json"`
	Recommendations           string    `gorm:"type:json"`
	OptimalStartDate          time.Time `gorm:"not null"`
	WorkingHoursUntilDeadline float64   `gorm:"default:0"`

	// Actual Results
	ActualDuration     *int
	ActualCompletion   *time.Time
	WasAccurate        *bool
	AccuracyScore      *float64
	DurationDifference *int

	// Metadata
	AnalysisDate    time.Time `gorm:"not null;index"`
	ModelVersion    string    `gorm:"default:'gemini-2.0-flash';type:varchar(100)"`
	ConfidenceScore int       `gorm:"default:0"`

	// Relationships
	CollaborativeTask CollaborativeTask `gorm:"foreignKey:CollaborativeTaskID;constraint:OnDelete:CASCADE"`
}
