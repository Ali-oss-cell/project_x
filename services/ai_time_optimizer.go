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

type AITimeOptimizer struct {
	DB           *gorm.DB
	client       *genai.Client
	model        *genai.GenerativeModel
	WorkSchedule *config.WorkScheduleConfig
}

type TimeAnalysis struct {
	TaskID                    uint      `json:"task_id"`
	TaskTitle                 string    `json:"task_title"`
	EstimatedDuration         int       `json:"estimated_duration_hours"`
	ActualDuration            int       `json:"actual_duration_hours"`
	DeadlineRisk              string    `json:"deadline_risk"` // low, medium, high, critical
	RiskFactors               []string  `json:"risk_factors"`
	Recommendations           []string  `json:"recommendations"`
	OptimalStartDate          time.Time `json:"optimal_start_date"`
	PredictedCompletion       time.Time `json:"predicted_completion"`
	WorkingHoursUntilDeadline float64   `json:"working_hours_until_deadline"`
	IsWithinWorkingHours      bool      `json:"is_within_working_hours"`
}

type ProjectTimeReport struct {
	ProjectID            uint           `json:"project_id"`
	ProjectTitle         string         `json:"project_title"`
	OverallRisk          string         `json:"overall_risk"`
	PredictedDelay       int            `json:"predicted_delay_days"`
	CriticalPath         []uint         `json:"critical_path_task_ids"`
	TimeOptimizations    []string       `json:"time_optimizations"`
	ResourceConflicts    []string       `json:"resource_conflicts"`
	TaskAnalyses         []TimeAnalysis `json:"task_analyses"`
	Summary              string         `json:"ai_summary"`
	WorkingDaysRemaining int            `json:"working_days_remaining"`
	TotalWorkingHours    float64        `json:"total_working_hours_required"`
}

type UserWorkloadAnalysis struct {
	UserID              uint     `json:"user_id"`
	Username            string   `json:"username"`
	TotalHoursAssigned  int      `json:"total_hours_assigned"`
	CapacityUtilization float64  `json:"capacity_utilization_percent"`
	OverloadRisk        string   `json:"overload_risk"`
	TimeConflicts       []string `json:"time_conflicts"`
	Recommendations     []string `json:"recommendations"`
	WeeklyCapacity      float64  `json:"weekly_capacity_hours"`
	CurrentWorkload     float64  `json:"current_workload_hours"`
}

func NewAITimeOptimizer(db *gorm.DB) *AITimeOptimizer {
	ctx := context.Background()

	// Get API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Printf("Warning: GEMINI_API_KEY not set, AI time optimization will be disabled")
		return &AITimeOptimizer{
			DB:           db,
			WorkSchedule: config.GetDefaultWorkSchedule(),
		}
	}

	// Initialize Gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Error creating Gemini client: %v", err)
		return &AITimeOptimizer{
			DB:           db,
			WorkSchedule: config.GetDefaultWorkSchedule(),
		}
	}

	// Configure the model
	model := client.GenerativeModel("gemini-2.0-flash")
	model.SetTemperature(0.3) // Lower temperature for more consistent analysis
	model.SetTopP(0.8)
	model.SetTopK(40)
	model.SetMaxOutputTokens(2048)

	return &AITimeOptimizer{
		DB:           db,
		client:       client,
		model:        model,
		WorkSchedule: config.GetDefaultWorkSchedule(),
	}
}

// AnalyzeTaskTimeRisks analyzes all tasks for time-related risks using Arabic working hours
func (a *AITimeOptimizer) AnalyzeTaskTimeRisks() ([]TimeAnalysis, error) {
	if a.model == nil {
		return nil, fmt.Errorf("AI service not available")
	}

	// Get all active tasks with time data
	var tasks []models.Task
	err := a.DB.Where("status IN ?", []string{"pending", "in_progress"}).
		Preload("User").
		Preload("Project").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	var analyses []TimeAnalysis

	for _, task := range tasks {
		analysis, err := a.analyzeIndividualTask(task)
		if err != nil {
			log.Printf("Error analyzing task %d: %v", task.ID, err)
			continue
		}
		analyses = append(analyses, *analysis)
	}

	return analyses, nil
}

