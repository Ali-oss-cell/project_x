# Bulk Operations Guide

## Overview

The task management system now supports comprehensive bulk operations that allow **Managers and Admins** to update **all fields** of multiple tasks at once, not just the status.

## Available Endpoints

### 1. Smart Bulk Operations (Recommended)
**Endpoint:** `POST /api/tasks/bulk?action=bulk_update`

This is the main bulk operations endpoint that supports updating all task fields.

### 2. Legacy Status-Only Bulk Update
**Endpoint:** `POST /api/tasks/legacy/bulk-update`

This endpoint only updates the status of multiple tasks (for backward compatibility).

## Bulk Update All Fields

### Request Format

```json
{
  "task_ids": [1, 2, 3, 4, 5],
  "task_type": "regular",  // or "collaborative"
  "title": "Updated Title",           // Optional
  "description": "Updated Description", // Optional
  "status": "in_progress",            // Optional: "pending", "in_progress", "completed", "cancelled"
  "due_date": "2024-12-31T23:59:59Z", // Optional
  // StartTime is immutable and cannot be updated
  "end_time": "2024-12-01T17:00:00Z",   // Optional
  "priority": "high",                  // Optional: "high", "medium", "low" (collaborative tasks only)
  "complexity": "medium",              // Optional: "simple", "medium", "complex" (collaborative tasks only)
  "assigned_to": 123,                  // Optional: reassign regular tasks to user ID
  "lead_user_id": 456,                 // Optional: change lead for collaborative tasks
  "project_id": 789                    // Optional: change project
}
```

### Examples

#### 1. Update Status Only (Fast)
```json
{
  "task_ids": [1, 2, 3],
  "task_type": "regular",
  "status": "completed"
}
```

#### 2. Update Multiple Fields
```json
{
  "task_ids": [1, 2, 3],
  "task_type": "regular",
  "title": "New Title for All Tasks",
  "description": "Updated description",
  "due_date": "2024-12-31T23:59:59Z",
  "status": "in_progress"
}
```

#### 3. Reassign Tasks
```json
{
  "task_ids": [1, 2, 3],
  "task_type": "regular",
  "assigned_to": 123,
  "status": "pending"
}
```

#### 4. Update Collaborative Tasks
```json
{
  "task_ids": [4, 5, 6],
  "task_type": "collaborative",
  "title": "Updated Collaborative Task",
  "priority": "high",
  "complexity": "complex",
  "lead_user_id": 456
}
```

#### 5. Change Project for Multiple Tasks
```json
{
  "task_ids": [1, 2, 3, 4, 5],
  "task_type": "regular",
  "project_id": 789
}
```

## Response Format

### Success Response
```json
{
  "action": "bulk_updated",
  "type": "regular",
  "updated_count": 3,
  "total": 3,
  "updated_fields": {
    "title": true,
    "description": false,
    "status": true,
    "due_date": false,
    // start_time is immutable
    "end_time": false,
    "priority": false,
    "complexity": false,
    "assigned_to": false,
    "lead_user_id": false,
    "project_id": false
  }
}
```

### Response with Failures
```json
{
  "action": "bulk_updated",
  "type": "regular",
  "updated_count": 2,
  "total": 3,
  "failed_tasks": [3],
  "permission_denied_tasks": [4],
  "updated_fields": {
    "title": true,
    "status": true
  }
}
```

### Response with Permission Issues
```json
{
  "action": "bulk_updated",
  "type": "regular",
  "updated_count": 1,
  "total": 3,
  "permission_denied_tasks": [2, 3],
  "updated_fields": {
    "title": true,
    "status": true
  }
}
```

## Permissions

### Role-Based Access Control

The bulk operations respect the same role-based permissions as individual task updates:

#### **Regular Tasks:**
- **Task Assignee**: Can only update status of their assigned tasks
- **Admin/Manager**: Can update title, description, due date, end time, and reassign tasks to other users
- **Employee/Head**: Can only update status of tasks assigned to them

#### **Collaborative Tasks:**
- **Lead User**: Can only update status of collaborative tasks they lead
- **Admin/Manager**: Can update title, description, due date, end time, priority, complexity, and change lead users
- **Employee/Head**: Can only update status of collaborative tasks they lead

#### **Bulk Operations:**
- **Manager+ only**: Only users with Manager or Admin roles can perform bulk operations
- **Employees and Heads**: Cannot perform bulk operations (even on their own tasks)

## Notifications

When tasks are reassigned via bulk operations, the system automatically sends notifications to:
- The newly assigned user(s)
- The previous assignee(s) (if applicable)

## Performance Notes

- **Status-only updates**: Use the optimized bulk status update for better performance
- **Multi-field updates**: Each task is updated individually to ensure data integrity
- **Large batches**: Consider breaking very large batches into smaller chunks

## Error Handling

- **Invalid task IDs**: Are ignored and reported in `failed_tasks`
- **Invalid field values**: Return a 400 error with specific validation messages
- **Permission violations**: Tasks are skipped and reported in `permission_denied_tasks`
- **Role restrictions**: Employees and Heads cannot perform bulk operations (403 error)
- **Individual task permissions**: Each task is checked for user permissions before updating

## Migration from Legacy

If you're currently using the legacy bulk update endpoint, you can migrate by:

1. **Status-only updates**: Continue using the legacy endpoint
2. **Multi-field updates**: Switch to the new smart bulk endpoint
3. **Gradual migration**: Both endpoints are supported simultaneously

## Best Practices

1. **Use specific field updates**: Only include fields you want to change
2. **Batch size**: Keep batches under 100 tasks for optimal performance
3. **Validation**: Always validate task IDs before bulk operations
4. **Notifications**: Be aware that reassignments trigger notifications
5. **Testing**: Test bulk operations on a small batch first 