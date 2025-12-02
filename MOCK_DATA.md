  # Mock Data for Frontend Development

This file contains mock data that frontend developers can use to work locally without the backend API.

## 🔐 Authentication Mock Data

### Login Response (All Roles)

#### Admin Login
```json
{
  "token": "mock_admin_token_12345",
  "user": {
    "id": 1,
    "username": "admin",
    "role": "admin",
    "department": "IT",
    "skills": "[\"System Administration\", \"Database Management\", \"Security\"]",
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

#### Manager Login
```json
{
  "token": "mock_manager_token_12345",
  "user": {
    "id": 2,
    "username": "manager",
    "role": "manager",
    "department": "Engineering",
    "skills": "[\"Project Management\", \"Team Leadership\", \"Agile\"]",
    "created_at": "2024-01-16T09:00:00Z"
  }
}
```

#### Head Login
```json
{
  "token": "mock_head_token_12345",
  "user": {
    "id": 3,
    "username": "head",
    "role": "head",
    "department": "Engineering",
    "skills": "[\"Backend Development\", \"API Design\", \"Code Review\"]",
    "created_at": "2024-01-17T08:00:00Z"
  }
}
```

#### Employee Login
```json
{
  "token": "mock_employee_token_12345",
  "user": {
    "id": 4,
    "username": "employee",
    "role": "employee",
    "department": "Engineering",
    "skills": "[\"Frontend Development\", \"React\", \"TypeScript\"]",
    "created_at": "2024-01-18T09:30:00Z"
  }
}
```

#### HR Login
```json
{
  "token": "mock_hr_token_12345",
  "user": {
    "id": 5,
    "username": "hr",
    "role": "hr",
    "department": "Human Resources",
    "skills": "[\"HR Management\", \"Recruitment\", \"Employee Relations\"]",
    "created_at": "2024-01-19T10:00:00Z"
  }
}
```

---

## 👥 Users Mock Data

### Get All Users (Manager+/HR+)
```json
{
  "users": [
    {
      "id": 1,
      "username": "admin",
      "role": "admin",
      "department": "IT",
      "skills": "[\"System Administration\"]",
      "created_at": "2024-01-15T10:00:00Z"
    },
    {
      "id": 2,
      "username": "manager",
      "role": "manager",
      "department": "Engineering",
      "skills": "[\"Project Management\"]",
      "created_at": "2024-01-16T09:00:00Z"
    },
    {
      "id": 3,
      "username": "head",
      "role": "head",
      "department": "Engineering",
      "skills": "[\"Backend Development\"]",
      "created_at": "2024-01-17T08:00:00Z"
    },
    {
      "id": 4,
      "username": "employee",
      "role": "employee",
      "department": "Engineering",
      "skills": "[\"Frontend Development\"]",
      "created_at": "2024-01-18T09:30:00Z"
    },
    {
      "id": 5,
      "username": "hr",
      "role": "hr",
      "department": "Human Resources",
      "skills": "[\"HR Management\"]",
      "created_at": "2024-01-19T10:00:00Z"
    }
  ]
}
```

---

## 📋 Tasks Mock Data

### Get User Tasks (Regular Tasks)
```json
{
  "tasks": [
    {
      "id": 1,
      "title": "Implement user authentication",
      "description": "Add JWT-based authentication system",
      "status": "in_progress",
      "project_id": 1,
      "assigned_at": "2024-01-20T09:00:00Z",
      "due_date": "2024-01-25T17:00:00Z",
      "created_at": "2024-01-20T09:00:00Z"
    },
    {
      "id": 2,
      "title": "Design database schema",
      "description": "Create ERD for project management system",
      "status": "completed",
      "project_id": 1,
      "assigned_at": "2024-01-18T10:00:00Z",
      "due_date": "2024-01-22T17:00:00Z",
      "created_at": "2024-01-18T10:00:00Z"
    },
    {
      "id": 3,
      "title": "Write API documentation",
      "description": "Document all REST API endpoints",
      "status": "pending",
      "project_id": 2,
      "assigned_at": "2024-01-21T11:00:00Z",
      "due_date": "2024-01-28T17:00:00Z",
      "created_at": "2024-01-21T11:00:00Z"
    }
  ]
}
```

### Get User Tasks (With Full Details - Enhanced)
```json
{
  "tasks": [
    {
      "id": 1,
      "title": "Implement user authentication",
      "description": "Add JWT-based authentication system with role-based access control",
      "status": "in_progress",
      "user": {
        "id": 4,
        "username": "employee"
      },
      "project": {
        "id": 1,
        "title": "Project X"
      },
      "start_time": "2024-01-20T09:00:00Z",
      "end_time": "2024-01-25T17:00:00Z",
      "due_date": "2024-01-25T17:00:00Z",
      "assigned_at": "2024-01-20T09:00:00Z",
      "created_at": "2024-01-20T09:00:00Z",
      "updated_at": "2024-01-22T14:30:00Z"
    }
  ]
}
```

### Get Collaborative Tasks
```json
{
  "collaborative_tasks": [
    {
      "id": 1,
      "title": "Redesign user interface",
      "description": "Complete UI/UX redesign for mobile and web",
      "status": "in_progress",
      "lead_user_id": 3,
      "lead_user_name": "head",
      "project_id": 1,
      "assigned_at": "2024-01-19T10:00:00Z",
      "due_date": "2024-01-30T17:00:00Z",
      "created_at": "2024-01-19T10:00:00Z",
      "priority": "high",
      "progress": 45,
      "complexity": "complex",
      "max_participants": 5,
      "participants": [
        {
          "id": 1,
          "user_id": 3,
          "username": "head",
          "role": "lead",
          "status": "active"
        },
        {
          "id": 2,
          "user_id": 4,
          "username": "employee",
          "role": "contributor",
          "status": "active"
        }
      ]
    }
  ]
}
```

### Get Project Tasks
```json
{
  "tasks": [
    {
      "id": 1,
      "title": "Setup development environment",
      "description": "Configure local development setup",
      "status": "completed",
      "user_id": 4,
      "assigned_at": "2024-01-15T09:00:00Z",
      "due_date": "2024-01-18T17:00:00Z",
      "created_at": "2024-01-15T09:00:00Z"
    },
    {
      "id": 2,
      "title": "Implement core features",
      "description": "Build main application features",
      "status": "in_progress",
      "user_id": 3,
      "assigned_at": "2024-01-16T10:00:00Z",
      "due_date": "2024-01-25T17:00:00Z",
      "created_at": "2024-01-16T10:00:00Z"
    }
  ]
}
```

---

## 📁 Projects Mock Data

### Get User Projects
```json
{
  "projects": [
    {
      "id": 1,
      "title": "Project X - Main System",
      "description": "Main project management system development",
      "status": "active",
      "created_by": 2,
      "start_date": "2024-01-01T00:00:00Z",
      "end_date": "2024-06-30T23:59:59Z",
      "created_at": "2024-01-01T08:00:00Z"
    },
    {
      "id": 2,
      "title": "Mobile App Development",
      "description": "iOS and Android mobile application",
      "status": "active",
      "created_by": 2,
      "start_date": "2024-02-01T00:00:00Z",
      "end_date": "2024-08-31T23:59:59Z",
      "created_at": "2024-02-01T09:00:00Z"
    }
  ]
}
```

### Get Project Details
```json
{
  "project": {
    "id": 1,
    "title": "Project X - Main System",
    "description": "Main project management system development",
    "status": "active",
    "created_by": 2,
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-06-30T23:59:59Z",
    "created_at": "2024-01-01T08:00:00Z",
    "creator": {
      "id": 2,
      "username": "manager",
      "role": "manager"
    },
    "member_count": 5,
    "task_count": 12
  }
}
```

### Get Project Members
```json
{
  "members": [
    {
      "id": 2,
      "username": "manager",
      "role": "manager",
      "department": "Engineering",
      "skills": "[\"Project Management\"]",
      "project_role": "manager",
      "job_role": "Project Manager"
    },
    {
      "id": 3,
      "username": "head",
      "role": "head",
      "department": "Engineering",
      "skills": "[\"Backend Development\"]",
      "project_role": "head",
      "job_role": "Backend Developer"
    },
    {
      "id": 4,
      "username": "employee",
      "role": "employee",
      "department": "Engineering",
      "skills": "[\"Frontend Development\"]",
      "project_role": "employee",
      "job_role": "Frontend Developer"
    }
  ]
}
```

### Get Project Statistics
```json
{
  "statistics": {
    "project_id": 1,
    "project_title": "Project X - Main System",
    "status": "active",
    "member_count": 5,
    "total_tasks": 12,
    "pending_tasks": 3,
    "in_progress_tasks": 5,
    "completed_tasks": 4,
    "cancelled_tasks": 0,
    "completion_rate": 33.33,
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-06-30T23:59:59Z"
  }
}
```

---

## 🔔 Notifications Mock Data

### Get User Notifications
```json
{
  "notifications": [
    {
      "id": 1,
      "user_id": 4,
      "type": "task_assigned",
      "title": "New Task Assigned",
      "message": "You have been assigned a new task: 'Implement user authentication' by manager",
      "data": "{\"task_id\":1,\"from_user_id\":2,\"task_title\":\"Implement user authentication\",\"from_user\":\"manager\"}",
      "is_read": false,
      "read_at": null,
      "related_task_id": 1,
      "from_user_id": 2,
      "created_at": "2024-01-20T09:00:00Z"
    },
    {
      "id": 2,
      "user_id": 4,
      "type": "task_updated",
      "title": "Task Updated",
      "message": "Task 'Design database schema' has been updated by manager (Changes: due date changed)",
      "data": "{\"task_id\":2,\"from_user_id\":2,\"task_title\":\"Design database schema\",\"from_user\":\"manager\"}",
      "is_read": true,
      "read_at": "2024-01-20T10:30:00Z",
      "related_task_id": 2,
      "from_user_id": 2,
      "created_at": "2024-01-20T10:00:00Z"
    },
    {
      "id": 3,
      "user_id": 4,
      "type": "task_completed",
      "title": "Task Completed",
      "message": "Task 'Design database schema' has been completed by employee",
      "data": "{\"task_id\":2,\"from_user_id\":4,\"task_title\":\"Design database schema\",\"from_user\":\"employee\"}",
      "is_read": false,
      "read_at": null,
      "related_task_id": 2,
      "from_user_id": 4,
      "created_at": "2024-01-22T15:00:00Z"
    }
  ],
  "unread_count": 2
}
```

### Notification Types
```typescript
type NotificationType = 
  | "task_assigned"
  | "task_updated"
  | "task_completed"
  | "task_commented"
  | "task_due_soon"
  | "project_created"
  | "user_joined"
  | "file_uploaded"
  | "hr_problem"
  | "hr_problem_update"
  | "hr_problem_assigned";
