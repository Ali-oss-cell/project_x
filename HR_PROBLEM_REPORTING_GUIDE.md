# HR Problem Reporting System

## Overview
The HR Problem Reporting System allows employees to report workplace issues directly to HR with a structured form, predefined categories, and real-time notifications. The system supports both anonymous and identified reporting with comprehensive tracking and resolution management.

## Features
- **Predefined Problem Categories** - 13 workplace issue categories
- **Priority Levels** - From low to critical priority
- **Anonymous Reporting** - Optional anonymous submissions
- **Real-time Notifications** - Instant notifications to HR team
- **Status Tracking** - Complete problem lifecycle management
- **Comments System** - Two-way communication between reporter and HR
- **HR Dashboard** - Statistics and management tools for HR
- **Resolution Tracking** - Final resolution documentation

## Problem Categories

### Available Categories:
1. **Harassment** - Harassment or bullying in the workplace
2. **Discrimination** - Discrimination based on race, gender, age, religion, etc.
3. **Work Environment** - Issues with workplace conditions or environment
4. **Management Issues** - Problems with management or supervision
5. **Workload** - Excessive workload or unrealistic expectations
6. **Safety Concerns** - Safety concerns or unsafe working conditions
7. **Payroll Issues** - Payroll, salary, or compensation issues
8. **Benefits** - Problems with benefits, insurance, or leave
9. **Workplace Conflict** - Conflicts with colleagues or team members
10. **Policy Violation** - Company policy violations or concerns
11. **Equipment Issues** - Equipment, tools, or technology issues
12. **Training Needs** - Training needs or development concerns
13. **Other** - Other workplace issues not listed above

### Priority Levels:
- **Low** - Can be addressed in normal timeframe
- **Medium** - Should be addressed within a week
- **High** - Needs attention within 2-3 days
- **Urgent** - Requires immediate attention within 24 hours
- **Critical** - Emergency situation requiring immediate action

## API Endpoints

### Public Endpoints (All Authenticated Users)

#### Get Problem Categories
```http
GET /api/hr-problems/categories
```
Returns all available problem categories and priority levels.

**Response:**
```json
{
  "categories": {
    "harassment": "Harassment or bullying in the workplace",
    "discrimination": "Discrimination based on race, gender, age, religion, etc.",
    // ... other categories
  },
  "priorities": {
    "low": "Low priority - can be addressed in normal timeframe",
    "medium": "Medium priority - should be addressed within a week",
    // ... other priorities
  }
}
```

#### Report a Problem
```http
POST /api/hr-problems
```

**Request Body:**
```json
{
  "title": "Workplace Safety Concern",
  "description": "Detailed description of the problem...",
  "category": "safety_concerns",
  "priority": "high",
  "is_anonymous": false,
  "contact_method": "email",
  "phone_number": "+1234567890",
  "preferred_time": "9 AM - 5 PM",
  "witness_info": "John Doe witnessed the incident",
  "location": "Factory Floor A",
  "incident_date": "2024-01-15T10:30:00Z",
  "previous_reports": false
}
```

**Response:**
```json
{
  "message": "Problem reported successfully",
  "problem": {
    "id": 1,
    "title": "Workplace Safety Concern",
    "category": "safety_concerns",
    "priority": "high",
    "status": "pending",
    "reported_at": "2024-01-15T14:30:00Z",
    "is_urgent": true
  }
}
```

#### Get My Problems
```http
GET /api/hr-problems/my-problems?page=1&limit=20
```
Returns all problems reported by the current user.

#### Get Specific Problem
```http
GET /api/hr-problems/{id}
```
Users can only see their own problems; HR/Admin can see all problems.

#### Add Comment
```http
POST /api/hr-problems/{id}/comments
```

**Request Body:**
```json
{
  "comment": "Additional information about the incident",
  "is_hr_only": false
}
```

### HR/Admin Only Endpoints

#### Get All Problems
```http
GET /api/hr-problems?page=1&limit=20&status=pending&category=harassment&priority=high
```
Returns all problems with optional filtering.

#### Get Assigned Problems
```http
GET /api/hr-problems/assigned?page=1&limit=20
```
Returns problems assigned to the current HR user.

#### Update Problem Status
```http
PATCH /api/hr-problems/{id}/status
```

