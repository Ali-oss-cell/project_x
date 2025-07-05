package middleware

import (
	"net/http"
	"os"
	"project-x/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// AuthMiddleware validates JWT token and sets user info in context
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*models.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		// Get user from database
		var user models.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user", &user)
		c.Set("userID", user.ID)
		c.Set("username", user.Username)
		c.Set("userRole", user.Role)

		c.Next()
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(requiredRole models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		role := userRole.(models.Role)
		if !hasPermission(role, requiredRole) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasPermission checks if user role has permission for required role
// Role hierarchy: admin > manager > head > employee
// HR has special permissions separate from hierarchy
func hasPermission(userRole, requiredRole models.Role) bool {
	// Admin has access to everything
	if userRole == models.RoleAdmin {
		return true
	}

	// HR has special permissions - only specific HR functions
	if userRole == models.RoleHR {
		// HR can access user management, reports, but not task assignment
		return requiredRole == models.RoleHR || requiredRole == models.RoleEmployee
	}

	// Normal hierarchy for other roles
	roleHierarchy := map[models.Role]int{
		models.RoleEmployee: 1,
		models.RoleHead:     2,
		models.RoleManager:  3,
		models.RoleAdmin:    4,
	}

	userLevel := roleHierarchy[userRole]
	requiredLevel := roleHierarchy[requiredRole]

	return userLevel >= requiredLevel
}

// RequireAdmin middleware - only admin can access
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

// RequireManagerOrHigher middleware - manager or admin can access
func RequireManagerOrHigher() gin.HandlerFunc {
	return RequireRole(models.RoleManager)
}

// RequireHeadOrHigher middleware - head, manager, or admin can access
func RequireHeadOrHigher() gin.HandlerFunc {
	return RequireRole(models.RoleHead)
}

// RequireHROrAdmin middleware - only HR or Admin can access
func RequireHROrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		role := userRole.(models.Role)
		if role != models.RoleHR && role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "HR or Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireHROrHigher middleware - hr, manager, or admin can access
func RequireHROrHigher() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		role := userRole.(models.Role)
		if role != models.RoleHR && role != models.RoleManager && role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "HR, Manager, or Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IsHRRole checks if user has HR role
func IsHRRole(userRole models.Role) bool {
	return userRole == models.RoleHR
}

// IsManagerOrHigher checks if user has manager or higher permissions (excluding HR)
func IsManagerOrHigher(userRole models.Role) bool {
	return userRole == models.RoleManager || userRole == models.RoleAdmin
}

// RequireSelfOrAdmin middleware - user can access their own data or admin can access any
func RequireSelfOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		// Admin can access any user's data
		if userRole.(models.Role) == models.RoleAdmin {
			c.Next()
			return
		}

		// Get the requested user ID from URL
		requestedUserID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			c.Abort()
			return
		}

		// User can only access their own data
		if userID.(uint) != uint(requestedUserID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}
