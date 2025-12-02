# 🤖 AI Chat Assistant Guide

## Overview

The AI Chat Assistant is an intelligent chatbot integrated into your project management system. It helps users manage tasks, get project insights, and receive personalized recommendations using natural language conversations. The assistant supports both **private one-on-one chats** and **team chat mentions**, and works in both **English and Arabic**.

---

## 🎯 Key Features

### 1. **Dual Chat Modes**
- **Private AI Chat Room**: Personal assistant for each user
- **Team Chat Integration**: Mention `@ai` in any chat room

### 2. **Natural Language Processing**
- Understands questions in plain English or Arabic
- Creates tasks from natural language descriptions
- Provides contextual answers based on your data

### 3. **Smart Commands**
- Create tasks
- Get task lists
- Check workload
- Get project status
- Receive recommendations

### 4. **Bilingual Support**
- English responses
- Arabic (العربية) responses
- Auto-detects language from your message

### 5. **Context-Aware**
- Knows your active tasks
- Understands your workload
- Remembers recent conversation
- Considers Arabic working hours

---

## 🚀 Quick Start

### **Option 1: Private AI Chat Room**

1. **Get your AI chat room:**
   ```http
   GET /api/chat/ai/room
   Authorization: Bearer {your_jwt_token}
   ```

2. **Send messages in that room** (via WebSocket or REST API)

3. **AI responds automatically!**

### **Option 2: Team Chat Mentions**

1. **In any chat room**, type:
   ```
   @ai what tasks do I have?
   ```

2. **AI detects the mention** and responds in the same room

---

## 📋 Supported Commands

### **1. Create Task**

**English:**
- "Create task to review API documentation by Friday"
- "New task: Fix login bug"
- "Add task to update database schema"

**Arabic:**
- "إنشاء مهمة لمراجعة وثائق API"
- "مهمة جديدة: إصلاح خطأ تسجيل الدخول"

**Example:**
```
User: "Create task to review API docs by Friday"
AI: ✅ Task created successfully!

Title: Review API docs
Status: Pending
Task ID: 42
Due Date: Friday, January 19, 2024
```

---

### **2. Get Tasks**

**English:**
- "What tasks do I have?"
- "Show my tasks"
- "List my active tasks"
- "What's on my todo list?"

**Arabic:**
- "مهامي"
- "ما هي المهام"
- "عرض مهامي"

**Example:**
```
User: "What tasks do I have?"
AI: You have 3 active tasks:

1. Review API documentation (pending)
   Due: Jan 19, 2024
2. Fix login bug (in_progress)
3. Update database schema (pending)
```

---

### **3. Check Workload**

**English:**
- "What's my workload?"
- "How busy am I?"
- "Check my capacity"
- "Am I overloaded?"

**Arabic:**
- "عبء العمل"
- "كم أنا مشغول"
- "ما هو عبء عملي"

**Example:**
```
User: "What's my workload?"
AI: ✅ Current workload: 85.0%

Status: healthy
Weekly capacity: 42 hours
Active tasks: 5
```

---

### **4. General Questions**

The AI can answer various questions about:
- Task management
- Project status
- Time management advice
- Priority suggestions
- Best practices

**Examples:**
```
User: "How should I prioritize my tasks?"
AI: Based on your workload and deadlines, here's my recommendation:

🔴 URGENT (Do today):
1. Complete API documentation (2 hours remaining)
2. Fix critical bug (3 hours)

🟡 IMPORTANT (This week):
3. Code review (4 hours)
4. Write tests (6 hours)

Your current capacity: 85% (within healthy range)
```

```
User: "Can I finish all my tasks this week?"
AI: Let me analyze your workload...

Total estimated hours: 45 hours
Your weekly capacity: 42 hours
Current utilization: 107% (OVERLOADED)

⚠️ Recommendation: You're overloaded by 3 hours. I suggest:
- Delegate 2 low-priority tasks
- Request deadline extension for 1 task
- Focus on critical tasks first
```

---

## 🔌 API Endpoints

### **1. Get or Create AI Chat Room**

Get your private AI chat room (creates one if it doesn't exist).

```http
GET /api/chat/ai/room
Authorization: Bearer {jwt_token}
```

**Response:**
```json
{
  "message": "AI chat room ready",
  "room": {
    "id": 5,
    "name": "AI Assistant - 123",
    "description": "Your personal AI assistant for task and project management",
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

---

### **2. Process AI Message**

Process an AI message directly (useful for API calls).

```http
POST /api/chat/ai/rooms/:roomId/message
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "message": "What tasks do I have?"
}
```

**Response:**
```json
{
  "message": "AI response generated",
  "response": {
    "message": "You have 3 active tasks:\n\n1. Review API docs...",
    "action": "get_tasks",
    "data": {
      "task_count": 3
    },
    "language": "en"
  }
}
```

---

## 💬 Usage Examples

### **Example 1: Private Chat Conversation**

```
[Private: AI Assistant - 123]

