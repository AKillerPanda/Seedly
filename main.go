package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

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

	// Register Liminal banking tools (read-only for this app)
	srv.AddTools(tools.LiminalTools(liminalExecutor)...)

	// ============================================
	// CUSTOM INVESTMENT TOOLS
	// ============================================

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

			current, _ := strconv.ParseFloat(params.CurrentAmount, 64)
			monthly, _ := strconv.ParseFloat(params.MonthlyCapacity, 64)

			recommendation := generateInvestmentPlan(params.Goal, params.TimeHorizon, current, monthly)
			return recommendation, nil
		}).
		Build()

	srv.AddTool(analyzeRecommendationsTool)

	// Tool 3: Calculate investment growth projection
	projectionTool := tools.New("calculate_investment_projection").
		Description("Calculate how much an investment could grow over time with compound interest").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"initial_amount":    tools.StringProperty("Starting amount in USD"),
			"monthly_addition":  tools.StringProperty("Amount added each month in USD"),
			"expected_return":   tools.StringProperty("Expected annual return percentage (e.g., '7' for 7%)"),
			"years":             tools.StringProperty("Number of years to project"),
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

			initial, _ := strconv.ParseFloat(params.InitialAmount, 64)
			monthly, _ := strconv.ParseFloat(params.MonthlyAddition, 64)
			returnRate, _ := strconv.ParseFloat(params.ExpectedReturn, 64)
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
			"age":                    tools.NumberProperty("User's age"),
			"years_to_retirement":    tools.NumberProperty("Years until retirement goal"),
			"market_downturn_comfort": tools.StringProperty("How comfortable with 20% market drops? ('very_uncomfortable', 'somewhat_uncomfortable', 'neutral', 'comfortable', 'very_comfortable')"),
			"previous_experience":    tools.StringProperty("Previous investment experience? ('none', 'minimal', 'moderate', 'extensive')"),
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
				MonthlyAmount string `json:"monthly_amount"`
				InvestmentType string `json:"investment_type"`
				Strategy      string `json:"strategy"`
				StartDate     string `json:"start_date"`
			}
			if err := json.Unmarshal(input, &params); err != nil {
				return nil, fmt.Errorf("invalid input: %w", err)
			}

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
					"projected_annual": calculateAnnualContribution(params.MonthlyAmount),
				},
			}, nil
		}).
		Build()

	srv.AddTool(startAutomatedInvestingTool)

	// Run the server
	port := ":8080"
	log.Printf("🚀 InvestMate Server starting on %s\n", port)
	log.Printf("📱 Connect via WebSocket at ws://localhost%s/ws\n", port)
	log.Printf("💡 Try asking: 'Help me start investing' or 'What's my investment profile?'\n")

	if err := srv.Run(port); err != nil {
		log.Fatal(err)
	}
}

// ============================================
// HELPER FUNCTIONS
// ============================================

func calculateRecommendedSavings(portfolio InvestmentPortfolio) float64 {
	// Rule of thumb: 20% of gross income for long-term investing
	// Simplified: 20% of total balance per year
	return portfolio.TotalBalance * 0.20
}

func generateInvestmentPlan(goal, timeHorizon string, currentAmount, monthlyCapacity float64) map[string]interface{} {
	// Simple allocation model based on time horizon
	var stocks, bonds, cash float64

	years := parseTimeHorizon(timeHorizon)
	if years <= 5 {
		stocks = 0.30
		bonds = 0.50
		cash = 0.20
	} else if years <= 15 {
		stocks = 0.60
		bonds = 0.30
		cash = 0.10
	} else {
		stocks = 0.80
		bonds = 0.15
		cash = 0.05
	}

	totalMonthly := monthlyCapacity * 12
	return map[string]interface{}{
		"goal":             goal,
		"time_horizon":     timeHorizon,
		"current_amount":   currentAmount,
		"recommended_allocation": map[string]interface{}{
			"stocks": fmt.Sprintf("%.0f%%", stocks*100),
			"bonds":  fmt.Sprintf("%.0f%%", bonds*100),
			"cash":   fmt.Sprintf("%.0f%%", cash*100),
		},
		"annual_contribution":      totalMonthly,
		"monthly_investment":       monthlyCapacity,
		"estimated_growth_rate":    "6-8% annually",
		"key_strategies":           []string{"Dollar-cost averaging", "Automatic rebalancing", "Tax-efficient investing"},
		"next_steps":               "Review fund options, set up automatic transfers, monitor quarterly",
	}
}

