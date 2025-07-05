package handlers

import (
	"net/http"
	"project-x/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ChatHandler struct {
	DB          *gorm.DB
	ChatService *services.ChatService
}

func NewChatHandler(db *gorm.DB, chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{
		DB:          db,
		ChatService: chatService,
	}
}

// GetOrCreateTeamChat creates or returns the main team chat room
func (h *ChatHandler) GetOrCreateTeamChat(c *gin.Context) {
	userID, _ := c.Get("userID")

	// Try to get existing team chat
	room, err := h.ChatService.GetTeamChatRoom()
	if err != nil {
		// Create team chat if it doesn't exist
		room, err = h.ChatService.CreateTeamChat(
			"Team Chat",
			"Main chat room for all team members",
			userID.(uint),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team chat"})
			return
		}
	}

	// Auto-join user to team chat
	h.ChatService.JoinTeamChat(room.ID, userID.(uint))

	c.JSON(http.StatusOK, gin.H{
		"message": "Team chat ready",
		"room": gin.H{
			"id":          room.ID,
			"name":        room.Name,
			"description": room.Description,
			"created_at":  room.CreatedAt,
		},
	})
}

// SendMessage sends a message to team chat
func (h *ChatHandler) SendMessage(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	var request struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	message, err := h.ChatService.SendMessage(uint(roomID), userID.(uint), request.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Message sent successfully",
		"chat_message": gin.H{
			"id":         message.ID,
			"content":    message.Content,
			"sender":     message.Sender.Username,
			"created_at": message.CreatedAt,
		},
	})
}

// GetMessages retrieves messages from team chat
func (h *ChatHandler) GetMessages(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	userID, _ := c.Get("userID")

	messages, err := h.ChatService.GetMessages(uint(roomID), userID.(uint), page, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var messageList []gin.H
	for _, message := range messages {
		messageList = append(messageList, gin.H{
			"id":         message.ID,
			"content":    message.Content,
			"sender":     message.Sender.Username,
			"created_at": message.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messageList,
		"page":     page,
		"limit":    limit,
	})
}

// GetRoomMembers gets all team chat members
func (h *ChatHandler) GetRoomMembers(c *gin.Context) {
	roomID, err := strconv.ParseUint(c.Param("roomId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	members, err := h.ChatService.GetRoomMembers(uint(roomID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get room members"})
		return
	}

	var memberList []gin.H
	for _, member := range members {
		memberList = append(memberList, gin.H{
			"id":       member.ID,
			"username": member.Username,
			"role":     member.Role,
		})
	}

	c.JSON(http.StatusOK, gin.H{"members": memberList})
}
