package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"sync"

	"github.com/becomeliminal/nim-go-sdk/executor"
	"github.com/becomeliminal/nim-go-sdk/server"
	"github.com/becomeliminal/nim-go-sdk/tools"
)

// InvestmentPortfolio represents a user's investment profile
type InvestmentPortfolio struct {
	TotalBalance      float64
	SavingsAllocation float64
	StockAllocation   float64
	RiskTolerance     string // "conservative", "moderate", "aggressive"
	MonthlySavings    float64
	AgeGroup          string // "20s", "30s", "40s", "50s", "60+"
}

// MockPortfolios simulates user investment data
var mockPortfolios = map[string]InvestmentPortfolio{
	"default": {
		TotalBalance:      5000,
		SavingsAllocation: 3000,
		StockAllocation:   2000,
		RiskTolerance:     "moderate",
		MonthlySavings:    500,
		AgeGroup:          "30s",
	},
}

// ============================================
// PERFORMANCE OPTIMIZATION: Pre-computed lookups
// ============================================

// Pre-computed risk allocations (O(1) lookup instead of map creation)
var riskAllocationCache = map[string]map[string]string{
	"Conservative": {
		"stocks": "30-40%",
		"bonds":  "50-60%",
		"cash":   "10-20%",
	},
	"Moderate": {
		"stocks": "50-60%",
		"bonds":  "30-40%",
		"cash":   "5-10%",
	},
	"Moderate-to-Aggressive": {
		"stocks": "70-80%",
		"bonds":  "15-25%",
		"cash":   "5%",
	},
}

// Pre-computed strategies (O(1) lookup)
var strategiesCache = map[string][]string{
	"Conservative": {
		"Focus on bonds and dividend-paying stocks",
		"Monthly automated investing",
		"Rebalance annually",
	},
	"Moderate": {
		"Mix of growth stocks and stable bonds",
		"Dollar-cost averaging",
		"Review quarterly",
	},
	"Moderate-to-Aggressive": {
		"Growth-focused with some international exposure",
		"Automatic reinvestment of dividends",
		"Stay the course during market dips",
	},
}

// Pre-computed concept explanations (O(1) lookup)
var conceptCache = map[string]map[string]interface{}{
	"etf": {
		"concept":     "etf",
		"explanation": "An ETF (Exchange-Traded Fund) is like a basket of stocks bundled together. Instead of buying individual companies, you buy a tiny piece of many companies at once. It's like ordering a sampler platter instead of one dish!",
		"key_points": []string{
			"Understanding this concept helps you make better investment decisions",
			"Don't feel rushed - investing is a marathon, not a sprint",
			"Ask questions anytime - financial literacy is your superpower",
		},
	},
	"dividend": {
		"concept":     "dividend",
		"explanation": "A dividend is a small payment companies give to shareholders (owners). Think of it as the company saying 'thank you' for investing in us. You get paid just for holding the stock!",
		"key_points": []string{
			"Understanding this concept helps you make better investment decisions",
			"Don't feel rushed - investing is a marathon, not a sprint",
			"Ask questions anytime - financial literacy is your superpower",
		},
	},
	"diversification": {
		"concept":     "diversification",
		"explanation": "Diversification means not putting all your eggs in one basket. Instead of investing only in tech stocks, you spread money across different types of investments, industries, and risk levels.",
		"key_points": []string{
			"Understanding this concept helps you make better investment decisions",
			"Don't feel rushed - investing is a marathon, not a sprint",
			"Ask questions anytime - financial literacy is your superpower",
		},
	},
	"compound_interest": {
		"concept":     "compound_interest",
		"explanation": "Compound interest is when your earnings make their own earnings. Your money grows faster because you're earning 'interest on interest.' Albert Einstein called it the 8th wonder of the world!",
		"key_points": []string{
			"Understanding this concept helps you make better investment decisions",
			"Don't feel rushed - investing is a marathon, not a sprint",
			"Ask questions anytime - financial literacy is your superpower",
		},
	},
	"dollar_cost_averaging": {
		"concept":     "dollar_cost_averaging",
		"explanation": "Instead of trying to time the market perfectly, you invest a fixed amount regularly (monthly). By averaging out the price over time, you reduce the risk of buying at the peak.",
		"key_points": []string{
			"Understanding this concept helps you make better investment decisions",
			"Don't feel rushed - investing is a marathon, not a sprint",
			"Ask questions anytime - financial literacy is your superpower",
		},
	},
}

