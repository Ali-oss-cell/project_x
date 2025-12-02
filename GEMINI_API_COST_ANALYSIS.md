# Google Gemini API Cost Analysis

## Overview
This project uses **Google Gemini 2.0 Flash** for three main AI features:
1. **AI Time Optimizer** - Task risk analysis and time predictions
2. **AI Chat Assistant** - Conversational AI assistant
3. **AI Project Task Generator** - Automatic task generation from project descriptions

---

## Gemini 2.0 Flash Pricing (as of 2024)

### Standard Pricing:
- **Input tokens**: $0.30 per million tokens
- **Output tokens**: $2.50 per million tokens

### Free Tier:
- **15 requests per minute (RPM)**
- **1,500 requests per day (RPD)**
- **32,000 tokens per minute (TPM)**

---

## Feature Breakdown & Token Usage

### 1. AI Time Optimizer

#### When it's called:
- When user requests task analysis (`GET /api/ai/time/analyze/:taskId`)
- When user requests all tasks analysis (`GET /api/ai/time/analyze`)
- When user requests project report (`GET /api/ai/time/project/:projectId`)
- When user requests workload analysis (`GET /api/ai/time/workload`)

#### Token estimation per call:

**Single Task Analysis:**
- **Input**: ~800-1,200 tokens
  - Task details: ~200 tokens
  - Historical data: ~300-500 tokens
  - User context: ~100 tokens
  - Project context: ~200 tokens
  - Prompt template: ~200 tokens
- **Output**: ~300-500 tokens
  - JSON response with analysis
- **Total per call**: ~1,100-1,700 tokens

**All Tasks Analysis:**
- **Input**: ~500 tokens (base) + (800 tokens × number of tasks)
- **Output**: ~300 tokens × number of tasks
- **Example (10 tasks)**: ~8,500 input + ~3,000 output = ~11,500 tokens

**Project Report:**
- **Input**: ~1,500-2,500 tokens
  - All task analyses: ~1,000-2,000 tokens
  - Project summary: ~300 tokens
  - Prompt: ~200 tokens
- **Output**: ~800-1,200 tokens
- **Total per call**: ~2,300-3,700 tokens

**Workload Analysis:**
- **Input**: ~1,000-1,500 tokens
- **Output**: ~500-800 tokens
- **Total per call**: ~1,500-2,300 tokens

#### Caching:
- ✅ **1-hour cache** for individual task analysis
- Reduces API calls by ~70-80% for repeated requests

#### Estimated monthly usage (small team: 5 users, 50 active tasks):
- Task analyses: 200 calls/month (with cache)
- Project reports: 20 calls/month
- Workload analyses: 30 calls/month
- **Total tokens**: ~250,000 input + ~100,000 output = ~350,000 tokens/month

---

### 2. AI Chat Assistant

#### When it's called:
- Every message in AI chat room
- Every `@ai` mention in team chat
- Commands (create task, get analytics, etc.)

#### Token estimation per call:

**Simple Question:**
- **Input**: ~500-1,000 tokens
  - User message: ~50-200 tokens
  - Chat history (last 10 messages): ~300-600 tokens
  - User context: ~100 tokens
  - System prompt: ~50 tokens
- **Output**: ~100-300 tokens
- **Total per call**: ~600-1,300 tokens

**Complex Question with Context:**
- **Input**: ~1,500-2,500 tokens
  - User message: ~200-500 tokens
  - Extended chat history: ~800-1,500 tokens
  - User tasks/projects: ~300-500 tokens
  - System prompt: ~200 tokens
- **Output**: ~300-600 tokens
- **Total per call**: ~1,800-3,100 tokens

**Command (create task, analytics):**
- **Input**: ~800-1,500 tokens
- **Output**: ~200-400 tokens
- **Total per call**: ~1,000-1,900 tokens

#### Estimated monthly usage (active team: 10 users):
- Messages per user per day: 5-10
- Total messages: 1,500-3,000/month
- **Total tokens**: ~1,500,000 input + ~600,000 output = ~2,100,000 tokens/month

---

### 3. AI Project Task Generator

