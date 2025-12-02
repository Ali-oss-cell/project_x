## Project X — Architecture Diagrams

This document provides system-level and feature-level diagrams. Diagrams use Mermaid; render in any Mermaid-enabled viewer or IDE.

### 1) System overview

```mermaid
flowchart LR
  subgraph Client
    UI[Web UI / Mobile]
  end

  subgraph Server(Go)
    Gin[REST API (Gin)]
    MW[Auth + RBAC Middleware]
    Handlers[Handlers]
    Services[Domain Services]
    WS[WebSocket Hub]
  end

  subgraph AI
    Gemini[Google Gemini 2.0 Flash]
  end

  subgraph DB
    PG[(PostgreSQL + GORM)]
  end

  UI <--> Gin
  Gin --> MW --> Handlers --> Services --> PG
  Services <---> Gemini
  UI <--> WS
```

### 2) Request lifecycle and RBAC

```mermaid
sequenceDiagram
  participant C as Client
  participant G as Gin Router
  participant MW as Auth/RBAC Middleware
  participant H as Handler
  participant S as Service
  participant DB as PostgreSQL

  C->>G: HTTP Request (+JWT)
  G->>MW: Validate JWT, load User
  MW-->>G: Check Role Permissions
  alt Authorized
    G->>H: Call handler
    H->>S: Domain call
    S->>DB: Query/Update
    DB-->>S: Result
    S-->>H: DTO/Result
    H-->>C: JSON Response
  else Forbidden/Unauthorized
    MW-->>C: 401/403
  end
```

### 3) Data model (key entities and relations)

```mermaid
classDiagram
  class User {
    uint ID
    string Username
    string Password
    Role Role
    string Department
    string Skills  // JSON
    // hasMany: Tasks
    // many2many: Projects via UserProject
  }

  class Project {
    uint ID
    string Title
    string Description
    time StartDate
    *time EndDate
    uint CreatedBy
  }

  class UserProject {
    uint UserID
    uint ProjectID
    time JoinedAt
    string Role        // manager/head/employee/member
    string JobRole     // project-specific (UX/UI, Backend, etc.)
  }

  class Task {
    uint ID
    uint ProjectID
    uint UserID
    string Title
    string Description
    string Status
    string Priority
    *time StartTime
    *time EndTime
    *time DueDate
  }

  class CollaborativeTask {
    uint ID
    uint ProjectID
    uint LeadUserID
    string Title
    string Description
    string Status
    string Priority
    *time StartTime
    *time EndTime
    *time DueDate
  }

  class CollaborativeTaskParticipant {
    uint ID
    uint CollaborativeTaskID
    uint UserID
    string Role
  }

  class AIAnalysis {
    uint ID
    uint TaskID
    string TaskType // regular/collaborative
    int PredictedDuration
    time PredictedCompletion
    string DeadlineRisk
    string RiskFactors // JSON
    string Recommendations // JSON
    time OptimalStartDate
    float WorkingHoursUntilDeadline
    *int ActualDuration
    *time ActualCompletion
    *bool WasAccurate
    *float AccuracyScore
    *int DurationDifference
    time AnalysisDate
    string ModelVersion
    int ConfidenceScore
  }

  class ChatRoom {
    uint ID
    string Name
    string Description
  }
  class ChatParticipant {
    uint ID
    uint RoomID
    uint UserID
  }
  class ChatMessage {
    uint ID
    uint RoomID
    uint UserID
    string Content
    time CreatedAt
  }

  class HRProblem {
    uint ID
    uint ReporterID
    string Title
    string Description
    string Status
  }
  class HRProblemComment {
    uint ID
    uint ProblemID
    uint UserID
    string Comment
  }
  class HRProblemUpdate {
    uint ID
    uint ProblemID
    string UpdateText
  }

  class Notification {
    uint ID
    uint UserID
    string Type
    string Payload // JSON
    bool Read
  }
  class UserNotificationPreference {
    uint ID
    uint UserID
    string Preferences // JSON
  }

  User "many" -- "many" Project : UserProject
  Project "1" -- "many" Task
  User "1" -- "many" Task
  Project "1" -- "many" CollaborativeTask
  CollaborativeTask "1" -- "many" CollaborativeTaskParticipant
  Task "1" -- "many" AIAnalysis
  ChatRoom "1" -- "many" ChatMessage
  ChatRoom "1" -- "many" ChatParticipant
  User "1" -- "many" ChatMessage
  User "1" -- "many" ChatParticipant
  HRProblem "1" -- "many" HRProblemComment
  HRProblem "1" -- "many" HRProblemUpdate
  User "1" -- "many" Notification
  User "1" -- "many" UserNotificationPreference
```

