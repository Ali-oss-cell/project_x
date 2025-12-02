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

type AIProjectTaskGenerator struct {
	DB           *gorm.DB
	client       *genai.Client
	model        *genai.GenerativeModel
	WorkSchedule *config.WorkScheduleConfig
	TaskService  *TaskService
}

type GeneratedTask struct {
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	AssignedToID   uint       `json:"assigned_to_user_id"`
	EstimatedHours int        `json:"estimated_hours"`
	Priority       string     `json:"priority"`     // high, medium, low
	Dependencies   []uint     `json:"dependencies"` // IDs of tasks this depends on
	StartTime      *time.Time `json:"start_time,omitempty"`
	EndTime        *time.Time `json:"end_time,omitempty"`
}

type ProjectTaskGenerationRequest struct {
	ProjectID    uint
	ProjectTitle string
	Description  string
	TeamMembers  []TeamMemberInfo
	StartDate    time.Time
	EndDate      *time.Time
}

type TeamMemberInfo struct {
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	JobRole    string `json:"job_role"` // Project-specific job role
	Role       string `json:"role"`     // Project management role
	Department string `json:"department"`
	Skills     string `json:"skills"` // User's skills (JSON array)
}

type TaskGenerationResponse struct {
	Tasks   []GeneratedTask `json:"tasks"`
	Summary string          `json:"summary"`
}

func NewAIProjectTaskGenerator(db *gorm.DB, taskService *TaskService) *AIProjectTaskGenerator {
	ctx := context.Background()

	// Get API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Printf("Warning: GEMINI_API_KEY not set, AI task generation will be disabled")
		return &AIProjectTaskGenerator{
			DB:           db,
			WorkSchedule: config.GetDefaultWorkSchedule(),
			TaskService:  taskService,
		}
	}

	// Initialize Gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Error creating Gemini client for task generation: %v", err)
		return &AIProjectTaskGenerator{
			DB:           db,
			WorkSchedule: config.GetDefaultWorkSchedule(),
			TaskService:  taskService,
		}
	}

	// Configure the model
	model := client.GenerativeModel("gemini-2.0-flash")
	model.SetTemperature(0.7) // Higher temperature for more creative task generation
	model.SetTopP(0.9)
	model.SetTopK(40)
	model.SetMaxOutputTokens(4096)

	return &AIProjectTaskGenerator{
		DB:           db,
		client:       client,
		model:        model,
		WorkSchedule: config.GetDefaultWorkSchedule(),
		TaskService:  taskService,
	}
}

