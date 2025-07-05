# Chat System Documentation

## Overview

The chat system provides real-time messaging capabilities for all users in the project management system. It supports multiple chat rooms, direct messages, file sharing, message reactions, and comprehensive moderation features.

## Features

### 🏠 Chat Rooms
- **General**: Company-wide discussions
- **Department**: Department-specific conversations
- **Project**: Project-related communications
- **Private**: Direct messages between users
- **Group**: Custom group chats
- **Announcement**: Official announcements (Admin/Manager only)

### 💬 Messaging
- Real-time messaging via WebSocket
- Message editing (15-minute window)
- Message deletion with proper permissions
- Reply to messages
- Message reactions (emojis)
- @mentions with notifications
- Typing indicators
- Message search

### 👥 User Management
- Role-based permissions (Owner, Admin, Moderator, Member, Read-only)
- Room capacity limits
- Mute/unmute notifications
- Block users from sending messages
- Invitation system for private rooms

### 🌍 Arabic Language Support
- Full UTF-8 support for Arabic text
- Right-to-left (RTL) text handling
- Arabic text search capabilities
- Mixed Arabic/English content support

## API Endpoints

### Authentication
All chat endpoints require JWT authentication via `Authorization: Bearer <token>` header.

### Room Management

#### Create Chat Room
```http
POST /api/chat/rooms
Content-Type: application/json

{
  "name": "Team Discussion",
  "description": "Main team communication channel",
  "type": "general",
  "is_private": false,
  "project_id": null,
  "department": "Engineering",
  "max_members": 100
}
```

**Room Types:**
- `general` - Company-wide chat
- `department` - Department-specific
- `project` - Project-related
- `private` - Direct messages
- `group` - Custom groups
- `announcement` - Official announcements

#### Get User's Chat Rooms
```http
GET /api/chat/rooms
```

#### Join Chat Room
```http
POST /api/chat/rooms/{roomId}/join
```

#### Leave Chat Room
```http
POST /api/chat/rooms/{roomId}/leave
```

#### Delete Chat Room (Owner only)
```http
DELETE /api/chat/rooms/{roomId}
```

### Message Management

#### Send Message
```http
POST /api/chat/rooms/{roomId}/messages
Content-Type: application/json

{
  "content": "Hello team! 👋",
  "type": "text",
  "reply_to_id": null,
  "metadata": {}
}
```

**Message Types:**
- `text` - Regular text message
- `image` - Image message
- `file` - File attachment
- `announcement` - Official announcement
- `reply` - Reply to another message

#### Get Room Messages (with pagination)
```http
GET /api/chat/rooms/{roomId}/messages?page=1&limit=50
```

#### Edit Message
```http
PUT /api/chat/messages/{messageId}
Content-Type: application/json

{
  "content": "Updated message content"
}
```

#### Delete Message
```http
DELETE /api/chat/messages/{messageId}
```

#### React to Message
```http
POST /api/chat/messages/{messageId}/react
Content-Type: application/json

{
  "emoji": "👍"
}
```

#### Search Messages
```http
GET /api/chat/rooms/{roomId}/search?q=search+term
```

### Room Participants

#### Get Room Participants
```http
GET /api/chat/rooms/{roomId}/participants
```

#### Update Participant Role
```http
PUT /api/chat/rooms/{roomId}/participants/{participantId}/role
Content-Type: application/json

{
  "role": "admin"
}
```

**Participant Roles:**
- `owner` - Room owner (can delete room)
- `admin` - Room admin (can manage participants)
- `moderator` - Can moderate messages
- `member` - Regular member
- `read_only` - Can only read messages

#### Mute/Unmute Room
```http
POST /api/chat/rooms/{roomId}/mute
Content-Type: application/json

{
  "is_muted": true
}
```

### Status & Statistics

#### WebSocket Status
```http
GET /api/chat/ws/status
```

#### Online Users
```http
GET /api/chat/online-users
```

