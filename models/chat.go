package models

import (
	"time"

	"gorm.io/gorm"
)

// MessageType represents different types of messages
type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeFile  MessageType = "file"
	MessageTypeVideo MessageType = "video"
	MessageTypeAudio MessageType = "audio"
)

// MessageStatus represents message status
type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

// ChatRole represents user role in chat
type ChatRole string

const (
	ChatRoleMember   ChatRole = "member"
	ChatRoleReadOnly ChatRole = "read_only"
)

// Simple team chat room
type ChatRoom struct {
	gorm.Model
	Name        string     `gorm:"not null;type:varchar(255) COLLATE \"default\""`
	Description string     `gorm:"type:text"`
	CreatedBy   uint       `gorm:"not null;index"`
	MaxMembers  int        `gorm:"default:1000"`
	LastMessage *time.Time `gorm:"default:null"`

	// Relationships
	Creator  User          `gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE"`
	Messages []ChatMessage `gorm:"foreignKey:ChatRoomID;constraint:OnDelete:CASCADE"`
	Members  []User        `gorm:"many2many:chat_participants;constraint:OnDelete:CASCADE"`
}

// Simple chat participant (just to track who's in the room)
type ChatParticipant struct {
	gorm.Model
	ChatRoomID uint      `gorm:"not null;index"`
	UserID     uint      `gorm:"not null;index"`
	JoinedAt   time.Time `gorm:"not null"`
	Role       ChatRole  `gorm:"default:'member'"`
	IsBlocked  bool      `gorm:"default:false"`

	// Relationships
	ChatRoom ChatRoom `gorm:"foreignKey:ChatRoomID;constraint:OnDelete:CASCADE"`
	User     User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// Simple chat message
type ChatMessage struct {
	gorm.Model
	ChatRoomID uint          `gorm:"not null;index"`
	SenderID   uint          `gorm:"not null;index"`
	Content    string        `gorm:"not null;type:text"`
	Type       MessageType   `gorm:"default:'text'"`
	Status     MessageStatus `gorm:"default:'sent'"`
	ReplyToID  *uint         `gorm:"default:null;index"`
	Metadata   string        `gorm:"type:text"`

	// Relationships
	ChatRoom ChatRoom     `gorm:"foreignKey:ChatRoomID;constraint:OnDelete:CASCADE"`
	Sender   User         `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
	ReplyTo  *ChatMessage `gorm:"foreignKey:ReplyToID;constraint:OnDelete:SET NULL"`
}