#### When it's called:
- When manager generates tasks for a project (`POST /api/projects/:id/generate-tasks`)
- Only called once per project (or when regenerating)

#### Token estimation per call:

**Task Generation (5 team members, 15 tasks):**
- **Input**: ~3,000-5,000 tokens
  - Project description: ~200-500 tokens
  - Team member info (with historical data): ~400-800 tokens per member
  - Working schedule context: ~200 tokens
  - Prompt instructions: ~500 tokens
- **Output**: ~2,000-4,000 tokens
  - JSON with 10-20 tasks: ~1,500-3,500 tokens
  - Summary: ~200-300 tokens
- **Total per call**: ~5,000-9,000 tokens

**Task Generation (10 team members, 30 tasks):**
- **Input**: ~6,000-10,000 tokens
- **Output**: ~4,000-7,000 tokens
- **Total per call**: ~10,000-17,000 tokens

#### Estimated monthly usage:
- New projects per month: 5-10
- Regenerations: 2-5
- **Total calls**: 7-15/month
- **Total tokens**: ~70,000 input + ~50,000 output = ~120,000 tokens/month

---

## Total Monthly Cost Estimation

### Scenario 1: Small Team (5-10 users, 50 tasks, 5 projects/month)

| Feature | Input Tokens | Output Tokens | Cost (Input) | Cost (Output) | Total Cost |
|---------|--------------|---------------|--------------|---------------|------------|
| AI Time Optimizer | 250,000 | 100,000 | $0.075 | $0.25 | **$0.33** |
| AI Chat Assistant | 1,500,000 | 600,000 | $0.45 | $1.50 | **$1.95** |
| AI Task Generator | 70,000 | 50,000 | $0.021 | $0.125 | **$0.15** |
| **TOTAL** | **1,820,000** | **750,000** | **$0.55** | **$1.88** | **~$2.43/month** |

### Scenario 2: Medium Team (20-30 users, 200 tasks, 15 projects/month)

| Feature | Input Tokens | Output Tokens | Cost (Input) | Cost (Output) | Total Cost |
|---------|--------------|---------------|--------------|---------------|------------|
| AI Time Optimizer | 1,000,000 | 400,000 | $0.30 | $1.00 | **$1.30** |
| AI Chat Assistant | 4,500,000 | 1,800,000 | $1.35 | $4.50 | **$5.85** |
| AI Task Generator | 200,000 | 150,000 | $0.06 | $0.375 | **$0.44** |
| **TOTAL** | **5,700,000** | **2,350,000** | **$1.71** | **$5.88** | **~$7.59/month** |

### Scenario 3: Large Team (50+ users, 500+ tasks, 30 projects/month)

| Feature | Input Tokens | Output Tokens | Cost (Input) | Cost (Output) | Total Cost |
|---------|--------------|---------------|--------------|---------------|------------|
| AI Time Optimizer | 2,500,000 | 1,000,000 | $0.75 | $2.50 | **$3.25** |
| AI Chat Assistant | 10,000,000 | 4,000,000 | $3.00 | $10.00 | **$13.00** |
| AI Task Generator | 400,000 | 300,000 | $0.12 | $0.75 | **$0.87** |
| **TOTAL** | **12,900,000** | **5,300,000** | **$3.87** | **$13.25** | **~$17.12/month** |

---

## Cost Optimization Features Already Implemented

### ✅ 1. Caching (AI Time Optimizer)
- **1-hour cache** for task analyses
- Reduces redundant API calls by ~70-80%
- **Savings**: ~$0.50-2.00/month depending on usage

### ✅ 2. Efficient Prompt Design
- Concise prompts with only necessary context
- Structured JSON responses (smaller output)
- **Savings**: ~20-30% token reduction

### ✅ 3. Batch Processing
- Project reports analyze multiple tasks in one call
- Workload analysis processes all users together
- **Savings**: ~15-25% token reduction

---

## Additional Cost Optimization Recommendations

### 1. Implement Chat History Limits
- **Current**: Last 10-20 messages
- **Recommendation**: Limit to last 5-10 messages for simple queries
- **Savings**: ~20-30% on chat tokens

