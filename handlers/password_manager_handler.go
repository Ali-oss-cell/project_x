package handlers

import (
	"net/http"
	"project-x/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PasswordManagerHandler struct {
	DB                     *gorm.DB
	PasswordManagerService *services.PasswordManagerService
}

func NewPasswordManagerHandler(db *gorm.DB) (*PasswordManagerHandler, error) {
	service, err := services.NewPasswordManagerService(db)
	if err != nil {
		return nil, err
	}

	return &PasswordManagerHandler{
		DB:                     db,
		PasswordManagerService: service,
	}, nil
}

// CreateCredential creates a new credential (Admin only)
func (h *PasswordManagerHandler) CreateCredential(c *gin.Context) {
	currentUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req struct {
		Platform string `json:"platform" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Username string `json:"username"`
		Password string `json:"password" binding:"required"`
		URL      string `json:"url"`
		Notes    string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	credential, err := h.PasswordManagerService.CreateCredential(
		currentUserID.(uint),
		req.Platform,
		req.Email,
		req.Username,
		req.Password,
		req.URL,
		req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Credential created successfully",
		"credential": credential,
	})
}

// GetAllCredentials returns all credentials (Admin only)
func (h *PasswordManagerHandler) GetAllCredentials(c *gin.Context) {
	currentUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	credentials, err := h.PasswordManagerService.GetAllCredentials(currentUserID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return credentials without decrypted passwords
	type CredentialResponse struct {
		ID             uint                     `json:"id"`
		Platform       string                   `json:"platform"`
		Email          string                   `json:"email"` // Still encrypted, just for display
		URL            string                   `json:"url"`
		Notes          string                   `json:"notes"`
		CreatedAt      string                   `json:"created_at"`
		LastAccessedAt *string                  `json:"last_accessed_at"`
		IsActive       bool                     `json:"is_active"`
		CreatedBy      map[string]interface{}   `json:"created_by"`
		SharedWith     []map[string]interface{} `json:"shared_with"`
	}

	var response []CredentialResponse
	for _, cred := range credentials {
		createdBy := map[string]interface{}{
			"id":       cred.CreatedBy.ID,
			"username": cred.CreatedBy.Username,
		}

		var sharedWith []map[string]interface{}
		for _, share := range cred.SharedWith {
			sharedWith = append(sharedWith, map[string]interface{}{
				"user_id":   share.UserID,
				"username":  share.User.Username,
				"can_view":  share.CanView,
				"can_copy":  share.CanCopy,
				"shared_at": share.SharedAt,
			})
		}

		lastAccessed := ""
		if cred.LastAccessedAt != nil {
			lastAccessed = cred.LastAccessedAt.Format("2006-01-02 15:04:05")
		}

		response = append(response, CredentialResponse{
			ID:             cred.ID,
			Platform:       cred.Platform,
			Email:          "***@***.***", // Masked for security
			URL:            cred.URL,
			Notes:          cred.Notes,
			CreatedAt:      cred.CreatedAt.Format("2006-01-02 15:04:05"),
			LastAccessedAt: &lastAccessed,
			IsActive:       cred.IsActive,
			CreatedBy:      createdBy,
			SharedWith:     sharedWith,
		})
	}

	c.JSON(http.StatusOK, gin.H{"credentials": response})
}

// GetCredentialDetails returns decrypted credential details (Admin or shared user)
func (h *PasswordManagerHandler) GetCredentialDetails(c *gin.Context) {
	currentUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	credentialID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	// Get IP address and user agent
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	details, err := h.PasswordManagerService.GetCredentialDetails(
		uint(credentialID),
		currentUserID.(uint),
		ipAddress,
		userAgent,
	)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"credential": details})
}

// UpdateCredential updates a credential (Admin only)
func (h *PasswordManagerHandler) UpdateCredential(c *gin.Context) {
	currentUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	credentialID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	var req struct {
		Platform string `json:"platform"`
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
		URL      string `json:"url"`
		Notes    string `json:"notes"`
		IsActive *bool  `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Platform != "" {
		updates["platform"] = req.Platform
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Username != "" || req.Username == "" {
		updates["username"] = req.Username
	}
	if req.Password != "" {
		updates["password"] = req.Password
	}
	if req.URL != "" {
		updates["url"] = req.URL
	}
	if req.Notes != "" {
		updates["notes"] = req.Notes
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	err = h.PasswordManagerService.UpdateCredential(uint(credentialID), currentUserID.(uint), updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Credential updated successfully"})
}

// DeleteCredential deletes a credential (Admin only)
func (h *PasswordManagerHandler) DeleteCredential(c *gin.Context) {
	currentUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	credentialID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	err = h.PasswordManagerService.DeleteCredential(uint(credentialID), currentUserID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Credential deleted successfully"})
}

// ShareCredential shares a credential with users (Admin only)
func (h *PasswordManagerHandler) ShareCredential(c *gin.Context) {
	currentUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	credentialID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	var req struct {
		UserIDs []uint `json:"user_ids" binding:"required"`
		CanView bool   `json:"can_view" binding:"required"`
		CanCopy bool   `json:"can_copy" binding:"required"`
		Notes   string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.PasswordManagerService.ShareCredential(
		uint(credentialID),
		currentUserID.(uint),
		req.UserIDs,
		req.CanView,
		req.CanCopy,
		req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Credential shared successfully"})
}

// UnshareCredential removes sharing permission (Admin only)
func (h *PasswordManagerHandler) UnshareCredential(c *gin.Context) {
	currentUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	credentialID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	err = h.PasswordManagerService.UnshareCredential(uint(credentialID), uint(userID), currentUserID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Credential unshared successfully"})
}

// GetMyCredentials returns credentials shared with the current user
func (h *PasswordManagerHandler) GetMyCredentials(c *gin.Context) {
	currentUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	credentials, err := h.PasswordManagerService.GetMyCredentials(currentUserID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return credentials without decrypted passwords
	type CredentialResponse struct {
		ID        uint   `json:"id"`
		Platform  string `json:"platform"`
		URL       string `json:"url"`
		Notes     string `json:"notes"`
		CreatedAt string `json:"created_at"`
		SharedBy  string `json:"shared_by"`
	}

	var response []CredentialResponse
	for _, cred := range credentials {
		response = append(response, CredentialResponse{
			ID:        cred.ID,
			Platform:  cred.Platform,
			URL:       cred.URL,
			Notes:     cred.Notes,
			CreatedAt: cred.CreatedAt.Format("2006-01-02 15:04:05"),
			SharedBy:  cred.CreatedBy.Username,
		})
	}

	c.JSON(http.StatusOK, gin.H{"credentials": response})
}

// CopyPassword logs a password copy action
func (h *PasswordManagerHandler) CopyPassword(c *gin.Context) {
	currentUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	credentialID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	// Get IP address and user agent
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	err = h.PasswordManagerService.CopyPassword(uint(credentialID), currentUserID.(uint), ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password copy logged"})
}

// GetAccessLogs returns access logs for a credential (Admin only)
func (h *PasswordManagerHandler) GetAccessLogs(c *gin.Context) {
	currentUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	credentialID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credential ID"})
		return
	}

	logs, err := h.PasswordManagerService.GetAccessLogs(uint(credentialID), currentUserID.(uint))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	type LogResponse struct {
		ID         uint   `json:"id"`
		UserID     uint   `json:"user_id"`
		Username   string `json:"username"`
		Action     string `json:"action"`
		IPAddress  string `json:"ip_address"`
		UserAgent  string `json:"user_agent"`
		AccessedAt string `json:"accessed_at"`
	}

	var response []LogResponse
	for _, log := range logs {
		response = append(response, LogResponse{
			ID:         log.ID,
			UserID:     log.UserID,
			Username:   log.User.Username,
			Action:     log.Action,
			IPAddress:  log.IPAddress,
			UserAgent:  log.UserAgent,
			AccessedAt: log.AccessedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"logs": response})
}