User: "What tasks do I have today?"
AI: "You have 3 tasks today:
     1. Review API documentation (High priority, due in 2 hours)
     2. Fix login bug (Medium priority, due tomorrow)
     3. Update project status (Low priority, no deadline)
     
     Based on your workload, I recommend starting with the API documentation review."

User: "Create a task to review the new API endpoints by Friday"
AI: "✅ Task created successfully!
     Title: Review new API endpoints
     Due Date: Friday, January 19, 2024
     Status: Pending
     Estimated time: 4 hours (based on similar tasks)
     
     I've added this to your task list. Would you like me to set a reminder?"

User: "How should I prioritize?"
AI: "Based on your workload and deadlines, here's my recommendation:
     
     🔴 URGENT (Do today):
     1. Complete API documentation (2 hours remaining)
     2. Fix critical bug (3 hours)
     
     🟡 IMPORTANT (This week):
     3. Code review (4 hours)
     4. Write tests (6 hours)
     
     🟢 CAN WAIT:
     5. Update documentation (2 hours)
     
     Your current capacity: 85% (within healthy range)"
```

---

### **Example 2: Team Chat with @ai Mention**

```
[Team Chat Room]

John: "Hey team, what's the project status?"
Sarah: "Good question!"
@ai: "The Mobile App project is 65% complete:
     - 12 tasks completed (60%)
     - 5 tasks in progress
     - 3 tasks pending
     - Overall risk: Medium
     - Predicted completion: 3 days ahead of schedule"

John: "@ai which tasks need urgent attention?"
AI: "There are 2 tasks requiring urgent attention:
     1. Complete user authentication (2 days overdue)
     2. Write unit tests (1 day overdue)
     
     Would you like me to help prioritize these?"
```

---

### **Example 3: Arabic Conversation**

```
[Private: AI Assistant - 123]

User: "ما هي مهامي اليوم؟"
AI: "لديك 3 مهام اليوم:
     1. مراجعة وثائق API (أولوية عالية)
     2. إصلاح خطأ تسجيل الدخول
     3. تحديث حالة المشروع
     
     بناءً على ساعات العمل العربية (9 صباحاً - 4 عصراً)،
     أنصحك بالبدء بالمهمة الأولى في الساعة 10 صباحاً."

User: "إنشاء مهمة لمراجعة الوثائق"
AI: "✅ تم إنشاء المهمة بنجاح!
     العنوان: مراجعة الوثائق
     الحالة: قيد الانتظار
     رقم المهمة: 42"
