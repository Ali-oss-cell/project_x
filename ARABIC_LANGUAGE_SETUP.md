# Arabic Language Support Setup Guide

This guide explains how to properly configure your PostgreSQL database and Go application to support Arabic text.

## Database Configuration

### 1. PostgreSQL Database Setup

Ensure your PostgreSQL database is created with UTF-8 encoding:

```sql
-- Create database with UTF-8 encoding
CREATE DATABASE your_database_name 
WITH ENCODING 'UTF8' 
LC_COLLATE='en_US.UTF-8' 
LC_CTYPE='en_US.UTF-8';

-- Or if using Arabic locale (recommended for Arabic text)
CREATE DATABASE your_database_name 
WITH ENCODING 'UTF8' 
LC_COLLATE='ar_SA.UTF-8' 
LC_CTYPE='ar_SA.UTF-8';
```

### 2. Connection String Configuration

The application now includes proper UTF-8 encoding in the connection string:

```go
dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC client_encoding=UTF8",
    config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort)
```

### 3. Model Updates

All string fields in models now explicitly specify UTF-8 support:

- **Text fields**: Use `type:text` for long content (descriptions, messages)
- **Varchar fields**: Use `type:varchar(N) COLLATE "default"` for shorter strings
- **Proper collation**: Ensures correct sorting and comparison of Arabic text

## Testing Arabic Language Support

### 1. Test Arabic Text Input

```go
// Example: Creating a task with Arabic title and description
task := models.Task{
    Title:       "مهمة جديدة",
    Description: "هذه مهمة تحتوي على نص عربي",
    Status:      models.TaskStatusPending,
    UserID:      1,
}
```

### 2. API Testing

Use the following JSON in your API requests:

```json
{
    "title": "مهمة جديدة",
    "description": "هذه مهمة تحتوي على نص عربي مع تفاصيل أكثر",
    "status": "pending"
}
```

### 3. Database Verification

Connect to your PostgreSQL database and verify Arabic text storage:

```sql
-- Check if Arabic text is stored correctly
SELECT id, title, description FROM tasks WHERE title LIKE '%مهمة%';

-- Verify character encoding
SHOW client_encoding;
SHOW server_encoding;
```

## Common Issues and Solutions

### Issue 1: Garbled Text Display
**Solution**: Ensure your client applications (web browsers, mobile apps) use UTF-8 encoding.

### Issue 2: Sorting Issues with Arabic Text
**Solution**: Use proper collation in your database queries:

```sql
SELECT * FROM tasks ORDER BY title COLLATE "ar_SA.UTF-8";
```

### Issue 3: Search Not Working with Arabic
**Solution**: Use PostgreSQL's full-text search with Arabic language support:

```sql
-- Enable Arabic text search
CREATE INDEX idx_tasks_arabic_search ON tasks USING gin(to_tsvector('arabic', title || ' ' || description));

-- Search Arabic text
SELECT * FROM tasks WHERE to_tsvector('arabic', title || ' ' || description) @@ plainto_tsquery('arabic', 'البحث');
```

## Best Practices

1. **Always use UTF-8**: Ensure all layers of your application use UTF-8 encoding
2. **Test thoroughly**: Test with various Arabic text inputs including:
   - Right-to-left text
   - Mixed Arabic and English text
   - Special Arabic characters and diacritics
3. **Backup considerations**: Ensure your backup tools preserve UTF-8 encoding
4. **Client-side handling**: Configure your frontend applications to handle RTL text properly

## Environment Variables

Update your `.env` file to include proper database settings:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_username
DB_PASSWORD=your_password
DB_NAME=your_database_name
DB_CHARSET=utf8
```

## Migration Commands

After updating the models, run the migration:

```bash
# The application will automatically migrate the database on startup
go run main.go
```

If you need to manually recreate tables with proper encoding:

```sql
-- Drop and recreate tables (WARNING: This will delete all data)
DROP TABLE IF EXISTS tasks CASCADE;
DROP TABLE IF EXISTS collaborative_tasks CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;
-- ... other tables

-- Then restart your Go application to recreate tables with proper encoding
```

## Verification Checklist

- [ ] Database created with UTF-8 encoding
- [ ] Connection string includes `client_encoding=UTF8`
- [ ] Models updated with proper field types
- [ ] Application successfully starts and migrates
- [ ] Arabic text can be inserted via API
- [ ] Arabic text displays correctly when retrieved
- [ ] Search functionality works with Arabic text
- [ ] Sorting works correctly with Arabic text

## Additional Resources

- [PostgreSQL Character Set Support](https://www.postgresql.org/docs/current/charset.html)
- [Go UTF-8 String Handling](https://golang.org/pkg/unicode/utf8/)
- [Arabic Language Support in Databases](https://unicode.org/reports/tr9/) 