// analyzeIndividualTask performs AI analysis on a single task with Arabic working hours
func (a *AITimeOptimizer) analyzeIndividualTask(task models.Task) (*TimeAnalysis, error) {
	ctx := context.Background()

	// Gather historical data for similar tasks
	historicalData, err := a.getHistoricalTaskData(task)
	if err != nil {
		return nil, err
	}

	// Calculate working hours until deadline
	now := time.Now()
	var workingHoursUntilDeadline float64
	var isWithinWorkingHours bool

	if task.DueDate != nil {
		workingHoursUntilDeadline = a.WorkSchedule.GetWorkingHoursInPeriod(now, *task.DueDate)
		isWithinWorkingHours = a.WorkSchedule.IsWorkingHour(now)
	}

	// Create AI prompt for task analysis with Arabic working context
	prompt := a.buildTaskAnalysisPromptArabic(task, historicalData, workingHoursUntilDeadline, isWithinWorkingHours)

	// Get AI analysis
	resp, err := a.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	// Parse AI response
	analysis, err := a.parseTaskAnalysisResponse(resp, task)
	if err != nil {
		return nil, err
	}

	// Add Arabic working hours context
	analysis.WorkingHoursUntilDeadline = workingHoursUntilDeadline
	analysis.IsWithinWorkingHours = isWithinWorkingHours

	// Calculate optimal start date based on Arabic working schedule
	analysis.OptimalStartDate = a.WorkSchedule.GetOptimalTaskSchedule(float64(analysis.EstimatedDuration), *task.DueDate)

	return analysis, nil
}

// buildTaskAnalysisPromptArabic creates the AI prompt with Arabic working hours context
func (a *AITimeOptimizer) buildTaskAnalysisPromptArabic(task models.Task, historicalData map[string]interface{}, workingHoursUntilDeadline float64, isWithinWorkingHours bool) string {
	now := time.Now()

	var daysUntilDue string
	var workingDaysUntilDue int
	if task.DueDate != nil {
		workingDaysUntilDue = a.WorkSchedule.GetWorkingDaysUntil(now, *task.DueDate)
		daysUntilDue = fmt.Sprintf("%d working days (%.1f working hours)", workingDaysUntilDue, workingHoursUntilDeadline)
	} else {
		daysUntilDue = "No deadline set"
	}

	var taskAge string
	age := int(now.Sub(task.CreatedAt).Hours() / 24)
	taskAge = fmt.Sprintf("%d calendar days", age)

	workingTimeStatus := "outside working hours"
	if isWithinWorkingHours {
		workingTimeStatus = "within working hours"
	}

	prompt := fmt.Sprintf(`
You are an AI project management expert analyzing tasks for an Arabic organization with specific working hours.

WORKING SCHEDULE CONTEXT:
- Working Days: Saturday to Thursday (6 days per week)
- Working Hours: 9:00 AM to 4:00 PM (7 hours with 1 hour lunch = 6 effective hours per day)
- Weekly Capacity: 42 working hours per person
- Timezone: Asia/Riyadh
- Current Time Status: %s

TASK DETAILS:
- Title: %s
- Description: %s
- Status: %s
- Created: %s ago
- Due Date: %s
- Assigned User: %s
- Working Hours Until Deadline: %.1f hours

HISTORICAL DATA:
- User's average task completion: %.1f days
- User's completion rate: %.1f%%%%
- User's total completed tasks: %.0f
- Current workload: %.0f active tasks
- Similar tasks completed: %.0f

PROJECT CONTEXT:
- Project total tasks: %.0f
- Project completed: %.0f
- Project overdue: %.0f

ANALYSIS REQUIRED (considering Arabic working schedule):
1. Estimate realistic completion time in working hours (6 hours per day, Saturday-Thursday)
2. Assess deadline risk level considering only working hours
3. Identify specific risk factors including working schedule constraints
4. Provide 3-5 actionable recommendations that respect working hours
5. Suggest optimal start time within working schedule
6. Predict actual completion date considering weekends (Friday off)

IMPORTANT CONSIDERATIONS:
- Friday is NOT a working day
- Only count working hours (9 AM - 4 PM, Saturday-Thursday)
- Account for lunch break (12-1 PM)
- Consider Ramadan/holiday adjustments if applicable
- Respect local business culture and practices

Respond in this exact JSON format:
{
  "estimated_duration_hours": <number>,
  "deadline_risk": "<low|medium|high|critical>",
  "risk_factors": ["factor1", "factor2", "factor3"],
  "recommendations": ["rec1", "rec2", "rec3"],
  "optimal_start_date": "<YYYY-MM-DD>",
  "predicted_completion": "<YYYY-MM-DD>",
  "confidence_score": <0-100>
}

Be specific about Arabic working schedule constraints in your recommendations.`,
		workingTimeStatus,
		task.Title,
		task.Description,
		task.Status,
		taskAge,
		daysUntilDue,
		task.User.Username,
		workingHoursUntilDeadline,
		historicalData["user_avg_completion_days"],
		historicalData["user_completion_rate"],
		historicalData["user_total_tasks"],
		historicalData["current_workload"],
		historicalData["similar_tasks_count"],
		historicalData["project_total_tasks"],
		historicalData["project_completed_tasks"],
		historicalData["project_overdue_tasks"],
	)

	return prompt
}