**Request Body:**
```json
{
  "status": "reviewing",
  "note": "Started investigation process"
}
```

#### Assign Problem to HR User
```http
PATCH /api/hr-problems/{id}/assign
```

**Request Body:**
```json
{
  "hr_user_id": 5
}
```

#### Update HR Notes
```http
PATCH /api/hr-problems/{id}/notes
```

**Request Body:**
```json
{
  "notes": "Internal HR notes about the case..."
}
```

#### Set Resolution
```http
PATCH /api/hr-problems/{id}/resolution
```

**Request Body:**
```json
{
  "resolution": "Issue resolved through mediation and policy clarification"
}
```

#### Get Statistics
```http
GET /api/hr-problems/statistics
```

**Response:**
```json
{
  "statistics": {
    "total_problems": 45,
    "by_status": {
      "pending": 5,
      "reviewing": 8,
      "in_progress": 12,
      "resolved": 18,
      "closed": 2
    },
    "by_category": {
      "harassment": 3,
      "safety_concerns": 8,
      "workload": 12
    },
    "by_priority": {
      "low": 10,
      "medium": 20,
      "high": 12,
      "urgent": 3
    },
    "urgent_open": 3,
    "recent_problems": 8
  }
}
```

## Problem Status Workflow

1. **Pending** - Initial status when problem is reported
2. **Reviewing** - HR is reviewing the problem
3. **In Progress** - HR is actively working on the problem
4. **Resolved** - Problem has been resolved
5. **Closed** - Problem is closed (no further action needed)
6. **Rejected** - Problem was rejected (invalid/duplicate)

## Notification System

### For HR Users:
- **New Problem Report** - Immediate notification when a problem is reported
- **Urgent Problems** - Special urgent notifications for high/critical priority issues
- **Problem Assignment** - Notification when assigned to handle a problem

### For Reporters:
- **Status Updates** - Notification when problem status changes
- **Comments** - Notification when HR adds non-private comments
- **Resolution** - Notification when problem is resolved

## Usage Examples

### JavaScript Client Example

```javascript
// Report a problem
async function reportProblem() {
  const response = await fetch('/api/hr-problems', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    },
    body: JSON.stringify({
      title: 'Workplace Harassment',
      description: 'Detailed description of the harassment incident...',
      category: 'harassment',
      priority: 'high',
      is_anonymous: false,
      contact_method: 'email',
      location: 'Office Building A',
      incident_date: new Date().toISOString(),
      previous_reports: false
    })
  });
  
  const result = await response.json();
  console.log('Problem reported:', result);
}

// Get problem categories for form
async function loadCategories() {
  const response = await fetch('/api/hr-problems/categories', {
    headers: {
      'Authorization': 'Bearer ' + token
    }
  });
  
  const data = await response.json();
  
  // Populate category dropdown
  const categorySelect = document.getElementById('category');
  Object.entries(data.categories).forEach(([key, description]) => {
    const option = document.createElement('option');
    option.value = key;
    option.textContent = description;
    categorySelect.appendChild(option);
  });
}

// HR: Update problem status
async function updateProblemStatus(problemId, status, note) {
  const response = await fetch(`/api/hr-problems/${problemId}/status`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    },
    body: JSON.stringify({
      status: status,
      note: note
    })
  });
  
  const result = await response.json();
  console.log('Status updated:', result);
}
```

### HTML Form Example

