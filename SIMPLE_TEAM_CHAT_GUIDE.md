# Simple Team Chat Guide

## Overview

A simple chat system where all team members can talk together in one place. No complex features - just basic messaging for team communication.

## Features

✅ **One Team Chat Room** - Everyone talks in the same place  
✅ **Real-time Messaging** - Messages appear instantly  
✅ **Arabic Language Support** - Full support for Arabic text  
✅ **Message History** - See previous messages  
✅ **Online Users** - See who's currently online  

## Quick Start

### 1. Get Team Chat Room
```http
GET /api/chat/team-chat
Authorization: Bearer <your-jwt-token>
```

This will:
- Create the team chat room if it doesn't exist
- Auto-join you to the room
- Return the room details

**Response:**
```json
{
  "message": "Team chat ready",
  "room": {
    "id": 1,
    "name": "Team Chat",
    "description": "Main chat room for all team members",
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

### 2. Connect to WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/chat');

ws.onopen = () => {
  console.log('Connected to team chat');
  
  // Join the team chat room
  ws.send(JSON.stringify({
    type: 'join_room',
    room_id: 1  // Use the room ID from step 1
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('New message:', data);
};
```

### 3. Send Messages
```javascript
// Send a message
ws.send(JSON.stringify({
  type: 'chat',
  room_id: 1,
  content: 'Hello team! 👋'
}));

// Send Arabic message
ws.send(JSON.stringify({
  type: 'chat',
  room_id: 1,
  content: 'مرحبا بالجميع!'
}));
```

## API Endpoints

### Get Team Chat
```http
GET /api/chat/team-chat
```
Creates or returns the main team chat room and auto-joins you.

### Send Message
```http
POST /api/chat/rooms/{roomId}/messages
Content-Type: application/json

{
  "content": "Hello everyone!"
}
```

### Get Messages
```http
GET /api/chat/rooms/{roomId}/messages?page=1&limit=50
```

### Get Team Members
```http
GET /api/chat/rooms/{roomId}/members
```

### Check Online Status
```http
GET /api/chat/ws/status
```

## WebSocket Messages

### Join Room
```json
{
  "type": "join_room",
  "room_id": 1
}
```

### Send Message
```json
{
  "type": "chat",
  "room_id": 1,
  "content": "Your message here"
}
```

### Receive Message
```json
{
  "type": "message",
  "message_id": 123,
  "room_id": 1,
  "sender_id": 456,
  "sender_name": "john_doe",
  "content": "Hello everyone!",
  "created_at": "2024-01-15T10:30:00Z"
}
```

## Complete Example

```html
<!DOCTYPE html>
<html>
<head>
    <title>Simple Team Chat</title>
    <meta charset="UTF-8">
</head>
<body>
    <div id="messages"></div>
    <input type="text" id="messageInput" placeholder="Type your message...">
    <button onclick="sendMessage()">Send</button>

    <script>
        let ws;
        let roomId;
        const token = localStorage.getItem('jwt_token'); // Your JWT token

        // 1. Get team chat room
        fetch('/api/chat/team-chat', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        })
        .then(response => response.json())
        .then(data => {
            roomId = data.room.id;
            console.log('Team chat ready:', data);
            connectWebSocket();
        });

        // 2. Connect to WebSocket
        function connectWebSocket() {
            ws = new WebSocket('ws://localhost:8080/ws/chat');
            
            ws.onopen = () => {
                console.log('Connected to chat');
                // Join team chat room
                ws.send(JSON.stringify({
                    type: 'join_room',
                    room_id: roomId
                }));
            };

            ws.onmessage = (event) => {
                const data = JSON.parse(event.data);
                if (data.type === 'message') {
                    displayMessage(data);
                }
            };
        }

        // 3. Send message
        function sendMessage() {
            const input = document.getElementById('messageInput');
            const content = input.value.trim();
            
            if (content && ws.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({
                    type: 'chat',
                    room_id: roomId,
                    content: content
                }));
                input.value = '';
            }
        }

        // 4. Display message
        function displayMessage(data) {
            const messages = document.getElementById('messages');
            const messageDiv = document.createElement('div');
            messageDiv.innerHTML = `<strong>${data.sender_name}:</strong> ${data.content}`;
            messages.appendChild(messageDiv);
            messages.scrollTop = messages.scrollHeight;
        }

        // Send message on Enter key
        document.getElementById('messageInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });
    </script>
</body>
</html>
```

## That's It!

Your simple team chat is ready. Everyone can:
1. Get the team chat room
2. Connect via WebSocket
3. Send and receive messages in real-time
4. Use Arabic text without any issues

No complex features, no multiple rooms, just simple team communication! 🎉 