// AnalyzeUserWorkload analyzes individual user's time allocation with Arabic working hours
func (a *AITimeOptimizer) AnalyzeUserWorkload(userID uint) (*UserWorkloadAnalysis, error) {
	if a.model == nil {
		return nil, fmt.Errorf("AI service not available")
	}

	// Get user with active tasks
	var user models.User
	err := a.DB.Preload("Tasks", "status IN ?", []string{"pending", "in_progress"}).
		First(&user, userID).Error
	if err != nil {
		return nil, err
	}

	// Analyze user's tasks for time estimates
	totalEstimatedHours := 0
	var taskRisks []string

	for _, task := range user.Tasks {
		analysis, err := a.analyzeIndividualTask(task)
		if err != nil {
			continue
		}

		totalEstimatedHours += analysis.EstimatedDuration
		if analysis.DeadlineRisk == "high" || analysis.DeadlineRisk == "critical" {
			taskRisks = append(taskRisks, fmt.Sprintf("Task '%s' at %s risk", task.Title, analysis.DeadlineRisk))
		}
	}

	// Calculate capacity utilization based on Arabic working schedule (42 hours/week)
	weeklyCapacity := a.WorkSchedule.WeeklyHours // 42 hours
	utilizationPercent := float64(totalEstimatedHours) / weeklyCapacity * 100

	// Determine overload risk
	overloadRisk := "low"
	if utilizationPercent > 120 {
		overloadRisk = "critical"
	} else if utilizationPercent > 100 {
		overloadRisk = "high"
	} else if utilizationPercent > 80 {
		overloadRisk = "medium"
	}

	// Generate AI recommendations with Arabic context
	recommendations, err := a.generateUserRecommendationsArabic(user, totalEstimatedHours, utilizationPercent, taskRisks)
	if err != nil {
		recommendations = []string{"Unable to generate AI recommendations"}
	}

	analysis := &UserWorkloadAnalysis{
		UserID:              user.ID,
		Username:            user.Username,
		TotalHoursAssigned:  totalEstimatedHours,
		CapacityUtilization: utilizationPercent,
		OverloadRisk:        overloadRisk,
		TimeConflicts:       taskRisks,
		Recommendations:     recommendations,
		WeeklyCapacity:      weeklyCapacity,
		CurrentWorkload:     float64(totalEstimatedHours),
	}

	return analysis, nil
}