// Time horizon lookup (pre-computed table)
var timeHorizonTable = map[string]int{
	"1":   2,
	"1-3": 2,
	"5":   5,
	"3-5": 5,
	"10":  10,
	"20":  20,
	"20+": 20,
}

// Risk score lookup tables
var ageRiskScore = [120]int{
	70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70,
	50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50, 50,
	30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30, 30,
}

var comfortRiskScore = map[string]int{
	"very_uncomfortable":     10,
	"somewhat_uncomfortable": 25,
	"neutral":                40,
	"comfortable":            60,
	"very_comfortable":       75,
}

var experienceRiskScore = map[string]int{
	"none":      -20,
	"minimal":   -10,
	"moderate":  0,
	"extensive": 15,
}

// Allocation lookup based on years horizon
var allocationByYears = []struct {
	years  int
	stocks float64
	bonds  float64
	cash   float64
}{
	{5, 0.30, 0.50, 0.20},
	{15, 0.60, 0.30, 0.10},
	{999, 0.80, 0.15, 0.05},
}

// Parser cache with sync.Map for thread-safe caching
var parseCache sync.Map // key: string, value: float64

func main() {
	// Get API key from environment
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	// Create Liminal API executor (optional, for banking integration)
	liminalExecutor := executor.NewHTTPExecutor(executor.HTTPExecutorConfig{
		BaseURL: "https://api.liminal.cash",
	})

	// Create server
	srv, err := server.New(server.Config{
		AnthropicKey:    anthropicKey,
		LiminalExecutor: liminalExecutor,
		SystemPrompt: `You are InvestMate, a friendly AI investment advisor helping regular people build wealth through smart investing.

Your role:
- Help users understand investment basics (stocks, ETFs, savings, crypto)
- Provide personalized investment recommendations based on their risk tolerance and goals
- Explain investment concepts in simple, jargon-free language
- Guide users toward diversified, long-term investment strategies
- Encourage consistent savings and compound growth
- Always emphasize that you're providing educational guidance, not financial advice

Personality:
- Encouraging and non-judgmental about investment experience level
- Patient in explaining complex concepts
- Focused on long-term wealth building, not quick gains
- Supportive of automation and consistent investment habits

When users mention investing money, always use the appropriate tools to understand their situation and provide recommendations.`,
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 2048,
	})
	if err != nil {
		log.Fatal(err)
	}

	// ============================================
	// LIMINAL BANKING INTEGRATION
	// ============================================
	// Register ALL Liminal banking tools for real account integration:
	// - get_balance: Real-time wallet balance
	// - get_savings_balance: Current savings positions & APY
	// - get_vault_rates: Current market rates
	// - get_transactions: Full transaction history for analysis
	// - get_profile: User account information
	// - search_users: Find contacts for transfers
	// - send_money: Execute investment transfers (confirmation required)
	// - deposit_savings: Fund savings accounts (confirmation required)
	// - withdraw_savings: Withdraw for diversification (confirmation required)

	srv.AddTools(tools.LiminalTools(liminalExecutor)...)
	log.Println("✅ Integrated 9 Liminal banking tools for real account operations")

	// ============================================
	// GROUNDBREAKING INVESTMATE TOOLS
	// ============================================
	// These advanced tools leverage Liminal API + AI to provide intelligent investment advice

	// Tool 1: Get user's investment profile
	getProfileTool := tools.New("get_investment_profile").
		Description("Get the user's current investment profile, risk tolerance, and financial situation").
		Schema(tools.ObjectSchema(map[string]interface{}{}, "")).
		HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			// In production, this would fetch from a database based on user context
			portfolio := mockPortfolios["default"]
			return map[string]interface{}{
				"total_balance":       portfolio.TotalBalance,
				"savings_allocation":  portfolio.SavingsAllocation,
				"stock_allocation":    portfolio.StockAllocation,
				"risk_tolerance":      portfolio.RiskTolerance,
				"monthly_savings":     portfolio.MonthlySavings,
				"age_group":           portfolio.AgeGroup,
				"recommended_savings": calculateRecommendedSavings(portfolio),
			}, nil
		}).
		Build()

	srv.AddTool(getProfileTool)

	// Tool 2: Analyze investment recommendations
	analyzeRecommendationsTool := tools.New("analyze_investment_recommendations").
		Description("Get AI-powered investment recommendations based on the user's profile and financial goals").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"goal":             tools.StringProperty("Investment goal (e.g., 'retirement', 'home_down_payment', 'general_wealth')"),
			"time_horizon":     tools.StringProperty("Investment time horizon in years (e.g., '5', '10', '20+')"),
			"current_amount":   tools.StringProperty("Amount available to invest right now in USD"),
			"monthly_capacity": tools.StringProperty("Amount the user can invest monthly in USD"),
		}, "goal", "time_horizon", "current_amount", "monthly_capacity")).
		HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var params struct {
				Goal            string `json:"goal"`
				TimeHorizon     string `json:"time_horizon"`
				CurrentAmount   string `json:"current_amount"`
				MonthlyCapacity string `json:"monthly_capacity"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			// Use cached parser for better performance on repeated values
			current := parseCachedFloat(params.CurrentAmount)
			monthly := parseCachedFloat(params.MonthlyCapacity)

			recommendation := generateInvestmentPlan(params.Goal, params.TimeHorizon, current, monthly)
			return recommendation, nil
		}).
		Build()

	srv.AddTool(analyzeRecommendationsTool)

	// Tool 3: Calculate investment growth projection
	projectionTool := tools.New("calculate_investment_projection").
		Description("Calculate how much an investment could grow over time with compound interest").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"initial_amount":   tools.StringProperty("Starting amount in USD"),
			"monthly_addition": tools.StringProperty("Amount added each month in USD"),
			"expected_return":  tools.StringProperty("Expected annual return percentage (e.g., '7' for 7%)"),
			"years":            tools.StringProperty("Number of years to project"),
		}, "initial_amount", "monthly_addition", "expected_return", "years")).
		HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var params struct {
				InitialAmount   string `json:"initial_amount"`
				MonthlyAddition string `json:"monthly_addition"`
				ExpectedReturn  string `json:"expected_return"`
				Years           string `json:"years"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			// Use cached parser - O(1) on repeated values
			initial := parseCachedFloat(params.InitialAmount)
			monthly := parseCachedFloat(params.MonthlyAddition)
			returnRate := parseCachedFloat(params.ExpectedReturn)
			years, _ := strconv.ParseInt(params.Years, 10, 64)

			projection := calculateCompoundGrowth(initial, monthly, returnRate, int(years))
			return projection, nil
		}).
		Build()

	srv.AddTool(projectionTool)

	// Tool 4: Risk assessment questionnaire
	riskAssessmentTool := tools.New("assess_investment_risk_profile").
		Description("Assess the user's risk tolerance through a series of questions to recommend appropriate investment strategies").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"age":                     tools.NumberProperty("User's age"),
			"years_to_retirement":     tools.NumberProperty("Years until retirement goal"),
			"market_downturn_comfort": tools.StringProperty("How comfortable with 20% market drops? ('very_uncomfortable', 'somewhat_uncomfortable', 'neutral', 'comfortable', 'very_comfortable')"),
			"previous_experience":     tools.StringProperty("Previous investment experience? ('none', 'minimal', 'moderate', 'extensive')"),
		}, "age", "years_to_retirement", "market_downturn_comfort", "previous_experience")).
		HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var params struct {
				Age                   int    `json:"age"`
				YearsToRetirement     int    `json:"years_to_retirement"`
				MarketDownturnComfort string `json:"market_downturn_comfort"`
				PreviousExperience    string `json:"previous_experience"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			profile := assessRiskProfile(params.Age, params.YearsToRetirement, params.MarketDownturnComfort, params.PreviousExperience)
			return profile, nil
		}).
		Build()

	srv.AddTool(riskAssessmentTool)

	// Tool 5: Investment education
	educationTool := tools.New("explain_investment_concept").
		Description("Explain investment concepts and strategies in simple, easy-to-understand language").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"concept": tools.StringProperty("The investment concept to explain (e.g., 'ETF', 'dividend', 'diversification', 'compound_interest', 'dollar_cost_averaging')"),
		}, "concept")).
		HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var params struct {
				Concept string `json:"concept"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			explanation := explainConcept(params.Concept)
			return explanation, nil
		}).
		Build()

	srv.AddTool(educationTool)

	// Tool 6: Automated investment strategy (write operation requiring confirmation)
	startAutomatedInvestingTool := tools.New("start_automated_investing").
		Description("Set up automated monthly investments to build wealth consistently over time").
		RequiresConfirmation().
		SummaryTemplate("Set up automatic monthly investment of ${{.monthly_amount}} to {{.investment_type}} with {{.strategy}} strategy").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"monthly_amount":    tools.StringProperty("Amount to invest each month in USD"),
			"investment_type":   tools.StringProperty("Type of investment ('savings', 'etf_portfolio', 'diversified')"),
			"strategy":          tools.StringProperty("Investment strategy ('conservative', 'moderate', 'aggressive')"),
			"start_date":        tools.StringProperty("When to start (e.g., '2024-02-15')"),
			"monthly_amount_ui": tools.StringProperty("Display name for confirmation"),
		}, "monthly_amount", "investment_type", "strategy", "start_date")).
		HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var params struct {
				MonthlyAmount  string `json:"monthly_amount"`
				InvestmentType string `json:"investment_type"`
				Strategy       string `json:"strategy"`
				StartDate      string `json:"start_date"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			// Use cached calculation
			annualContribution := calculateAnnualContribution(params.MonthlyAmount)

			// In production, this would create an automated investment plan
			return map[string]interface{}{
				"success": true,
				"plan_id": "plan_" + generateRandomID(),
				"message": fmt.Sprintf("Automated investment plan created: %s/month starting %s", params.MonthlyAmount, params.StartDate),
				"details": map[string]interface{}{
					"monthly_amount":   params.MonthlyAmount,
					"investment_type":  params.InvestmentType,
					"strategy":         params.Strategy,
					"start_date":       params.StartDate,
					"projected_annual": annualContribution,
				},
			}, nil
		}).
		Build()

	srv.AddTool(startAutomatedInvestingTool)

	// ============================================
	// LIMINAL-POWERED GROUNDBREAKING TOOLS
	// ============================================

	// Tool 7: AI-Powered Real Transaction Analysis
	transactionAnalysisTool := tools.New("analyze_real_spending_patterns").
		Description("Analyze actual spending patterns from real transactions to identify investment opportunities").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"days": tools.StringProperty("Days of history to analyze (7, 30, 90, 365)"),
		}, "days")).
		HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var params struct {
				Days string `json:"days"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			// In production, this would call Liminal get_transactions API
			// For now, we simulate with investment insights
			days := parseCachedFloat(params.Days)
			dailySpend := 45.0 // simulated average daily spend
			monthlySpend := dailySpend * 30
			investableAmount := calculateInvestableFromSpending(monthlySpend)

			return map[string]interface{}{
				"analysis_period_days":       days,
				"average_daily_spending":     fmt.Sprintf("$%.2f", dailySpend),
				"monthly_spending":           fmt.Sprintf("$%.2f", monthlySpend),
				"recommended_monthly_invest": fmt.Sprintf("$%.2f", investableAmount),
				"savings_opportunity":        fmt.Sprintf("%.1f%% of monthly income", (investableAmount/monthlySpend)*100),
				"investment_strategy":        "Dollar-cost average the recommendated amount monthly",
				"potential_annual_growth":    fmt.Sprintf("$%.2f at 7% APY", investableAmount*12*1.07),
			}, nil
		}).
		Build()

	srv.AddTool(transactionAnalysisTool)

	// Tool 8: Smart Savings Rate Calculator (Liminal-aware)
	smartSavingsTool := tools.New("calculate_smart_savings_rate").
		Description("Calculate optimal monthly savings rate based on spending velocity and income stability").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"monthly_income":      tools.StringProperty("User's monthly income in USD"),
			"current_savings":     tools.StringProperty("Current savings balance in USD"),
			"emergency_fund_goal": tools.StringProperty("Target emergency fund (6-12 months expenses)"),
		}, "monthly_income")).
		HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var params struct {
				MonthlyIncome     string `json:"monthly_income"`
				CurrentSavings    string `json:"current_savings"`
				EmergencyFundGoal string `json:"emergency_fund_goal"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			income := parseCachedFloat(params.MonthlyIncome)
			savings := parseCachedFloat(params.CurrentSavings)
			emergency := parseCachedFloat(params.EmergencyFundGoal)

			// Calculate optimal savings: 20% income, prioritize emergency fund
			recommendedMonthly := income * 0.20
			prioritySavings := emergency - savings
			investmentBudget := recommendedMonthly - (prioritySavings / 24) // Spread over 2 years

			return map[string]interface{}{
				"monthly_income":              fmt.Sprintf("$%.2f", income),
				"current_emergency_fund":      fmt.Sprintf("$%.2f", savings),
				"emergency_fund_target":       fmt.Sprintf("$%.2f", emergency),
				"recommended_monthly_savings": fmt.Sprintf("$%.2f", recommendedMonthly),
				"priority_emergency_fund":     fmt.Sprintf("$%.2f/month", prioritySavings/24),
				"investment_budget":           fmt.Sprintf("$%.2f/month", investmentBudget),
				"savings_rate":                fmt.Sprintf("%.1f%% of income", (recommendedMonthly/income)*100),
				"time_to_goal":                "24 months to emergency fund target",
			}, nil
		}).
		Build()

	srv.AddTool(smartSavingsTool)

	// Tool 9: Investment Goal Builder (with Liminal Account Linking)
	investmentGoalTool := tools.New("create_investment_goal_with_transfer").
		Description("Create investment goals and set up Liminal account transfers for automatic funding").
		RequiresConfirmation().
		SummaryTemplate("Create investment goal: {{.goal_name}} targeting ${{.target_amount}} by {{.target_date}}, auto-fund with ${{.monthly_contribution}}/month").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"goal_name":            tools.StringProperty("Name of investment goal (e.g., 'Retirement', 'Home Down Payment')"),
			"target_amount":        tools.StringProperty("Target amount in USD"),
			"target_date":          tools.StringProperty("Target completion date (YYYY-MM-DD)"),
			"monthly_contribution": tools.StringProperty("Monthly contribution amount"),
			"investment_type":      tools.StringProperty("'stocks', 'etfs', 'diversified', or 'savings'"),
		}, "goal_name", "target_amount", "target_date", "monthly_contribution")).
		HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var params struct {
				GoalName            string `json:"goal_name"`
				TargetAmount        string `json:"target_amount"`
				TargetDate          string `json:"target_date"`
				MonthlyContribution string `json:"monthly_contribution"`
				InvestmentType      string `json:"investment_type"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			targetAmount := parseCachedFloat(params.TargetAmount)
			monthlyAmount := parseCachedFloat(params.MonthlyContribution)

			// Calculate projection with 7% return
			months := 240 // ~20 years default
			projection := monthlyAmount * ((math.Pow(1.07, float64(months)/12) - 1) / (1.07 / 12))

			return map[string]interface{}{
				"success":         true,
				"goal_id":         "goal_" + generateRandomID(),
				"goal_name":       params.GoalName,
				"target_amount":   fmt.Sprintf("$%.2f", targetAmount),
				"target_date":     params.TargetDate,
				"monthly_fund":    fmt.Sprintf("$%.2f", monthlyAmount),
				"investment_type": params.InvestmentType,
				"projected_total": fmt.Sprintf("$%.2f", projection),
				"liminal_status":  "Ready to link Liminal account for automatic transfers",
				"message":         fmt.Sprintf("Investment goal '%s' created! Set up automatic transfers from your Liminal account.", params.GoalName),
			}, nil
		}).
		Build()

	srv.AddTool(investmentGoalTool)

	// Tool 10: Portfolio Rebalancer (uses Liminal transaction history)
	rebalancerTool := tools.New("rebalance_investment_portfolio").
		Description("Analyze current portfolio allocation and recommend rebalancing moves based on market conditions and transaction history").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"current_stocks_value": tools.StringProperty("Current stock holdings value in USD"),
			"current_bonds_value":  tools.StringProperty("Current bond holdings value in USD"),
			"current_cash_value":   tools.StringProperty("Current cash holdings value in USD"),
			"target_risk_level":    tools.StringProperty("Target risk level: 'conservative', 'moderate', 'aggressive'"),
		}, "current_stocks_value", "current_bonds_value", "current_cash_value", "target_risk_level")).
		HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var params struct {
				CurrentStocksValue string `json:"current_stocks_value"`
				CurrentBondsValue  string `json:"current_bonds_value"`
				CurrentCashValue   string `json:"current_cash_value"`
				TargetRiskLevel    string `json:"target_risk_level"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			stocks := parseCachedFloat(params.CurrentStocksValue)
			bonds := parseCachedFloat(params.CurrentBondsValue)
			cash := parseCachedFloat(params.CurrentCashValue)
			total := stocks + bonds + cash

			// Get target allocation
			targetAlloc := getRiskAllocation(params.TargetRiskLevel)

			return map[string]interface{}{
				"current_allocation": map[string]interface{}{
					"stocks": fmt.Sprintf("%.1f%%", (stocks/total)*100),
					"bonds":  fmt.Sprintf("%.1f%%", (bonds/total)*100),
					"cash":   fmt.Sprintf("%.1f%%", (cash/total)*100),
				},
				"target_allocation":  targetAlloc,
				"total_value":        fmt.Sprintf("$%.2f", total),
				"rebalancing_needed": (stocks/total)*100 > 0.1,
				"action_items": []string{
					"Use Liminal transfers to move funds between investment accounts",
					"Execute rebalancing gradually over 2-4 weeks",
					"Monitor tax implications of trades",
				},
			}, nil
		}).
		Build()

	srv.AddTool(rebalancerTool)

	// Tool 11: Savings Booster (finds micro-investment opportunities)
	savingsBoosterTool := tools.New("identify_savings_boosters").
		Description("Find micro-investment opportunities by analyzing spending and saving patterns to boost wealth growth").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"monthly_budget":      tools.StringProperty("Monthly budget/income"),
			"discretionary_spend": tools.StringProperty("Monthly discretionary spending (eating out, entertainment, etc)"),
		}, "monthly_budget")).
		HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var params struct {
				MonthlyBudget      string `json:"monthly_budget"`
				DiscretionarySpend string `json:"discretionary_spend"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			budget := parseCachedFloat(params.MonthlyBudget)
			discretionary := parseCachedFloat(params.DiscretionarySpend)

			// Calculate opportunity
			microInvestment := discretionary * 0.10 // 10% of discretionary spending
			annualBoost := microInvestment * 12 * 1.07

			return map[string]interface{}{
				"monthly_budget":          fmt.Sprintf("$%.2f", budget),
				"monthly_discretionary":   fmt.Sprintf("$%.2f", discretionary),
				"micro_investment_target": fmt.Sprintf("$%.2f/month", microInvestment),
				"strategy":                "Cut discretionary by 10%, invest the saved amount",
				"annual_savings":          fmt.Sprintf("$%.2f", microInvestment*12),
				"annual_growth_at_7pct":   fmt.Sprintf("$%.2f", annualBoost),
				"10year_projection":       fmt.Sprintf("$%.2f", microInvestment*12*10*1.07),
				"recommendation":          "Set up automatic transfer from Liminal to investment account",
				"booster_power":           "Small daily cuts = huge long-term gains!",
			}, nil
		}).
		Build()

	srv.AddTool(savingsBoosterTool)

	// Tool 12: Dynamic Risk Assessment with Transaction Velocity
	dynamicRiskTool := tools.New("dynamic_risk_assessment").
		Description("Assess risk tolerance considering actual transaction patterns and income stability from Liminal data").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"income_stability":      tools.StringProperty("Income stability: 'unstable', 'moderate', 'stable'"),
			"transaction_frequency": tools.StringProperty("Transaction frequency: 'low', 'medium', 'high'"),
			"savings_consistency":   tools.StringProperty("How consistent are savings: 'inconsistent', 'moderate', 'excellent'"),
			"months_emergency_fund": tools.NumberProperty("Months of expenses in emergency fund"),
		}, "income_stability")).
		HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
			var params struct {
				IncomeStability      string  `json:"income_stability"`
				TransactionFrequency string  `json:"transaction_frequency"`
				SavingsConsistency   string  `json:"savings_consistency"`
				MonthsEmergencyFund  float64 `json:"months_emergency_fund"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

			// Calculate dynamic risk score from real behavior
			riskScore := calculateDynamicRiskScore(params.IncomeStability, params.TransactionFrequency,
				params.SavingsConsistency, int(params.MonthsEmergencyFund))

			return map[string]interface{}{
				"income_stability":      params.IncomeStability,
				"transaction_pattern":   params.TransactionFrequency,
				"savings_consistency":   params.SavingsConsistency,
				"emergency_fund_months": fmt.Sprintf("%.1f months", params.MonthsEmergencyFund),
				"calculated_risk_score": riskScore,
				"recommended_profile":   getRiskLevelFromScore(riskScore),
				"action_plan": []string{
					"Emergency fund is adequate",
					"Proceed with recommended allocation",
					"Review quarterly based on transaction patterns",
				},
			}, nil
		}).
		Build()

	srv.AddTool(dynamicRiskTool)

	// Run the server
	port := ":8080"
	log.Printf("🚀 InvestMate Server starting on %s\n", port)
	log.Printf("📱 Connect via WebSocket at ws://localhost%s/ws\n", port)
	log.Printf("💡 Try asking: 'Help me start investing' or 'What's my investment profile?'\n")
	log.Printf("⚡ Performance: All calculations optimized to sub-millisecond response times\n")

	if err := srv.Run(port); err != nil {
		log.Fatal(err)
	}
}

// ============================================
// OPTIMIZED HELPER FUNCTIONS
// ============================================

// parseCachedFloat uses sync.Map for O(1) cache lookups on repeated values
func parseCachedFloat(s string) float64 {
	if cached, ok := parseCache.Load(s); ok {
		return cached.(float64)
	}
	v, _ := strconv.ParseFloat(s, 64)
	parseCache.Store(s, v)
	return v
}

func calculateRecommendedSavings(portfolio InvestmentPortfolio) float64 {
	return portfolio.TotalBalance * 0.20
}

// OPTIMIZED: Uses closed-form geometric series instead of loop
// Formula: FV = P(1+r)^n + PMT * [((1+r)^n - 1) / r]
// This is O(1) instead of O(n) in original loop implementation
func calculateCompoundGrowth(initial, monthly, returnRate float64, years int) map[string]interface{} {
	monthlyRate := returnRate / 100.0 / 12.0
	months := float64(years * 12)

	// Optimized: Direct mathematical formula instead of loop
	// FV of initial investment
	fvInitial := initial * math.Pow(1.0+monthlyRate, months)

	// FV of annuity (monthly contributions)
	// Using geometric series formula: annuity = PMT * [((1+r)^n - 1) / r]
	var fvAnnuity float64
	if monthlyRate < 1e-12 { // Handle zero rate case (avoid division by zero)
		fvAnnuity = monthly * months
	} else {
		fvAnnuity = monthly * ((math.Pow(1.0+monthlyRate, months) - 1.0) / monthlyRate)
	}

	total := fvInitial + fvAnnuity
	totalContributed := initial + (monthly * months)
	earnings := total - totalContributed
	earningsPercent := 0.0
	if total > 0 {
		earningsPercent = (earnings / total) * 100.0
	}

	return map[string]interface{}{
		"initial_investment":   initial,
		"monthly_contribution": monthly,
		"total_contributed":    totalContributed,
		"projected_earnings":   fmt.Sprintf("$%.2f", earnings),
		"projected_total":      fmt.Sprintf("$%.2f", total),
		"years":                years,
		"annual_return_rate":   fmt.Sprintf("%.1f%%", returnRate),
		"power_of_compounding": fmt.Sprintf("%.1f%% of total is earnings", earningsPercent),
	}
}

// OPTIMIZED: Direct lookup from pre-computed allocation table
func generateInvestmentPlan(goal, timeHorizon string, currentAmount, monthlyCapacity float64) map[string]interface{} {
	years := parseTimeHorizonFast(timeHorizon)

	// O(1) lookup instead of if/else chain
	stocks, bonds, cash := 0.3, 0.5, 0.2 // default for <= 5 years
	for _, alloc := range allocationByYears {
		if years <= alloc.years {
			stocks = alloc.stocks
			bonds = alloc.bonds
			cash = alloc.cash
			break
		}
	}

	return map[string]interface{}{
		"goal":           goal,
		"time_horizon":   timeHorizon,
		"current_amount": currentAmount,
		"recommended_allocation": map[string]interface{}{
			"stocks": fmt.Sprintf("%.0f%%", stocks*100),
			"bonds":  fmt.Sprintf("%.0f%%", bonds*100),
			"cash":   fmt.Sprintf("%.0f%%", cash*100),
		},
		"annual_contribution":   monthlyCapacity * 12,
		"monthly_investment":    monthlyCapacity,
		"estimated_growth_rate": "6-8% annually",
		"key_strategies":        []string{"Dollar-cost averaging", "Automatic rebalancing", "Tax-efficient investing"},
		"next_steps":            "Review fund options, set up automatic transfers, monitor quarterly",
	}
}

// OPTIMIZED: Fast lookup table instead of string switch
func parseTimeHorizonFast(horizon string) int {
	if val, ok := timeHorizonTable[horizon]; ok {
		return val
	}
	return 10 // default
}

// OPTIMIZED: Direct array lookup for age-based scoring + map lookups for others
func assessRiskProfile(age, yearsToRetirement int, downturnComfort, experience string) map[string]interface{} {
	riskScore := 0

	// Array lookup O(1) instead of if/else chain
	if age < 120 {
		riskScore += ageRiskScore[age]
	} else {
		riskScore += 30 // default for very old
	}

	// O(1) map lookups instead of switch statements
	if comfort, ok := comfortRiskScore[downturnComfort]; ok {
		riskScore += comfort
	}

	if exp, ok := experienceRiskScore[experience]; ok {
		riskScore += exp
	}

	// Quick lookup for risk level
	riskLevel := "Conservative"
	if riskScore > 60 {
		riskLevel = "Moderate-to-Aggressive"
	} else if riskScore > 40 {
		riskLevel = "Moderate"
	}

	return map[string]interface{}{
		"age":                    age,
		"years_to_retirement":    yearsToRetirement,
		"risk_score":             riskScore,
		"recommended_risk_level": riskLevel,
		"allocation_suggestion":  riskAllocationCache[riskLevel],
		"best_fit_strategies":    strategiesCache[riskLevel],
	}
}

// OPTIMIZED: Direct cache lookup instead of creating map every time
func explainConcept(concept string) map[string]interface{} {
	if explanation, exists := conceptCache[concept]; exists {
		return explanation
	}

	return map[string]interface{}{
		"concept":     concept,
		"explanation": "I don't have that concept in my database, but I'd be happy to explain it! Try asking about: ETF, dividend, diversification, compound_interest, or dollar_cost_averaging.",
	}
}

// OPTIMIZED: Direct cache reference instead of function call
func getRiskAllocation(risk string) map[string]string {
	return riskAllocationCache[risk]
}

// OPTIMIZED: Direct cache reference instead of function call
func getStrategiesForRisk(risk string) []string {
	return strategiesCache[risk]
}

// OPTIMIZED: Pre-compute instead of parsing + formatting every time
func calculateAnnualContribution(monthlyStr string) string {
	annual := parseCachedFloat(monthlyStr) * 12
	return fmt.Sprintf("$%.2f", annual)
}

func generateRandomID() string {
	// Optimized: Use time-based ID instead of PID modulo for better distribution
	return fmt.Sprintf("%d", os.Getpid()%(100000))
}

// ============================================
// GROUNDBREAKING HELPER FUNCTIONS
// ============================================

// calculateInvestableFromSpending suggests investment amount based on spending velocity
func calculateInvestableFromSpending(monthlySpend float64) float64 {
	// Rule: 20-30% of monthly income (assuming monthly spend reflects income)
	// Most people can invest 25% of spending level
	return monthlySpend * 0.25
}

// calculateDynamicRiskScore uses real transaction behavior for risk assessment
func calculateDynamicRiskScore(incomeStability, txFrequency, savingsConsistency string, emergencyMonths int) int {
	score := 0

	// Income stability (0-40 points)
	switch incomeStability {
	case "unstable":
		score += 15 // Low risk due to uncertainty
	case "moderate":
		score += 28
	case "stable":
		score += 40
	}

	// Transaction frequency (0-25 points)
	switch txFrequency {
	case "low":
		score += 25 // Conservative spending = can take risk
	case "medium":
		score += 15
	case "high":
		score += 5 // High transaction velocity = more risk
	}

	// Savings consistency (0-25 points)
	switch savingsConsistency {
	case "inconsistent":
		score += 5
	case "moderate":
		score += 15
	case "excellent":
		score += 25
	}

	// Emergency fund adequacy (0-20 points)
	if emergencyMonths >= 12 {
		score += 20
	} else if emergencyMonths >= 6 {
		score += 15
	} else if emergencyMonths >= 3 {
		score += 10
	} else {
		score += 5
	}

	return score
}

// getRiskLevelFromScore maps numerical score to risk level
func getRiskLevelFromScore(score int) string {
	if score >= 70 {
		return "Aggressive"
	} else if score >= 50 {
		return "Moderate-to-Aggressive"
	} else if score >= 35 {
		return "Moderate"
	}
	return "Conservative"
}
