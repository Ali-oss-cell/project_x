# Head Role Permission Update

## 🎉 New Feature: Enhanced Head Permissions

**Date:** Updated permissions for Head role  
**Impact:** Frontend & Backend changes

---

## What Changed

### Before ❌
Heads could only:
- View tasks in their projects
- Update status of their own tasks only
- Create new tasks and assign to employees

### After ✅
Heads can now:
- **Edit ANY task** in projects they're members of
- **Delete ANY task** in projects they're members of
- Update status of their own tasks
- Create new tasks and assign to employees

---

## Technical Implementation

### Backend Changes

#### 1. Updated `UpdateTask` Handler (`handlers/task_handler.go`)

**New Logic:**
```go
// Check permissions: Admin/Manager can update any task
// Head can update tasks in their projects
// Employees and HR cannot update task details
if currentUserRole != models.RoleAdmin && 
   currentUserRole != models.RoleManager && 
   currentUserRole != models.RoleHead {
    return error
}

// If user is a Head, check if task belongs to a project they're in
if currentUserRole == models.RoleHead {
    if task.ProjectID == nil {
        return error("Heads can only edit tasks that belong to a project")
    }

    // Check if the Head is a member of the project
    var userProject models.UserProject
    err := db.Where("project_id = ? AND user_id = ?", 
                    task.ProjectID, userID).First(&userProject).Error
    if err != nil {
        return error("You can only edit tasks in projects you're a member of")
    }
}
```

#### 2. Updated `DeleteTask` Handler (`handlers/task_handler.go`)

**New Logic:**
```go
// Admin/Manager can delete any task
// Head can delete tasks in their projects
// Others can only delete their own tasks
if userRole == models.RoleHead {
    // Head must be in the project to delete the task
    if task.ProjectID == nil {
        return error("Heads can only delete tasks that belong to a project")
    }

    var userProject models.UserProject
    err := db.Where("project_id = ? AND user_id = ?", 
                    task.ProjectID, userID).First(&userProject).Error
    if err != nil {
        return error("You can only delete tasks in projects you're a member of")
    }

    // Head has permission - proceed with deletion
    db.Delete(&task)
}
```

#### 3. Updated `DeleteCollaborativeTask` Handler

Same logic applied to collaborative tasks.

---

## Frontend Changes

### UI Components to Update

#### 1. Task Card Component

**Add Conditional Edit/Delete Buttons for Heads:**

```tsx
// Before
const TaskCard = ({ task, user }) => {
  const canEdit = user.role === 'admin' || user.role === 'manager';
  const canDelete = user.role === 'admin' || user.role === 'manager';
  
  return (
    <Card>
      {/* ... task content ... */}
      {canEdit && <Button onClick={handleEdit}>Edit</Button>}
      {canDelete && <Button onClick={handleDelete}>Delete</Button>}
    </Card>
  );
};
```

```tsx
// After
const TaskCard = ({ task, user, userProjects }) => {
  const isInUserProject = task.project_id && 
                          userProjects.some(p => p.id === task.project_id);
  
  const canEdit = user.role === 'admin' || 
                  user.role === 'manager' ||
                  (user.role === 'head' && isInUserProject);
  
  const canDelete = user.role === 'admin' || 
                    user.role === 'manager' ||
                    (user.role === 'head' && isInUserProject);
  
  return (
    <Card>
      {/* ... task content ... */}
      {canEdit && <Button onClick={handleEdit}>Edit</Button>}
      {canDelete && <Button onClick={handleDelete}>Delete</Button>}
    </Card>
  );
};
```

#### 2. Task List Page

**Show Edit Icons for Heads:**

```tsx
const TaskList = ({ tasks, user }) => {
  // Fetch user's projects
  const { data: userProjects } = useQuery('userProjects', fetchUserProjects);
  
  const showEditIcon = (task) => {
    if (user.role === 'admin' || user.role === 'manager') {
      return true;
    }
    
    // NEW: Show for Heads if task is in their project
    if (user.role === 'head') {
      return task.project_id && 
             userProjects?.some(p => p.id === task.project_id);
    }
    
    return false;
  };
  
  return (
    <div>
      {tasks.map(task => (
        <TaskCard 
          key={task.id} 
          task={task}
          showEditIcon={showEditIcon(task)}
        />
      ))}
    </div>
  );
};
```

#### 3. Context Menu (Right-Click)

**Update Menu Options:**

```tsx
const getContextMenuOptions = (task, user, userProjects) => {
  const isInUserProject = task.project_id && 
                          userProjects.some(p => p.id === task.project_id);
  
  const options = [
    { label: 'View Details', action: 'view', always: true }
  ];
  
  // Edit and Delete for Admin/Manager
  if (user.role === 'admin' || user.role === 'manager') {
    options.push(
      { label: 'Edit Task', action: 'edit' },
      { label: 'Delete Task', action: 'delete' }
    );
  }
  
  // NEW: Edit and Delete for Heads in their projects
  if (user.role === 'head' && isInUserProject) {
    options.push(
      { label: 'Edit Task', action: 'edit' },
      { label: 'Delete Task', action: 'delete' }
    );
  }
  
  return options;
};
```

#### 4. Task Detail Modal