```html
<!DOCTYPE html>
<html>
<head>
    <title>Report Workplace Problem</title>
    <style>
        .form-container { max-width: 600px; margin: 20px auto; padding: 20px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input, select, textarea { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        textarea { height: 100px; resize: vertical; }
        .urgent { color: red; font-weight: bold; }
        .submit-btn { background: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="form-container">
        <h2>Report Workplace Problem</h2>
        <p>Your report will be sent directly to HR. All information is confidential.</p>
        
        <form id="problemForm">
            <div class="form-group">
                <label for="title">Problem Title *</label>
                <input type="text" id="title" name="title" required>
            </div>
            
            <div class="form-group">
                <label for="category">Category *</label>
                <select id="category" name="category" required>
                    <option value="">Select a category...</option>
                </select>
            </div>
            
            <div class="form-group">
                <label for="priority">Priority</label>
                <select id="priority" name="priority">
                    <option value="medium">Medium (default)</option>
                    <option value="low">Low</option>
                    <option value="high">High</option>
                    <option value="urgent" class="urgent">Urgent (24 hours)</option>
                    <option value="critical" class="urgent">Critical (Emergency)</option>
                </select>
            </div>
            
            <div class="form-group">
                <label for="description">Description *</label>
                <textarea id="description" name="description" required 
                          placeholder="Please provide detailed information about the problem..."></textarea>
            </div>
            
            <div class="form-group">
                <label for="location">Location</label>
                <input type="text" id="location" name="location" 
                       placeholder="Where did this occur?">
            </div>
            
            <div class="form-group">
                <label for="incidentDate">Incident Date</label>
                <input type="datetime-local" id="incidentDate" name="incident_date">
            </div>
            
            <div class="form-group">
                <label for="witnessInfo">Witness Information</label>
                <textarea id="witnessInfo" name="witness_info" 
                          placeholder="Names and contact info of any witnesses..."></textarea>
            </div>
            
            <div class="form-group">
                <label for="contactMethod">Preferred Contact Method</label>
                <select id="contactMethod" name="contact_method">
                    <option value="email">Email</option>
                    <option value="phone">Phone</option>
                    <option value="in_person">In Person</option>
                </select>
            </div>
            
            <div class="form-group">
                <label for="phoneNumber">Phone Number (for urgent issues)</label>
                <input type="tel" id="phoneNumber" name="phone_number">
            </div>
            
            <div class="form-group">
                <label for="preferredTime">Preferred Contact Time</label>
                <input type="text" id="preferredTime" name="preferred_time" 
                       placeholder="e.g., 9 AM - 5 PM weekdays">
            </div>
            
            <div class="form-group">
                <label>
                    <input type="checkbox" id="isAnonymous" name="is_anonymous">
                    Submit anonymously (HR won't know who reported this)
                </label>
            </div>
            
            <div class="form-group">
                <label>
                    <input type="checkbox" id="previousReports" name="previous_reports">
                    This issue has been reported before
                </label>
            </div>
            
            <button type="submit" class="submit-btn">Submit Report</button>
        </form>
    </div>

    <script>
        // Load categories when page loads
        document.addEventListener('DOMContentLoaded', loadCategories);
        
        // Handle form submission
        document.getElementById('problemForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const formData = new FormData(e.target);
            const data = Object.fromEntries(formData.entries());
            
            // Convert checkboxes to boolean
            data.is_anonymous = document.getElementById('isAnonymous').checked;
            data.previous_reports = document.getElementById('previousReports').checked;
            
            // Convert datetime-local to ISO string
            if (data.incident_date) {
                data.incident_date = new Date(data.incident_date).toISOString();
            }
            
            try {
                await reportProblem(data);
                alert('Problem reported successfully! You will receive updates via your preferred contact method.');
                e.target.reset();
            } catch (error) {
                alert('Error reporting problem: ' + error.message);
            }
        });
    </script>
</body>
</html>
```

## Security and Privacy

### Data Protection:
- All problem reports are encrypted in transit and at rest
- Anonymous reports cannot be traced back to the reporter
- HR-only comments are never visible to reporters
- Access is strictly controlled by role-based permissions

### Privacy Features:
- **Anonymous Reporting** - Complete anonymity option
- **Confidential Comments** - HR can add internal notes
- **Secure Access** - Only authorized HR personnel can access reports
- **Audit Trail** - Complete history of all actions and updates

## Best Practices

### For Users:
1. **Be Specific** - Provide detailed descriptions and relevant information
2. **Choose Correct Priority** - Use urgent/critical only for true emergencies
3. **Include Evidence** - Mention witnesses, dates, and locations
4. **Follow Up** - Check status and respond to HR comments

### For HR:
1. **Respond Quickly** - Acknowledge receipt within 24 hours
2. **Update Status** - Keep reporters informed of progress
3. **Document Everything** - Use HR notes for internal tracking
4. **Maintain Confidentiality** - Protect reporter identity and information
5. **Follow Company Policy** - Ensure all actions comply with HR policies

This HR Problem Reporting System provides a comprehensive solution for workplace issue management with proper security, tracking, and communication features. 