### 2. Add Response Caching for Common Queries
- Cache common AI chat responses (e.g., "how to create a task")
- **Savings**: ~10-15% on chat tokens

### 3. Use Batch API for Task Generation
- If generating tasks for multiple projects, batch them
- **Savings**: ~5-10% on generation tokens

### 4. Implement Rate Limiting
- Prevent abuse and excessive API calls
- **Savings**: Prevents unexpected costs

### 5. Monitor Usage
- Track API calls and token usage
- Set up alerts for high usage
- **Benefit**: Early detection of cost spikes

---

## Free Tier Coverage

### Free Tier Limits:
- **15 requests per minute (RPM)**
- **1,500 requests per day (RPD)**
- **32,000 tokens per minute (TPM)**

### Coverage Analysis:

**Small Team (Scenario 1):**
- Daily requests: ~100-150/day ✅ (within 1,500/day limit)
- Requests per minute: ~1-2/min ✅ (within 15/min limit)
- Tokens per minute: ~5,000-8,000/min ✅ (within 32,000/min limit)
- **Result**: ✅ **Fully covered by free tier!**

**Medium Team (Scenario 2):**
- Daily requests: ~300-400/day ✅ (within 1,500/day limit)
- Requests per minute: ~3-5/min ✅ (within 15/min limit)
- Tokens per minute: ~15,000-25,000/min ✅ (within 32,000/min limit)
- **Result**: ✅ **Fully covered by free tier!**

**Large Team (Scenario 3):**
- Daily requests: ~800-1,000/day ✅ (within 1,500/day limit)
- Requests per minute: ~8-12/min ✅ (within 15/min limit)
- Tokens per minute: ~30,000-35,000/min ⚠️ (close to 32,000 limit)
- **Result**: ⚠️ **Mostly covered, may need rate limiting**

---

## Real-World Cost Examples

### Example 1: Startup (5 users, 2 projects/month)
- **Monthly cost**: ~$1.50-2.00
- **Annual cost**: ~$18-24
- **Free tier**: ✅ Fully covered

### Example 2: Small Business (15 users, 10 projects/month)
- **Monthly cost**: ~$4.00-5.00
- **Annual cost**: ~$48-60
- **Free tier**: ✅ Fully covered

### Example 3: Growing Company (30 users, 20 projects/month)
- **Monthly cost**: ~$8.00-10.00
- **Annual cost**: ~$96-120
- **Free tier**: ⚠️ May exceed on busy days

### Example 4: Enterprise (100+ users, 50+ projects/month)
- **Monthly cost**: ~$20.00-30.00
- **Annual cost**: ~$240-360
- **Free tier**: ❌ Will exceed, need paid tier

---

## Cost Monitoring Setup

### Recommended Monitoring:

1. **Track API Calls**
   - Log all Gemini API calls
   - Track tokens per call
   - Monitor daily/weekly/monthly totals

2. **Set Up Alerts**
   - Alert when daily requests > 1,000
   - Alert when daily tokens > 20M
   - Alert when monthly cost > $10

3. **Usage Dashboard**
   - Show API calls per feature
   - Show token usage breakdown
   - Show cost projections

---

## Summary

### ✅ Good News:
- **Small to medium teams**: Likely **FREE** (within free tier)
- **Cost is very low**: Even for large teams, ~$10-20/month
- **Optimizations in place**: Caching reduces costs significantly

### 💡 Recommendations:
1. **Start with free tier** - Monitor usage first
2. **Implement rate limiting** - Prevent abuse
3. **Monitor costs** - Track usage and set alerts
4. **Optimize prompts** - Keep them concise
5. **Use caching** - Already implemented ✅

### 📊 Bottom Line:
For most projects, **Google Gemini API costs will be $0-20/month**, which is extremely affordable for the AI capabilities provided!

---

## Getting Your API Key

1. Go to [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Create a new API key
3. Set it as environment variable:
   ```bash
   export GEMINI_API_KEY=your_api_key_here
   ```
4. Start using AI features!

---

## Questions?

- Check [Google Gemini Pricing](https://ai.google.dev/pricing)
- Monitor usage in Google Cloud Console
- Set up billing alerts in Google Cloud