// generateUserRecommendationsArabic generates AI-powered recommendations considering Arabic working schedule
func (a *AITimeOptimizer) generateUserRecommendationsArabic(user models.User, totalHours int, utilization float64, risks []string) ([]string, error) {
	ctx := context.Background()

	prompt := fmt.Sprintf(`
You are an AI workload optimization expert for an Arabic organization. Analyze this user's workload considering Arabic working schedule.

WORKING SCHEDULE CONTEXT:
- Working Days: Saturday to Thursday (6 days per week)
- Working Hours: 9:00 AM to 4:00 PM (7 hours with 1 hour lunch = 6 effective hours per day)
- Weekly Capacity: 42 working hours
- Friday is OFF (weekend)
- Lunch break: 12:00-13:00 PM
- Timezone: Asia/Riyadh

USER DETAILS:
- Username: %s
- Role: %s
- Department: %s
- Total Estimated Hours: %d
- Capacity Utilization: %.1f%% (based on 42-hour work week)
- Active Tasks: %d
- Time Conflicts: %s

CULTURAL CONSIDERATIONS:
- Respect for work-life balance
- Friday is sacred day off
- Potential Ramadan schedule adjustments
- Local business practices and customs

Provide 3-5 specific, actionable recommendations considering Arabic working schedule and culture.
Focus on practical solutions that respect working hours and cultural values.

Respond with a JSON array of recommendation strings:
["recommendation1", "recommendation2", "recommendation3"]

Examples of culturally appropriate recommendations:
- "Redistribute 2 tasks to reduce workload to 85%% of 42-hour capacity"
- "Schedule complex tasks during peak hours (10 AM - 12 PM)"
- "Plan task completion before Thursday to avoid weekend pressure"
- "Use Saturday morning for planning and Wednesday afternoon for reviews"`,
		user.Username,
		user.Role,
		user.Department,
		totalHours,
		utilization,
		len(user.Tasks),
		strings.Join(risks, "; "),
	)

	resp, err := a.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from AI")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	var recommendations []string
	err = json.Unmarshal([]byte(responseText), &recommendations)
	if err != nil {
		return []string{"Unable to parse AI recommendations"}, nil
	}

	return recommendations, nil
}

// getHistoricalTaskData retrieves historical data for similar tasks
func (a *AITimeOptimizer) getHistoricalTaskData(task models.Task) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Get user's historical performance
	var userStats struct {
		AvgCompletionDays float64
		CompletionRate    float64
		TotalTasks        int64
	}

	// Calculate average completion time for this user
	a.DB.Raw(`
		SELECT 
			AVG(EXTRACT(EPOCH FROM (updated_at - created_at))/86400) as avg_completion_days,
			COUNT(CASE WHEN status = 'completed' THEN 1 END)::float / COUNT(*)::float * 100 as completion_rate,
			COUNT(*) as total_tasks
		FROM tasks 
		WHERE user_id = ? AND status IN ('completed', 'cancelled')
	`, task.UserID).Scan(&userStats)

	data["user_avg_completion_days"] = userStats.AvgCompletionDays
	data["user_completion_rate"] = userStats.CompletionRate
	data["user_total_tasks"] = userStats.TotalTasks

	// Get similar tasks (same user, similar keywords)
	var similarTasks []models.Task
	a.DB.Where("user_id = ? AND status = 'completed' AND id != ?", task.UserID, task.ID).
		Limit(5).Find(&similarTasks)

	data["similar_tasks_count"] = len(similarTasks)

	// Get current workload
	var currentWorkload int64
	a.DB.Model(&models.Task{}).
		Where("user_id = ? AND status IN ?", task.UserID, []string{"pending", "in_progress"}).
		Count(&currentWorkload)

	data["current_workload"] = currentWorkload

	// Get project timeline pressure
	if task.ProjectID != nil {
		var projectPressure struct {
			TotalTasks     int64
			CompletedTasks int64
			OverdueTasks   int64
		}

		a.DB.Raw(`
			SELECT 
				COUNT(*) as total_tasks,
				COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_tasks,
				COUNT(CASE WHEN due_date < NOW() AND status != 'completed' THEN 1 END) as overdue_tasks
			FROM tasks 
			WHERE project_id = ?
		`, *task.ProjectID).Scan(&projectPressure)

		data["project_total_tasks"] = projectPressure.TotalTasks
		data["project_completed_tasks"] = projectPressure.CompletedTasks
		data["project_overdue_tasks"] = projectPressure.OverdueTasks
	}

	return data, nil
}