### 4) AI Project Task Generation (preview then confirm)

```mermaid
sequenceDiagram
  participant M as Manager
  participant H as Project Handler
  participant Gen as AI Project Task Generator
  participant DB as PostgreSQL
  participant G as Gemini

  M->>H: Generate tasks (projectId)
  H->>Gen: Build prompt (desc + JobRoles + skills + history)
  Gen->>G: GenerateContent(prompt)
  G-->>Gen: JSON { tasks, summary }
  Gen-->>H: TaskGenerationResponse
  H-->>M: Preview (editable list)

  M->>H: Confirm with edited tasks
  H->>Gen: CreateTasksFromGeneration(projectId, tasks)
  Gen->>DB: Insert tasks (validated, assigned members)
  DB-->>Gen: OK
  Gen-->>H: Created tasks
  H-->>M: Success response
```

### 5) AI Time Optimizer (caching + learning)

```mermaid
sequenceDiagram
  participant U as User/Manager
  participant H as AI Time Handler
  participant T as AI Time Optimizer
  participant DB as PostgreSQL
  participant G as Gemini

  U->>H: Analyze task(s)/project/workload
  H->>T: Analyze
  T->>DB: Check AIAnalysis cache (within 1h)
  alt Cached
    DB-->>T: Latest cached result
    T-->>H: Return cached analysis
  else Not Cached
    T->>DB: Gather historical data
    T->>G: GenerateContent(task prompt + Arabic working hours)
    G-->>T: JSON analysis
    T->>DB: Save AIAnalysis (prediction)
    T-->>H: Analysis result
  end

  Note over T,DB: On task completion, update AIAnalysis with actuals,\ncompute accuracy and duration difference
```

### 6) AI Chat Assistant (private room + @ai mentions)

```mermaid
sequenceDiagram
  participant U as User
  participant WS as WebSocket (Chat)
  participant CS as Chat Service
  participant AIC as AI Chat Service
  participant DB as PostgreSQL
  participant G as Gemini

  U->>WS: Send message (in AI room or '@ai ...')
  WS->>CS: OnMessage
  CS->>AIC: IsAIMessage? If yes, process
  AIC->>DB: Load user context (tasks, projects, workload)
  AIC->>G: GenerateContent(chat prompt + context)
  G-->>AIC: AI response / action
  alt Action: create task
    AIC->>DB: Insert task via TaskService
    DB-->>AIC: OK
  end
  AIC-->>CS: Response text + optional action result
  CS-->>WS: Broadcast AI reply to room
```

### 7) Real-time chat and membership

```mermaid
sequenceDiagram
  participant U as User
  participant WS as WebSocket Hub
  participant CS as Chat Service
  participant DB as PostgreSQL

  U->>WS: Connect with JWT
  WS->>CS: Authenticate and join room
  U->>WS: Send message
  WS->>CS: Persist message
  CS->>DB: Insert ChatMessage
  DB-->>CS: OK
  CS-->>WS: Broadcast to room participants
```

### 8) Notifications (current DB-level + extensible channels)

```mermaid
sequenceDiagram
  participant S as Services
  participant N as Notification Service
  participant DB as PostgreSQL
  participant U as User

  S->>N: Trigger(notification event)
  N->>DB: Save Notification (respect preferences)
  DB-->>N: OK
  Note over N,U: Future: fan-out via Email/SMS/Push\nbased on UserNotificationPreference
```

### 9) Arabic working schedule constraints

```mermaid
flowchart LR
  A[Task with DueDate] --> B{Working Hours Calc}
  B -->|9:00-16:00, Sat-Thu| C[Effective hours/day = 6]
  C --> D[Optimal start computed]
  B -->|Friday excluded| E[Skip non-working day]
  D --> F[AI Prompts include:\n- Working days left\n- Working hours until due\n- Capacity/Workload context]
```

### 10) Deployment and configuration

```mermaid
flowchart LR
  subgraph Docker
    App[project-x:latest]
  end

  subgraph Env
    K1[GEMINI_API_KEY]
    K2[JWT_SECRET]
    K3[DB_* settings]
  end

  App -->|GORM| PG[(PostgreSQL)]
  App -->|Gemini SDK| Gemini[Google AI]
  Env --> App
```