// GenerateProjectTasks generates tasks for a project based on description and team members
func (a *AIProjectTaskGenerator) GenerateProjectTasks(req ProjectTaskGenerationRequest) (*TaskGenerationResponse, error) {
	if a.model == nil {
		return nil, fmt.Errorf("AI service not available")
	}

	// Build prompt for AI
	prompt := a.buildTaskGenerationPrompt(req)

	// Get AI response
	ctx := context.Background()
	resp, err := a.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty AI response")
	}

	// Parse AI response
	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	responseText = strings.TrimSpace(responseText)

	// Extract JSON from response (AI might wrap it in markdown code blocks)
	jsonStart := strings.Index(responseText, "{")
	jsonEnd := strings.LastIndex(responseText, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("invalid AI response format")
	}

	jsonStr := responseText[jsonStart : jsonEnd+1]

	var aiResponse struct {
		Tasks   []GeneratedTask `json:"tasks"`
		Summary string          `json:"summary"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &aiResponse); err != nil {
		log.Printf("Failed to parse AI response: %v\nResponse: %s", err, jsonStr)
		return nil, fmt.Errorf("failed to parse AI response: %v", err)
	}

	return &TaskGenerationResponse{
		Tasks:   aiResponse.Tasks,
		Summary: aiResponse.Summary,
	}, nil
}

// getTeamMemberHistoricalData retrieves historical performance data for a team member
func (a *AIProjectTaskGenerator) getTeamMemberHistoricalData(userID uint) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Get user's historical performance
	var userStats struct {
		AvgCompletionDays float64
		CompletionRate    float64
		TotalTasks        int64
		AvgTaskHours      float64
	}

	// Calculate average completion time for this user
	a.DB.Raw(`
		SELECT 
			AVG(EXTRACT(EPOCH FROM (updated_at - created_at))/86400) as avg_completion_days,
			COUNT(CASE WHEN status = 'completed' THEN 1 END)::float / COUNT(*)::float * 100 as completion_rate,
			COUNT(*) as total_tasks,
			AVG(EXTRACT(EPOCH FROM (end_time - start_time))/3600) as avg_task_hours
		FROM tasks 
		WHERE user_id = ? AND status IN ('completed', 'cancelled') AND start_time IS NOT NULL AND end_time IS NOT NULL
	`, userID).Scan(&userStats)

	data["avg_completion_days"] = userStats.AvgCompletionDays
	data["completion_rate"] = userStats.CompletionRate
	data["total_tasks"] = userStats.TotalTasks
	data["avg_task_hours"] = userStats.AvgTaskHours

	// Get current workload
	var currentWorkload int64
	a.DB.Model(&models.Task{}).
		Where("user_id = ? AND status IN ?", userID, []string{"pending", "in_progress"}).
		Count(&currentWorkload)

	data["current_workload"] = currentWorkload

	// Get AI prediction accuracy for this user
	var aiAccuracy struct {
		AvgAccuracyScore    float64
		TotalPredictions    int64
		AccuratePredictions int64
		AvgDurationDiff     float64
	}

	a.DB.Raw(`
		SELECT 
			COALESCE(AVG(aa.accuracy_score), 0) as avg_accuracy_score,
			COUNT(*) as total_predictions,
			COUNT(CASE WHEN aa.was_accurate = true THEN 1 END) as accurate_predictions,
			COALESCE(AVG(aa.duration_difference), 0) as avg_duration_diff
		FROM ai_analyses aa
		JOIN tasks t ON aa.task_id = t.id
		WHERE t.user_id = ? AND aa.actual_duration IS NOT NULL
	`, userID).Scan(&aiAccuracy)

	data["ai_avg_accuracy"] = aiAccuracy.AvgAccuracyScore
	data["ai_total_predictions"] = aiAccuracy.TotalPredictions
	data["ai_accuracy_rate"] = 0.0
	if aiAccuracy.TotalPredictions > 0 {
		data["ai_accuracy_rate"] = float64(aiAccuracy.AccuratePredictions) / float64(aiAccuracy.TotalPredictions) * 100
	}
	data["ai_avg_duration_diff"] = aiAccuracy.AvgDurationDiff

	// Get similar completed tasks for this user (for pattern recognition)
	var similarTasksCount int64
	a.DB.Model(&models.Task{}).
		Where("user_id = ? AND status = 'completed'", userID).
		Count(&similarTasksCount)

	data["completed_tasks_count"] = similarTasksCount

	return data, nil
}

// buildTaskGenerationPrompt creates the AI prompt for task generation
func (a *AIProjectTaskGenerator) buildTaskGenerationPrompt(req ProjectTaskGenerationRequest) string {
	// Build team members info with historical data
	teamInfo := ""
	for i, member := range req.TeamMembers {
		// Parse skills JSON
		var skills []string
		if member.Skills != "" {
			json.Unmarshal([]byte(member.Skills), &skills)
		}
		skillsStr := strings.Join(skills, ", ")
		if skillsStr == "" {
			skillsStr = "None specified"
		}

		// Get historical performance data
		historicalData, err := a.getTeamMemberHistoricalData(member.UserID)
		if err != nil {
			log.Printf("Warning: Failed to get historical data for user %d: %v", member.UserID, err)
			historicalData = make(map[string]interface{})
		}

		// Format historical data
		avgCompletionDays := 0.0
		completionRate := 0.0
		totalTasks := int64(0)
		avgTaskHours := 0.0
		currentWorkload := int64(0)
		aiAccuracyRate := 0.0
		aiAvgDurationDiff := 0.0

		if val, ok := historicalData["avg_completion_days"].(float64); ok {
			avgCompletionDays = val
		}
		if val, ok := historicalData["completion_rate"].(float64); ok {
			completionRate = val
		}
		if val, ok := historicalData["total_tasks"].(int64); ok {
			totalTasks = val
		}
		if val, ok := historicalData["avg_task_hours"].(float64); ok {
			avgTaskHours = val
		}
		if val, ok := historicalData["current_workload"].(int64); ok {
			currentWorkload = val
		}
		if val, ok := historicalData["ai_accuracy_rate"].(float64); ok {
			aiAccuracyRate = val
		}
		if val, ok := historicalData["ai_avg_duration_diff"].(float64); ok {
			aiAvgDurationDiff = val
		}

		// Build historical info string
		historicalInfo := ""
		if totalTasks > 0 {
			historicalInfo = fmt.Sprintf(`
   - Historical Performance:
     * Total Completed Tasks: %d
     * Average Completion Time: %.1f days
     * Completion Rate: %.1f%%
     * Average Task Duration: %.1f hours
     * Current Active Tasks: %d
     * AI Prediction Accuracy: %.1f%% (avg difference: %.1f hours)`,
				totalTasks,
				avgCompletionDays,
				completionRate,
				avgTaskHours,
				currentWorkload,
				aiAccuracyRate,
				aiAvgDurationDiff,
			)
		} else {
			historicalInfo = "\n   - Historical Performance: No previous task data (new user)"
		}

		teamInfo += fmt.Sprintf(`
%d. %s (ID: %d)
   - Job Role in Project: %s
   - Project Management Role: %s
   - Department: %s
   - Available Skills: %s%s
`, i+1, member.Username, member.UserID, member.JobRole, member.Role, member.Department, skillsStr, historicalInfo)
	}

	// Calculate working days
	workingDaysRemaining := 0
	if req.EndDate != nil {
		workingDaysRemaining = a.WorkSchedule.GetWorkingDaysUntil(req.StartDate, *req.EndDate)
	}

	// Build prompt
	prompt := fmt.Sprintf(`
You are an AI project planning assistant for an Arabic organization. Analyze this project and generate specific, actionable tasks.

PROJECT INFORMATION:
- Title: %s
- Description: %s
- Start Date: %s
- End Date: %s
- Working Days Remaining: %d days

TEAM MEMBERS (with historical performance data):
%s

WORKING SCHEDULE CONTEXT:
- Working Days: Saturday to Thursday (6 days per week)
- Working Hours: 9:00 AM to 4:00 PM (7 hours with 1 hour lunch = 6 effective hours per day)
- Weekly Capacity: 42 working hours per person
- Friday is OFF (weekend)
- Timezone: Asia/Riyadh

INSTRUCTIONS:
1. Break down the project into specific, actionable tasks
2. Assign tasks to team members based on their JOB ROLE in the project (not their project management role)
3. Create task dependencies (which tasks must be completed before others can start)
4. Estimate duration in hours for each task:
   - Use historical performance data to inform estimates
   - If user has completed similar tasks, use their average completion time as reference
   - Consider AI prediction accuracy: if user's tasks typically take longer/shorter than predicted, adjust accordingly
   - Consider current workload: users with more active tasks may need more time
   - Always consider Arabic working hours (6 effective hours per day)
5. Set priorities (high/medium/low) based on project critical path
6. Consider the working schedule and available time
7. Ensure tasks are specific and measurable
8. Balance workload across team members (consider their current active tasks)
9. Create logical task sequences with proper dependencies

OUTPUT FORMAT (JSON only, no markdown):
{
  "tasks": [
    {
      "title": "Task title (specific and actionable)",
      "description": "Detailed description of what needs to be done",
      "assigned_to_user_id": 123,
      "estimated_hours": 8,
      "priority": "high",
      "dependencies": [],
      "start_time": "2024-01-15T09:00:00Z",
      "end_time": "2024-01-16T17:00:00Z"
    }
  ],
  "summary": "Brief summary of the generated task plan"
}

IMPORTANT:
- Assign tasks based on JOB ROLE (e.g., "UX/UI Designer" gets design tasks, "Backend Developer" gets API tasks)
- Use historical performance data to create realistic time estimates:
  * If user averages 2 days for similar tasks, estimate accordingly
  * If AI predictions for this user are typically off by +2 hours, add buffer time
  * If user has high current workload, add 10-20%% more time
- Create realistic time estimates based on task complexity AND user's historical performance
- Ensure dependencies make logical sense
- Balance work across all team members (consider their current workload)
- Consider Arabic working hours in time estimates

Generate tasks now:
`,
		req.ProjectTitle,
		req.Description,
		req.StartDate.Format("2006-01-02"),
		formatDate(req.EndDate),
		workingDaysRemaining,
		teamInfo,
	)

	return prompt
}

// CreateTasksFromGeneration creates actual tasks in the database from generated tasks
func (a *AIProjectTaskGenerator) CreateTasksFromGeneration(projectID uint, generatedTasks []GeneratedTask) ([]models.Task, error) {
	var createdTasks []models.Task
	taskMap := make(map[uint]*models.Task) // Map for dependency resolution

	// First pass: Create all tasks without dependencies
	for i, genTask := range generatedTasks {
		task, err := a.TaskService.CreateTask(
			genTask.Title,
			genTask.Description,
			genTask.AssignedToID,
			&projectID,
			genTask.StartTime,
			genTask.EndTime,
			nil, // DueDate - can be calculated from EndTime
		)
		if err != nil {
			log.Printf("Failed to create task %d: %v", i+1, err)
			continue
		}

		// Store task with a temporary ID for dependency mapping
		// We'll use the index as temporary ID
		taskMap[uint(i)] = task
		createdTasks = append(createdTasks, *task)
	}

	// Note: Dependencies are stored in the GeneratedTask but we don't have a dependency system
	// in the current Task model. This could be enhanced in the future.
	// For now, tasks are created and can be manually ordered if needed.

	return createdTasks, nil
}

// Helper function to format date
func formatDate(date *time.Time) string {
	if date == nil {
		return "Not specified"
	}
	return date.Format("2006-01-02")
}

// Close closes the AI client
func (a *AIProjectTaskGenerator) Close() error {
	if a.client != nil {
		return a.client.Close()
	}
	return nil
}