// parseTaskAnalysisResponse parses the AI response into TimeAnalysis struct
func (a *AITimeOptimizer) parseTaskAnalysisResponse(resp *genai.GenerateContentResponse, task models.Task) (*TimeAnalysis, error) {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from AI")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Clean the response (remove markdown code blocks if present)
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	// Parse JSON response
	var aiResponse struct {
		EstimatedDurationHours int      `json:"estimated_duration_hours"`
		DeadlineRisk           string   `json:"deadline_risk"`
		RiskFactors            []string `json:"risk_factors"`
		Recommendations        []string `json:"recommendations"`
		OptimalStartDate       string   `json:"optimal_start_date"`
		PredictedCompletion    string   `json:"predicted_completion"`
		ConfidenceScore        int      `json:"confidence_score"`
	}

	err := json.Unmarshal([]byte(responseText), &aiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %v", err)
	}

	// Parse dates
	optimalStart, _ := time.Parse("2006-01-02", aiResponse.OptimalStartDate)
	predictedCompletion, _ := time.Parse("2006-01-02", aiResponse.PredictedCompletion)

	analysis := &TimeAnalysis{
		TaskID:              task.ID,
		TaskTitle:           task.Title,
		EstimatedDuration:   aiResponse.EstimatedDurationHours,
		DeadlineRisk:        aiResponse.DeadlineRisk,
		RiskFactors:         aiResponse.RiskFactors,
		Recommendations:     aiResponse.Recommendations,
		OptimalStartDate:    optimalStart,
		PredictedCompletion: predictedCompletion,
	}

	return analysis, nil
}

// AnalyzeProjectTimeRisks analyzes entire project for time optimization
func (a *AITimeOptimizer) AnalyzeProjectTimeRisks(projectID uint) (*ProjectTimeReport, error) {
	if a.model == nil {
		return nil, fmt.Errorf("AI service not available")
	}

	// Get project with all tasks
	var project models.Project
	err := a.DB.Preload("Tasks.User").
		Preload("CollaborativeTasks.LeadUser").
		First(&project, projectID).Error
	if err != nil {
		return nil, err
	}

	// Analyze individual tasks
	var taskAnalyses []TimeAnalysis
	totalWorkingHours := 0.0
	for _, task := range project.Tasks {
		if task.Status == "completed" || task.Status == "cancelled" {
			continue
		}

		analysis, err := a.analyzeIndividualTask(task)
		if err != nil {
			log.Printf("Error analyzing task %d: %v", task.ID, err)
			continue
		}
		taskAnalyses = append(taskAnalyses, *analysis)
		totalWorkingHours += float64(analysis.EstimatedDuration)
	}

	// Calculate working days remaining until project deadline
	workingDaysRemaining := 0
	if project.EndDate != nil {
		workingDaysRemaining = a.WorkSchedule.GetWorkingDaysUntil(time.Now(), *project.EndDate)
	}

	// Generate project-level analysis
	projectReport, err := a.generateProjectAnalysis(project, taskAnalyses)
	if err != nil {
		return nil, err
	}

	// Add Arabic working schedule context
	projectReport.WorkingDaysRemaining = workingDaysRemaining
	projectReport.TotalWorkingHours = totalWorkingHours

	return projectReport, nil
}

