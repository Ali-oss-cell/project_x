package services

import (
	"errors"
	"project-x/models"
	"time"

	"gorm.io/gorm"
)

type PasswordManagerService struct {
	DB         *gorm.DB
	Encryption *EncryptionService
}

func NewPasswordManagerService(db *gorm.DB) (*PasswordManagerService, error) {
	encryption, err := NewEncryptionService()
	if err != nil {
		return nil, err
	}

	return &PasswordManagerService{
		DB:         db,
		Encryption: encryption,
	}, nil
}

// CreateCredential creates a new credential (Admin only)
func (s *PasswordManagerService) CreateCredential(adminID uint, platform, email, username, password, url, notes string) (*models.Credential, error) {
	// Verify admin role
	var admin models.User
	if err := s.DB.First(&admin, adminID).Error; err != nil {
		return nil, errors.New("admin not found")
	}
	if admin.Role != models.RoleAdmin {
		return nil, errors.New("only admin can create credentials")
	}

	// Encrypt sensitive data
	encryptedEmail, err := s.Encryption.Encrypt(email)
	if err != nil {
		return nil, errors.New("failed to encrypt email")
	}

	encryptedPassword, err := s.Encryption.Encrypt(password)
	if err != nil {
		return nil, errors.New("failed to encrypt password")
	}

	var encryptedUsername string
	if username != "" {
		encryptedUsername, err = s.Encryption.Encrypt(username)
		if err != nil {
			return nil, errors.New("failed to encrypt username")
		}
	}

	// Create credential
	credential := &models.Credential{
		Platform:    platform,
		Email:       encryptedEmail,
		Username:    encryptedUsername,
		Password:    encryptedPassword,
		URL:         url,
		Notes:       notes,
		CreatedByID: adminID,
		IsActive:    true,
	}

	if err := s.DB.Create(credential).Error; err != nil {
		return nil, err
	}

	return credential, nil
}

// UpdateCredential updates a credential (Admin only)
func (s *PasswordManagerService) UpdateCredential(credentialID, adminID uint, updates map[string]interface{}) error {
	// Verify admin role
	var admin models.User
	if err := s.DB.First(&admin, adminID).Error; err != nil {
		return errors.New("admin not found")
	}
	if admin.Role != models.RoleAdmin {
		return errors.New("only admin can update credentials")
	}

	// Get credential
	var credential models.Credential
	if err := s.DB.First(&credential, credentialID).Error; err != nil {
		return errors.New("credential not found")
	}

	// Encrypt sensitive fields if they're being updated
	if email, ok := updates["email"].(string); ok && email != "" {
		encryptedEmail, err := s.Encryption.Encrypt(email)
		if err != nil {
			return errors.New("failed to encrypt email")
		}
		updates["email"] = encryptedEmail
	}

	if password, ok := updates["password"].(string); ok && password != "" {
		encryptedPassword, err := s.Encryption.Encrypt(password)
		if err != nil {
			return errors.New("failed to encrypt password")
		}
		updates["password"] = encryptedPassword
	}

	if username, ok := updates["username"].(string); ok {
		if username == "" {
			updates["username"] = ""
		} else {
			encryptedUsername, err := s.Encryption.Encrypt(username)
			if err != nil {
				return errors.New("failed to encrypt username")
			}
			updates["username"] = encryptedUsername
		}
	}

	// Update credential
	return s.DB.Model(&credential).Updates(updates).Error
}

// DeleteCredential deletes a credential (Admin only)
func (s *PasswordManagerService) DeleteCredential(credentialID, adminID uint) error {
	// Verify admin role
	var admin models.User
	if err := s.DB.First(&admin, adminID).Error; err != nil {
		return errors.New("admin not found")
	}
	if admin.Role != models.RoleAdmin {
		return errors.New("only admin can delete credentials")
	}

	// Get credential
	var credential models.Credential
	if err := s.DB.First(&credential, credentialID).Error; err != nil {
		return errors.New("credential not found")
	}

	// Delete credential (shares and logs will be deleted due to CASCADE)
	return s.DB.Delete(&credential).Error
}

// ShareCredential shares a credential with users (Admin only)
func (s *PasswordManagerService) ShareCredential(credentialID, adminID uint, userIDs []uint, canView, canCopy bool, notes string) error {
	// Verify admin role
	var admin models.User
	if err := s.DB.First(&admin, adminID).Error; err != nil {
		return errors.New("admin not found")
	}
	if admin.Role != models.RoleAdmin {
		return errors.New("only admin can share credentials")
	}

	// Get credential
	var credential models.Credential
	if err := s.DB.First(&credential, credentialID).Error; err != nil {
		return errors.New("credential not found")
	}

	// Share with each user
	for _, userID := range userIDs {
		// Check if user exists
		var user models.User
		if err := s.DB.First(&user, userID).Error; err != nil {
			continue // Skip non-existent users
		}

		// Check if already shared
		var existingShare models.CredentialShare
		err := s.DB.Where("credential_id = ? AND user_id = ?", credentialID, userID).First(&existingShare).Error
		if err == nil {
			// Update existing share
			existingShare.CanView = canView
			existingShare.CanCopy = canCopy
			existingShare.Notes = notes
			existingShare.SharedByID = adminID
			existingShare.SharedAt = time.Now()
			s.DB.Save(&existingShare)
		} else {
			// Create new share
			share := &models.CredentialShare{
				CredentialID: credentialID,
				UserID:       userID,
				SharedByID:   adminID,
				SharedAt:     time.Now(),
				CanView:      canView,
				CanCopy:      canCopy,
				Notes:        notes,
			}
			s.DB.Create(share)
		}
	}

	return nil
}

