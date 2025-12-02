package services

import (
	"errors"
	"fmt"
	"log"
	"project-x/models"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type ChatService struct {
	db               *gorm.DB
	websocketService *WebSocketService
	aiChatService    *AIChatService
}

func NewChatService(db *gorm.DB, wsService *WebSocketService, notificationService *NotificationService) *ChatService {
	return &ChatService{
		db:               db,
		websocketService: wsService,
	}
}

// SetAIChatService sets the AI chat service
func (cs *ChatService) SetAIChatService(aiChatService *AIChatService) {
	cs.aiChatService = aiChatService
}

// CreateTeamChat creates the main team chat room
func (cs *ChatService) CreateTeamChat(name, description string, createdBy uint) (*models.ChatRoom, error) {
	// Check if team chat already exists
	var existingRoom models.ChatRoom
	if err := cs.db.Where("name = ?", name).First(&existingRoom).Error; err == nil {
		return &existingRoom, nil // Return existing room
	}

	// Create new team chat room
	room := &models.ChatRoom{
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		MaxMembers:  1000,
	}

	if err := cs.db.Create(room).Error; err != nil {
		return nil, err
	}

	// Auto-join the creator
	cs.JoinTeamChat(room.ID, createdBy)

	return room, nil
}

// JoinTeamChat joins a user to the team chat
func (cs *ChatService) JoinTeamChat(roomID, userID uint) error {
	// Check if user is already in the room
	var participant models.ChatParticipant
	if err := cs.db.Where("chat_room_id = ? AND user_id = ?", roomID, userID).First(&participant).Error; err == nil {
		return nil // User already in room
	}

	// Add user to room
	participant = models.ChatParticipant{
		ChatRoomID: roomID,
		UserID:     userID,
		JoinedAt:   time.Now(),
	}

	return cs.db.Create(&participant).Error
}

// SendMessage sends a message to the team chat
func (cs *ChatService) SendMessage(roomID, senderID uint, content string) (*models.ChatMessage, error) {
	// Check if user is in the room
	var participant models.ChatParticipant
	if err := cs.db.Where("chat_room_id = ? AND user_id = ?", roomID, senderID).First(&participant).Error; err != nil {
		return nil, errors.New("user is not in the chat room")
	}

	// Create message
	message := &models.ChatMessage{
		ChatRoomID: roomID,
		SenderID:   senderID,
		Content:    content,
	}

	if err := cs.db.Create(message).Error; err != nil {
		return nil, err
	}

	// Load sender info
	cs.db.Preload("Sender").First(message, message.ID)

	// Broadcast message via WebSocket if available
	if cs.websocketService != nil {
		cs.websocketService.BroadcastToRoom(strconv.FormatUint(uint64(roomID), 10), Notification{
			Type:    "message",
			Title:   "New Message",
			Message: content,
			Data: map[string]interface{}{
				"message_id": message.ID,
				"room_id":    roomID,
				"sender":     message.Sender.Username,
				"content":    content,
			},
			Timestamp: time.Now(),
		})
	}

	// Check if this is an AI message and process it
	if cs.aiChatService != nil && cs.aiChatService.IsAIMessage(content, roomID) {
		go cs.processAIMessage(roomID, senderID, content)
	}

	return message, nil
}

// processAIMessage processes AI messages asynchronously
func (cs *ChatService) processAIMessage(roomID, senderID uint, content string) {
	// Process AI message
	aiResponse, err := cs.aiChatService.ProcessAIMessage(senderID, content, roomID)
	if err != nil {
		log.Printf("Error processing AI message: %v", err)
		return
	}

	// Create AI user (system user) for AI messages
	// We'll use a special sender ID (0) or create a system user
	// For now, we'll create a message with a special marker
	aiMessage := &models.ChatMessage{
		ChatRoomID: roomID,
		SenderID:   senderID, // Keep original sender for now, can be enhanced
		Content:    aiResponse.Message,
		Metadata:   fmt.Sprintf(`{"ai_response":true,"action":"%s"}`, aiResponse.Action),
	}

	if err := cs.db.Create(aiMessage).Error; err != nil {
		log.Printf("Error saving AI response: %v", err)
		return
	}

	// Broadcast AI response via WebSocket
	if cs.websocketService != nil {
		cs.websocketService.BroadcastToRoom(strconv.FormatUint(uint64(roomID), 10), Notification{
			Type:    "ai_message",
			Title:   "AI Assistant",
			Message: aiResponse.Message,
			Data: map[string]interface{}{
				"message_id": aiMessage.ID,
				"room_id":    roomID,
				"sender":     "AI Assistant",
				"content":    aiResponse.Message,
				"action":     aiResponse.Action,
				"data":       aiResponse.Data,
			},
			Timestamp: time.Now(),
		})
	}
}

// GetMessages retrieves messages from the team chat
func (cs *ChatService) GetMessages(roomID, userID uint, page, limit int) ([]models.ChatMessage, error) {
	// Check if user is in the room
	var participant models.ChatParticipant
	if err := cs.db.Where("chat_room_id = ? AND user_id = ?", roomID, userID).First(&participant).Error; err != nil {
		return nil, errors.New("user is not in the chat room")
	}

	var messages []models.ChatMessage
	offset := (page - 1) * limit

	err := cs.db.Where("chat_room_id = ?", roomID).
		Preload("Sender").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error

	return messages, err
}

// GetTeamChatRoom gets the main team chat room
func (cs *ChatService) GetTeamChatRoom() (*models.ChatRoom, error) {
	var room models.ChatRoom
	err := cs.db.Where("name = ?", "Team Chat").First(&room).Error
	return &room, err
}

// GetRoomMembers gets all members of the team chat
func (cs *ChatService) GetRoomMembers(roomID uint) ([]models.User, error) {
	var users []models.User
	err := cs.db.Table("users").
		Joins("JOIN chat_participants ON users.id = chat_participants.user_id").
		Where("chat_participants.chat_room_id = ?", roomID).
		Find(&users).Error
	return users, err
}
