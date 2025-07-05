# Command Line Utilities

## Create Admin User

The `create_admin.go` utility allows you to create admin users for the Project X application.

### Usage

1. Make sure your database is running and your `.env` file is configured with the correct database credentials.

2. Run the utility:
   ```bash
   go run cmd/create_admin.go
   ```

3. Follow the prompts to enter:
   - Username (must be unique)
   - Password (will be hashed securely)
   - Password confirmation
   - Department

### Features

- ✅ Validates that username is unique
- ✅ Securely hashes passwords using bcrypt
- ✅ Confirms password entry
- ✅ Automatically sets role to "admin"
- ✅ Creates user in the database
- ✅ Provides feedback on successful creation

### Example Output

```
=== Admin User Creation Utility ===

Enter username: admin
Enter password: mysecurepassword
Confirm password: mysecurepassword
Enter department: IT

✅ Admin user created successfully!
Username: admin
Role: admin
Department: IT
User ID: 1
```

### Requirements

- Go 1.23 or higher
- PostgreSQL database running
- Proper `.env` configuration
- All required Go dependencies installed

### Security Notes

- Passwords are hashed using bcrypt with default cost
- Username uniqueness is enforced at the database level
- The utility validates all inputs before creating the user 