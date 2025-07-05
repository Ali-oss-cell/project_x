# Railway Deployment Guide

This guide will help you deploy your Project X backend to Railway for testing and development.

## 🚀 Quick Start

### 1. Prerequisites
- Railway account (sign up at [railway.app](https://railway.app))
- Git repository with your code
- Railway CLI (optional but recommended)

### 2. Deploy to Railway

#### Option A: Deploy from GitHub (Recommended)

1. **Connect GitHub Repository**
   - Go to [Railway Dashboard](https://railway.app/dashboard)
   - Click "New Project"
   - Select "Deploy from GitHub repo"
   - Choose your repository

2. **Add PostgreSQL Database**
   - In your Railway project dashboard
   - Click "New Service"
   - Select "Database" → "PostgreSQL"
   - Railway will automatically create a database

3. **Configure Environment Variables**
   - Go to your backend service
   - Click "Variables" tab
   - Add the following variables:

```env
# Database Configuration (Railway will auto-populate these)
DB_HOST=${{ Postgres.PGHOST }}
DB_PORT=${{ Postgres.PGPORT }}
DB_USER=${{ Postgres.PGUSER }}
DB_PASSWORD=${{ Postgres.PGPASSWORD }}
DB_NAME=${{ Postgres.PGDATABASE }}

# JWT Secret (generate a strong random string)
JWT_SECRET=your_super_secret_jwt_key_here_make_it_long_and_random

# Port (Railway will set this automatically)
PORT=${{ PORT }}
```

#### Option B: Deploy via Railway CLI

1. **Install Railway CLI**
```bash
# macOS
brew install railway

# Windows (via npm)
npm install -g @railway/cli

# Linux
curl -fsSL https://railway.app/install.sh | sh
```

2. **Login and Deploy**
```bash
# Login to Railway
railway login

# Initialize project
railway init

# Add PostgreSQL database
railway add postgresql

# Deploy
railway up
```

### 3. Configure Database

Railway will automatically set up PostgreSQL, but you need to configure the environment variables:

1. **Get Database Credentials**
   - Go to your PostgreSQL service in Railway
   - Click "Variables" tab
   - Copy the connection details

2. **Set Environment Variables**
   - Use the template above
   - Railway's reference syntax `${{ Postgres.PGHOST }}` will automatically populate values

### 4. Deploy and Test

1. **Deployment**
   - Railway will automatically build and deploy your application
   - Build process uses the included `Dockerfile`
   - Health checks ensure the service is running

2. **Get Your URL**
   - Go to your service in Railway dashboard
   - Click "Settings" → "Networking"
   - Copy the public URL (e.g., `https://your-app.railway.app`)

3. **Test the Deployment**
```bash
# Test health endpoint
curl https://your-app.railway.app/health

# Expected response:
# {"status":"healthy","message":"Service is running"}
```

## 🔧 Configuration Details

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_HOST` | Database host | `containers-us-west-1.railway.app` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `generated_password` |
| `DB_NAME` | Database name | `railway` |
| `JWT_SECRET` | JWT signing secret | `your_super_secret_key` |
| `PORT` | Server port | `8080` |

### Railway-Specific Features

1. **Automatic SSL/TLS**
   - Railway provides HTTPS automatically
   - No additional configuration needed

2. **Environment-based Configuration**
   - Railway automatically sets `PORT` environment variable
   - Database credentials are auto-populated

3. **Zero-downtime Deployments**
   - Railway handles deployments without downtime
   - Health checks ensure service availability

## 🧪 Testing Your Deployment

### 1. Create Admin User

Use the Railway console to create an admin user:

```bash
# Connect to your Railway service
railway shell

# Run the admin creation utility
go run cmd/create_admin.go
```

### 2. Test API Endpoints

```bash
# Base URL (replace with your Railway URL)
BASE_URL="https://your-app.railway.app"

# Health check
curl $BASE_URL/health

# Register a user
curl -X POST $BASE_URL/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123",
    "role": "employee",
    "department": "Engineering"
  }'

# Login
curl -X POST $BASE_URL/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "testpass123"
  }'
```

### 3. Test WebSocket Connection

```javascript
// Test WebSocket connection
const ws = new WebSocket('wss://your-app.railway.app/ws/chat');

ws.onopen = function() {
    console.log('WebSocket connected');
};

ws.onmessage = function(event) {
    console.log('Message:', event.data);
};
```

## 📊 Monitoring and Logs

### View Logs
```bash
# Using Railway CLI
railway logs

# Or in Railway dashboard
# Go to your service → "Deployments" → Click on a deployment
```

### Monitor Performance
- Railway provides built-in metrics
- CPU, memory, and network usage
- Request/response times
- Error rates

## 🔧 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check environment variables are set correctly
   - Verify PostgreSQL service is running
   - Check database credentials

2. **Build Failed**
   - Check Go version compatibility
   - Verify all dependencies in `go.mod`
   - Check Dockerfile syntax

3. **Health Check Failed**
   - Verify `/health` endpoint is accessible
   - Check database connectivity
   - Review application logs

### Debug Commands

```bash
# Check service status
railway status

# View environment variables
railway variables

# Connect to database
railway connect postgresql

# Run commands in Railway environment
railway run go version
```

## 🎯 Production Considerations

### Security
- Use strong JWT secrets
- Enable CORS properly for your frontend
- Consider rate limiting
- Use HTTPS only (Railway provides this)

### Performance
- Monitor resource usage
- Consider horizontal scaling for high traffic
- Use connection pooling for database
- Implement caching where appropriate

### Backup
- Railway provides automatic backups for PostgreSQL
- Consider additional backup strategies for production

## 📚 API Documentation

Your deployed API will be available at:
- Base URL: `https://your-app.railway.app`
- Health Check: `GET /health`
- Auth: `POST /api/auth/login`, `POST /api/auth/register`
- Tasks: `GET /api/tasks`, `POST /api/tasks`
- Projects: `GET /api/projects`, `POST /api/projects`
- Chat: `GET /api/chat/team-chat`, WebSocket at `/ws/chat`
- HR Problems: `GET /api/hr-problems/categories`, `POST /api/hr-problems`

## 🎉 Next Steps

1. **Frontend Integration**
   - Update your frontend to use the Railway URL
   - Configure CORS if needed
   - Test all API endpoints

2. **Custom Domain** (Optional)
   - Add custom domain in Railway settings
   - Configure DNS records
   - SSL certificate is automatic

3. **Team Access**
   - Invite team members to Railway project
   - Set up different environments (dev, staging, prod)
   - Configure deployment triggers

## 💡 Tips

- Use Railway's environment variables for sensitive data
- Monitor your usage to avoid unexpected costs
- Use Railway's built-in PostgreSQL for development
- Consider upgrading to Railway Pro for production use
- Set up deployment notifications
- Use Railway's preview deployments for testing

---

**Happy Deploying! 🚀**

Your Project X backend is now ready for testing on Railway. The platform provides excellent developer experience with automatic deployments, built-in databases, and comprehensive monitoring. 