// generateProjectAnalysis creates AI-powered project-level time analysis
func (a *AITimeOptimizer) generateProjectAnalysis(project models.Project, taskAnalyses []TimeAnalysis) (*ProjectTimeReport, error) {
	ctx := context.Background()

	// Build project analysis prompt
	prompt := a.buildProjectAnalysisPrompt(project, taskAnalyses)

	// Get AI analysis
	resp, err := a.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	// Parse response
	report, err := a.parseProjectAnalysisResponse(resp, project, taskAnalyses)
	if err != nil {
		return nil, err
	}

	return report, nil
}

// buildProjectAnalysisPrompt creates prompt for project-level analysis
func (a *AITimeOptimizer) buildProjectAnalysisPrompt(project models.Project, taskAnalyses []TimeAnalysis) string {
	now := time.Now()

	var projectDeadline string
	var workingDaysRemaining int
	if project.EndDate != nil {
		workingDaysRemaining = a.WorkSchedule.GetWorkingDaysUntil(now, *project.EndDate)
		projectDeadline = fmt.Sprintf("%d working days", workingDaysRemaining)
	} else {
		projectDeadline = "No deadline set"
	}

	// Summarize task analyses
	criticalTasks := 0
	highRiskTasks := 0
	totalEstimatedHours := 0

	for _, analysis := range taskAnalyses {
		totalEstimatedHours += analysis.EstimatedDuration
		switch analysis.DeadlineRisk {
		case "critical":
			criticalTasks++
		case "high":
			highRiskTasks++
		}
	}

	prompt := fmt.Sprintf(`
You are an AI project management expert analyzing a project for an Arabic organization.

WORKING SCHEDULE CONTEXT:
- Working Days: Saturday to Thursday (6 days per week)
- Working Hours: 9:00 AM to 4:00 PM (7 hours with 1 hour lunch = 6 effective hours per day)
- Weekly Capacity: 42 working hours per person
- Friday is OFF (weekend)

PROJECT DETAILS:
- Title: %s
- Description: %s
- Status: %s
- Project Deadline: %s
- Working Days Remaining: %d
- Total Active Tasks: %d
- Total Estimated Hours: %d
- Critical Risk Tasks: %d
- High Risk Tasks: %d

INDIVIDUAL TASK ANALYSES:
%s

ANALYSIS REQUIRED (considering Arabic working schedule):
1. Overall project risk assessment considering working days only
2. Predicted delay in working days (not calendar days)
3. Identify critical path considering Arabic work week
4. Suggest 5-7 specific time optimizations respecting working schedule
5. Identify resource conflicts within working hours
6. Provide executive summary considering cultural context

Respond in this exact JSON format:
{
  "overall_risk": "<low|medium|high|critical>",
  "predicted_delay_days": <number>,
  "critical_path_task_ids": [<task_id1>, <task_id2>],
  "time_optimizations": ["optimization1", "optimization2"],
  "resource_conflicts": ["conflict1", "conflict2"],
  "executive_summary": "<2-3 sentence summary considering Arabic business context>"
}

Focus on actionable recommendations that respect Arabic working schedule and business culture.`,
		project.Title,
		project.Description,
		project.Status,
		projectDeadline,
		workingDaysRemaining,
		len(taskAnalyses),
		totalEstimatedHours,
		criticalTasks,
		highRiskTasks,
		a.formatTaskAnalysesForPrompt(taskAnalyses),
	)

	return prompt
}

// formatTaskAnalysesForPrompt formats task analyses for the AI prompt
func (a *AITimeOptimizer) formatTaskAnalysesForPrompt(analyses []TimeAnalysis) string {
	var formatted strings.Builder

	for i, analysis := range analyses {
		if i >= 10 { // Limit to first 10 tasks to avoid prompt length issues
			formatted.WriteString("... (additional tasks omitted)\n")
			break
		}

		formatted.WriteString(fmt.Sprintf(
			"Task %d: %s (Risk: %s, Est: %dh, Working Hours Until Deadline: %.1f, Factors: %s)\n",
			analysis.TaskID,
			analysis.TaskTitle,
			analysis.DeadlineRisk,
			analysis.EstimatedDuration,
			analysis.WorkingHoursUntilDeadline,
			strings.Join(analysis.RiskFactors, ", "),
		))
	}

	return formatted.String()
}

