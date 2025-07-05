# HR Role Permissions

## Overview
The HR role has been implemented as a **special role** that is separate from the normal role hierarchy. HR has specific permissions focused on human resources management without the ability to manage tasks or projects.

## Role Hierarchy
```
Admin > Manager > Head > Employee
HR (Special Role - separate from hierarchy)
```

## HR Permissions

### ✅ What HR CAN Do:
1. **User Management (Read-Only mostly)**
   - View all users (`GET /api/users`)
   - View users by role (`GET /api/users/role/:role`)
   - View users by department (`GET /api/users/department/:department`)
   - Update user departments (`PATCH /api/users/:id/department`)

2. **Reports & Analytics**
   - Access task reports for HR analytics (`GET /api/tasks?period=weekly/monthly`)
   - View department-based task information
   - Access user performance reports

3. **Chat System**
   - Full access to team chat like other users
   - Can participate in real-time messaging

### ❌ What HR CANNOT Do:
1. **Task Management**
   - Cannot assign tasks to other users
   - Cannot create tasks for other users
   - Cannot update task details (title, description, etc.)
   - Cannot perform bulk task operations
   - Cannot update task status for others

2. **User Management (Administrative)**
   - Cannot create new users
   - Cannot change user roles
   - Cannot delete users

3. **Project Management**
   - Cannot create or manage projects
   - Cannot assign users to projects

## Implementation Details

### Middleware Functions
- `RequireHROrAdmin()` - Only HR or Admin can access
- `RequireHROrHigher()` - HR, Manager, or Admin can access
- `IsHRRole()` - Check if user has HR role
- `IsManagerOrHigher()` - Check Manager+ permissions (excluding HR)

### Permission Logic
```go
// HR has special permissions - only specific HR functions
if userRole == models.RoleHR {
    // HR can access user management, reports, but not task assignment
    return requiredRole == models.RoleHR || requiredRole == models.RoleEmployee
}
```

## Use Cases
HR role is perfect for:
- Human Resources personnel who need to view employee information
- HR analytics and reporting
- Department management
- Employee onboarding support
- Performance monitoring (read-only)

## Security Notes
- HR cannot escalate privileges or assign tasks
- HR has read-only access to most sensitive operations
- HR cannot modify core business processes (tasks/projects)
- HR permissions are explicitly defined and cannot be bypassed through role hierarchy

## API Endpoints Available to HR

### User Management
- `GET /api/users` - List all users
- `GET /api/users/role/:role` - Get users by role
- `GET /api/users/department/:department` - Get users by department
- `PATCH /api/users/:id/department` - Update user department

### Reports
- `GET /api/tasks?period=weekly` - Weekly task reports
- `GET /api/tasks?period=monthly` - Monthly task reports
- `GET /api/tasks?department=:dept` - Department task overview

### Chat
- `GET /api/chat/team-chat` - Access team chat
- `POST /api/chat/rooms/:roomId/messages` - Send messages
- `GET /api/chat/rooms/:roomId/messages` - View messages
- `GET /ws/chat` - WebSocket connection

This implementation ensures HR has the necessary permissions for human resources functions while maintaining security boundaries for task and project management. 