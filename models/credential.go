package models

import (
	"time"

	"gorm.io/gorm"
)

// Credential represents a stored credential (password, email, username) for external platforms
type Credential struct {
	gorm.Model
	Platform       string     `gorm:"not null;index;type:varchar(255) COLLATE \"default\""` // "Instagram", "Stripe", "GitHub", etc.
	Email          string     `gorm:"not null;type:text"`                                   // Encrypted email
	Username       string     `gorm:"type:text"`                                            // Encrypted username (optional)
	Password       string     `gorm:"not null;type:text"`                                   // Encrypted password
	URL            string     `gorm:"type:varchar(500)"`                                    // Platform URL (optional)
	Notes          string     `gorm:"type:text"`                                            // Optional notes
	CreatedByID    uint       `gorm:"not null;index"`                                       // Admin who created it
	LastAccessedAt *time.Time `gorm:"index"`                                                // Last time anyone accessed
	IsActive       bool       `gorm:"default:true;index"`                                   // Can be deactivated

	// Relationships
	CreatedBy  User                  `gorm:"foreignKey:CreatedByID;constraint:OnDelete:SET NULL"`
	SharedWith []CredentialShare     `gorm:"foreignKey:CredentialID;constraint:OnDelete:CASCADE"`
	AccessLogs []CredentialAccessLog `gorm:"foreignKey:CredentialID;constraint:OnDelete:CASCADE"`
}

// CredentialShare represents sharing permissions for a credential
type CredentialShare struct {
	gorm.Model
	CredentialID uint      `gorm:"not null;index"`
	UserID       uint      `gorm:"not null;index"`
	SharedByID   uint      `gorm:"not null;index"` // Admin who shared it
	SharedAt     time.Time `gorm:"not null;index"`
	CanView      bool      `gorm:"default:true"` // Can view credentials
	CanCopy      bool      `gorm:"default:true"` // Can copy password
	Notes        string    `gorm:"type:text"`    // Optional sharing notes

	// Relationships
	Credential Credential `gorm:"foreignKey:CredentialID;constraint:OnDelete:CASCADE"`
	User       User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	SharedBy   User       `gorm:"foreignKey:SharedByID;constraint:OnDelete:SET NULL"`
}

// CredentialAccessLog represents an audit log entry for credential access
type CredentialAccessLog struct {
	gorm.Model
	CredentialID uint      `gorm:"not null;index"`
	UserID       uint      `gorm:"not null;index"`
	Action       string    `gorm:"not null;index;type:varchar(50)"` // "view", "copy", "decrypt"
	IPAddress    string    `gorm:"type:varchar(45)"`                // User's IP address
	UserAgent    string    `gorm:"type:text"`                       // Browser/client info
	AccessedAt   time.Time `gorm:"not null;index"`

	// Relationships
	Credential Credential `gorm:"foreignKey:CredentialID;constraint:OnDelete:CASCADE"`
	User       User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// CredentialDetails represents decrypted credential data (for API responses)
type CredentialDetails struct {
	ID             uint       `json:"id"`
	Platform       string     `json:"platform"`
	Email          string     `json:"email"`
	Username       string     `json:"username"`
	Password       string     `json:"password"`
	URL            string     `json:"url"`
	Notes          string     `json:"notes"`
	CreatedAt      time.Time  `json:"created_at"`
	LastAccessedAt *time.Time `json:"last_accessed_at"`
	IsActive       bool       `json:"is_active"`
	SharedWith     []uint     `json:"shared_with"` // User IDs who have access
}
