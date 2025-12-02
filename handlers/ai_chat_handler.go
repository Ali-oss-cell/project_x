package handlers

import (
	"net/http"
	"project-x/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AIChatHandler struct {
	DB            *gorm.DB
	AIChatService *services.AIChatService
}

func NewAIChatHandler(db *gorm.DB, aiChatService *services.AIChatService) *AIChatHandler {
	return &AIChatHandler{
		DB:            db,
		AIChatService: aiChatService,
	}
}

// GetOrCreateAIChatRoom gets or creates a private AI chat room for the user
func (h *AIChatHandler) GetOrCreateAIChatRoom(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	room, err := h.AIChatService.GetOrCreateAIChatRoom(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get or create AI chat room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "AI chat room ready",
		"room": gin.H{
			"id":          room.ID,
			"name":        room.Name,
			"description": room.Description,
			"created_at":  room.CreatedAt,
		},
	})
}

// ProcessAIMessage processes an AI message and returns response
func (h *AIChatHandler) ProcessAIMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	var request struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process AI message
	aiResponse, err := h.AIChatService.ProcessAIMessage(userID.(uint), request.Message, uint(roomID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process AI message", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "AI response generated",
		"response": gin.H{
			"message":     aiResponse.Message,
			"action":      aiResponse.Action,
			"data":        aiResponse.Data,
			"suggestions": aiResponse.Suggestions,
			"language":    aiResponse.Language,
		},
	})
}
