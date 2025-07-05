package services

import (
	"encoding/json"
	"log"
	"net/http"
	"project-x/models"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// WebSocket connection manager
type WebSocketService struct {
	db        *gorm.DB
	upgrader  websocket.Upgrader
	clients   map[uint]*Client
	rooms     map[string]*Room
	mutex     sync.RWMutex
	broadcast chan Notification
}

// Client represents a connected user
type Client struct {
	ID       uint
	Username string
	Role     string
	Conn     *websocket.Conn
	Send     chan []byte
	Rooms    map[string]bool
	LastSeen time.Time
}

// Room represents a chat room for WebSocket connections
type Room struct {
	ID      uint
	Name    string
	Clients map[uint]*Client
	mutex   sync.RWMutex
}

// Notification represents a message to be sent
type Notification struct {
	Type      string      `json:"type"`
	Title     string      `json:"title"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	UserID    uint        `json:"user_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ChatMessage represents a WebSocket chat message
type WebSocketChatMessage struct {
	Type        string                 `json:"type"`
	Action      string                 `json:"action"`
	RoomID      uint                   `json:"room_id"`
	MessageID   uint                   `json:"message_id,omitempty"`
	Content     string                 `json:"content,omitempty"`
	MessageType models.MessageType     `json:"message_type,omitempty"`
	ReplyToID   *uint                  `json:"reply_to_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService(db *gorm.DB) *WebSocketService {
	return &WebSocketService{
		db: db,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		clients:   make(map[uint]*Client),
		rooms:     make(map[string]*Room),
		broadcast: make(chan Notification, 100),
	}
}

// HandleWebSocket handles WebSocket connections with authentication
func (ws *WebSocketService) HandleWebSocket(c *gin.Context) {
	// Get user info from middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	username, _ := c.Get("username")
	userRole, _ := c.Get("userRole")

	// Upgrade HTTP connection to WebSocket
	conn, err := ws.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Create new client
	client := &Client{
		ID:       userID.(uint),
		Username: username.(string),
		Role:     userRole.(string),
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Rooms:    make(map[string]bool),
		LastSeen: time.Now(),
	}

	// Register client
	ws.registerClient(client)

	// Send welcome message
	ws.sendWelcomeMessage(client)

	// Start goroutines for reading and writing
	go ws.readPump(client)
	go ws.writePump(client)

	log.Printf("Client %s (ID: %d) connected", client.Username, client.ID)
}

// registerClient adds a client to the service
func (ws *WebSocketService) registerClient(client *Client) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	ws.clients[client.ID] = client
}

// unregisterClient removes a client from the service
func (ws *WebSocketService) unregisterClient(client *Client) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	// Remove client from all rooms
	for roomID := range client.Rooms {
		ws.leaveRoom(client, roomID)
	}

	delete(ws.clients, client.ID)
	close(client.Send)

	log.Printf("Client %s (ID: %d) disconnected", client.Username, client.ID)
}

// readPump handles reading messages from the client
func (ws *WebSocketService) readPump(client *Client) {
	defer func() {
		ws.unregisterClient(client)
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Update last seen
		client.LastSeen = time.Now()

		// Parse incoming message
		var wsMessage WebSocketChatMessage
		if err := json.Unmarshal(messageBytes, &wsMessage); err != nil {
			log.Printf("Failed to parse WebSocket message: %v", err)
			continue
		}

		// Handle the message based on type
		ws.handleWebSocketMessage(client, &wsMessage)
	}
}

// writePump handles writing messages to the client
func (ws *WebSocketService) writePump(client *Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleWebSocketMessage processes incoming WebSocket messages
func (ws *WebSocketService) handleWebSocketMessage(client *Client, wsMessage *WebSocketChatMessage) {
	switch wsMessage.Type {
	case "chat":
		ws.handleChatMessage(client, wsMessage)
	case "join_room":
		ws.handleJoinRoom(client, wsMessage)
	case "leave_room":
		ws.handleLeaveRoom(client, wsMessage)
	case "typing":
		ws.handleTypingIndicator(client, wsMessage)
	case "ping":
		ws.handlePing(client)
	default:
		log.Printf("Unknown WebSocket message type: %s", wsMessage.Type)
	}
}

// handleChatMessage processes chat messages
func (ws *WebSocketService) handleChatMessage(client *Client, wsMessage *WebSocketChatMessage) {
	// Verify user is in the room
	roomKey := strconv.FormatUint(uint64(wsMessage.RoomID), 10)
	if !client.Rooms[roomKey] {
		ws.sendErrorMessage(client, "You are not in this room")
		return
	}

	// Validate message content
	if wsMessage.Content == "" {
		ws.sendErrorMessage(client, "Message content cannot be empty")
		return
	}

	// Check if user can send messages to this room
	var participant models.ChatParticipant
	if err := ws.db.Where("chat_room_id = ? AND user_id = ?", wsMessage.RoomID, client.ID).First(&participant).Error; err != nil {
		ws.sendErrorMessage(client, "You are not a participant in this room")
		return
	}

	if participant.IsBlocked || participant.Role == models.ChatRoleReadOnly {
		ws.sendErrorMessage(client, "You cannot send messages to this room")
		return
	}

	// Create and save the message to database
	messageType := wsMessage.MessageType
	if messageType == "" {
		messageType = models.MessageTypeText
	}

	message := &models.ChatMessage{
		ChatRoomID: wsMessage.RoomID,
		SenderID:   client.ID,
		Content:    wsMessage.Content,
		Type:       messageType,
		Status:     models.MessageStatusSent,
		ReplyToID:  wsMessage.ReplyToID,
	}

	if wsMessage.Metadata != nil {
		metadataBytes, _ := json.Marshal(wsMessage.Metadata)
		message.Metadata = string(metadataBytes)
	}

	if err := ws.db.Create(message).Error; err != nil {
		ws.sendErrorMessage(client, "Failed to save message")
		return
	}

	// Update room's last message time
	ws.db.Model(&models.ChatRoom{}).Where("id = ?", wsMessage.RoomID).Update("last_message", time.Now())

	// Load sender info for broadcast
	ws.db.Preload("Sender").First(message, message.ID)

	// Broadcast to room
	ws.broadcastToRoom(roomKey, map[string]interface{}{
		"type":         "message",
		"message_id":   message.ID,
		"room_id":      message.ChatRoomID,
		"sender_id":    message.SenderID,
		"sender_name":  client.Username,
		"content":      message.Content,
		"message_type": message.Type,
		"reply_to_id":  message.ReplyToID,
		"created_at":   message.CreatedAt,
		"timestamp":    time.Now(),
	})
}

// handleJoinRoom handles room joining
func (ws *WebSocketService) handleJoinRoom(client *Client, wsMessage *WebSocketChatMessage) {
	// Check if user is a participant in the room
	var participant models.ChatParticipant
	if err := ws.db.Where("chat_room_id = ? AND user_id = ?", wsMessage.RoomID, client.ID).First(&participant).Error; err != nil {
		ws.sendErrorMessage(client, "You are not a participant in this room")
		return
	}

	roomKey := strconv.FormatUint(uint64(wsMessage.RoomID), 10)
	ws.joinRoom(client, roomKey, wsMessage.RoomID)

	// Send confirmation
	ws.sendMessage(client, map[string]interface{}{
		"type":    "room_joined",
		"room_id": wsMessage.RoomID,
		"message": "Successfully joined room",
	})

	// Notify other participants
	ws.broadcastToRoom(roomKey, map[string]interface{}{
		"type":      "user_joined",
		"room_id":   wsMessage.RoomID,
		"user_id":   client.ID,
		"username":  client.Username,
		"timestamp": time.Now(),
	})
}

// handleLeaveRoom handles room leaving
func (ws *WebSocketService) handleLeaveRoom(client *Client, wsMessage *WebSocketChatMessage) {
	roomKey := strconv.FormatUint(uint64(wsMessage.RoomID), 10)
	ws.leaveRoom(client, roomKey)

	// Send confirmation
	ws.sendMessage(client, map[string]interface{}{
		"type":    "room_left",
		"room_id": wsMessage.RoomID,
		"message": "Successfully left room",
	})

	// Notify other participants
	ws.broadcastToRoom(roomKey, map[string]interface{}{
		"type":      "user_left",
		"room_id":   wsMessage.RoomID,
		"user_id":   client.ID,
		"username":  client.Username,
		"timestamp": time.Now(),
	})
}

// handleTypingIndicator handles typing indicators
func (ws *WebSocketService) handleTypingIndicator(client *Client, wsMessage *WebSocketChatMessage) {
	roomKey := strconv.FormatUint(uint64(wsMessage.RoomID), 10)

	// Check if user is in the room
	if !client.Rooms[roomKey] {
		return
	}

	// Broadcast typing indicator to room (except sender)
	ws.broadcastToRoomExcept(roomKey, client.ID, map[string]interface{}{
		"type":      "typing",
		"room_id":   wsMessage.RoomID,
		"user_id":   client.ID,
		"username":  client.Username,
		"timestamp": time.Now(),
	})
}

// handlePing handles ping messages
func (ws *WebSocketService) handlePing(client *Client) {
	ws.sendMessage(client, map[string]interface{}{
		"type":      "pong",
		"timestamp": time.Now(),
	})
}

// Room management methods
func (ws *WebSocketService) joinRoom(client *Client, roomKey string, roomID uint) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	// Create room if it doesn't exist
	if _, exists := ws.rooms[roomKey]; !exists {
		ws.rooms[roomKey] = &Room{
			ID:      roomID,
			Name:    roomKey,
			Clients: make(map[uint]*Client),
		}
	}

	room := ws.rooms[roomKey]
	room.mutex.Lock()
	room.Clients[client.ID] = client
	room.mutex.Unlock()

	client.Rooms[roomKey] = true
}

func (ws *WebSocketService) leaveRoom(client *Client, roomKey string) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if room, exists := ws.rooms[roomKey]; exists {
		room.mutex.Lock()
		delete(room.Clients, client.ID)
		room.mutex.Unlock()

		// Remove room if empty
		if len(room.Clients) == 0 {
			delete(ws.rooms, roomKey)
		}
	}

	delete(client.Rooms, roomKey)
}

// Broadcasting methods
func (ws *WebSocketService) broadcastToRoom(roomKey string, data interface{}) {
	ws.mutex.RLock()
	room, exists := ws.rooms[roomKey]
	ws.mutex.RUnlock()

	if !exists {
		return
	}

	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	for _, client := range room.Clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(room.Clients, client.ID)
		}
	}
}

func (ws *WebSocketService) broadcastToRoomExcept(roomKey string, exceptUserID uint, data interface{}) {
	ws.mutex.RLock()
	room, exists := ws.rooms[roomKey]
	ws.mutex.RUnlock()

	if !exists {
		return
	}

	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	for _, client := range room.Clients {
		if client.ID != exceptUserID {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(room.Clients, client.ID)
			}
		}
	}
}

// Utility methods
func (ws *WebSocketService) sendMessage(client *Client, data interface{}) {
	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case client.Send <- message:
	default:
		close(client.Send)
		ws.unregisterClient(client)
	}
}

func (ws *WebSocketService) sendErrorMessage(client *Client, errorMsg string) {
	ws.sendMessage(client, map[string]interface{}{
		"type":      "error",
		"message":   errorMsg,
		"timestamp": time.Now(),
	})
}

func (ws *WebSocketService) sendWelcomeMessage(client *Client) {
	ws.sendMessage(client, map[string]interface{}{
		"type":      "welcome",
		"message":   "Connected to chat server",
		"user_id":   client.ID,
		"username":  client.Username,
		"timestamp": time.Now(),
	})
}

// GetConnectionCount returns number of active connections
func (ws *WebSocketService) GetConnectionCount() int {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	return len(ws.clients)
}

// GetRoomCount returns number of active rooms
func (ws *WebSocketService) GetRoomCount() int {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	return len(ws.rooms)
}

// GetOnlineUsers returns list of online users
func (ws *WebSocketService) GetOnlineUsers() []map[string]interface{} {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	users := make([]map[string]interface{}, 0, len(ws.clients))
	for _, client := range ws.clients {
		users = append(users, map[string]interface{}{
			"id":        client.ID,
			"username":  client.Username,
			"role":      client.Role,
			"last_seen": client.LastSeen,
		})
	}
	return users
}

// SendNotification sends a notification to a specific user
func (ws *WebSocketService) SendNotification(userID uint, notification Notification) {
	ws.mutex.RLock()
	client, exists := ws.clients[userID]
	ws.mutex.RUnlock()

	if !exists {
		log.Printf("User %d not connected, notification not sent", userID)
		return
	}

	// Convert notification to JSON
	message, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal notification: %v", err)
		return
	}

	// Send to client
	select {
	case client.Send <- message:
		log.Printf("Notification sent to user %d", userID)
	default:
		log.Printf("Failed to send notification to user %d: channel full", userID)
	}
}

// BroadcastToRoom sends a notification to all users in a room
func (ws *WebSocketService) BroadcastToRoom(roomName string, notification Notification) {
	ws.mutex.RLock()
	room, exists := ws.rooms[roomName]
	ws.mutex.RUnlock()

	if !exists {
		log.Printf("Room %s not found", roomName)
		return
	}

	// Convert notification to JSON
	message, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Failed to marshal notification: %v", err)
		return
	}

	room.mutex.RLock()
	defer room.mutex.RUnlock()

	for _, client := range room.Clients {
		select {
		case client.Send <- message:
			log.Printf("Notification sent to user %d in room %s", client.ID, roomName)
		default:
			log.Printf("Failed to send notification to user %d in room %s: channel full", client.ID, roomName)
		}
	}
}