```

---

## 🎨 Response Actions

The AI can return different types of responses:

### **Action Types:**

1. **`answer`** - General question/answer
2. **`create_task`** - Task was created
3. **`get_tasks`** - Task list retrieved
4. **`get_workload`** - Workload information
5. **`get_analytics`** - Analytics data

### **Response Structure:**

```json
{
  "message": "Response text",
  "action": "answer",
  "data": {
    // Additional data (task_id, task_count, etc.)
  },
  "suggestions": [
    // Optional suggestions
  ],
  "language": "en" // or "ar"
}
```

---

## 🔍 How AI Detection Works

The AI automatically detects messages directed to it in two ways:

### **1. Message Prefix Detection**
Messages starting with:
- `@ai`
- `@assistant`
- `ai:`
- `assistant:`

### **2. AI Chat Room Detection**
Any message in a chat room with "AI Assistant" in the name

---

## 🌍 Language Support

### **Auto-Detection**
The AI automatically detects the language of your message:
- **Arabic**: If message contains Arabic characters (U+0600–U+06FF)
- **English**: Default for other messages

### **Bilingual Responses**
- Responds in the same language as your message
- Supports mixed Arabic/English conversations
- Understands Arabic task management terms

---

## 🧠 AI Context Awareness

The AI has access to:

1. **Your Profile**
   - Username, role, department
   - Current workload percentage
   - Active tasks count

2. **Your Tasks**
   - Active tasks (pending, in_progress)
   - Task statuses and deadlines
   - Task history

3. **Your Projects**
   - Projects you're part of
   - Project status and progress

4. **Working Schedule**
   - Arabic working hours (Saturday-Thursday, 9 AM-4 PM)
   - 42-hour weekly capacity
   - Friday is off

5. **Recent Conversation**
   - Last 5 messages in the chat room
   - Conversation context

---

## ⚙️ Configuration

### **Environment Variables**

Make sure you have the Gemini API key set:

```bash
GEMINI_API_KEY=your_gemini_api_key_here
```

### **AI Model Settings**

- **Model**: `gemini-2.0-flash`
- **Temperature**: 0.7 (conversational)
- **Max Tokens**: 2048

---

## 🔒 Security & Permissions

### **Access Control**
- All AI chat endpoints require authentication
- Users can only access their own private AI chat room
- Team chat mentions work in rooms where user is a member

### **Data Privacy**
- AI only accesses your own tasks and projects
- Private AI chat rooms are only visible to the owner
- AI responses respect role-based permissions

---

## 📊 Response Examples

### **Task Creation Response**

```json
{
  "message": "AI response generated",
  "response": {
    "message": "✅ Task created successfully!\n\nTitle: Review API docs\nStatus: Pending\nTask ID: 42",
    "action": "create_task",
    "data": {
      "task_id": 42,
      "title": "Review API docs"
    },
    "language": "en"
  }
}
```

### **Task List Response**

```json
{
  "message": "AI response generated",
  "response": {
    "message": "You have 3 active tasks:\n\n1. Review API docs (pending)\n2. Fix bug (in_progress)\n3. Update docs (pending)",
    "action": "get_tasks",
    "data": {
      "task_count": 3
    },
    "language": "en"
  }
}
```

### **Workload Response**

```json
{
  "message": "AI response generated",
  "response": {
    "message": "✅ Current workload: 85.0%\n\nStatus: healthy\nWeekly capacity: 42 hours\nActive tasks: 5",
    "action": "get_workload",
    "data": {
      "workload_percent": 85.0,
      "status": "healthy"
    },
    "language": "en"
  }
}
```

---

## 🚨 Troubleshooting

### **AI Not Responding**

1. **Check API Key**
   ```bash
   # Verify GEMINI_API_KEY is set
   echo $GEMINI_API_KEY
   ```

2. **Check Logs**
   - Look for "AI service is currently unavailable" messages
   - Check for API key errors

3. **Verify Room Access**
   - Ensure you're in the chat room
   - Check room permissions

### **AI Not Understanding Commands**

- Be specific with commands
- Use supported keywords (see Commands section)
- Try rephrasing your question

### **Language Detection Issues**

- If AI responds in wrong language, explicitly mention language:
  - "Answer in Arabic" / "أجب بالعربية"
  - "Answer in English"

---

## 💡 Best Practices

### **For Users**

1. **Be Specific**
   - "Create task to review API docs by Friday" ✅
   - "Create task" ❌ (too vague)

2. **Use Natural Language**
   - The AI understands conversational language
   - No need for special syntax

3. **Ask Follow-up Questions**
   - AI remembers recent conversation
   - Build on previous answers

4. **Use Private Chat for Personal Questions**
   - Keeps team chat clean
   - Better for personal task management

### **For Developers**

1. **Monitor API Usage**
   - Track Gemini API calls
   - Implement rate limiting if needed

2. **Cache Responses**
   - Consider caching common queries
   - Reduce API costs

3. **Error Handling**
   - Always handle AI service unavailability
   - Provide fallback responses

---

## 🔄 Integration with Existing Features

The AI Chat Assistant integrates with:

1. **Task Management**
   - Creates tasks via TaskService
   - Queries user tasks
   - Uses AI time analysis data

2. **Chat System**
   - Uses existing chat infrastructure
   - Works with WebSocket for real-time responses
   - Respects chat room permissions

3. **AI Time Optimizer**
   - Uses workload analysis
   - Considers Arabic working hours
   - Provides time management advice

4. **Notifications**
   - Can trigger notifications (future enhancement)
   - Integrates with notification system

---

## 📈 Future Enhancements

Planned features:

1. **More Commands**
   - Update task status
   - Assign tasks to team members
   - Get project analytics

2. **Proactive Suggestions**
   - Daily task recommendations
   - Deadline reminders
   - Workload alerts

3. **Advanced Analytics**
   - Project risk analysis
   - Team performance insights
   - Predictive recommendations

4. **Voice Support**
   - Voice-to-text input
   - Text-to-speech responses

---

## 📞 Support

For issues or questions:
- Check troubleshooting section above
- Review API documentation
- Check server logs for errors
- Verify GEMINI_API_KEY configuration

---

## 🎯 Quick Reference

### **Common Commands**

| Command | Example |
|---------|---------|
| Create Task | "Create task to review docs by Friday" |
| Get Tasks | "What tasks do I have?" |
| Check Workload | "What's my workload?" |
| Get Help | "How should I prioritize?" |

### **API Endpoints**

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/chat/ai/room` | GET | Get/create AI chat room |
| `/api/chat/ai/rooms/:roomId/message` | POST | Process AI message |

### **Trigger Words**

- `@ai` - Mention AI in team chat
- `@assistant` - Alternative mention
- `create task` - Create new task
- `my tasks` - Get task list
- `workload` - Check workload

---

**Remember**: The AI learns from your data and improves over time. The more you use it, the better it understands your work patterns! 🚀