// UnshareCredential removes sharing permission (Admin only)
func (s *PasswordManagerService) UnshareCredential(credentialID, userID, adminID uint) error {
	// Verify admin role
	var admin models.User
	if err := s.DB.First(&admin, adminID).Error; err != nil {
		return errors.New("admin not found")
	}
	if admin.Role != models.RoleAdmin {
		return errors.New("only admin can unshare credentials")
	}

	// Delete share
	return s.DB.Where("credential_id = ? AND user_id = ?", credentialID, userID).Delete(&models.CredentialShare{}).Error
}

// GetMyCredentials returns credentials shared with the current user
func (s *PasswordManagerService) GetMyCredentials(userID uint) ([]models.Credential, error) {
	var credentials []models.Credential

	// Get credentials shared with this user
	err := s.DB.
		Joins("JOIN credential_shares ON credentials.id = credential_shares.credential_id").
		Where("credential_shares.user_id = ? AND credentials.is_active = ?", userID, true).
		Preload("CreatedBy").
		Preload("SharedWith").
		Find(&credentials).Error

	return credentials, err
}

// GetCredentialDetails decrypts and returns credential details (with access logging)
func (s *PasswordManagerService) GetCredentialDetails(credentialID, userID uint, ipAddress, userAgent string) (*models.CredentialDetails, error) {
	// Get credential
	var credential models.Credential
	if err := s.DB.Preload("SharedWith").First(&credential, credentialID).Error; err != nil {
		return nil, errors.New("credential not found")
	}

	// Check if user has access
	var hasAccess bool
	if credential.CreatedByID == userID {
		hasAccess = true // Creator always has access
	} else {
		// Check if shared with user
		var share models.CredentialShare
		err := s.DB.Where("credential_id = ? AND user_id = ?", credentialID, userID).First(&share).Error
		hasAccess = (err == nil && share.CanView)
	}

	if !hasAccess {
		return nil, errors.New("access denied")
	}

	// Decrypt sensitive data
	email, err := s.Encryption.Decrypt(credential.Email)
	if err != nil {
		return nil, errors.New("failed to decrypt email")
	}

	password, err := s.Encryption.Decrypt(credential.Password)
	if err != nil {
		return nil, errors.New("failed to decrypt password")
	}

	var username string
	if credential.Username != "" {
		username, err = s.Encryption.Decrypt(credential.Username)
		if err != nil {
			return nil, errors.New("failed to decrypt username")
		}
	}

	// Get shared user IDs
	var sharedUserIDs []uint
	for _, share := range credential.SharedWith {
		sharedUserIDs = append(sharedUserIDs, share.UserID)
	}

	// Log access
	now := time.Now()
	accessLog := &models.CredentialAccessLog{
		CredentialID: credentialID,
		UserID:       userID,
		Action:       "view",
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		AccessedAt:   now,
	}
	s.DB.Create(accessLog)

	// Update last accessed time
	s.DB.Model(&credential).Update("last_accessed_at", now)

	// Return decrypted details
	return &models.CredentialDetails{
		ID:             credential.ID,
		Platform:       credential.Platform,
		Email:          email,
		Username:       username,
		Password:       password,
		URL:            credential.URL,
		Notes:          credential.Notes,
		CreatedAt:      credential.CreatedAt,
		LastAccessedAt: &now,
		IsActive:       credential.IsActive,
		SharedWith:     sharedUserIDs,
	}, nil
}

// CopyPassword logs a password copy action
func (s *PasswordManagerService) CopyPassword(credentialID, userID uint, ipAddress, userAgent string) error {
	// Check if user has copy permission
	var credential models.Credential
	if err := s.DB.First(&credential, credentialID).Error; err != nil {
		return errors.New("credential not found")
	}

	var hasCopyPermission bool
	if credential.CreatedByID == userID {
		hasCopyPermission = true // Creator always has permission
	} else {
		var share models.CredentialShare
		err := s.DB.Where("credential_id = ? AND user_id = ?", credentialID, userID).First(&share).Error
		hasCopyPermission = (err == nil && share.CanCopy)
	}

	if !hasCopyPermission {
		return errors.New("copy permission denied")
	}

	// Log copy action
	accessLog := &models.CredentialAccessLog{
		CredentialID: credentialID,
		UserID:       userID,
		Action:       "copy",
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		AccessedAt:   time.Now(),
	}
	return s.DB.Create(accessLog).Error
}

// GetAllCredentials returns all credentials (Admin only)
func (s *PasswordManagerService) GetAllCredentials(adminID uint) ([]models.Credential, error) {
	// Verify admin role
	var admin models.User
	if err := s.DB.First(&admin, adminID).Error; err != nil {
		return nil, errors.New("admin not found")
	}
	if admin.Role != models.RoleAdmin {
		return nil, errors.New("only admin can view all credentials")
	}

	var credentials []models.Credential
	err := s.DB.
		Preload("CreatedBy").
		Preload("SharedWith.User").
		Order("created_at DESC").
		Find(&credentials).Error

	return credentials, err
}

// GetAccessLogs returns access logs for a credential (Admin only)
func (s *PasswordManagerService) GetAccessLogs(credentialID, adminID uint) ([]models.CredentialAccessLog, error) {
	// Verify admin role
	var admin models.User
	if err := s.DB.First(&admin, adminID).Error; err != nil {
		return nil, errors.New("admin not found")
	}
	if admin.Role != models.RoleAdmin {
		return nil, errors.New("only admin can view access logs")
	}

	var logs []models.CredentialAccessLog
	err := s.DB.
		Where("credential_id = ?", credentialID).
		Preload("User").
		Order("accessed_at DESC").
		Find(&logs).Error

	return logs, err
}
