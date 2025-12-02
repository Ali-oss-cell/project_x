# 🤖 AI Chat Assistant - Quick Reference

## 🚀 Quick Start

### **Get Your AI Chat Room**
```bash
GET /api/chat/ai/room
Authorization: Bearer {token}
```

### **Send Message to AI**
```bash
POST /api/chat/ai/rooms/:roomId/message
Authorization: Bearer {token}
Content-Type: application/json

{
  "message": "What tasks do I have?"
}
```

### **Or Mention in Team Chat**
```
@ai what tasks do I have?
```

---

## 📝 Common Commands

| Command | Example | Response |
|---------|---------|----------|
| **Create Task** | `"Create task to review docs by Friday"` | Creates task and returns confirmation |
| **Get Tasks** | `"What tasks do I have?"` | Lists active tasks |
| **Check Workload** | `"What's my workload?"` | Shows workload percentage and status |
| **Get Help** | `"How should I prioritize?"` | Provides prioritization recommendations |

---

## 🌍 Language Support

| Language | Trigger | Example |
|----------|---------|---------|
| **English** | Default | `"What tasks do I have?"` |
| **Arabic** | Auto-detected | `"ما هي مهامي؟"` |

---

## 🎯 AI Detection

AI responds to:
- ✅ Messages starting with `@ai` or `@assistant`
- ✅ Messages in AI chat rooms (name contains "AI Assistant")
- ✅ Direct API calls to `/api/chat/ai/rooms/:roomId/message`

---

## 📊 Response Actions

| Action | Description | Data Included |
|--------|-------------|---------------|
| `answer` | General question/answer | None |
| `create_task` | Task was created | `task_id`, `title` |
| `get_tasks` | Task list retrieved | `task_count` |
| `get_workload` | Workload information | `workload_percent`, `status` |

---

## 💡 Tips

1. **Be Specific**: "Create task to review API docs by Friday" ✅
2. **Use Natural Language**: No special syntax needed
3. **Ask Follow-ups**: AI remembers recent conversation
4. **Private Chat**: Use for personal questions
5. **Team Chat**: Use `@ai` for team discussions

---

## 🔗 Related Documentation

- Full Guide: [AI_CHAT_ASSISTANT_GUIDE.md](./AI_CHAT_ASSISTANT_GUIDE.md)
- API Collection: `Project_X_API_Collection.postman_collection.json`

---

**Need help?** Check the full guide or server logs for detailed error messages.

