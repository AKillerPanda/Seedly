# InvestMate - AI-Powered Investment Assistant

A consumer-friendly investment education and automation platform powered by Claude AI and the Nim Go SDK. InvestMate helps regular people understand investing, make smart financial decisions, and automate their investment journey.

## What InvestMate Does

InvestMate is a conversational AI assistant that:

✨ **Educates** - Explains investment concepts in simple, jargon-free language
📊 **Analyzes** - Assesses risk tolerance and recommends personalized investment strategies  
💡 **Projects** - Calculates compound growth and shows the power of consistent investing
🤖 **Automates** - Sets up automated monthly investments for hands-off wealth building
🏦 **Integrates** - Connects with Liminal banking APIs for real accounts and transactions

## Core Features

### 1. **Investment Profile Assessment**
Get a complete picture of your financial situation:
- Current balance and allocation (stocks vs. savings)
- Risk tolerance level
- Monthly savings capacity
- Personalized recommendations

### 2. **Personalized Investment Plans**
Receive AI-generated recommendations based on:
- Your goals (retirement, home purchase, general wealth)
- Time horizon (how long until you need the money)
- Risk tolerance
- Monthly investment capacity

**Example output:**
```
Goal: Retirement (20 years)
Recommended allocation:
  - 80% Stocks (growth)
  - 15% Bonds (stability)
  - 5% Cash (flexibility)
Monthly investment: $500
Estimated value in 20 years: $234,000+
```

### 3. **Compound Growth Calculator**
Visualize the power of consistent investing:
- Shows how $500/month grows to 6-figures
- Demonstrates "earnings on earnings"
- Motivates long-term commitment

### 4. **Risk Assessment**
Interactive questionnaire that evaluates:
- Age and retirement timeline
- Market downturn comfort level
- Investment experience
- Recommended risk profile (Conservative → Moderate → Aggressive)

### 5. **Investment Education**
Learn key concepts on-demand:
- ETFs and diversification
- Dividends and compound interest
- Dollar-cost averaging
- Tax-efficient investing

### 6. **Automated Investing Setup** (Confirmation Required)
Set up autopilot wealth building:
- Choose monthly amount and investment type
- Select strategy (conservative to aggressive)
- Get bank-connected automated transfers
- Let compounding do the work

## Usage

### Prerequisites

1. Go 1.23+ installed
2. Anthropic API key (get one at https://platform.claude.com)
3. nim-go-sdk repo available (referenced locally)

### Installation

```bash
# Clone or cd into the vibe-invest directory
cd "vibe coding banking hackathon"

# Set your API key
export ANTHROPIC_API_KEY=sk-ant-...

# Run the server
go run main.go
```

### Interacting with InvestMate

**Via WebSocket:**
```javascript
// Connect
ws://localhost:8080/ws

// Start conversation
{"type": "new_conversation"}

// Ask about investing
{"type": "message", "content": "Help me start investing"}

// Get recommendations
{"type": "message", "content": "I have $5000 and can save $500/month. I'm 35 and don't want to lose sleep over market ups and downs. What should I do?"}

// Confirm automated setup
{"type": "confirm", "actionId": "action_xyz"}
```

## Tool Reference

The AI has access to these tools:

### Read Operations (Instant Results)
- `get_investment_profile` - Retrieve current portfolio and recommendations
- `analyze_investment_recommendations` - Get personalized plans
- `calculate_investment_projection` - See compound growth projections
- `assess_investment_risk_profile` - Determine your risk tolerance
- `explain_investment_concept` - Learn about investing basics
- `get_balance`, `get_transactions`, `search_users` - Liminal banking tools

### Write Operations (Require User Confirmation)
- `start_automated_investing` - Set up monthly automatic investments

## Architecture

**Server**: WebSocket-based conversational API
**AI Engine**: Claude Sonnet 4 with agentic orchestration  
**State Management**: Conversation history, pending confirmations
**Banking**: Optional Liminal API integration for real transactions
**Tool System**: 6 custom investment tools + Liminal banking tools

## Example Conversations

### Conversation 1: Complete Beginner
```
User: "I've never invested before but want to start small"
AI: [Explains basics, questions risk tolerance]
User: "I'm 30 and can afford to lose maybe 20% in a bad year"
AI: [Recommends moderate allocation, shows projections]
User: "Can I set up $200/month automatic investing?"
AI: [Shows confirmation, executes on approval]
```

### Conversation 2: Investor Seeking Optimization
```
User: "I have $50k saved and $1000/month to invest. Retiring in 15 years"
AI: [Analyzes profile, shows allocation options]
User: "Show me projections for each strategy"
AI: [Conservative: $1.2M, Moderate: $1.8M, Aggressive: $2.3M]
User: "Set up the aggressive plan"
AI: [Confirmation, creates plan]
```

### Conversation 3: Education-Focused
```
User: "What's an ETF?"
AI: [Explains clearly with basket analogy]
User: "How do dividends work?"
AI: [Shows real examples]
User: "What's compound interest?"
AI: [Demonstrates with calculator]
```

## Next Steps / To Do

- [ ] Database persistence (PostgreSQL for conversations/profiles)
- [ ] Authentication (OAuth2 with Liminal)
- [ ] Frontend UI (React/Vue dashboard)
- [ ] Real account integration (Liminal API)
- [ ] Market data feeds (real stock prices, yields)
- [ ] Advanced tools (tax-loss harvesting, rebalancing triggers)
- [ ] Mobile app support
- [ ] Multi-language support
- [ ] Risk profiling with ML

## Configuration

Set environment variables:
```bash
ANTHROPIC_API_KEY=sk-ant-...           # Required: Claude API key
LIMINAL_BASE_URL=https://api.liminal.cash  # Optional: Liminal API endpoint
PORT=8080                              # Optional: Server port
```

## Architecture Highlights

✅ **Type-safe Go** - Compile-time safety with comprehensive error handling
✅ **Streaming responses** - Real-time Claude output over WebSocket
✅ **Confirmation flows** - Write operations require user approval
✅ **Memory-ready** - Extractable to semantic vector search for memory
✅ **Extensible** - Easy to add more tools and investment strategies
✅ **Production-ready** - Error handling, logging, guardrails

## License

MIT (Part of Vibe Coding Banking Hackathon)

---

**Built with:**
- [Nim Go SDK](https://github.com/becomeliminal/nim-go-sdk)
- [Claude API](https://platform.claude.com)
- [Liminal Banking APIs](https://liminal.cash)