**Update Edit Button Visibility:**

```tsx
const TaskDetailModal = ({ task, user, userProjects }) => {
  const isInUserProject = task.project_id && 
                          userProjects.some(p => p.id === task.project_id);
  
  const canEdit = user.role === 'admin' || 
                  user.role === 'manager' ||
                  (user.role === 'head' && isInUserProject);
  
  return (
    <Modal>
      <ModalHeader>
        <Title>{task.title}</Title>
        {canEdit && (
          <Button onClick={handleEdit}>
            <Icon name="edit" /> Edit Task
          </Button>
        )}
      </ModalHeader>
      {/* ... rest of modal ... */}
    </Modal>
  );
};
```

---

## API Endpoint Updates

### PUT /api/tasks/:id
**Before:** Admin/Manager only  
**After:** Admin/Manager/Head (if in project)

**Request:**
```json
{
  "title": "Updated title",
  "description": "Updated description",
  "due_date": "2024-01-30T17:00:00Z"
}
```

**Response for Head not in project:**
```json
{
  "error": "You can only edit tasks in projects you're a member of"
}
```

### DELETE /api/tasks/:id
**Before:** Admin/Manager only  
**After:** Admin/Manager/Head (if in project)

**Response for Head not in project:**
```json
{
  "error": "You can only delete tasks in projects you're a member of"
}
```

---

## Validation Rules

### For Heads to Edit/Delete Tasks:

1. ✅ User must have `head` role
2. ✅ Task must belong to a project (not standalone)
3. ✅ Head must be a member of that project (in `user_projects` table)

### Validation Flow:

```
1. Check if user is Head
   ↓
2. Check if task has project_id
   ↓ (if null → deny)
3. Query user_projects table
   WHERE project_id = task.project_id
   AND user_id = head.id
   ↓ (if not found → deny)
4. Allow edit/delete
```

---

## Frontend Implementation Checklist

- [ ] Update TaskCard component to show edit/delete for Heads
- [ ] Update TaskList to show edit icons for Heads
- [ ] Update context menu options for Heads
- [ ] Update Task Detail Modal edit button visibility
- [ ] Update permission checks in task actions
- [ ] Fetch user's projects list when logged in as Head
- [ ] Cache user's projects in global state/context
- [ ] Update tooltips (show "Must be in project" for disabled buttons)
- [ ] Test edit functionality for Heads
- [ ] Test delete functionality for Heads
- [ ] Test error handling (not in project)

---

## Testing Scenarios

### Test Case 1: Head Edits Task in Their Project
1. Login as Head user
2. Head is member of "Project X"
3. Navigate to task in "Project X"
4. Edit button should be **visible**
5. Click edit, modify task
6. Should succeed ✅

### Test Case 2: Head Tries to Edit Task Outside Their Project
1. Login as Head user
2. Head is NOT member of "Project Y"
3. Try to edit task in "Project Y"
4. Should show error: "You can only edit tasks in projects you're a member of" ❌

### Test Case 3: Head Deletes Task in Their Project
1. Login as Head user
2. Head is member of "Project X"
3. Navigate to task in "Project X"
4. Delete button should be **visible**
5. Click delete, confirm
6. Should succeed ✅

### Test Case 4: Head Tries to Delete Standalone Task
1. Login as Head user
2. Try to delete task with no project_id
3. Should show error: "Heads can only delete tasks that belong to a project" ❌

### Test Case 5: Head Edits Someone Else's Task in Same Project
1. Login as Head user
2. Head is member of "Project X"
3. Navigate to task assigned to Employee in "Project X"
4. Edit button should be **visible**
5. Click edit, modify task
6. Should succeed ✅ (This is the new feature!)

---

## Benefits

### For Teams:
- ✅ Heads have more autonomy to manage project tasks
- ✅ Reduces bottleneck (don't always need Manager approval)
- ✅ Faster task updates and corrections
- ✅ Better project management at team level

### For Heads:
- ✅ Can fix mistakes in task details immediately
- ✅ Can update tasks when team members are unavailable
- ✅ More control over project execution
- ✅ Can delete duplicate or incorrect tasks

### For System:
- ✅ Better delegation of responsibilities
- ✅ Clearer role boundaries (project-based)
- ✅ Maintains security (only project tasks)

---

## Important Notes

⚠️ **Heads can ONLY edit/delete tasks in projects they're members of**  
⚠️ **Heads CANNOT edit/delete standalone tasks (tasks without project_id)**  
⚠️ **Heads still CANNOT reassign tasks to other users (Manager+ only)**  
⚠️ **Heads still CANNOT create projects (Manager+ only)**  

---

## Summary

| Permission | Before | After |
|------------|--------|-------|
| Edit own tasks | ✅ (status only) | ✅ (status only) |
| Edit team tasks (in project) | ❌ | ✅ **NEW** |
| Edit tasks (no project) | ❌ | ❌ |
| Delete own tasks | ❌ | ❌ |
| Delete team tasks (in project) | ❌ | ✅ **NEW** |
| Delete tasks (no project) | ❌ | ❌ |
| Reassign tasks | ❌ | ❌ |
| Create projects | ❌ | ❌ |

---

**This update gives Heads more responsibility while maintaining proper access control through project membership.**