// parseProjectAnalysisResponse parses project analysis response
func (a *AITimeOptimizer) parseProjectAnalysisResponse(resp *genai.GenerateContentResponse, project models.Project, taskAnalyses []TimeAnalysis) (*ProjectTimeReport, error) {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from AI")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	var aiResponse struct {
		OverallRisk         string   `json:"overall_risk"`
		PredictedDelayDays  int      `json:"predicted_delay_days"`
		CriticalPathTaskIDs []uint   `json:"critical_path_task_ids"`
		TimeOptimizations   []string `json:"time_optimizations"`
		ResourceConflicts   []string `json:"resource_conflicts"`
		ExecutiveSummary    string   `json:"executive_summary"`
	}

	err := json.Unmarshal([]byte(responseText), &aiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %v", err)
	}

	report := &ProjectTimeReport{
		ProjectID:         project.ID,
		ProjectTitle:      project.Title,
		OverallRisk:       aiResponse.OverallRisk,
		PredictedDelay:    aiResponse.PredictedDelayDays,
		CriticalPath:      aiResponse.CriticalPathTaskIDs,
		TimeOptimizations: aiResponse.TimeOptimizations,
		ResourceConflicts: aiResponse.ResourceConflicts,
		TaskAnalyses:      taskAnalyses,
		Summary:           aiResponse.ExecutiveSummary,
	}

	return report, nil
}

// GetCriticalTimeAlerts returns urgent time-related alerts considering Arabic working schedule
func (a *AITimeOptimizer) GetCriticalTimeAlerts() ([]map[string]interface{}, error) {
	var alerts []map[string]interface{}

	// Get tasks with critical deadline risks
	analyses, err := a.AnalyzeTaskTimeRisks()
	if err != nil {
		return nil, err
	}

	for _, analysis := range analyses {
		switch analysis.DeadlineRisk {
		case "critical", "high":
			alert := map[string]interface{}{
				"type":                    "deadline_risk",
				"severity":                analysis.DeadlineRisk,
				"task_id":                 analysis.TaskID,
				"task_title":              analysis.TaskTitle,
				"message":                 fmt.Sprintf("Task '%s' at %s deadline risk (%.1f working hours remaining)", analysis.TaskTitle, analysis.DeadlineRisk, analysis.WorkingHoursUntilDeadline),
				"risk_factors":            analysis.RiskFactors,
				"actions":                 analysis.Recommendations,
				"working_hours_remaining": analysis.WorkingHoursUntilDeadline,
				"is_within_working_hours": analysis.IsWithinWorkingHours,
				"created_at":              time.Now(),
			}
			alerts = append(alerts, alert)
		}
	}

	// Check for overloaded users based on 42-hour work week
	var users []models.User
	a.DB.Find(&users)

	for _, user := range users {
		workload, err := a.AnalyzeUserWorkload(user.ID)
		if err != nil {
			continue
		}

		switch workload.OverloadRisk {
		case "critical", "high":
			alert := map[string]interface{}{
				"type":             "user_overload",
				"severity":         workload.OverloadRisk,
				"user_id":          user.ID,
				"username":         user.Username,
				"message":          fmt.Sprintf("User %s at %s overload risk (%.1f%% of 42-hour capacity)", user.Username, workload.OverloadRisk, workload.CapacityUtilization),
				"utilization":      workload.CapacityUtilization,
				"weekly_capacity":  workload.WeeklyCapacity,
				"current_workload": workload.CurrentWorkload,
				"actions":          workload.Recommendations,
				"created_at":       time.Now(),
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// Close closes the AI client connection
func (a *AITimeOptimizer) Close() error {
	if a.client != nil {
		return a.client.Close()
	}
	return nil
}
