# Head Role Permissions - Quick Reference

## ✨ What's New

**Heads can now edit and delete ANY task in projects they're members of!**

---

## Quick Comparison

| Action | Before | After | Condition |
|--------|:------:|:-----:|-----------|
| **View tasks in project** | ✅ | ✅ | Must be project member |
| **Create task** | ✅ | ✅ | Can assign to employees |
| **Update own task status** | ✅ | ✅ | Own tasks only |
| **Edit own task details** | ❌ | ❌ | Still restricted |
| **Edit ANY task in project** | ❌ | ✅ ✨ | Must be project member |
| **Delete ANY task in project** | ❌ | ✅ ✨ | Must be project member |
| **Reassign tasks** | ❌ | ❌ | Still Manager+ only |
| **Edit tasks outside project** | ❌ | ❌ | Not allowed |

---

## Rules for Heads

### ✅ Heads CAN:
1. Edit **any task** in projects they're members of
2. Delete **any task** in projects they're members of
3. Edit title, description, due date, start/end time
4. Edit both regular tasks and collaborative tasks
5. Edit tasks assigned to anyone (in their projects)

### ❌ Heads CANNOT:
1. Edit tasks outside their projects
2. Edit standalone tasks (no project_id)
3. Reassign tasks to other users
4. Create projects
5. Delete projects
6. Add/remove project members
7. Change task priority (Manager+ only)

---

## Examples

### ✅ Allowed Scenario 1
```
Head: John
Project: "Mobile App" (John is member)
Task: "Fix login bug" (assigned to Sarah, in "Mobile App")

Action: John can edit this task ✅
Reason: Task is in his project
```

### ✅ Allowed Scenario 2
```
Head: John
Project: "Mobile App" (John is member)
Task: "Update API" (assigned to Mike, in "Mobile App")

Action: John can delete this task ✅
Reason: Task is in his project
```

### ❌ Denied Scenario 1
```
Head: John
Project: "Web Portal" (John is NOT member)
Task: "Design homepage" (in "Web Portal")

Action: John tries to edit this task ❌
Error: "You can only edit tasks in projects you're a member of"
```

### ❌ Denied Scenario 2
```
Head: John
Task: "Personal task" (no project_id)

Action: John tries to delete this task ❌
Error: "Heads can only delete tasks that belong to a project"
```

---

## Frontend UI Changes

### Task Cards - What Heads See Now:

**In Their Projects:**
```
┌─────────────────────────────┐
│ Task: Implement feature X   │
│ Assigned to: Sarah          │
│ Project: Mobile App         │
│                             │
│ [View] [Edit] [Delete]      │  ← Edit & Delete visible ✨
└─────────────────────────────┘
```

**Outside Their Projects:**
```
┌─────────────────────────────┐
│ Task: Design homepage       │
│ Assigned to: Mike           │
│ Project: Web Portal         │
│                             │
│ [View]                      │  ← Only View button
└─────────────────────────────┘
```

### Right-Click Menu:

**In Their Projects:**
```
┌─────────────────────┐
│ View Details        │
│ Edit Task          │  ← NEW ✨
│ Delete Task        │  ← NEW ✨
│ Update Status      │
│ Add Comment        │
└─────────────────────┘
```

**Outside Their Projects:**
```
┌─────────────────────┐
│ View Details        │
│ Add Comment        │
└─────────────────────┘
```

---

## API Responses

### Success Response
```json
{
  "message": "Task updated successfully"
}
```

### Error Responses

**Not in Project:**
```json
{
  "error": "You can only edit tasks in projects you're a member of"
}
```

**No Project ID:**
```json
{
  "error": "Heads can only edit tasks that belong to a project"
}
```

---

## Code Snippet for Frontend

### Check if Head Can Edit Task

```typescript
const canHeadEditTask = (task: Task, user: User, userProjects: Project[]) => {
  // Only for Heads
  if (user.role !== 'head') return false;
  
  // Task must have a project
  if (!task.project_id) return false;
  
  // Head must be member of that project
  return userProjects.some(project => project.id === task.project_id);
};

// Usage
const TaskCard = ({ task }) => {
  const { user, userProjects } = useAuth();
  
  const canEdit = 
    user.role === 'admin' ||
    user.role === 'manager' ||
    canHeadEditTask(task, user, userProjects);
  
  return (
    <Card>
      {/* ... */}
      {canEdit && <Button onClick={handleEdit}>Edit</Button>}
    </Card>
  );
};
```

---

## Testing Quick Checklist

- [ ] Head can edit task in their project ✅
- [ ] Head can delete task in their project ✅
- [ ] Head gets error for task outside project ❌
- [ ] Head gets error for standalone task ❌
- [ ] Edit button shows for Heads in their projects
- [ ] Delete button shows for Heads in their projects
- [ ] Context menu shows edit/delete for Heads
- [ ] Error messages display correctly

---

## Impact Summary

### Positive Changes:
✅ More autonomy for team leads  
✅ Faster task management  
✅ Reduced dependency on managers  
✅ Better project control  

### Security Maintained:
🔒 Only project-based access  
🔒 Cannot affect other projects  
🔒 Cannot reassign tasks  
🔒 Cannot delete standalone tasks  

---

**Remember: Heads are project-level managers now. They can manage ALL tasks within their projects, but nothing outside their scope.**

