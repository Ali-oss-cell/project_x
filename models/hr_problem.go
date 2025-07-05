package models

import (
	"time"

	"gorm.io/gorm"
)

// ProblemCategory represents different types of workplace problems
type ProblemCategory string

const (
	ProblemCategoryHarassment      ProblemCategory = "harassment"
	ProblemCategoryDiscrimination  ProblemCategory = "discrimination"
	ProblemCategoryWorkEnvironment ProblemCategory = "work_environment"
	ProblemCategoryManagement      ProblemCategory = "management_issues"
	ProblemCategoryWorkload        ProblemCategory = "workload"
	ProblemCategorySafety          ProblemCategory = "safety_concerns"
	ProblemCategoryPayroll         ProblemCategory = "payroll_issues"
	ProblemCategoryBenefits        ProblemCategory = "benefits"
	ProblemCategoryConflict        ProblemCategory = "workplace_conflict"
	ProblemCategoryPolicyViolation ProblemCategory = "policy_violation"
	ProblemCategoryEquipment       ProblemCategory = "equipment_issues"
	ProblemCategoryTraining        ProblemCategory = "training_needs"
	ProblemCategoryOther           ProblemCategory = "other"
)

// ProblemStatus represents the status of a problem report
type ProblemStatus string

const (
	ProblemStatusPending    ProblemStatus = "pending"
	ProblemStatusReviewing  ProblemStatus = "reviewing"
	ProblemStatusInProgress ProblemStatus = "in_progress"
	ProblemStatusResolved   ProblemStatus = "resolved"
	ProblemStatusClosed     ProblemStatus = "closed"
	ProblemStatusRejected   ProblemStatus = "rejected"
)

// ProblemPriority represents the priority level of a problem
type ProblemPriority string

const (
	ProblemPriorityLow       ProblemPriority = "low"
	ProblemPriorityMedium    ProblemPriority = "medium"
	ProblemPriorityHigh      ProblemPriority = "high"
	ProblemPriorityUrgent    ProblemPriority = "urgent"
	ProblemPritorityCritical ProblemPriority = "critical"
)

// HRProblem represents a workplace problem reported by a user
type HRProblem struct {
	gorm.Model
	Title           string          `gorm:"not null;type:varchar(500) COLLATE \"default\""`
	Description     string          `gorm:"not null;type:text"`
	Category        ProblemCategory `gorm:"not null;index;type:varchar(100)"`
	Priority        ProblemPriority `gorm:"default:'medium';index;type:varchar(50)"`
	Status          ProblemStatus   `gorm:"default:'pending';index;type:varchar(50)"`
	ReporterID      uint            `gorm:"not null;index"`
	AssignedHRID    *uint           `gorm:"index"` // HR person handling this
	IsAnonymous     bool            `gorm:"default:false"`
	IsUrgent        bool            `gorm:"default:false;index"`
	ReportedAt      time.Time       `gorm:"not null;index"`
	ResolvedAt      *time.Time      `gorm:"index"`
	HRNotes         string          `gorm:"type:text"`                             // HR internal notes
	Resolution      string          `gorm:"type:text"`                             // Final resolution description
	FollowUpDate    *time.Time      `gorm:"index"`                                 // When to follow up
	AttachmentPath  string          `gorm:"type:varchar(500)"`                     // Path to uploaded files
	ContactMethod   string          `gorm:"type:varchar(100);default:'email'"`     // How user wants to be contacted
	PhoneNumber     string          `gorm:"type:varchar(20)"`                      // Optional phone for urgent issues
	PreferredTime   string          `gorm:"type:varchar(100)"`                     // Preferred contact time
	WitnessInfo     string          `gorm:"type:text"`                             // Witness information if any
	PreviousReports bool            `gorm:"default:false"`                         // Has this been reported before
	Location        string          `gorm:"type:varchar(255) COLLATE \"default\""` // Where the incident occurred
	IncidentDate    *time.Time      `gorm:"index"`                                 // When the incident occurred

	// Relationships
	Reporter   User               `gorm:"foreignKey:ReporterID;constraint:OnDelete:CASCADE"`
	AssignedHR *User              `gorm:"foreignKey:AssignedHRID;constraint:OnDelete:SET NULL"`
	Comments   []HRProblemComment `gorm:"foreignKey:ProblemID;constraint:OnDelete:CASCADE"`
	Updates    []HRProblemUpdate  `gorm:"foreignKey:ProblemID;constraint:OnDelete:CASCADE"`
}

// HRProblemComment represents comments on a problem report
type HRProblemComment struct {
	gorm.Model
	ProblemID uint   `gorm:"not null;index"`
	UserID    uint   `gorm:"not null;index"`
	Comment   string `gorm:"not null;type:text"`
	IsHROnly  bool   `gorm:"default:false"` // Only HR can see this comment

	// Relationships
	Problem HRProblem `gorm:"foreignKey:ProblemID;constraint:OnDelete:CASCADE"`
	User    User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// HRProblemUpdate represents status updates on a problem report
type HRProblemUpdate struct {
	gorm.Model
	ProblemID   uint          `gorm:"not null;index"`
	UpdatedBy   uint          `gorm:"not null;index"`
	OldStatus   ProblemStatus `gorm:"type:varchar(50)"`
	NewStatus   ProblemStatus `gorm:"type:varchar(50)"`
	UpdateNote  string        `gorm:"type:text"`
	IsAutomatic bool          `gorm:"default:false"` // System-generated update

	// Relationships
	Problem       HRProblem `gorm:"foreignKey:ProblemID;constraint:OnDelete:CASCADE"`
	UpdatedByUser User      `gorm:"foreignKey:UpdatedBy;constraint:OnDelete:CASCADE"`
}

// GetProblemCategories returns all available problem categories with descriptions
func GetProblemCategories() map[ProblemCategory]string {
	return map[ProblemCategory]string{
		ProblemCategoryHarassment:      "Harassment or bullying in the workplace",
		ProblemCategoryDiscrimination:  "Discrimination based on race, gender, age, religion, etc.",
		ProblemCategoryWorkEnvironment: "Issues with workplace conditions or environment",
		ProblemCategoryManagement:      "Problems with management or supervision",
		ProblemCategoryWorkload:        "Excessive workload or unrealistic expectations",
		ProblemCategorySafety:          "Safety concerns or unsafe working conditions",
		ProblemCategoryPayroll:         "Payroll, salary, or compensation issues",
		ProblemCategoryBenefits:        "Problems with benefits, insurance, or leave",
		ProblemCategoryConflict:        "Conflicts with colleagues or team members",
		ProblemCategoryPolicyViolation: "Company policy violations or concerns",
		ProblemCategoryEquipment:       "Equipment, tools, or technology issues",
		ProblemCategoryTraining:        "Training needs or development concerns",
		ProblemCategoryOther:           "Other workplace issues not listed above",
	}
}

// GetProblemPriorities returns all available priorities with descriptions
func GetProblemPriorities() map[ProblemPriority]string {
	return map[ProblemPriority]string{
		ProblemPriorityLow:       "Low priority - can be addressed in normal timeframe",
		ProblemPriorityMedium:    "Medium priority - should be addressed within a week",
		ProblemPriorityHigh:      "High priority - needs attention within 2-3 days",
		ProblemPriorityUrgent:    "Urgent - requires immediate attention within 24 hours",
		ProblemPritorityCritical: "Critical - emergency situation requiring immediate action",
	}
}
