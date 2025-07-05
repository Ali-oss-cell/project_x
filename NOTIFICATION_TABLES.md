# Notification System Database Tables

## Overview
The notification system uses two main tables to manage notifications and user preferences. This document provides detailed information about the table structures, relationships, and usage.

---

## 📋 Table 1: `notifications`

### Table Structure

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `BIGINT UNSIGNED` | `PRIMARY KEY, AUTO_INCREMENT` | Unique notification identifier |
| `created_at` | `TIMESTAMP` | `NOT NULL` | When notification was created |
| `updated_at` | `TIMESTAMP` | `NOT NULL` | When notification was last updated |
| `deleted_at` | `TIMESTAMP NULL` | `NULL` | Soft delete timestamp |
| `user_id` | `BIGINT UNSIGNED` | `NOT NULL, INDEX` | ID of user receiving the notification |
| `type` | `VARCHAR(50)` | `NOT NULL, INDEX` | Type of notification (see types below) |
| `title` | `VARCHAR(255)` | `NOT NULL` | Notification title |
| `message` | `TEXT` | `NOT NULL` | Notification message content |
| `data` | `JSON` | `NULL` | Additional structured data |
| `is_read` | `BOOLEAN` | `DEFAULT FALSE, INDEX` | Whether notification has been read |
| `read_at` | `TIMESTAMP NULL` | `NULL` | When notification was read |
| `related_task_id` | `BIGINT UNSIGNED NULL` | `INDEX` | Related task ID (if applicable) |
| `related_project_id` | `BIGINT UNSIGNED NULL` | `INDEX` | Related project ID (if applicable) |
| `from_user_id` | `BIGINT UNSIGNED NULL` | `INDEX` | ID of user who triggered the notification |

### Notification Types

| Type | Value | Description | Usage |
|------|-------|-------------|-------|
| `task_assigned` | `"task_assigned"` | New task assigned to user | Task creation/assignment |
| `task_updated` | `"task_updated"` | Task details modified | Task updates (title, description, due date) |
| `task_completed` | `"task_completed"` | Task marked as completed | Status change to completed |
| `task_commented` | `"task_commented"` | Comment added to task | Task comments (future feature) |
| `task_due_soon` | `"task_due_soon"` | Task due date approaching | Due date reminders |
| `project_created` | `"project_created"` | New project created | Project creation |
| `user_joined` | `"user_joined"` | New user joined system | User registration |
| `file_uploaded` | `"file_uploaded"` | File uploaded to task | File uploads |

### Relationships

```sql
-- Foreign Key Relationships
notifications.user_id → users.id (CASCADE DELETE)
notifications.related_task_id → tasks.id (SET NULL ON DELETE)
notifications.related_project_id → projects.id (SET NULL ON DELETE)
notifications.from_user_id → users.id (SET NULL ON DELETE)
```

### Data JSON Structure

The `data` column contains structured JSON with the following format:

```json
{
  "task_id": 123,
  "project_id": 456,
  "from_user_id": 789,
  "task_title": "Complete project documentation",
  "project_name": "Project X",
  "from_user": "John Manager",
  "due_date": "2024-12-31T23:59:59Z",
  "file_url": "https://drive.google.com/file/...",
  "file_name": "document.pdf"
}
```

### Sample Records

| ID | User ID | Type | Title | Message | Is Read | Related Task ID | From User ID |
|----|---------|------|-------|---------|---------|-----------------|--------------|
| 1 | 5 | `task_assigned` | "New Task Assigned" | "You have been assigned a new task: 'Complete documentation' by John Manager" | false | 123 | 2 |
| 2 | 5 | `task_updated` | "Task Updated" | "Task 'Complete documentation' has been updated by John Manager (Changes: title, due date changed)" | false | 123 | 2 |
| 3 | 5 | `task_completed` | "Task Completed" | "Task 'Complete documentation' has been completed by Jane Employee" | false | 123 | 3 |
| 4 | 5 | `task_due_soon` | "Task Due Soon" | "Task 'Complete documentation' is due in 24 hours" | false | 123 | null |

---

## 📋 Table 2: `user_notification_preferences`

