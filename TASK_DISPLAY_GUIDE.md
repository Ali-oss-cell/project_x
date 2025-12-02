# Task Display Guide - Weekly vs Monthly Views

## Overview

This document explains how tasks are displayed in projects, what information should be shown to users, and how weekly/monthly views work.

## Current Task Model Structure

### Regular Tasks
Tasks have the following fields available:

```go
type Task struct {
    ID          uint       // Task ID
    Title       string     // Task title
    Description string     // Task description
    Status      TaskStatus // pending, in_progress, completed, cancelled
    UserID      uint       // Assigned user ID
    ProjectID   *uint      // Optional: project this task belongs to
    AssignedAt  time.Time  // When task was assigned
    StartTime   *time.Time // When task should start (optional)
    EndTime     *time.Time // When task should end (optional)
    DueDate     *time.Time // Optional due date
    CreatedAt   time.Time  // When task was created
    UpdatedAt   time.Time  // Last update time
}
```

### Collaborative Tasks
Collaborative tasks have additional fields:

```go
type CollaborativeTask struct {
    // Same as Task, plus:
    LeadUserID      uint       // Main person responsible
    Priority        string     // high, medium, low
    Progress        int        // 0-100 percentage
    Complexity      string     // simple, medium, complex
    MaxParticipants int        // Maximum number of participants
    Participants    []CollaborativeTaskParticipant // Team members
}
```

## Current API Response - What's Currently Shown

### Regular Tasks Response
Currently, when fetching tasks, the API returns:

```json
{
  "id": 1,
  "title": "Task Title",
  "description": "Task Description",
  "status": "pending",
  "user_id": 5,
  "project_id": 2,
  "assigned_at": "2024-01-15T10:00:00Z",
  "due_date": "2024-01-20T17:00:00Z",
  "created_at": "2024-01-15T10:00:00Z"
}
```

**Missing Fields in Current Response:**
- ❌ `start_time` - Not included (but exists in model)
- ❌ `end_time` - Not included (but exists in model)
- ❌ `updated_at` - Not included
- ❌ User name/username - Only user_id is shown
- ❌ Project name - Only project_id is shown

### Collaborative Tasks Response
```json
{
  "id": 1,
  "title": "Collaborative Task",
  "description": "Description",
  "status": "in_progress",
  "lead_user_id": 5,
  "lead_user_name": "john_doe",  // ✅ Included
  "project_id": 2,
  "assigned_at": "2024-01-15T10:00:00Z",
  "due_date": "2024-01-20T17:00:00Z",
  "created_at": "2024-01-15T10:00:00Z"
}
```

**Missing Fields:**
- ❌ `start_time` - Not included
- ❌ `end_time` - Not included
- ❌ `priority` - Not included (but exists in model)
- ❌ `progress` - Not included (but exists in model)
- ❌ `complexity` - Not included (but exists in model)
- ❌ `participants` - Not included

## Weekly vs Monthly Views

### Current Implementation

The system supports **weekly** and **monthly** periods for **reports only**, not for task grouping/display:

#### Project Reports
```
GET /api/tasks/reports/project/:projectId?period=weekly
GET /api/tasks/reports/project/:projectId?period=monthly
```

**What Reports Show:**
- Statistics (counts by status)
- Tasks created in the period
- Tasks completed in the period
- User performance metrics
- Completion rates

**What Reports DON'T Show:**
- ❌ Actual task list grouped by week/month
- ❌ Tasks organized by time periods
- ❌ Calendar/timeline view data

### What's Missing for Weekly/Monthly Task Display

Currently, there's **NO endpoint** that:
1. Groups tasks by week or month
2. Returns tasks organized in a calendar/timeline format
3. Filters tasks by specific week/month ranges

## Recommended Task Information to Show Users

### Essential Information (Must Show)

1. **Basic Task Info**
   - ✅ Task ID
   - ✅ Title
   - ✅ Description
   - ✅ Status (with color coding: pending=yellow, in_progress=blue, completed=green, cancelled=red)

2. **Time Information**
   - ✅ Start Time (`start_time`) - **Currently missing in response**
   - ✅ End Time (`end_time`) - **Currently missing in response**
   - ✅ Due Date (`due_date`)
   - ✅ Assigned At (`assigned_at`)
   - ✅ Created At (`created_at`)
   - ✅ Updated At (`updated_at`) - **Currently missing**

3. **Assignment Information**
   - ✅ Assigned User (name/username, not just ID)
   - ✅ Project Name (not just ID)
   - ✅ For collaborative tasks: Lead User name

4. **For Collaborative Tasks**
   - ✅ Priority (high/medium/low) - **Currently missing**
   - ✅ Progress (0-100%) - **Currently missing**
   - ✅ Complexity (simple/medium/complex) - **Currently missing**
   - ✅ Participants list - **Currently missing**

### Nice-to-Have Information