#### Chat Statistics (Admin/Manager only)
```http
GET /api/chat/statistics
```

## WebSocket Real-time Communication

### Connection
Connect to WebSocket endpoint with authentication:
```javascript
const token = localStorage.getItem('jwt_token');
const ws = new WebSocket(`ws://localhost:8080/ws/chat?token=${token}`);
```

### Message Types

#### Join Room
```json
{
  "type": "join_room",
  "room_id": 1
}
```

#### Send Message
```json
{
  "type": "chat",
  "room_id": 1,
  "content": "Hello everyone!",
  "message_type": "text",
  "reply_to_id": null,
  "metadata": {}
}
```

#### Leave Room
```json
{
  "type": "leave_room",
  "room_id": 1
}
```

#### Typing Indicator
```json
{
  "type": "typing",
  "room_id": 1
}
```

#### Ping/Pong
```json
{
  "type": "ping"
}
```

### Incoming Message Types

#### Welcome Message
```json
{
  "type": "welcome",
  "message": "Connected to chat server",
  "user_id": 123,
  "username": "john_doe",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### New Message
```json
{
  "type": "message",
  "message_id": 456,
  "room_id": 1,
  "sender_id": 123,
  "sender_name": "john_doe",
  "content": "Hello everyone!",
  "message_type": "text",
  "reply_to_id": null,
  "created_at": "2024-01-15T10:30:00Z",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### User Joined/Left
```json
{
  "type": "user_joined",
  "room_id": 1,
  "user_id": 123,
  "username": "john_doe",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Typing Indicator
```json
{
  "type": "typing",
  "room_id": 1,
  "user_id": 123,
  "username": "john_doe",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Error Messages
```json
{
  "type": "error",
  "message": "You are not in this room",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Usage Examples

### JavaScript Client Example

```javascript
class ChatClient {
  constructor(token) {
    this.token = token;
    this.ws = null;
    this.currentRoom = null;
  }

  connect() {
    this.ws = new WebSocket(`ws://localhost:8080/ws/chat`);
    
    this.ws.onopen = () => {
      console.log('Connected to chat server');
    };

    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      this.handleMessage(data);
    };

    this.ws.onclose = () => {
      console.log('Disconnected from chat server');
      // Implement reconnection logic
    };
  }

  joinRoom(roomId) {
    this.currentRoom = roomId;
    this.send({
      type: 'join_room',
      room_id: roomId
    });
  }

  sendMessage(content, replyToId = null) {
    if (!this.currentRoom) return;
    
    this.send({
      type: 'chat',
      room_id: this.currentRoom,
      content: content,
      message_type: 'text',
      reply_to_id: replyToId
    });
  }

  sendTyping() {
    if (!this.currentRoom) return;
    
    this.send({
      type: 'typing',
      room_id: this.currentRoom
    });
  }

  send(data) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  handleMessage(data) {
    switch (data.type) {
      case 'welcome':
        console.log('Welcome:', data.message);
        break;
      case 'message':
        this.displayMessage(data);
        break;
      case 'typing':
        this.showTypingIndicator(data);
        break;
      case 'error':
        console.error('Chat error:', data.message);
        break;
    }
  }

  displayMessage(data) {
    // Implement message display logic
    console.log(`${data.sender_name}: ${data.content}`);
  }

  showTypingIndicator(data) {
    // Show typing indicator for user
    console.log(`${data.username} is typing...`);
  }
}

// Usage
const chatClient = new ChatClient(localStorage.getItem('jwt_token'));
chatClient.connect();

// Join a room
chatClient.joinRoom(1);

// Send a message
chatClient.sendMessage("Hello everyone!");

// Send typing indicator
chatClient.sendTyping();
```

### Arabic Language Example

```javascript
// Send Arabic message
chatClient.sendMessage("مرحبا بالجميع! كيف حالكم؟");