### Table Structure

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | `BIGINT UNSIGNED` | `PRIMARY KEY, AUTO_INCREMENT` | Unique preference identifier |
| `created_at` | `TIMESTAMP` | `NOT NULL` | When preference was created |
| `updated_at` | `TIMESTAMP` | `NOT NULL` | When preference was last updated |
| `deleted_at` | `TIMESTAMP NULL` | `NULL` | Soft delete timestamp |
| `user_id` | `BIGINT UNSIGNED` | `UNIQUE INDEX` | ID of user (one preference per user) |
| `task_assigned` | `BOOLEAN` | `DEFAULT TRUE` | Receive task assignment notifications |
| `task_updated` | `BOOLEAN` | `DEFAULT TRUE` | Receive task update notifications |
| `task_completed` | `BOOLEAN` | `DEFAULT TRUE` | Receive task completion notifications |
| `task_commented` | `BOOLEAN` | `DEFAULT TRUE` | Receive task comment notifications |
| `task_due_soon` | `BOOLEAN` | `DEFAULT TRUE` | Receive due date reminder notifications |
| `project_created` | `BOOLEAN` | `DEFAULT TRUE` | Receive project creation notifications |
| `user_joined` | `BOOLEAN` | `DEFAULT TRUE` | Receive user joined notifications |
| `file_uploaded` | `BOOLEAN` | `DEFAULT TRUE` | Receive file upload notifications |
| `email_notifications` | `BOOLEAN` | `DEFAULT FALSE` | Enable email notifications |
| `push_notifications` | `BOOLEAN` | `DEFAULT TRUE` | Enable push notifications |
| `in_app_notifications` | `BOOLEAN` | `DEFAULT TRUE` | Enable in-app notifications |

### Relationships

```sql
-- Foreign Key Relationships
user_notification_preferences.user_id → users.id (CASCADE DELETE)
```

### Sample Records

| ID | User ID | Task Assigned | Task Updated | Task Completed | Task Due Soon | Email | Push | In App |
|----|---------|---------------|--------------|----------------|---------------|-------|------|--------|
| 1 | 5 | true | true | true | true | false | true | true |
| 2 | 10 | true | false | true | false | true | false | true |
| 3 | 15 | false | false | false | false | false | false | false |

---

## 🔍 Database Queries Examples

### Get Unread Notifications for User
```sql
SELECT * FROM notifications 
WHERE user_id = ? AND is_read = false 
ORDER BY created_at DESC;
```

### Get Notification Count by Type
```sql
SELECT type, COUNT(*) as count 
FROM notifications 
WHERE user_id = ? AND is_read = false 
GROUP BY type;
```

### Get User's Notification Preferences
```sql
SELECT * FROM user_notification_preferences 
WHERE user_id = ?;
```

### Mark Notification as Read
```sql
UPDATE notifications 
SET is_read = true, read_at = NOW() 
WHERE id = ? AND user_id = ?;
```

### Get Notifications with Related Data
```sql
SELECT n.*, u.username as from_username, t.title as task_title 
FROM notifications n
LEFT JOIN users u ON n.from_user_id = u.id
LEFT JOIN tasks t ON n.related_task_id = t.id
WHERE n.user_id = ?
ORDER BY n.created_at DESC;
```

---

## 📊 Indexes and Performance

### Primary Indexes
- `notifications.id` (Primary Key)
- `user_notification_preferences.id` (Primary Key)

### Secondary Indexes
- `notifications.user_id` - For filtering by user
- `notifications.type` - For filtering by notification type
- `notifications.is_read` - For filtering read/unread notifications
- `notifications.related_task_id` - For task-related queries
- `notifications.related_project_id` - For project-related queries
- `notifications.from_user_id` - For sender-related queries
- `user_notification_preferences.user_id` - Unique index for user preferences

### Performance Considerations
1. **Soft Deletes**: Both tables use soft deletes (`deleted_at`) for data retention
2. **Indexing Strategy**: Indexes on frequently queried columns for optimal performance
3. **JSON Data**: The `data` column uses JSON type for flexible structured data storage
4. **Cascade Deletes**: User deletion cascades to their notifications and preferences

---

## 🚀 Migration Scripts

### Create Notifications Table
```sql
CREATE TABLE notifications (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSON NULL,
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP NULL,
    related_task_id BIGINT UNSIGNED NULL,
    related_project_id BIGINT UNSIGNED NULL,
    from_user_id BIGINT UNSIGNED NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_type (type),
    INDEX idx_is_read (is_read),
    INDEX idx_related_task_id (related_task_id),
    INDEX idx_related_project_id (related_project_id),
    INDEX idx_from_user_id (from_user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (related_task_id) REFERENCES tasks(id) ON DELETE SET NULL,
    FOREIGN KEY (related_project_id) REFERENCES projects(id) ON DELETE SET NULL,
    FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE SET NULL
);
```

### Create User Notification Preferences Table
```sql
CREATE TABLE user_notification_preferences (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    task_assigned BOOLEAN DEFAULT TRUE,
    task_updated BOOLEAN DEFAULT TRUE,
    task_completed BOOLEAN DEFAULT TRUE,
    task_commented BOOLEAN DEFAULT TRUE,
    task_due_soon BOOLEAN DEFAULT TRUE,
    project_created BOOLEAN DEFAULT TRUE,
    user_joined BOOLEAN DEFAULT TRUE,
    file_uploaded BOOLEAN DEFAULT TRUE,
    email_notifications BOOLEAN DEFAULT FALSE,
    push_notifications BOOLEAN DEFAULT TRUE,
    in_app_notifications BOOLEAN DEFAULT TRUE,
    UNIQUE INDEX idx_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

---

**Last Updated**: December 2024  
**Database Engine**: PostgreSQL (via GORM)  
**Status**: ✅ Production Ready 