5. **Time Context** (Arabic Working Hours)
   - Working hours until deadline
   - Working days until deadline
   - Deadline risk level
   - Is overdue flag

6. **Project Context**
   - Project status
   - Project start/end dates
   - User's role in project

## Recommended Display Formats

### Weekly View

**What to Show:**
- Tasks grouped by day (Sunday-Saturday)
- Tasks with `start_time` or `due_date` in that week
- Visual timeline showing task duration
- Color-coded by status
- Priority indicators for collaborative tasks

**Example Structure:**
```json
{
  "week": {
    "start_date": "2024-01-15",
    "end_date": "2024-01-21",
    "days": [
      {
        "date": "2024-01-15",
        "day_name": "Monday",
        "tasks": [
          {
            "id": 1,
            "title": "Task 1",
            "start_time": "2024-01-15T09:00:00Z",
            "end_time": "2024-01-15T16:00:00Z",
            "status": "in_progress",
            "priority": "high",
            "assigned_to": "John Doe",
            "project": "Project X"
          }
        ]
      }
    ]
  }
}
```

### Monthly View

**What to Show:**
- Tasks grouped by week within the month
- Or tasks grouped by day (calendar view)
- Summary statistics per week
- Overdue tasks highlighted
- Upcoming deadlines

**Example Structure:**
```json
{
  "month": {
    "year": 2024,
    "month": 1,
    "weeks": [
      {
        "week_number": 1,
        "start_date": "2024-01-01",
        "end_date": "2024-01-07",
        "task_count": 15,
        "completed_count": 8,
        "tasks": [...]
      }
    ]
  }
}
```

## API Endpoints Needed

### 1. Get Project Tasks by Week
```
GET /api/tasks/project/:projectId/weekly?week=2024-W03
GET /api/tasks/project/:projectId/weekly?start_date=2024-01-15&end_date=2024-01-21
```

### 2. Get Project Tasks by Month
```
GET /api/tasks/project/:projectId/monthly?year=2024&month=1
GET /api/tasks/project/:projectId/monthly?start_date=2024-01-01&end_date=2024-01-31
```

### 3. Enhanced Task Response
All task endpoints should include:
- `start_time` and `end_time`
- User names (not just IDs)
- Project names (not just IDs)
- For collaborative tasks: priority, progress, complexity, participants

## Implementation Recommendations

### 1. Update Task Response Format

**Current:**
```go
taskList = append(taskList, gin.H{
    "id":          task.ID,
    "title":       task.Title,
    "description": task.Description,
    "status":      task.Status,
    "user_id":     task.UserID,
    "project_id":  task.ProjectID,
    "assigned_at": task.AssignedAt,
    "due_date":    task.DueDate,
    "created_at":  task.CreatedAt,
})
```

**Recommended:**
```go
taskList = append(taskList, gin.H{
    "id":          task.ID,
    "title":       task.Title,
    "description": task.Description,
    "status":      task.Status,
    "user": gin.H{
        "id":       task.User.ID,
        "username": task.User.Username,
    },
    "project": gin.H{
        "id":    task.Project.ID,
        "title": task.Project.Title,
    },
    "start_time":  task.StartTime,
    "end_time":    task.EndTime,
    "due_date":    task.DueDate,
    "assigned_at": task.AssignedAt,
    "created_at":  task.CreatedAt,
    "updated_at":  task.UpdatedAt,
})
```

### 2. Add Weekly/Monthly Grouping Functions

Create new service methods:
- `GetProjectTasksByWeek(projectID, weekStart, weekEnd)`
- `GetProjectTasksByMonth(projectID, year, month)`
- `GroupTasksByWeek(tasks []Task)`
- `GroupTasksByMonth(tasks []Task)`

### 3. Frontend Display Recommendations

**Weekly View:**
- Calendar grid (7 days)
- Drag-and-drop task scheduling
- Time slots (9 AM - 4 PM working hours)
- Color coding by status/priority
- Task duration bars

**Monthly View:**
- Calendar grid (full month)
- Task count per day
- Click to see day details
- Summary cards (total, completed, pending)
- Progress indicators

## Summary

### What We Have Now:
✅ Basic task information (id, title, description, status)
✅ Project reports with weekly/monthly statistics
✅ Task assignment information (user_id, project_id)
✅ Due dates and assignment dates

### What's Missing:
❌ `start_time` and `end_time` in responses
❌ User names and project names (only IDs)
❌ Weekly/monthly task grouping endpoints
❌ Calendar/timeline view data
❌ Collaborative task details (priority, progress, complexity, participants)
❌ Task duration visualization data

### Priority Actions:
1. **High Priority:** Add `start_time` and `end_time` to all task responses
2. **High Priority:** Include user names and project names in responses
3. **Medium Priority:** Create weekly/monthly grouping endpoints
4. **Medium Priority:** Add collaborative task details to responses
5. **Low Priority:** Add calendar view specific endpoints

