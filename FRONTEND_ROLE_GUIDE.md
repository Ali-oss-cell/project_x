# Complete Frontend Role-Based UI Guide

**Purpose:** This guide explains exactly what UI elements, pages, buttons, and features each user role should see.

---

## 📋 Table of Contents

1. [Role Overview](#role-overview)
2. [Admin Role](#1-admin-role---full-system-control)
3. [Manager Role](#2-manager-role---team--project-management)
4. [Head Role](#3-head-role---team-leader)
5. [Employee Role](#4-employee-role---individual-contributor)
6. [HR Role](#5-hr-role---human-resources)
7. [Quick Reference Tables](#quick-reference-tables)
8. [Page-by-Page Breakdown](#page-by-page-breakdown)

---

## Role Overview

### Role Hierarchy
```
ADMIN (Level 5) ─────┐
    │                │ Full Access
MANAGER (Level 4) ───┤
    │                │ Management
HEAD (Level 3) ──────┤
    │                │ Team Work
EMPLOYEE (Level 2) ──┘

HR (Special) ────────→ Separate (HR-only features)
```

### Core Principle
**Higher roles can do everything lower roles can do, PLUS additional features**

Exception: HR is separate and has unique permissions

---

## 1. ADMIN Role - Full System Control

### 🎯 What Admin Sees

#### Navigation Menu
```
✅ Dashboard
✅ My Tasks
✅ Projects
✅ Team Management
✅ Users (Admin Panel)
✅ System Settings
✅ Analytics & Reports
✅ HR Dashboard
✅ Chat
✅ Notifications
```

#### Dashboard Page
**Visible Sections:**
- System overview (total users, projects, tasks)
- Recent activities (all system activities)
- User management quick access
- Project management quick access
- System health indicators
- All departments overview

**Action Buttons:**
```
✅ Create User
✅ Create Project
✅ Create Task
✅ Assign Task to Anyone
✅ View All Reports
✅ System Settings
✅ Export Data
✅ Manage Roles
```

#### Users Page
**What Admin Sees:**
- Complete user list (all roles, all departments)
- User search and filters
- Role filter dropdown: admin, manager, head, employee, hr
- Department filter

**User Card/Row Shows:**
```json
{
  "id": 1,
  "username": "john_doe",
  "role": "employee",
  "department": "Engineering",
  "skills": ["React", "TypeScript"],
  "created_at": "2024-01-15T10:00:00Z",
  "task_count": 8,
  "project_count": 3
}
```

**Action Buttons on Each User:**
```
✅ Edit User
✅ Change Role (dropdown: admin/manager/head/employee/hr)
✅ Change Department
✅ Update Skills
✅ Assign Task to This User
✅ View User Tasks
✅ View User Report
✅ Delete User
✅ Reset Password
```

#### Projects Page
**What Admin Sees:**
- All projects (not just their own)
- Project status: active, paused, completed, cancelled

**Project Card Shows:**
```json
{
  "id": 1,
  "title": "Project X",
  "status": "active",
  "creator": "manager",
  "member_count": 8,
  "task_count": 25,
  "completion_rate": 60.5,
  "start_date": "2024-01-01",
  "end_date": "2024-06-30"
}
```

**Action Buttons on Each Project:**
```
✅ View Details
✅ Edit Project
✅ Add Members
✅ Remove Members
✅ Create Tasks
✅ Generate AI Tasks
✅ View Statistics
✅ View Timeline
✅ Change Status
✅ Delete Project
✅ Export Project Data
```

#### Tasks Page
**What Admin Sees:**
- All tasks (all users, all projects)
- Filter options:
  - By user (dropdown of all users)
  - By project (dropdown of all projects)
  - By status (pending/in_progress/completed/cancelled)
  - By type (regular/collaborative)
  - By department

**Task Card Shows:**
```json
{
  "id": 1,
  "title": "Implement authentication",
  "description": "Add JWT authentication",
  "status": "in_progress",
  "assigned_to": "employee",
  "project": "Project X",
  "start_time": "2024-01-20T09:00:00Z",
  "end_time": "2024-01-25T17:00:00Z",
  "due_date": "2024-01-25T17:00:00Z",
  "created_at": "2024-01-20T09:00:00Z",
  "updated_at": "2024-01-22T14:30:00Z"
}
```

**Action Buttons on Each Task:**
```
✅ View Details
✅ Edit Task (title, description, due date)
✅ Reassign to Another User
✅ Change Project
✅ Change Status
✅ Delete Task
✅ View Time Analysis
✅ Add to Bulk Selection
```

**Bulk Actions Section:**
```
✅ Select All
✅ Select by Status
✅ Bulk Update Status
✅ Bulk Reassign
✅ Bulk Change Project
✅ Bulk Delete
```

#### Analytics & Reports Page
**Visible Sections:**
```
✅ System Statistics
✅ User Performance Reports (all users)
✅ Project Performance Reports (all projects)
✅ Department Analytics
✅ Weekly/Monthly Reports (all)
✅ Task Completion Trends
✅ AI Performance Metrics
✅ Time Analysis Dashboard
```

**Report Filters:**
```
✅ Period: Weekly / Monthly
✅ User: All / Select specific
✅ Project: All / Select specific
✅ Department: All / Select specific
✅ Date Range: Custom range picker
```

#### Settings Page (Admin Only)
**Visible Settings:**
```
✅ System Configuration
✅ API Keys Management
✅ Database Settings
✅ Email/SMS Configuration
✅ Working Hours Configuration
✅ Arabic Language Settings
✅ Notification Settings (Global)
✅ Security Settings
✅ Backup & Restore
```

---

## 2. MANAGER Role - Team & Project Management

### 🎯 What Manager Sees

#### Navigation Menu
```
✅ Dashboard
✅ My Tasks
✅ Projects
✅ Team Management
✅ Analytics & Reports
✅ Chat
✅ Notifications
❌ Users (Admin Panel) - HIDDEN
❌ System Settings - HIDDEN
```

#### Dashboard Page
**Visible Sections:**
- Personal overview (my projects, my teams)
- Team performance
- Project statistics (projects I manage)
- Recent activities (my teams and projects)

**Action Buttons:**
```
✅ Create Project
✅ Create Task
✅ Assign Task to Team
✅ View Team Reports
✅ Generate AI Tasks
❌ Create User - HIDDEN
❌ System Settings - HIDDEN
```

#### Users Page
**What Manager Sees:**
- Users in their department/teams ONLY
- Read-only view (can't create/delete users)

**User Card Shows:**
```json
{
  "id": 4,
  "username": "employee",
  "role": "employee",
  "department": "Engineering",
  "skills": ["React", "TypeScript"]
}
```

**Action Buttons on Each User:**
```
✅ View User Tasks
✅ View User Report
✅ Assign Task to This User
✅ Add to Project
❌ Edit User - HIDDEN
❌ Change Role - HIDDEN
❌ Delete User - HIDDEN
```

#### Projects Page
**What Manager Sees:**
- Projects they created
- Projects they're a member of
- Can create new projects

**Project Card Shows:**
```json
{
  "id": 1,
  "title": "Project X",
  "status": "active",
  "member_count": 5,
  "task_count": 12,
  "my_role": "manager"
}
```

**Action Buttons on Each Project:**
```
✅ View Details
✅ Edit Project (if creator)
✅ Add Members
✅ Remove Members
✅ Create Tasks
✅ Generate AI Tasks
✅ View Statistics
✅ Change Status
❌ Delete Project - HIDDEN (Admin only)
```

#### Tasks Page
**What Manager Sees:**
- Tasks in their projects
- Tasks in their department
- Filter by project/status/user

**Task Card Shows:**
```json
{
  "id": 1,
  "title": "Implement authentication",
  "status": "in_progress",
  "assigned_to": "employee",
  "project": "Project X"
}
```

**Action Buttons on Each Task:**
```
✅ View Details
✅ Edit Task
✅ Reassign to Another User
✅ Change Status
✅ Delete Task
✅ View Time Analysis
✅ Bulk Operations
❌ Delete System Tasks - LIMITED (only their project tasks)
```

#### Analytics & Reports Page
**Visible Sections:**
```
✅ My Projects Statistics
✅ Team Performance Reports
✅ Department Analytics (if department head)
✅ Weekly/Monthly Reports (my projects)
✅ AI Performance Metrics (my projects)
❌ System-Wide Reports - HIDDEN
❌ All Users Reports - LIMITED (only team members)
```

---

## 3. HEAD Role - Team Leader

### 🎯 What Head Sees

**NEW UPDATE:** Heads can now **edit and delete** any task in projects they're members of!

#### Navigation Menu
```
✅ Dashboard
✅ My Tasks
✅ Team Tasks
✅ Projects (where I'm a member)
✅ Chat
✅ Notifications
❌ Projects (Create) - HIDDEN
❌ Team Management - LIMITED
❌ Analytics & Reports - HIDDEN
❌ Users - HIDDEN
```

#### Dashboard Page
**Visible Sections:**
- My tasks overview
- Team tasks I assigned
- Projects I'm part of
- Upcoming deadlines

**Action Buttons:**
```
✅ Create Task (for myself)
✅ Create Task (assign to employees in my team)
✅ View My Tasks
✅ View Team Tasks
❌ Create Project - HIDDEN
❌ View Reports - HIDDEN
❌ System Settings - HIDDEN
```

#### My Tasks Page
**What Head Sees:**
- Tasks assigned to them
- Tasks they created for themselves

**Task Card Shows:**
```json
{
  "id": 1,
  "title": "Review code pull requests",
  "status": "in_progress",
  "project": "Project X",
  "due_date": "2024-01-25T17:00:00Z"
}
```

**Action Buttons on Each Task:**
```
✅ View Details
✅ Update Status (my tasks only)
✅ Add Comment
✅ Edit Task Details (if task is in my project)
✅ Delete Task (if task is in my project)
❌ Reassign - HIDDEN (Manager+ only)
```

**Important:** Heads can now edit/delete ANY task in projects they're members of, not just their own tasks!

#### Team Tasks Page (Heads Only)
**What Head Sees:**
- Tasks they assigned to employees
- Collaborative tasks they lead

**Task Card Shows:**
```json
{
  "id": 5,
  "title": "Write unit tests",
  "status": "pending",
  "assigned_to": "employee",
  "assigned_by": "me",
  "project": "Project X"
}
```

**Action Buttons:**
```
✅ View Details
✅ View Progress
✅ Add Comment
✅ Edit Task Details (if task is in my project)
✅ Delete Task (if task is in my project)
❌ Reassign - HIDDEN (Manager+ only)
```

#### Projects Page
**What Head Sees:**
- Only projects they're a member of
- Can't create projects

**Project Card Shows:**
```json
{
  "id": 1,
  "title": "Project X",
  "my_role": "head",
  "task_count": 8,
  "my_tasks": 3
}
```

**Action Buttons on Each Project:**
```
✅ View Details
✅ View Tasks
✅ View Members
❌ Edit Project - HIDDEN
❌ Add/Remove Members - HIDDEN
❌ Delete Project - HIDDEN
❌ Generate AI Tasks - HIDDEN
```

#### Create/Edit Task Form
**What Head Can Set:**
```
✅ Title
✅ Description
✅ Assign to: [Dropdown of employees only] (create only)
✅ Project: [Projects I'm in]
✅ Due Date
✅ Start/End Time
✅ Edit ANY task field (if task is in my project)
❌ Assign to Manager/Admin - DISABLED
❌ Reassign existing task - HIDDEN (Manager+ only)
❌ Priority - HIDDEN (Manager+ only)
```

---

## 4. EMPLOYEE Role - Individual Contributor

### 🎯 What Employee Sees

#### Navigation Menu
```
✅ Dashboard
✅ My Tasks
✅ My Projects
✅ Chat
✅ Notifications
❌ Team Tasks - HIDDEN
❌ Team Management - HIDDEN
❌ Analytics - HIDDEN
❌ Users - HIDDEN
❌ Reports - HIDDEN
```

#### Dashboard Page
**Visible Sections:**
- My tasks (pending, in_progress, completed)
- My projects
- Upcoming deadlines
- Recent notifications

**Action Buttons:**
```
✅ Create Task (for myself only)
✅ View My Tasks
✅ View My Projects
❌ Assign Task to Others - HIDDEN
❌ Create Project - HIDDEN
❌ View Team - HIDDEN
❌ View Reports - HIDDEN
```

#### My Tasks Page
**What Employee Sees:**
- Only tasks assigned to them
- No other users' tasks visible

**Task Card Shows:**
```json
{
  "id": 1,
  "title": "Implement login page",
  "status": "in_progress",
  "project": "Project X",
  "due_date": "2024-01-25T17:00:00Z"
}
```

**Action Buttons on Each Task:**
```
✅ Update Status (dropdown: pending/in_progress/completed)
✅ View Details
✅ Add Comment
❌ Edit Title/Description - HIDDEN
❌ Delete - HIDDEN
❌ Reassign - HIDDEN
❌ Change Project - HIDDEN
```

#### Projects Page
**What Employee Sees:**
- Only projects they're a member of
- Read-only view

**Project Card Shows:**
```json
{
  "id": 1,
  "title": "Project X",
  "my_tasks": 5,
  "my_completed": 2
}
```

**Action Buttons on Each Project:**
```
✅ View Details
✅ View My Tasks in This Project
✅ View Project Timeline
❌ Edit Project - HIDDEN
❌ Add Members - HIDDEN
❌ Create Tasks for Others - HIDDEN
❌ View All Project Tasks - HIDDEN (only see own tasks)
```

#### Create Task Form
**What Employee Can Set:**
```
✅ Title
✅ Description
✅ Project: [Projects I'm in]
✅ Due Date
❌ Assign to: HIDDEN (automatically assigned to self)
❌ Start/End Time - OPTIONAL
```

**Important Note:**
- Employee can only create tasks for themselves
- "Assigned to" field doesn't appear
- Task is automatically assigned to them

---

## 5. HR Role - Human Resources

### 🎯 What HR Sees

#### Navigation Menu
```
✅ Dashboard (HR Dashboard)
✅ HR Problems
✅ Users (View Only)
✅ Reports (Read Only)
✅ Chat
✅ Notifications
❌ My Tasks - HIDDEN
❌ Projects - HIDDEN
❌ Team Management - HIDDEN
```

#### HR Dashboard Page
**Visible Sections:**
- HR problems overview
- Department statistics
- User analytics
- Recent HR activities

**Action Buttons:**
```
✅ View All Users
✅ View HR Problems
✅ View Department Reports
✅ View User Analytics
❌ Create Project - HIDDEN
❌ Assign Tasks - HIDDEN
❌ Create Users - HIDDEN
```

#### Users Page (HR View)
**What HR Sees:**
- All users (all roles, all departments)
- User statistics
- Department distribution

**User Card Shows:**
```json
{
  "id": 4,
  "username": "employee",
  "role": "employee",
  "department": "Engineering",
  "task_count": 8,
  "completed_count": 5
}
```

**Action Buttons on Each User:**
```
✅ View Details
✅ View User Report
✅ Update Department
❌ Edit User - HIDDEN
❌ Change Role - HIDDEN
❌ Delete User - HIDDEN
❌ Assign Task - HIDDEN
```

#### HR Problems Page
**What HR Sees:**
- All HR problems
- Filter by status/priority/category

**Problem Card Shows:**
```json
{
  "id": 1,
  "title": "Harassment complaint",
  "reporter": "employee",
  "status": "open",
  "priority": "high",
  "category": "harassment",
  "assigned_to": "me",
  "created_at": "2024-01-20T10:00:00Z"
}
```

**Action Buttons on Each Problem:**
```
✅ View Details
✅ Add Comment
✅ Update Status
✅ Change Priority
✅ Assign to Another HR
✅ Mark as Resolved
```

#### Reports Page (HR View)
**What HR Sees:**
```
✅ User task statistics (read-only)
✅ Department performance (read-only)
✅ Weekly/Monthly summaries (read-only)
❌ Cannot modify anything
❌ Cannot assign tasks based on reports
```

---

## Quick Reference Tables

### Feature Comparison by Role

| Feature | Admin | Manager | Head | Employee | HR |
|---------|:-----:|:-------:|:----:|:--------:|:--:|
| Create User | ✅ | ❌ | ❌ | ❌ | ❌ |
| Create Project | ✅ | ✅ | ❌ | ❌ | ❌ |
| Assign Task to Anyone | ✅ | ✅ Team | ⚠️ Employees | ❌ | ❌ |
| Edit Any Task | ✅ | ⚠️ Own | ✅ In Projects | ❌ | ❌ |
| Delete Task | ✅ | ⚠️ Own | ✅ In Projects | ❌ | ❌ |
| View Analytics | ✅ Full | ✅ Team | ❌ | ❌ | ✅ View |
| Bulk Operations | ✅ | ✅ | ❌ | ❌ | ❌ |
| AI Task Generation | ✅ | ✅ | ❌ | ❌ | ❌ |
| System Settings | ✅ | ❌ | ❌ | ❌ | ❌ |
| HR Problems | ✅ | ❌ | ❌ | ✅ Own | ✅ All |

**Notes:**
- **Edit Any Task / Delete Task for Head:** Can edit/delete ANY task in projects they're members of (NEW!)
- **Edit Any Task / Delete Task for Manager:** Can edit/delete tasks in their own projects
- **Assign Task to Anyone for Head:** Can only assign to employees (not managers/admins)

### Page Access by Role

| Page | Admin | Manager | Head | Employee | HR |
|------|:-----:|:-------:|:----:|:--------:|:--:|
| Dashboard | ✅ Full | ✅ Team | ✅ Personal | ✅ Personal | ✅ HR |
| My Tasks | ✅ | ✅ | ✅ | ✅ | ❌ |
| Team Tasks | ✅ | ✅ | ✅ | ❌ | ❌ |
| All Tasks | ✅ | ✅* | ❌ | ❌ | ❌ |
| Projects List | ✅ All | ✅ All | ✅ Mine | ✅ Mine | ❌ |
| Create Project | ✅ | ✅ | ❌ | ❌ | ❌ |
| Users List | ✅ Full | ✅ Team | ❌ | ❌ | ✅ View |
| Create User | ✅ | ❌ | ❌ | ❌ | ❌ |
| Analytics | ✅ Full | ✅ Team | ❌ | ❌ | ✅ View |
| Reports | ✅ Full | ✅ Team | ❌ | ❌ | ✅ View |
| HR Problems | ✅ All | ❌ | ❌ | ✅ Mine | ✅ All |
| Settings | ✅ | ❌ | ❌ | ❌ | ❌ |
| Chat | ✅ | ✅ | ✅ | ✅ | ✅ |

*Manager can see tasks in their projects and department

### Button Visibility by Role

| Button | Admin | Manager | Head | Employee | HR |
|--------|:-----:|:-------:|:----:|:--------:|:--:|
| **User Management** |
| Create User | ✅ | ❌ | ❌ | ❌ | ❌ |
| Edit User | ✅ | ❌ | ❌ | ❌ | ❌ |
| Delete User | ✅ | ❌ | ❌ | ❌ | ❌ |
| Change Role | ✅ | ❌ | ❌ | ❌ | ❌ |
| Update Dept | ✅ | ✅ | ❌ | ❌ | ✅ |
| **Project Management** |
| Create Project | ✅ | ✅ | ❌ | ❌ | ❌ |
| Edit Project | ✅ | ✅ | ❌ | ❌ | ❌ |
| Delete Project | ✅ | ❌ | ❌ | ❌ | ❌ |
| Add Members | ✅ | ✅ | ❌ | ❌ | ❌ |
| Remove Members | ✅ | ✅ | ❌ | ❌ | ❌ |
| Generate AI Tasks | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Task Management** |
| Create Task | ✅ | ✅ | ✅ | ✅* | ❌ |
| Assign to Others | ✅ | ✅ | ✅** | ❌ | ❌ |
| Edit Task Details | ✅ | ✅ | ✅*** | ❌ | ❌ |
| Update Status | ✅ | ✅ | ✅**** | ✅**** | ❌ |
| Delete Task | ✅ | ✅ | ✅*** | ❌ | ❌ |
| Bulk Operations | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Analytics** |
| View Statistics | ✅ | ✅ | ❌ | ❌ | ✅ |
| Export Reports | ✅ | ✅ | ❌ | ❌ | ❌ |

*Employee can only create for self  
**Head can only assign to employees  
***Head can edit/delete ANY task in their projects (NEW!)  
****Only for own tasks

---

## Page-by-Page Breakdown

### Dashboard Page Components

#### Admin Dashboard
```typescript
interface AdminDashboard {
  widgets: [
    'SystemOverview',      // Total users, projects, tasks
    'RecentActivities',    // All system activities
    'UserManagement',      // Quick user actions
    'ProjectManagement',   // Quick project actions
    'SystemHealth',        // Server status, DB status
    'DepartmentOverview'   // All departments
  ];
  quickActions: [
    'CreateUser',
    'CreateProject',
    'ViewAllReports',
    'SystemSettings'
  ];
}
```

#### Manager Dashboard
```typescript
interface ManagerDashboard {
  widgets: [
    'MyProjectsOverview',   // Projects I manage
    'TeamPerformance',      // My team stats
    'TasksSummary',         // Tasks in my projects
    'RecentActivities'      // My team activities
  ];
  quickActions: [
    'CreateProject',
    'AssignTask',
    'ViewTeamReports',
    'GenerateAITasks'
  ];
}
```

#### Head Dashboard
```typescript
interface HeadDashboard {
  widgets: [
    'MyTasks',              // My assigned tasks
    'TeamTasks',            // Tasks I assigned
    'MyProjects',           // Projects I'm in
    'UpcomingDeadlines'     // My deadlines
  ];
  quickActions: [
    'CreateTask',
    'AssignTaskToTeam',
    'ViewMyTasks'
  ];
}
```

#### Employee Dashboard
```typescript
interface EmployeeDashboard {
  widgets: [
    'MyTasks',              // My tasks only
    'TasksBreakdown',       // By status
    'MyProjects',           // Projects I'm in
    'UpcomingDeadlines'     // My deadlines
  ];
  quickActions: [
    'CreateTask',           // For myself
    'ViewMyTasks'
  ];
}
```

#### HR Dashboard
```typescript
interface HRDashboard {
  widgets: [
    'HRProblemsOverview',   // All HR problems
    'DepartmentStats',      // Department analytics
    'UserAnalytics',        // User statistics
    'RecentHRActivities'    // HR activities
  ];
  quickActions: [
    'ViewHRProblems',
    'ViewUsers',
    'ViewReports'
  ];
}
```

---

### Task List Page Components

#### Filter Bar (Role-based visibility)

**Admin Sees:**
```typescript
interface AdminTaskFilters {
  filterBy: {
    user: 'All Users Dropdown',       // ✅
    project: 'All Projects Dropdown',  // ✅
    department: 'All Departments',     // ✅
    status: 'All Statuses',           // ✅
    type: 'Regular/Collaborative',    // ✅
    dateRange: 'Custom Date Picker'   // ✅
  };
  sortBy: ['due_date', 'status', 'created_at', 'priority'];
}
```

**Manager Sees:**
```typescript
interface ManagerTaskFilters {
  filterBy: {
    user: 'My Team Users Dropdown',    // ✅ LIMITED
    project: 'My Projects Dropdown',   // ✅ LIMITED
    department: 'My Department',       // ✅ LIMITED
    status: 'All Statuses',           // ✅
    type: 'Regular/Collaborative'     // ✅
  };
  sortBy: ['due_date', 'status', 'created_at'];
}
```

**Head Sees:**
```typescript
interface HeadTaskFilters {
  filterBy: {
    status: 'All Statuses',           // ✅
    project: 'My Projects',           // ✅ LIMITED
    type: 'My Tasks/Team Tasks'       // ✅
  };
  sortBy: ['due_date', 'status'];
}
```

**Employee Sees:**
```typescript
interface EmployeeTaskFilters {
  filterBy: {
    status: 'All Statuses',           // ✅
    project: 'My Projects'            // ✅ LIMITED
  };
  sortBy: ['due_date', 'status'];
}
```

**HR Sees:**
```typescript
// HR doesn't have a tasks page - they see reports only
```

---

### Task Card Components (What's Visible)

#### Task Card for Admin
```tsx
<TaskCard>
  <Header>
    <Title>{task.title}</Title>
    <StatusBadge status={task.status} />
    <PriorityBadge priority={task.priority} />  {/* ✅ Shown */}
  </Header>
  
  <Body>
    <Description>{task.description}</Description>
    <AssignedTo user={task.user} />              {/* ✅ Show username */}
    <Project project={task.project} />            {/* ✅ Show project name */}
    <Timeline 
      start={task.start_time} 
      end={task.end_time} 
      due={task.due_date} 
    />                                            {/* ✅ Full timeline */}
  </Body>
  
  <Actions>
    <Button onClick={edit}>Edit</Button>          {/* ✅ */}
    <Button onClick={reassign}>Reassign</Button>  {/* ✅ */}
    <Button onClick={delete}>Delete</Button>      {/* ✅ */}
    <StatusDropdown onChange={updateStatus} />    {/* ✅ */}
  </Actions>
</TaskCard>
```

#### Task Card for Manager
```tsx
<TaskCard>
  <Header>
    <Title>{task.title}</Title>
    <StatusBadge status={task.status} />
    <PriorityBadge priority={task.priority} />  {/* ✅ Shown */}
  </Header>
  
  <Body>
    <Description>{task.description}</Description>
    <AssignedTo user={task.user} />
    <Project project={task.project} />
    <Timeline start={task.start_time} end={task.end_time} due={task.due_date} />
  </Body>
  
  <Actions>
    <Button onClick={edit}>Edit</Button>          {/* ✅ If in their project */}
    <Button onClick={reassign}>Reassign</Button>  {/* ✅ If in their project */}
    <Button onClick={delete}>Delete</Button>      {/* ✅ If in their project */}
    <StatusDropdown onChange={updateStatus} />    {/* ⚠️ Only if task owner */}
  </Actions>
</TaskCard>
```

#### Task Card for Head
```tsx
<TaskCard>
  <Header>
    <Title>{task.title}</Title>
    <StatusBadge status={task.status} />
    {/* ❌ Priority badge hidden */}
  </Header>
  
  <Body>
    <Description>{task.description}</Description>
    <Project project={task.project} />
    <DueDate date={task.due_date} />
  </Body>
  
  <Actions>
    {isMyTask && (
      <StatusDropdown onChange={updateStatus} />  {/* ✅ Only own tasks */}
    )}
    {isInMyProject && (
      <>
        <Button onClick={edit}>Edit</Button>        {/* ✅ NEW: Can edit any task in my projects */}
        <Button onClick={delete}>Delete</Button>    {/* ✅ NEW: Can delete any task in my projects */}
      </>
    )}
    {/* ❌ No Reassign button */}
  </Actions>
</TaskCard>
```

#### Task Card for Employee
```tsx
<TaskCard>
  <Header>
    <Title>{task.title}</Title>
    <StatusBadge status={task.status} />
    {/* ❌ Priority badge hidden */}
  </Header>
  
  <Body>
    <Description>{task.description}</Description>
    <Project project={task.project} />
    <DueDate date={task.due_date} />
  </Body>
  
  <Actions>
    <StatusDropdown onChange={updateStatus} />    {/* ✅ Only status */}
    {/* ❌ No Edit button */}
    {/* ❌ No Delete button */}
    {/* ❌ No Reassign button */}
  </Actions>
</TaskCard>
```

---

### Project Detail Page Components

#### Admin/Manager View
```tsx
<ProjectDetailPage>
  <Header>
    <Title>{project.title}</Title>
    <StatusBadge status={project.status} />
    <Actions>
      <Button onClick={edit}>Edit Project</Button>           {/* ✅ */}
      <Button onClick={addMembers}>Add Members</Button>      {/* ✅ */}
      <Button onClick={generateTasks}>Generate AI Tasks</Button> {/* ✅ */}
      {isAdmin && <Button onClick={delete}>Delete</Button>}  {/* ✅ Admin only */}
    </Actions>
  </Header>
  
  <Tabs>
    <Tab name="Overview">
      <ProjectInfo />
      <MembersList />
      <Statistics />
    </Tab>
    
    <Tab name="Tasks">
      <TasksList showAll={true} />                          {/* ✅ All project tasks */}
      <CreateTaskButton />                                   {/* ✅ */}
    </Tab>
    
    <Tab name="Timeline">
      <ProjectTimeline />                                    {/* ✅ */}
    </Tab>
    
    <Tab name="Analytics">
      <ProjectAnalytics />                                   {/* ✅ */}
      <UserPerformance />                                    {/* ✅ */}
    </Tab>
    
    <Tab name="Settings">
      <ProjectSettings />                                    {/* ✅ */}
    </Tab>
  </Tabs>
</ProjectDetailPage>
```

#### Head View
```tsx
<ProjectDetailPage>
  <Header>
    <Title>{project.title}</Title>
    <StatusBadge status={project.status} />
    {/* ❌ No action buttons */}
  </Header>
  
  <Tabs>
    <Tab name="Overview">
      <ProjectInfo />
      <MembersList />
    </Tab>
    
    <Tab name="My Tasks">
      <TasksList showOnlyMine={true} />                    {/* ✅ Only my tasks */}
    </Tab>
    
    {/* ❌ No Timeline tab */}
    {/* ❌ No Analytics tab */}
    {/* ❌ No Settings tab */}
  </Tabs>
</ProjectDetailPage>
```

#### Employee View
```tsx
<ProjectDetailPage>
  <Header>
    <Title>{project.title}</Title>
    <StatusBadge status={project.status} />
    {/* ❌ No action buttons */}
  </Header>
  
  <Tabs>
    <Tab name="Overview">
      <ProjectInfo />                                        {/* ✅ Basic info */}
      <MembersList />                                        {/* ✅ See team */}
    </Tab>
    
    <Tab name="My Tasks">
      <TasksList showOnlyMine={true} />                    {/* ✅ Only my tasks */}
    </Tab>
    
    {/* ❌ No Timeline tab */}
    {/* ❌ No Analytics tab */}
    {/* ❌ No All Tasks tab */}
  </Tabs>
</ProjectDetailPage>
```

---

### Create/Edit Forms

#### Create Task Form - Visibility by Role

**Admin/Manager:**
```tsx
<CreateTaskForm>
  <Input name="title" required />                           {/* ✅ */}
  <Textarea name="description" required />                  {/* ✅ */}
  <Select name="assign_to" options={allUsers} />           {/* ✅ All users */}
  <Select name="project" options={allProjects} />          {/* ✅ All projects */}
  <DatePicker name="start_time" />                         {/* ✅ */}
  <DatePicker name="end_time" />                           {/* ✅ */}
  <DatePicker name="due_date" />                           {/* ✅ */}
  <Select name="priority" options={['high','medium','low']} /> {/* ✅ For collab */}
  <Select name="complexity" options={['simple','medium','complex']} /> {/* ✅ For collab */}
  <NumberInput name="max_participants" />                  {/* ✅ For collab */}
</CreateTaskForm>
```

**Head:**
```tsx
<CreateTaskForm>
  <Input name="title" required />                           {/* ✅ */}
  <Textarea name="description" required />                  {/* ✅ */}
  <Select name="assign_to" options={employeesOnly} />      {/* ✅ Limited */}
  <Select name="project" options={myProjects} />           {/* ✅ Limited */}
  <DatePicker name="due_date" />                           {/* ✅ */}
  {/* ❌ No priority */}
  {/* ❌ No complexity */}
  {/* ❌ No max_participants */}
</CreateTaskForm>
```

**Employee:**
```tsx
<CreateTaskForm>
  <Input name="title" required />                           {/* ✅ */}
  <Textarea name="description" required />                  {/* ✅ */}
  <Select name="project" options={myProjects} />           {/* ✅ Limited */}
  <DatePicker name="due_date" />                           {/* ✅ */}
  {/* ❌ No assign_to (auto-assigned to self) */}
  {/* ❌ No priority */}
  {/* ❌ No start/end time */}
</CreateTaskForm>
```

---

### Navigation Sidebar (Complete Structure)

```tsx
// Sidebar.tsx - Conditional rendering by role
const Sidebar = () => {
  const { user, isAdmin, isManagerOrHigher, isHeadOrHigher, isHR } = useAuth();

  return (
    <nav className="sidebar">
      {/* All Users */}
      <NavLink to="/dashboard">
        <Icon name="home" />
        Dashboard
      </NavLink>

      {/* Employee+ (not HR) */}
      {!isHR() && (
        <NavLink to="/my-tasks">
          <Icon name="tasks" />
          My Tasks
        </NavLink>
      )}

      {/* Head+ */}
      {isHeadOrHigher() && (
        <NavLink to="/team-tasks">
          <Icon name="team" />
          Team Tasks
        </NavLink>
      )}

      {/* All Users except HR */}
      {!isHR() && (
        <NavLink to="/projects">
          <Icon name="folder" />
          Projects
        </NavLink>
      )}

      {/* Manager+ only */}
      {isManagerOrHigher() && (
        <>
          <NavLink to="/analytics">
            <Icon name="chart" />
            Analytics
          </NavLink>
          <NavLink to="/reports">
            <Icon name="file" />
            Reports
          </NavLink>
        </>
      )}

      {/* Admin only */}
      {isAdmin() && (
        <>
          <NavLink to="/users">
            <Icon name="users" />
            Users
          </NavLink>
          <NavLink to="/settings">
            <Icon name="settings" />
            Settings
          </NavLink>
        </>
      )}

      {/* HR only */}
      {isHR() && (
        <>
          <NavLink to="/users">
            <Icon name="users" />
            Users (View)
          </NavLink>
          <NavLink to="/hr-problems">
            <Icon name="alert" />
            HR Problems
          </NavLink>
        </>
      )}

      {/* All Users */}
      <NavLink to="/chat">
        <Icon name="message" />
        Chat
      </NavLink>

      <NavLink to="/notifications">
        <Icon name="bell" />
        Notifications
        {unreadCount > 0 && <Badge>{unreadCount}</Badge>}
      </NavLink>

      {/* User Menu */}
      <UserMenu>
        <Avatar user={user} />
        <Dropdown>
          <MenuItem onClick={viewProfile}>Profile</MenuItem>
          <MenuItem onClick={settings}>Settings</MenuItem>
          <MenuItem onClick={logout}>Logout</MenuItem>
        </Dropdown>
      </UserMenu>
    </nav>
  );
};
```

---

### Header/Top Bar Components

```tsx
// Header.tsx - Role-specific action buttons
const Header = ({ page }: { page: string }) => {
  const { isAdmin, isManagerOrHigher, isHeadOrHigher } = useAuth();

  // Dashboard page header
  if (page === 'dashboard') {
    return (
      <header>
        <h1>Dashboard</h1>
        <Actions>
          {isHeadOrHigher() && (
            <Button onClick={createTask}>
              <Icon name="plus" /> Create Task
            </Button>
          )}
          {isManagerOrHigher() && (
            <Button onClick={createProject}>
              <Icon name="folder-plus" /> Create Project
            </Button>
          )}
          {isAdmin() && (
            <Button onClick={createUser}>
              <Icon name="user-plus" /> Create User
            </Button>
          )}
        </Actions>
      </header>
    );
  }

  // Tasks page header
  if (page === 'tasks') {
    return (
      <header>
        <h1>Tasks</h1>
        <Actions>
          {isHeadOrHigher() && (
            <Button onClick={createTask}>
              <Icon name="plus" /> New Task
            </Button>
          )}
          {isManagerOrHigher() && (
            <Button onClick={bulkActions}>
              <Icon name="layers" /> Bulk Actions
            </Button>
          )}
        </Actions>
      </header>
    );
  }

  // Projects page header
  if (page === 'projects') {
    return (
      <header>
        <h1>Projects</h1>
        <Actions>
          {isManagerOrHigher() && (
            <Button onClick={createProject}>
              <Icon name="plus" /> New Project
            </Button>
          )}
        </Actions>
      </header>
    );
  }

  return <header><h1>{page}</h1></header>;
};
```

---

## Detailed Differences Summary

### Key UI Differences

#### 1. Task Assignment Field

| Role | Task Assignment UI |
|------|-------------------|
| **Admin** | Dropdown shows: All users (admin, manager, head, employee, hr) |
| **Manager** | Dropdown shows: All users except other managers/admins |
| **Head** | Dropdown shows: Only employees in their department |
| **Employee** | No dropdown - tasks auto-assigned to self |
| **HR** | No task creation |

#### 2. Project Creation Access

| Role | Can Create Project | Can See "Create Project" Button |
|------|:------------------:|:-------------------------------:|
| **Admin** | ✅ Yes | ✅ Visible everywhere |
| **Manager** | ✅ Yes | ✅ Visible everywhere |
| **Head** | ❌ No | ❌ Hidden |
| **Employee** | ❌ No | ❌ Hidden |
| **HR** | ❌ No | ❌ Hidden |

#### 3. User List View

| Role | What Users They See | Can Edit Users |
|------|---------------------|:--------------:|
| **Admin** | All users | ✅ Full edit |
| **Manager** | Team/department users | ⚠️ Limited (assign tasks) |
| **Head** | No user list page | ❌ |
| **Employee** | No user list page | ❌ |
| **HR** | All users | ⚠️ View + update dept |

#### 4. Analytics Access

| Role | Analytics Page | What They See |
|------|:--------------:|---------------|
| **Admin** | ✅ Full access | System-wide analytics, all projects, all users |
| **Manager** | ✅ Limited | Their projects, their teams, their departments |
| **Head** | ❌ No access | Page hidden |
| **Employee** | ❌ No access | Page hidden |
| **HR** | ✅ Read-only | User stats, department stats (no task assignment) |

#### 5. Task Status Update Permission

| Role | Can Update Status | Conditions |
|------|:----------------:|------------|
| **Admin** | ✅ Any task | Can change any task status |
| **Manager** | ⚠️ Limited | Only tasks in their projects |
| **Head** | ⚠️ Own tasks | Only tasks assigned to them |
| **Employee** | ⚠️ Own tasks | Only tasks assigned to them |
| **HR** | ❌ No | Cannot update task status |

#### 6. Bulk Operations Access

| Role | Bulk Select Checkboxes | Bulk Actions Menu |
|------|:---------------------:|:-----------------:|
| **Admin** | ✅ On all tasks | ✅ Full menu (update/delete/reassign) |
| **Manager** | ✅ On their tasks | ✅ Full menu (their scope) |
| **Head** | ❌ Hidden | ❌ Hidden |
| **Employee** | ❌ Hidden | ❌ Hidden |
| **HR** | ❌ Hidden | ❌ Hidden |

---

## Context Menu (Right-click) Differences

### On Task Card Right-Click

**Admin:**
```
├─ View Details
├─ Edit Task
├─ Change Status ►
│  ├─ Pending
│  ├─ In Progress
│  ├─ Completed
│  └─ Cancelled
├─ Reassign ►
│  └─ [List of all users]
├─ Move to Project ►
│  └─ [List of all projects]
├─ View Time Analysis
├─ Add to Bulk Selection
├─ Duplicate Task
└─ Delete Task
```

**Manager:**
```
├─ View Details
├─ Edit Task (if in their project)
├─ Reassign ► (if in their project)
│  └─ [List of team users]
├─ Move to Project ► (if in their project)
│  └─ [List of their projects]
├─ View Time Analysis
└─ Delete Task (if in their project)
```

**Head:**
```
├─ View Details
├─ Update Status (if own task) ►
│  ├─ Pending
│  ├─ In Progress
│  └─ Completed
├─ Edit Task (if task in my project) ✨ NEW
├─ Delete Task (if task in my project) ✨ NEW
└─ View Task Info
```

**Employee:**
```
├─ View Details
└─ Update Status ►
   ├─ Pending
   ├─ In Progress
   └─ Completed
```

**HR:**
```
└─ View Details (read-only)
```

---

## Modal/Dialog Differences

### Create Task Modal

**Admin:**
```tsx
<Modal title="Create Task">
  <Form>
    <Input label="Title" required />
    <Textarea label="Description" required />
    
    {/* Assignment Section - FULL ACCESS */}
    <FormSection title="Assignment">
      <Select label="Assign to" options={allUsers} />
      <Select label="Project" options={allProjects} />
    </FormSection>
    
    {/* Timing Section */}
    <FormSection title="Schedule">
      <DateTimePicker label="Start Time" />
      <DateTimePicker label="End Time" />
      <DatePicker label="Due Date" />
    </FormSection>
    
    {/* Task Type */}
    <RadioGroup label="Task Type">
      <Radio value="regular">Regular Task</Radio>
      <Radio value="collaborative">Collaborative Task</Radio>
    </RadioGroup>
    
    {/* Collaborative Options (if collaborative selected) */}
    <FormSection title="Collaboration" show={isCollaborative}>
      <Select label="Priority" options={['high','medium','low']} />
      <Select label="Complexity" options={['simple','medium','complex']} />
      <NumberInput label="Max Participants" default={5} />
      <MultiSelect label="Add Participants" options={allUsers} />
    </FormSection>
    
    <Actions>
      <Button type="submit">Create Task</Button>
      <Button onClick={cancel}>Cancel</Button>
    </Actions>
  </Form>
</Modal>
```

**Manager:**
```tsx
<Modal title="Create Task">
  <Form>
    <Input label="Title" required />
    <Textarea label="Description" required />
    
    {/* Assignment - LIMITED */}
    <FormSection title="Assignment">
      <Select label="Assign to" options={teamUsers} />     {/* ⚠️ Limited users */}
      <Select label="Project" options={myProjects} />      {/* ⚠️ My projects */}
    </FormSection>
    
    {/* Timing Section */}
    <FormSection title="Schedule">
      <DateTimePicker label="Start Time" />
      <DateTimePicker label="End Time" />
      <DatePicker label="Due Date" />
    </FormSection>
    
    {/* Same collaborative options as Admin */}
    
    <Actions>
      <Button type="submit">Create Task</Button>
      <Button onClick={cancel}>Cancel</Button>
    </Actions>
  </Form>
</Modal>
```

**Head:**
```tsx
<Modal title="Create Task">
  <Form>
    <Input label="Title" required />
    <Textarea label="Description" required />
    
    {/* Assignment - MORE LIMITED */}
    <FormSection title="Assignment">
      <Select label="Assign to" options={employeesOnly} /> {/* ⚠️ Employees only */}
      <Select label="Project" options={myProjects} />      {/* ⚠️ My projects */}
    </FormSection>
    
    <DatePicker label="Due Date" />                        {/* ✅ */}
    
    {/* ❌ No start/end time */}
    {/* ❌ No priority/complexity */}
    {/* ❌ No collaborative options */}
    
    <Actions>
      <Button type="submit">Create Task</Button>
      <Button onClick={cancel}>Cancel</Button>
    </Actions>
  </Form>
</Modal>
```

**Employee:**
```tsx
<Modal title="Create Task">
  <Form>
    <Input label="Title" required />
    <Textarea label="Description" required />
    <Select label="Project" options={myProjects} />       {/* ✅ My projects */}
    <DatePicker label="Due Date" />                        {/* ✅ */}
    
    {/* ❌ No assign_to field */}
    {/* ❌ No start/end time */}
    {/* ❌ No priority */}
    
    <Actions>
      <Button type="submit">Create Task for Me</Button>   {/* Different text */}
      <Button onClick={cancel}>Cancel</Button>
    </Actions>
  </Form>
</Modal>
```

---

## API Endpoints Access Summary

### Task Endpoints

| Endpoint | Admin | Manager | Head | Employee | HR |
|----------|:-----:|:-------:|:----:|:--------:|:--:|
| `GET /api/tasks` | ✅ All | ✅ Team | ✅ Own | ✅ Own | ❌ |
| `POST /api/tasks` | ✅ | ✅ | ✅ | ✅ | ❌ |
| `PUT /api/tasks/:id` | ✅ | ✅ | ✅* | ❌ | ❌ |
| `DELETE /api/tasks/:id` | ✅ | ✅ | ✅* | ❌ | ❌ |
| `PATCH /api/tasks/:id/status` | ✅ | ✅* | ✅* | ✅* | ❌ |
| `POST /api/tasks/bulk` | ✅ | ✅ | ❌ | ❌ | ❌ |
| `GET /api/tasks/statistics` | ✅ | ✅ | ❌ | ❌ | ❌ |
| `GET /api/tasks/arabic-context` | ✅ | ✅ | ✅ | ✅ | ❌ |
| `POST /api/tasks/arabic-schedule` | ✅ | ✅ | ✅ | ✅ | ❌ |
| `GET /api/tasks/:id/time-analysis` | ✅ | ✅ | ✅ | ✅ | ❌ |

*Only tasks in projects they're members of

### Project Endpoints

| Endpoint | Admin | Manager | Head | Employee | HR |
|----------|:-----:|:-------:|:----:|:--------:|:--:|
| `GET /api/projects/my-projects` | ✅ | ✅ | ✅ | ✅ | ❌ |
| `POST /api/projects` | ✅ | ✅ | ❌ | ❌ | ❌ |
| `GET /api/projects/:id` | ✅ | ✅ | ✅* | ✅* | ❌ |
| `GET /api/projects/:id/members` | ✅ | ✅ | ✅* | ✅* | ❌ |
| `POST /api/projects/:id/members` | ✅ | ✅ | ❌ | ❌ | ❌ |
| `DELETE /api/projects/:id/members/:userId` | ✅ | ✅ | ❌ | ❌ | ❌ |
| `PATCH /api/projects/:id/status` | ✅ | ✅ | ❌ | ❌ | ❌ |
| `GET /api/projects/:id/statistics` | ✅ | ✅ | ✅* | ✅* | ❌ |
| `POST /api/projects/:id/generate-tasks` | ✅ | ✅ | ❌ | ❌ | ❌ |
| `POST /api/projects/:id/confirm-tasks` | ✅ | ✅ | ❌ | ❌ | ❌ |
| `DELETE /api/projects/:id` | ✅ | ❌ | ❌ | ❌ | ❌ |

*Only projects they're members of

### User Endpoints

| Endpoint | Admin | Manager | Head | Employee | HR |
|----------|:-----:|:-------:|:----:|:--------:|:--:|
| `GET /api/users` | ✅ | ✅ | ❌ | ❌ | ✅ |
| `POST /api/users` | ✅ | ❌ | ❌ | ❌ | ❌ |
| `GET /api/users/:id` | ✅ | ✅* | ✅** | ✅** | ✅ |
| `PATCH /api/users/:id/role` | ✅ | ❌ | ❌ | ❌ | ❌ |
| `PATCH /api/users/:id/department` | ✅ | ✅ | ❌ | ❌ | ✅ |
| `DELETE /api/users/:id` | ✅ | ❌ | ❌ | ❌ | ❌ |
| `GET /api/users/role/:role` | ✅ | ❌ | ❌ | ❌ | ✅ |
| `GET /api/users/department/:dept` | ✅ | ❌ | ❌ | ❌ | ✅ |
| `PATCH /api/users/:id/password` | ✅ | ✅** | ✅** | ✅** | ✅** |
| `PATCH /api/users/:id/skills` | ✅ | ✅** | ✅** | ✅** | ✅** |

*Team members only  
**Self only (unless Admin)

---

## Small Click Differences

### Delete Button Color/Confirmation

**Admin:**
- Delete button: Red, no special warning
- Confirmation: "Are you sure? This will delete the task."

**Manager:**
- Delete button: Red, shown only for their tasks
- Confirmation: "Delete this task from your project?"

**Head/Employee:**
- Delete button: HIDDEN (no delete access)

### Edit Icon Visibility

**On Task Card:**
- Admin: ✅ Pencil icon always visible
- Manager: ✅ Pencil icon on their project tasks only
- Head: ✅ Pencil icon on tasks in projects they're members of ✨ NEW
- Employee: ❌ No pencil icon (can only change status)
- HR: ❌ No edit access at all

### Dropdown Options

**Status Dropdown:**
- Admin/Manager: Shows all 4 statuses (pending, in_progress, completed, cancelled)
- Head/Employee: Shows 3 statuses (pending, in_progress, completed) - no "cancelled"

**Priority Dropdown (Collaborative Tasks):**
- Admin/Manager: ✅ Can set priority (high/medium/low)
- Head/Employee: ❌ Field hidden, priority is auto-set to "medium"
- HR: ❌ No access

### Tooltip Differences

**On Disabled Buttons:**
- Admin: No disabled buttons
- Manager: "This feature requires Admin access"
- Head: "This feature requires Manager access" OR "This task must be in a project you're a member of"
- Employee: "This feature requires Head access"
- HR: "HR cannot perform task operations"

---

## Complete Implementation Checklist

### For Each Page, Check:

#### ✅ Dashboard
- [ ] Shows correct widgets for role
- [ ] Action buttons match role permissions
- [ ] Stats show correct scope (system/team/personal)
- [ ] Quick actions are role-appropriate

#### ✅ Tasks Page
- [ ] Filter options match role permissions
- [ ] Task cards show correct actions
- [ ] Bulk selection shown/hidden correctly
- [ ] Create button visible for Head+
- [ ] Edit buttons only for Manager+
- [ ] Status dropdown for task owners only

#### ✅ Projects Page
- [ ] Shows correct projects (all/team/mine)
- [ ] Create button for Manager+ only
- [ ] Project cards show correct actions
- [ ] Member management for Manager+ only
- [ ] AI generation for Manager+ only

#### ✅ Users Page
- [ ] Shown for Admin/Manager/HR only
- [ ] Hidden for Head/Employee
- [ ] Edit actions for Admin only
- [ ] Department update for Admin/Manager/HR
- [ ] Role change for Admin only

#### ✅ Analytics Page
- [ ] Shown for Admin/Manager/HR only
- [ ] Hidden for Head/Employee
- [ ] HR sees read-only view
- [ ] Correct scope (system/team)

#### ✅ Navigation
- [ ] Menu items match role permissions
- [ ] Badges/counters show correct data
- [ ] Links navigate to allowed pages only

#### ✅ Forms
- [ ] Assignment field shown/hidden correctly
- [ ] Dropdown options match permissions
- [ ] Advanced fields for Manager+ only
- [ ] Auto-assignment for Employee

---

**This guide should be checked twice before implementation. Test each role thoroughly to ensure all UI elements match these specifications.**

