package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleEmployee Role = "employee"
	RoleManager  Role = "manager"
	RoleHead     Role = "head"
	RoleHR       Role = "hr"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type ProjectStatus string

const (
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusPaused    ProjectStatus = "paused"
	ProjectStatusCompleted ProjectStatus = "completed"
	ProjectStatusCancelled ProjectStatus = "cancelled"
)

type User struct {
	gorm.Model
	Username   string     `gorm:"unique;not null;index;type:varchar(255) COLLATE \"default\""`
	Password   string     `gorm:"not null;type:varchar(255)"`
	Role       Role       `gorm:"not null;index;type:varchar(50)"`
	Department string     `gorm:"not null;index;type:varchar(255) COLLATE \"default\""`
	Skills     string     `gorm:"type:json"` // JSON array of skills: ["UX/UI Design", "Frontend Development", etc.]
	LastLogin  *time.Time `gorm:"index"`     // Last login timestamp

	// Relationships
	Tasks              []Task              `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	CollaborativeTasks []CollaborativeTask `gorm:"foreignKey:LeadUserID;constraint:OnDelete:CASCADE"`   // Tasks where user is the lead
	Projects           []Project           `gorm:"many2many:user_projects;constraint:OnDelete:CASCADE"` // Many-to-many relationship
	CreatedProjects    []Project           `gorm:"foreignKey:CreatedBy;constraint:OnDelete:SET NULL"`   // Projects created by this user

	// Collaborative task participation
	CollaborativeTaskParticipations []CollaborativeTaskParticipant `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type Task struct {
	gorm.Model
	Title       string     `gorm:"not null;index;type:varchar(500) COLLATE \"default\""`
	Description string     `gorm:"not null;type:text"`
	Status      TaskStatus `gorm:"not null;default:'pending';index;type:varchar(50)"`
	UserID      uint       `gorm:"not null;index"`
	ProjectID   *uint      `gorm:"index"` // Optional: task can belong to a project
	AssignedAt  time.Time  `gorm:"not null;index"`
	StartTime   *time.Time `gorm:"index"` // When task should start
	EndTime     *time.Time `gorm:"index"` // When task should end
	DueDate     *time.Time `gorm:"index"` // Optional due date (legacy, can be removed later)

	// Relationships
	User    User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Project *Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:SET NULL"`
}

type CollaborativeTask struct {
	gorm.Model
	Title           string     `gorm:"not null;index;type:varchar(500) COLLATE \"default\""`
	Description     string     `gorm:"not null;type:text"`
	Status          TaskStatus `gorm:"not null;default:'pending';index;type:varchar(50)"`
	LeadUserID      uint       `gorm:"not null;index"` // Main person responsible
	ProjectID       *uint      `gorm:"index"`          // Optional: collaborative task can belong to a project
	AssignedAt      time.Time  `gorm:"not null;index"`
	StartTime       *time.Time `gorm:"index"`                                   // When task should start
	EndTime         *time.Time `gorm:"index"`                                   // When task should end
	DueDate         *time.Time `gorm:"index"`                                   // Optional due date (legacy, can be removed later)
	Priority        string     `gorm:"default:'medium';index;type:varchar(50)"` // high, medium, low
	Progress        int        `gorm:"default:0;index"`                         // 0-100 percentage
	Complexity      string     `gorm:"default:'medium';index;type:varchar(50)"` // simple, medium, complex
	MaxParticipants int        `gorm:"default:5;index"`                         // Maximum number of participants

	// Relationships
	LeadUser     User                           `gorm:"foreignKey:LeadUserID;constraint:OnDelete:SET NULL"`
	Project      *Project                       `gorm:"foreignKey:ProjectID;constraint:OnDelete:SET NULL"`
	Participants []CollaborativeTaskParticipant `gorm:"foreignKey:CollaborativeTaskID;constraint:OnDelete:CASCADE"`
}

// CollaborativeTaskParticipant represents team members working on a collaborative task
type CollaborativeTaskParticipant struct {
	gorm.Model
	CollaborativeTaskID uint       `gorm:"not null;index"`
	UserID              uint       `gorm:"not null;index"`
	Role                string     `gorm:"not null;default:'contributor';index;type:varchar(100)"` // lead, contributor, reviewer, observer
	Status              string     `gorm:"not null;default:'active';index;type:varchar(50)"`       // active, inactive, completed
	AssignedAt          time.Time  `gorm:"not null;index"`
	CompletedAt         *time.Time `gorm:"index"`
	Contribution        string     `gorm:"index;type:text"` // Description of their contribution

	// Relationships
	CollaborativeTask CollaborativeTask `gorm:"foreignKey:CollaborativeTaskID;constraint:OnDelete:CASCADE"`
	User              User              `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type Project struct {
	gorm.Model
	Title       string        `gorm:"not null;index;type:varchar(500) COLLATE \"default\""`
	Description string        `gorm:"not null;type:text"`
	Status      ProjectStatus `gorm:"not null;default:'active';index;type:varchar(50)"`
	CreatedBy   uint          `gorm:"not null;index"` // Who created the project
	StartDate   time.Time     `gorm:"not null;index"`
	EndDate     *time.Time    `gorm:"index"` // Optional end date

	// Relationships
	Creator            User                `gorm:"foreignKey:CreatedBy;constraint:OnDelete:SET NULL"`
	Users              []User              `gorm:"many2many:user_projects;constraint:OnDelete:CASCADE"` // Many-to-many relationship
	Tasks              []Task              `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	CollaborativeTasks []CollaborativeTask `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

// UserProject represents the many-to-many relationship between users and projects
type UserProject struct {
	UserID    uint      `gorm:"primaryKey;index"`
	ProjectID uint      `gorm:"primaryKey;index"`
	JoinedAt  time.Time `gorm:"not null;index"`
	Role      string    `gorm:"not null;default:'member';index;type:varchar(100)"` // Project management role: manager, head, employee, member
	JobRole   string    `gorm:"type:varchar(100)"`                                 // Project-specific job role: "UX/UI Designer", "Backend Developer", etc.

	// Relationships
	User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Project Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

type Claims struct {
	UserID uint `json:"userId"`
	Role   Role `json:"role"`
	jwt.RegisteredClaims
}