// Search for Arabic text
fetch('/api/chat/rooms/1/search?q=مرحبا', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
})
.then(response => response.json())
.then(data => console.log('Search results:', data));
```

## Permissions & Security

### Role-based Access Control

| Action | Employee | Head | Manager | Admin |
|--------|----------|------|---------|-------|
| Create general rooms | ❌ | ❌ | ✅ | ✅ |
| Create department rooms | ❌ | ✅ | ✅ | ✅ |
| Create private/group rooms | ✅ | ✅ | ✅ | ✅ |
| Create announcement rooms | ❌ | ❌ | ✅ | ✅ |
| Delete own messages | ✅ | ✅ | ✅ | ✅ |
| Delete others' messages | ❌ | Room Admin+ | Room Admin+ | Room Admin+ |
| Moderate rooms | ❌ | Room Moderator+ | Room Moderator+ | Room Moderator+ |
| View chat statistics | ❌ | ❌ | ✅ | ✅ |

### Message Permissions

- **Edit messages**: Only sender, within 15 minutes
- **Delete messages**: Sender always, Room Admin+ for others
- **React to messages**: All participants
- **Mention users**: All participants (must be in same room)

## Database Schema

### Chat Tables

```sql
-- Chat rooms
CREATE TABLE chat_rooms (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    is_private BOOLEAN DEFAULT FALSE,
    created_by INTEGER NOT NULL,
    project_id INTEGER,
    department VARCHAR(255),
    max_members INTEGER DEFAULT 100,
    is_archived BOOLEAN DEFAULT FALSE,
    last_message TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Chat participants
CREATE TABLE chat_participants (
    id SERIAL PRIMARY KEY,
    chat_room_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    role VARCHAR(50) DEFAULT 'member',
    joined_at TIMESTAMP NOT NULL,
    last_read_at TIMESTAMP,
    is_muted BOOLEAN DEFAULT FALSE,
    is_blocked BOOLEAN DEFAULT FALSE,
    invited_by INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Chat messages
CREATE TABLE chat_messages (
    id SERIAL PRIMARY KEY,
    chat_room_id INTEGER NOT NULL,
    sender_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    type VARCHAR(50) DEFAULT 'text',
    status VARCHAR(50) DEFAULT 'sent',
    reply_to_id INTEGER,
    edited_at TIMESTAMP,
    is_system_message BOOLEAN DEFAULT FALSE,
    metadata JSON,
    reactions_count INTEGER DEFAULT 0,
    replies_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## Performance Considerations

### WebSocket Optimization
- Connection pooling and management
- Message queuing for offline users
- Automatic reconnection with exponential backoff
- Heartbeat/ping-pong to detect dead connections

### Database Optimization
- Indexed frequently queried fields
- Message pagination to limit memory usage
- Soft delete for messages (status = 'deleted')
- Archive old rooms to improve performance

### Scaling Considerations
- Horizontal scaling with Redis for WebSocket state
- Message queue system for high-volume notifications
- CDN for file attachments
- Database read replicas for message history

## Troubleshooting

### Common Issues

1. **WebSocket Connection Failed**
   - Check authentication token
   - Verify network connectivity
   - Check server WebSocket endpoint

2. **Messages Not Appearing**
   - Verify user is in the room
   - Check WebSocket connection status
   - Ensure proper room joining

3. **Permission Denied**
   - Verify user role and permissions
   - Check room participant status
   - Validate authentication

4. **Arabic Text Issues**
   - Ensure UTF-8 encoding throughout stack
   - Check database collation settings
   - Verify client-side UTF-8 handling

### Debug Commands

```bash
# Check WebSocket connections
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/chat/ws/status

# Check online users
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/chat/online-users

# Get chat statistics
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/chat/statistics
```

## Future Enhancements

- [ ] File upload and sharing
- [ ] Voice messages
- [ ] Message threads
- [ ] Message scheduling
- [ ] Chat bots and automation
- [ ] Advanced search with filters
- [ ] Message translation
- [ ] Video/audio calling integration
- [ ] Mobile push notifications
- [ ] Message encryption 