package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"project-x/config"
	"project-x/models"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

type AIChatService struct {
	DB           *gorm.DB
	client       *genai.Client
	model        *genai.GenerativeModel
	WorkSchedule *config.WorkScheduleConfig
	TaskService  *TaskService
}

type UserChatContext struct {
	UserID          uint
	Username        string
	Role            string
	Department      string
	ActiveTasks     int
	Projects        []string
	CurrentWorkload float64
	Language        string // "ar" or "en"
}

type AIResponse struct {
	Message     string                 `json:"message"`
	Action      string                 `json:"action"` // "answer", "create_task", "get_analytics", etc.
	Data        map[string]interface{} `json:"data,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Language    string                 `json:"language"`
}

func NewAIChatService(db *gorm.DB, taskService *TaskService) *AIChatService {
	ctx := context.Background()

	// Get API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Printf("Warning: GEMINI_API_KEY not set, AI chat will be disabled")
		return &AIChatService{
			DB:           db,
			WorkSchedule: config.GetDefaultWorkSchedule(),
			TaskService:  taskService,
		}
	}

	// Initialize Gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Error creating Gemini client for chat: %v", err)
		return &AIChatService{
			DB:           db,
			WorkSchedule: config.GetDefaultWorkSchedule(),
			TaskService:  taskService,
		}
	}

	// Configure the model for chat (more conversational)
	model := client.GenerativeModel("gemini-2.0-flash")
	model.SetTemperature(0.7) // Higher temperature for more natural conversation
	model.SetTopP(0.9)
	model.SetTopK(40)
	model.SetMaxOutputTokens(2048)

	return &AIChatService{
		DB:           db,
		client:       client,
		model:        model,
		WorkSchedule: config.GetDefaultWorkSchedule(),
		TaskService:  taskService,
	}
}

// GetOrCreateAIChatRoom gets or creates a private AI chat room for a user
func (a *AIChatService) GetOrCreateAIChatRoom(userID uint) (*models.ChatRoom, error) {
	// Check if AI chat room already exists for this user
	var room models.ChatRoom
	roomName := fmt.Sprintf("AI Assistant - %d", userID)

	err := a.DB.Where("name = ? AND created_by = ?", roomName, userID).First(&room).Error
	if err == nil {
		// Room exists, ensure user is a participant
		var participant models.ChatParticipant
		if err := a.DB.Where("chat_room_id = ? AND user_id = ?", room.ID, userID).First(&participant).Error; err != nil {
			// User not in room, add them
			a.DB.Create(&models.ChatParticipant{
				ChatRoomID: room.ID,
				UserID:     userID,
				JoinedAt:   time.Now(),
				Role:       models.ChatRoleMember,
			})
		}
		return &room, nil
	}

	// Get user info
	var user models.User
	if err := a.DB.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	// Create new AI chat room
	room = models.ChatRoom{
		Name:        roomName,
		Description: "Your personal AI assistant for task and project management",
		CreatedBy:   userID,
		MaxMembers:  2, // Only user and AI
	}

	if err := a.DB.Create(&room).Error; err != nil {
		return nil, err
	}

	// Add user as participant
	a.DB.Create(&models.ChatParticipant{
		ChatRoomID: room.ID,
		UserID:     userID,
		JoinedAt:   time.Now(),
		Role:       models.ChatRoleMember,
	})

	log.Printf("✅ Created AI chat room for user %d", userID)
	return &room, nil
}

// IsAIMessage checks if a message is directed to AI (starts with @ai, @assistant, or in AI room)
func (a *AIChatService) IsAIMessage(message string, roomID uint) bool {
	// Check if message mentions AI
	messageLower := strings.ToLower(strings.TrimSpace(message))
	if strings.HasPrefix(messageLower, "@ai") ||
		strings.HasPrefix(messageLower, "@assistant") ||
		strings.HasPrefix(messageLower, "ai:") ||
		strings.HasPrefix(messageLower, "assistant:") {
		return true
	}

	// Check if it's in an AI chat room
	var room models.ChatRoom
	if err := a.DB.First(&room, roomID).Error; err == nil {
		if strings.Contains(room.Name, "AI Assistant") {
			return true
		}
	}

	return false
}

// ProcessAIMessage processes a message and returns AI response
func (a *AIChatService) ProcessAIMessage(userID uint, message string, roomID uint) (*AIResponse, error) {
	if a.model == nil {
		return &AIResponse{
			Message:  "AI service is currently unavailable. Please check your GEMINI_API_KEY configuration.",
			Action:   "answer",
			Language: "en",
		}, nil
	}

	// Get user context
	context, err := a.getUserContext(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %v", err)
	}

	// Detect language (simple detection)
	context.Language = a.detectLanguage(message)

	// Check if it's a command (create task, get analytics, etc.)
	if response := a.handleCommand(message, context); response != nil {
		return response, nil
	}

	// Process as a question/chat
	return a.processQuestion(message, context, roomID)
}

// getUserContext gathers user context for AI
func (a *AIChatService) getUserContext(userID uint) (*UserChatContext, error) {
	var user models.User
	if err := a.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	// Get active tasks count
	var activeTasksCount int64
	a.DB.Model(&models.Task{}).
		Where("user_id = ? AND status IN ?", userID, []string{"pending", "in_progress"}).
		Count(&activeTasksCount)

	// Get user's projects
	var projects []models.Project
	a.DB.Joins("JOIN user_projects ON projects.id = user_projects.project_id").
		Where("user_projects.user_id = ?", userID).
		Find(&projects)

	projectNames := make([]string, len(projects))
	for i, p := range projects {
		projectNames[i] = p.Title
	}

	// Get current workload
	var totalHours int64
	a.DB.Raw(`
		SELECT COALESCE(SUM(predicted_duration), 0)
		FROM ai_analyses aa
		JOIN tasks t ON aa.task_id = t.id
		WHERE t.user_id = ? AND t.status IN ('pending', 'in_progress')
	`, userID).Scan(&totalHours)

	workload := float64(totalHours) / a.WorkSchedule.WeeklyHours * 100

	return &UserChatContext{
		UserID:          user.ID,
		Username:        user.Username,
		Role:            string(user.Role),
		Department:      user.Department,
		ActiveTasks:     int(activeTasksCount),
		Projects:        projectNames,
		CurrentWorkload: workload,
		Language:        "en", // Default, will be detected
	}, nil
}

// detectLanguage detects if message is in Arabic or English
func (a *AIChatService) detectLanguage(message string) string {
	// Simple detection: check for Arabic characters
	for _, r := range message {
		if r >= 0x0600 && r <= 0x06FF {
			return "ar"
		}
	}
	return "en"
}

// handleCommand processes commands like "create task", "get analytics", etc.
func (a *AIChatService) handleCommand(message string, userContext *UserChatContext) *AIResponse {
	messageLower := strings.ToLower(message)

	// Create task command
	if strings.Contains(messageLower, "create task") ||
		strings.Contains(messageLower, "new task") ||
		strings.Contains(messageLower, "add task") ||
		strings.Contains(messageLower, "إنشاء مهمة") ||
		strings.Contains(messageLower, "مهمة جديدة") {
		return a.handleCreateTaskCommand(message, userContext)
	}

	// Get tasks command
	if strings.Contains(messageLower, "my tasks") ||
		strings.Contains(messageLower, "what tasks") ||
		strings.Contains(messageLower, "show tasks") ||
		strings.Contains(messageLower, "مهامي") ||
		strings.Contains(messageLower, "ما هي المهام") {
		return a.handleGetTasksCommand(userContext)
	}

	// Get analytics/workload command
	if strings.Contains(messageLower, "workload") ||
		strings.Contains(messageLower, "capacity") ||
		strings.Contains(messageLower, "how busy") ||
		strings.Contains(messageLower, "عبء العمل") ||
		strings.Contains(messageLower, "كم أنا مشغول") {
		return a.handleGetWorkloadCommand(userContext)
	}

	return nil // Not a command, process as question
}

// processQuestion processes general questions using AI
func (a *AIChatService) processQuestion(message string, userContext *UserChatContext, roomID uint) (*AIResponse, error) {
	ctx := context.Background()

	// Get recent chat history for context
	recentMessages, _ := a.getRecentChatHistory(roomID, 5)

	// Build prompt with context
	prompt := a.buildChatPrompt(message, userContext, recentMessages)

	// Get AI response
	resp, err := a.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from AI")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	responseText = strings.TrimSpace(responseText)

	return &AIResponse{
		Message:  responseText,
		Action:   "answer",
		Language: userContext.Language,
	}, nil
}

// buildChatPrompt creates the AI prompt for chat
func (a *AIChatService) buildChatPrompt(message string, userContext *UserChatContext, recentMessages []models.ChatMessage) string {
	// Build context string
	projectsStr := strings.Join(userContext.Projects, ", ")
	if projectsStr == "" {
		projectsStr = "None"
	}

	// Build recent messages context
	recentContext := ""
	for i := len(recentMessages) - 1; i >= 0; i-- {
		msg := recentMessages[i]
		if msg.SenderID == userContext.UserID {
			recentContext += fmt.Sprintf("User: %s\n", msg.Content)
		} else {
			recentContext += fmt.Sprintf("AI: %s\n", msg.Content)
		}
	}

	languageInstruction := "Respond in English."
	if userContext.Language == "ar" {
		languageInstruction = "Respond in Arabic (العربية)."
	}

	prompt := fmt.Sprintf(`
You are an AI assistant for a project management system in an Arabic organization.

WORKING SCHEDULE CONTEXT:
- Working Days: Saturday to Thursday (6 days per week)
- Working Hours: 9:00 AM to 4:00 PM (7 hours with 1 hour lunch = 6 effective hours per day)
- Weekly Capacity: 42 working hours per person
- Friday is OFF (weekend)
- Timezone: Asia/Riyadh

USER CONTEXT:
- Username: %s
- Role: %s
- Department: %s
- Active Tasks: %d
- Projects: %s
- Current Workload: %.1f%% of weekly capacity

RECENT CONVERSATION:
%s

USER QUESTION/MESSAGE:
%s

INSTRUCTIONS:
1. %s
2. Be helpful, concise, and friendly
3. Provide actionable advice when possible
4. Consider Arabic working hours in your responses
5. If asked about tasks, provide specific details
6. If asked about time/workload, consider the 42-hour work week
7. You can help with:
   - Task management questions
   - Project status inquiries
   - Time management advice
   - Workload analysis
   - Priority suggestions
   - Best practices

Respond naturally and helpfully. If you don't understand something, ask for clarification.
`,
		userContext.Username,
		userContext.Role,
		userContext.Department,
		userContext.ActiveTasks,
		projectsStr,
		userContext.CurrentWorkload,
		recentContext,
		message,
		languageInstruction,
	)

	return prompt
}

// getRecentChatHistory gets recent messages from chat room
func (a *AIChatService) getRecentChatHistory(roomID uint, limit int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := a.DB.Where("chat_room_id = ?", roomID).
		Preload("Sender").
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, err
}

// handleCreateTaskCommand handles task creation from chat
func (a *AIChatService) handleCreateTaskCommand(message string, userContext *UserChatContext) *AIResponse {
	// Extract task details from message using AI
	taskDetails, err := a.extractTaskDetails(message, userContext)
	if err != nil {
		return &AIResponse{
			Message:  "I couldn't understand the task details. Please provide: task title, description (optional), and due date (optional).",
			Action:   "answer",
			Language: userContext.Language,
		}
	}

	// Get description (handle nil case)
	description := ""
	if taskDetails["description"] != nil {
		description = taskDetails["description"].(string)
	}

	// Create task using TaskService
	task, err := a.TaskService.CreateTask(
		taskDetails["title"].(string),
		description,
		userContext.UserID,
		nil, // projectID - can be enhanced later
		nil, // startTime
		nil, // endTime
		taskDetails["dueDate"].(*time.Time),
	)

	if err != nil {
		return &AIResponse{
			Message:  fmt.Sprintf("Failed to create task: %v", err),
			Action:   "answer",
			Language: userContext.Language,
		}
	}

	responseMsg := fmt.Sprintf("✅ Task created successfully!\n\nTitle: %s\nStatus: Pending\nTask ID: %d", task.Title, task.ID)
	if task.DueDate != nil {
		responseMsg += fmt.Sprintf("\nDue Date: %s", task.DueDate.Format("January 2, 2006"))
	}

	if userContext.Language == "ar" {
		responseMsg = fmt.Sprintf("✅ تم إنشاء المهمة بنجاح!\n\nالعنوان: %s\nالحالة: قيد الانتظار\nرقم المهمة: %d", task.Title, task.ID)
		if task.DueDate != nil {
			responseMsg += fmt.Sprintf("\nتاريخ الاستحقاق: %s", task.DueDate.Format("2 January 2006"))
		}
	}

	return &AIResponse{
		Message: responseMsg,
		Action:  "create_task",
		Data: map[string]interface{}{
			"task_id": task.ID,
			"title":   task.Title,
		},
		Language: userContext.Language,
	}
}

// extractTaskDetails uses AI to extract task information from natural language
func (a *AIChatService) extractTaskDetails(message string, userContext *UserChatContext) (map[string]interface{}, error) {
	ctx := context.Background()

	prompt := fmt.Sprintf(`
Extract task information from this message: "%s"

Return JSON with:
- title: string (required)
- description: string (optional, can be empty)
- dueDate: string in format "YYYY-MM-DD" or null

Example response:
{
  "title": "Review API documentation",
  "description": "Review and update API documentation for new endpoints",
  "dueDate": "2024-01-19"
}

Only return valid JSON, no other text.
`, message)

	resp, err := a.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, err
	}

	// Parse due date
	var dueDate *time.Time
	if result["dueDate"] != nil && result["dueDate"].(string) != "" {
		parsed, err := time.Parse("2006-01-02", result["dueDate"].(string))
		if err == nil {
			dueDate = &parsed
		}
	}

	return map[string]interface{}{
		"title":       result["title"],
		"description": result["description"],
		"dueDate":     dueDate,
	}, nil
}

// handleGetTasksCommand handles getting user's tasks
func (a *AIChatService) handleGetTasksCommand(userContext *UserChatContext) *AIResponse {
	var tasks []models.Task
	a.DB.Where("user_id = ? AND status IN ?", userContext.UserID, []string{"pending", "in_progress"}).
		Order("created_at DESC").
		Limit(10).
		Find(&tasks)

	if len(tasks) == 0 {
		msg := "You have no active tasks. Great job! 🎉"
		if userContext.Language == "ar" {
			msg = "ليس لديك مهام نشطة. عمل رائع! 🎉"
		}
		return &AIResponse{
			Message:  msg,
			Action:   "get_tasks",
			Language: userContext.Language,
		}
	}

	var taskList strings.Builder
	if userContext.Language == "ar" {
		taskList.WriteString(fmt.Sprintf("لديك %d مهام نشطة:\n\n", len(tasks)))
	} else {
		taskList.WriteString(fmt.Sprintf("You have %d active tasks:\n\n", len(tasks)))
	}

	for i, task := range tasks {
		status := string(task.Status)
		if userContext.Language == "ar" {
			if status == "pending" {
				status = "قيد الانتظار"
			} else if status == "in_progress" {
				status = "قيد التنفيذ"
			}
		}

		taskList.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, task.Title, status))
		if task.DueDate != nil {
			taskList.WriteString(fmt.Sprintf("   Due: %s\n", task.DueDate.Format("Jan 2, 2006")))
		}
	}

	return &AIResponse{
		Message: taskList.String(),
		Action:  "get_tasks",
		Data: map[string]interface{}{
			"task_count": len(tasks),
		},
		Language: userContext.Language,
	}
}

// handleGetWorkloadCommand handles workload queries
func (a *AIChatService) handleGetWorkloadCommand(userContext *UserChatContext) *AIResponse {
	workload := userContext.CurrentWorkload
	status := "healthy"
	statusEmoji := "✅"

	if workload > 120 {
		status = "critical overload"
		statusEmoji = "🔴"
	} else if workload > 100 {
		status = "overloaded"
		statusEmoji = "🟠"
	} else if workload > 80 {
		status = "high"
		statusEmoji = "🟡"
	}

	var msg string
	if userContext.Language == "ar" {
		msg = fmt.Sprintf("%s عبء العمل الحالي: %.1f%%\n\nالحالة: %s\nالسعة الأسبوعية: 42 ساعة\nالمهام النشطة: %d",
			statusEmoji, workload, status, userContext.ActiveTasks)
	} else {
		msg = fmt.Sprintf("%s Current workload: %.1f%%\n\nStatus: %s\nWeekly capacity: 42 hours\nActive tasks: %d",
			statusEmoji, workload, status, userContext.ActiveTasks)
	}

	return &AIResponse{
		Message: msg,
		Action:  "get_workload",
		Data: map[string]interface{}{
			"workload_percent": workload,
			"status":           status,
		},
		Language: userContext.Language,
	}
}

// Close closes the AI client
func (a *AIChatService) Close() error {
	if a.client != nil {
		return a.client.Close()
	}
	return nil
}