func calculateCompoundGrowth(initial, monthly, returnRate float64, years int) map[string]interface{} {
	monthlyRate := returnRate / 100 / 12
	months := years * 12

	// Future value calculation with monthly contributions
	fv := initial * math.Pow(1+monthlyRate, float64(months))
	monthlyContributions := 0.0
	for i := 0; i < months; i++ {
		monthlyContributions += monthly * math.Pow(1+monthlyRate, float64(months-i-1))
	}

	total := fv + monthlyContributions
	totalContributed := initial + (monthly * float64(months))
	earnings := total - totalContributed

	return map[string]interface{}{
		"initial_investment":  initial,
		"monthly_contribution": monthly,
		"total_contributed":   totalContributed,
		"projected_earnings":  fmt.Sprintf("$%.2f", earnings),
		"projected_total":     fmt.Sprintf("$%.2f", total),
		"years":               years,
		"annual_return_rate":  fmt.Sprintf("%.1f%%", returnRate),
		"power_of_compounding": fmt.Sprintf("%.1f%% of total is earnings", (earnings/total)*100),
	}
}

func assessRiskProfile(age, yearsToRetirement int, downturnComfort, experience string) map[string]interface{} {
	riskScore := 0

	// Age-based scoring
	if age < 35 {
		riskScore += 70
	} else if age < 50 {
		riskScore += 50
	} else {
		riskScore += 30
	}

	// Market downturn comfort
	switch downturnComfort {
	case "very_uncomfortable":
		riskScore += 10
	case "somewhat_uncomfortable":
		riskScore += 25
	case "neutral":
		riskScore += 40
	case "comfortable":
		riskScore += 60
	case "very_comfortable":
		riskScore += 75
	}

	// Experience level
	switch experience {
	case "none":
		riskScore -= 20
	case "minimal":
		riskScore -= 10
	case "moderate":
		// neutral
	case "extensive":
		riskScore += 15
	}

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
		"allocation_suggestion":  getRiskAllocation(riskLevel),
		"best_fit_strategies":    getStrategiesForRisk(riskLevel),
	}
}

func explainConcept(concept string) map[string]interface{} {
	explanations := map[string]string{
		"etf": "An ETF (Exchange-Traded Fund) is like a basket of stocks bundled together. Instead of buying individual companies, you buy a tiny piece of many companies at once. It's like ordering a sampler platter instead of one dish!",
		"dividend": "A dividend is a small payment companies give to shareholders (owners). Think of it as the company saying 'thank you' for investing in us. You get paid just for holding the stock!",
		"diversification": "Diversification means not putting all your eggs in one basket. Instead of investing only in tech stocks, you spread money across different types of investments, industries, and risk levels.",
		"compound_interest": "Compound interest is when your earnings make their own earnings. Your money grows faster because you're earning 'interest on interest.' Albert Einstein called it the 8th wonder of the world!",
		"dollar_cost_averaging": "Instead of trying to time the market perfectly, you invest a fixed amount regularly (monthly). By averaging out the price over time, you reduce the risk of buying at the peak.",
	}

	if explanation, exists := explanations[concept]; exists {
		return map[string]interface{}{
			"concept":     concept,
			"explanation": explanation,
			"key_points": []string{
				"Understanding this concept helps you make better investment decisions",
				"Don't feel rushed - investing is a marathon, not a sprint",
				"Ask questions anytime - financial literacy is your superpower",
			},
		}
	}

	return map[string]interface{}{
		"concept": concept,
		"explanation": "I don't have that concept in my database, but I'd be happy to explain it! Try asking about: ETF, dividend, diversification, compound_interest, or dollar_cost_averaging.",
	}
}

func parseTimeHorizon(horizon string) int {
	switch horizon {
	case "1", "1-3":
		return 2
	case "5", "3-5":
		return 5
	case "10":
		return 10
	case "20", "20+":
		return 20
	default:
		return 10
	}
}

func getRiskAllocation(risk string) map[string]string {
	allocations := map[string]map[string]string{
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
	return allocations[risk]
}

func getStrategiesForRisk(risk string) []string {
	strategies := map[string][]string{
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
	return strategies[risk]
}

func calculateAnnualContribution(monthlyStr string) string {
	monthly, _ := strconv.ParseFloat(monthlyStr, 64)
	annual := monthly * 12
	return fmt.Sprintf("$%.2f", annual)
}

func generateRandomID() string {
	return fmt.Sprintf("%d%d%d%d%d", os.Getpid()%10, os.Getpid()%10, os.Getpid()%10, os.Getpid()%10, os.Getpid()%10)
}