```

---

## 📊 Reports Mock Data

### Project Report (Weekly)
```json
{
  "report": {
    "project": {
      "id": 1,
      "title": "Project X - Main System",
      "description": "Main project management system",
      "status": "active",
      "start_date": "2024-01-01T00:00:00Z",
      "end_date": "2024-06-30T23:59:59Z"
    },
    "period": {
      "type": "weekly",
      "start_date": "2024-01-15T00:00:00Z",
      "end_date": "2024-01-22T23:59:59Z"
    },
    "statistics": {
      "regular_tasks": {
        "total": 8,
        "pending": 2,
        "in_progress": 4,
        "completed": 2,
        "cancelled": 0,
        "created_in_period": 3
      },
      "collaborative_tasks": {
        "total": 4,
        "pending": 1,
        "in_progress": 2,
        "completed": 1,
        "cancelled": 0,
        "created_in_period": 1
      },
      "overall": {
        "total_tasks": 12,
        "completed_tasks": 3,
        "completion_rate": 25.0
      }
    },
    "user_performance": [
      {
        "user_id": 3,
        "username": "head",
        "role": "head",
        "department": "Engineering",
        "total_tasks": 5,
        "completed_tasks": 2,
        "completion_rate": 40.0,
        "regular_tasks": {
          "total": 3,
          "completed": 1
        },
        "collaborative_tasks": {
          "total": 2,
          "completed": 1
        }
      },
      {
        "user_id": 4,
        "username": "employee",
        "role": "employee",
        "department": "Engineering",
        "total_tasks": 7,
        "completed_tasks": 1,
        "completion_rate": 14.29,
        "regular_tasks": {
          "total": 5,
          "completed": 1
        },
        "collaborative_tasks": {
          "total": 2,
          "completed": 0
        }
      }
    ]
  }
}
```

### User Report (Monthly)
```json
{
  "report": {
    "user": {
      "id": 4,
      "username": "employee",
      "role": "employee",
      "department": "Engineering"
    },
    "period": {
      "type": "monthly",
      "start_date": "2024-01-01T00:00:00Z",
      "end_date": "2024-01-31T23:59:59Z"
    },
    "statistics": {
      "regular_tasks": {
        "total": 15,
        "pending": 3,
        "in_progress": 5,
        "completed": 7,
        "cancelled": 0,
        "created_in_period": 8,
        "completed_in_period": 7
      },
      "collaborative_tasks": {
        "total": 5,
        "pending": 1,
        "in_progress": 2,
        "completed": 2,
        "cancelled": 0,
        "created_in_period": 3,
        "completed_in_period": 2
      },
      "overall": {
        "total_tasks": 20,
        "completed_tasks": 9,
        "completion_rate": 45.0
      },
      "period_performance": {
        "total_tasks": 11,
        "completed_tasks": 9,
        "completion_rate": 81.82
      }
    },
    "project_performance": [
      {
        "project_id": 1,
        "project_title": "Project X - Main System",
        "project_status": "active",
        "total_tasks": 12,
        "completed_tasks": 6,
        "completion_rate": 50.0,
        "regular_tasks": {
          "total": 8,
          "completed": 4
        },
        "collaborative_tasks": {
          "total": 4,
          "completed": 2
        }
      }
    ]
  }
}
```

---

## 💬 Chat Mock Data

### Get Team Chat
```json
{
  "room": {
    "id": 1,
    "name": "Team Chat",
    "type": "team",
    "created_at": "2024-01-01T08:00:00Z"
  },
  "members": [
    {
      "id": 2,
      "username": "manager",
      "role": "manager"
    },
    {
      "id": 3,
      "username": "head",
      "role": "head"
    },
    {
      "id": 4,
      "username": "employee",
      "role": "employee"
    }
  ]
}
```

### Get Chat Messages
```json
{
  "messages": [
    {
      "id": 1,
      "room_id": 1,
      "user_id": 2,
      "username": "manager",
      "message": "Welcome to the team chat!",
      "created_at": "2024-01-20T09:00:00Z"
    },
    {
      "id": 2,
      "room_id": 1,
      "user_id": 3,
      "username": "head",
      "message": "Thanks! Looking forward to working together.",
      "created_at": "2024-01-20T09:05:00Z"
    },
    {
      "id": 3,
      "room_id": 1,
      "user_id": 4,
      "username": "employee",
      "message": "Hello everyone! 👋",
      "created_at": "2024-01-20T09:10:00Z"
    }
  ]
}
```

---

## 🏢 HR Problems Mock Data

### Get HR Problems (HR/Admin)
```json
{
  "problems": [
    {
      "id": 1,
      "reported_by": 4,
      "reporter_name": "employee",
      "title": "Workplace harassment complaint",
      "description": "Experiencing inappropriate behavior from a colleague",
      "status": "open",
      "priority": "high",
      "category": "harassment",
      "assigned_to": 5,
      "assigned_to_name": "hr",
      "created_at": "2024-01-20T10:00:00Z",
      "updated_at": "2024-01-20T10:00:00Z"
    },
    {
      "id": 2,
      "reported_by": 3,
      "reporter_name": "head",
      "title": "Request for flexible working hours",
      "description": "Would like to adjust working hours for better work-life balance",
      "status": "in_progress",
      "priority": "medium",
      "category": "work_conditions",
      "assigned_to": 5,
      "assigned_to_name": "hr",
      "created_at": "2024-01-18T14:00:00Z",
      "updated_at": "2024-01-19T11:00:00Z"
    }
  ]
}
```

### Get User's HR Problems
```json
{
  "problems": [
    {
      "id": 1,
      "reported_by": 4,
      "reporter_name": "employee",
      "title": "Workplace harassment complaint",
      "description": "Experiencing inappropriate behavior from a colleague",
      "status": "open",
      "priority": "high",
      "category": "harassment",
      "assigned_to": 5,
      "assigned_to_name": "hr",
      "created_at": "2024-01-20T10:00:00Z",
      "updated_at": "2024-01-20T10:00:00Z",
      "comments": [
        {
          "id": 1,
          "problem_id": 1,
          "user_id": 5,
          "username": "hr",
          "comment": "We take this seriously. An investigation will be conducted.",
          "created_at": "2024-01-20T11:00:00Z"
        }
      ]
    }
  ]
}
```

---

## 📈 Statistics Mock Data

### Get Task Statistics (Manager+)
```json
{
  "statistics": {
    "regular_tasks": {
      "total": 50,
      "pending": 15,
      "in_progress": 20,
      "completed": 14,
      "cancelled": 1
    },
    "collaborative_tasks": {
      "total": 20,
      "pending": 5,
      "in_progress": 8,
      "completed": 7,
      "cancelled": 0
    },
    "overall": {
      "total_tasks": 70,
      "completed_tasks": 21,
      "completion_rate": 30.0
    }
  }
}
```

---

## 🎨 UI State Mock Data

### Dashboard Data (Role-based)

#### Admin Dashboard
```json
{
  "user": {
    "id": 1,
    "username": "admin",
    "role": "admin"
  },
  "stats": {
    "total_users": 25,
    "total_projects": 8,
    "total_tasks": 70,
    "active_projects": 6
  },
  "recent_activities": [
    {
      "type": "user_created",
      "message": "New user 'john.doe' created",
      "timestamp": "2024-01-22T10:00:00Z"
    },
    {
      "type": "project_created",
      "message": "Project 'Mobile App' created by manager",
      "timestamp": "2024-01-22T09:30:00Z"
    }
  ]
}
```

#### Manager Dashboard
```json
{
  "user": {
    "id": 2,
    "username": "manager",
    "role": "manager"
  },
  "stats": {
    "my_projects": 5,
    "team_members": 12,
    "total_tasks": 45,
    "overdue_tasks": 3
  },
  "projects": [
    {
      "id": 1,
      "title": "Project X",
      "progress": 65,
      "tasks_completed": 13,
      "tasks_total": 20
    }
  ]
}
```

#### Employee Dashboard
```json
{
  "user": {
    "id": 4,
    "username": "employee",
    "role": "employee"
  },
  "stats": {
    "my_tasks": 8,
    "pending": 2,
    "in_progress": 4,
    "completed": 2,
    "overdue": 0
  },
  "upcoming_deadlines": [
    {
      "task_id": 1,
      "title": "Implement user authentication",
      "due_date": "2024-01-25T17:00:00Z",
      "days_remaining": 3
    }
  ]
}
```

---

## 🔧 Mock API Service (TypeScript)

Create a mock API service for local development:

```typescript
// services/mockApi.ts
export const mockApi = {
  login: async (username: string, password: string) => {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 500));
    
    const mockUsers = {
      admin: {
        token: "mock_admin_token_12345",
        user: {
          id: 1,
          username: "admin",
          role: "admin",
          department: "IT",
          skills: "[]",
          created_at: "2024-01-15T10:00:00Z"
        }
      },
      manager: {
        token: "mock_manager_token_12345",
        user: {
          id: 2,
          username: "manager",
          role: "manager",
          department: "Engineering",
          skills: "[]",
          created_at: "2024-01-16T09:00:00Z"
        }
      },
      // ... add other roles
    };
    
    if (mockUsers[username] && password === "password123") {
      return { data: mockUsers[username] };
    }
    
    throw { response: { data: { error: "invalid credentials" } } };
  },
  
  getTasks: async () => {
    await new Promise(resolve => setTimeout(resolve, 300));
    return {
      data: {
        tasks: [
          {
            id: 1,
            title: "Implement user authentication",
            description: "Add JWT-based authentication system",
            status: "in_progress",
            project_id: 1,
            assigned_at: "2024-01-20T09:00:00Z",
            due_date: "2024-01-25T17:00:00Z",
            created_at: "2024-01-20T09:00:00Z"
          }
          // ... more tasks
        ]
      }
    };
  },
  
  // Add more mock methods as needed
};
```

---

## 📝 Usage Instructions

1. **For Login Testing:**
   - Use any username from the mock data
   - Password is always: `password123`
   - Response includes both token and user info

2. **For Role-based UI:**
   - Test with different roles: admin, manager, head, employee, hr
   - Each role has different permissions and UI visibility

3. **For Local Development:**
   - Create a mock API service that returns this data
   - Use environment variable to switch between mock and real API
   - Example: `const API = process.env.USE_MOCK ? mockApi : realApi;`

4. **Date Format:**
   - All dates are in ISO 8601 format: `2024-01-20T09:00:00Z`
   - Use `new Date(dateString)` in JavaScript

5. **Skills Field:**
   - Skills are stored as JSON string
   - Parse with: `JSON.parse(user.skills)`

---

## 🎯 Quick Test Credentials

| Role | Username | Password |
|------|----------|----------|
| Admin | `admin` | `password123` |
| Manager | `manager` | `password123` |
| Head | `head` | `password123` |
| Employee | `employee` | `password123` |
| HR | `hr` | `password123` |

---

**Note:** This mock data is for frontend development only. Always test with real backend API before production deployment.

