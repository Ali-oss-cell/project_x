# Login API Reference - Updated Response

## 🔐 Login Endpoint

**Endpoint:** `POST /api/auth/login`

**Request:**
```json
{
  "username": "string",
  "password": "string"
}
```

## ✅ New Response Format

The login endpoint now returns both the token and user information:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "role": "admin",
    "department": "IT",
    "skills": "[\"Python\", \"Go\", \"PostgreSQL\"]",
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

## 📋 Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `token` | string | JWT token for authentication |
| `user.id` | number | User ID |
| `user.username` | string | Username |
| `user.role` | string | User role: `admin`, `manager`, `head`, `employee`, or `hr` |
| `user.department` | string | User's department |
| `user.skills` | string | JSON string array of skills |
| `user.created_at` | string | Account creation timestamp (ISO 8601) |

## 💻 Frontend Implementation

### TypeScript/TSX Example

```typescript
// types/auth.ts
export interface LoginResponse {
  token: string;
  user: {
    id: number;
    username: string;
    role: 'admin' | 'manager' | 'head' | 'employee' | 'hr';
    department: string;
    skills: string; // JSON string - parse it
    created_at: string;
  };
}

// Login component
const handleLogin = async (username: string, password: string) => {
  try {
    const response = await axios.post<LoginResponse>(
      `${API_BASE_URL}/auth/login`,
      { username, password }
    );

    const { token, user } = response.data;

    // Store token
    localStorage.setItem('token', token);
    
    // Store user info (no need to decode token!)
    localStorage.setItem('user', JSON.stringify(user));
    
    // Parse skills if needed
    const skills = user.skills ? JSON.parse(user.skills) : [];
    
    // Redirect based on role
    redirectBasedOnRole(user.role);
    
    return { token, user };
  } catch (error) {
    console.error('Login failed:', error);
    throw error;
  }
};
```

### React Hook Example

```typescript
// hooks/useAuth.ts
import { useState, useEffect } from 'react';

interface User {
  id: number;
  username: string;
  role: 'admin' | 'manager' | 'head' | 'employee' | 'hr';
  department: string;
  skills: string[];
  created_at: string;
}

export const useAuth = () => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Load user from localStorage on mount
    const storedUser = localStorage.getItem('user');
    const token = localStorage.getItem('token');
    
    if (storedUser && token) {
      try {
        const userData = JSON.parse(storedUser);
        // Parse skills JSON string
        userData.skills = userData.skills ? JSON.parse(userData.skills) : [];
        setUser(userData);
      } catch (error) {
        console.error('Failed to parse user data:', error);
        localStorage.removeItem('user');
        localStorage.removeItem('token');
      }
    }
    setLoading(false);
  }, []);

  const login = async (username: string, password: string) => {
    const response = await axios.post<LoginResponse>(
      '/api/auth/login',
      { username, password }
    );
    
    const { token, user: userData } = response.data;
    
    // Parse skills
    const user = {
      ...userData,
      skills: userData.skills ? JSON.parse(userData.skills) : [],
    };
    
    localStorage.setItem('token', token);
    localStorage.setItem('user', JSON.stringify(user));
    setUser(user);
    
    return { token, user };
  };

  const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    setUser(null);
  };

  const hasRole = (roles: string[]) => {
    return user ? roles.includes(user.role) : false;
  };

  return {
    user,
    loading,
    isAuthenticated: !!user,
    login,
    logout,
    hasRole,
    isAdmin: () => hasRole(['admin']),
    isManagerOrHigher: () => hasRole(['admin', 'manager']),
    isHeadOrHigher: () => hasRole(['admin', 'manager', 'head']),
    isHR: () => hasRole(['hr']),
  };
};
```

### Complete Login Component

```tsx
// Login.tsx
import React, { useState } from 'react';
import { useAuth } from '../hooks/useAuth';
import { useNavigate } from 'react-router-dom';

export const Login: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const { user } = await login(username, password);
      
      // Redirect based on role
      if (user.role === 'admin') {
        navigate('/admin/dashboard');
      } else if (user.role === 'manager') {
        navigate('/manager/dashboard');
      } else if (user.role === 'head') {
        navigate('/head/dashboard');
      } else if (user.role === 'hr') {
        navigate('/hr/dashboard');
      } else {
        navigate('/dashboard');
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        placeholder="Username"
        value={username}
        onChange={(e) => setUsername(e.target.value)}
        required
      />
      <input
        type="password"
        placeholder="Password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        required
      />
      <button type="submit" disabled={loading}>
        {loading ? 'Logging in...' : 'Login'}
      </button>
      {error && <div className="error">{error}</div>}
    </form>
  );
};
```

## 🔄 Migration from Old Format

### Old Code (decoding token):
```typescript
// ❌ Old way - decoding token
const token = response.data.token;
const payload = JSON.parse(atob(token.split('.')[1]));
const role = payload.role; // Only had role, no other user info
```

### New Code (using user object):
```typescript
// ✅ New way - direct user object
const { token, user } = response.data;
const role = user.role; // Full user info available immediately
```

## ✅ Benefits

1. **No token decoding needed** - User info is directly available
2. **Full user data** - Get username, department, skills immediately
3. **Single request** - Everything in one response
4. **Better UX** - Can show user name, department right away
5. **Type safety** - Clear TypeScript interfaces

## 🧪 Testing

```bash
# Test login endpoint
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password123"}'

# Expected response:
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "role": "admin",
    "department": "IT",
    "skills": "[]",
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

## 📝 Notes

- The `skills` field is a JSON string - parse it with `JSON.parse()` in frontend
- Token still contains `userId` and `role` in JWT claims (for backend validation)
- User info is returned for convenience - backend still validates token signature
- All existing authentication flows remain the same - only response format changed

