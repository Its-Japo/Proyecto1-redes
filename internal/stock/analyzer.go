package stock

import (
	"fmt"
	"math"
	"sort"
	"time"

	"proyecto-mcp-bolsa/pkg/models"
)

type Analyzer struct {
	apiClient *APIClient
}

func NewAnalyzer(apiClient *APIClient) *Analyzer {
	return &Analyzer{
		apiClient: apiClient,
	}
}

func (a *Analyzer) AnalyzePortfolio(symbols []string, timeframe string) (*models.PortfolioAnalysis, error) {
	portfolio := models.Portfolio{
		Name:    "Analysis Portfolio",
		Symbols: symbols,
		Stocks:  make([]models.Stock, 0, len(symbols)),
	}

	analyses := make([]models.StockAnalysis, 0, len(symbols))
	
	for _, symbol := range symbols {
		analysis, err := a.AnalyzeStock(symbol, timeframe)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze %s: %w", symbol, err)
		}
		
		analyses = append(analyses, *analysis)
		portfolio.Stocks = append(portfolio.Stocks, analysis.Stock)
	}

	// Calculate overall portfolio metrics
	overallScore := a.calculateOverallScore(analyses)
	overallRisk := a.calculateOverallRisk(analyses)
	recommendations := a.generatePortfolioRecommendations(analyses)

	return &models.PortfolioAnalysis{
		Portfolio:         portfolio,
		StockAnalyses:     analyses,
		OverallScore:      overallScore,
		OverallRisk:       overallRisk,
		Recommendations:   recommendations,
		GeneratedAt:       time.Now(),
	}, nil
}

func (a *Analyzer) AnalyzeStock(symbol, timeframe string) (*models.StockAnalysis, error) {
	// Get current quote
	stock, err := a.apiClient.GetQuote(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	// Get historical data for technical analysis
	timeSeries, err := a.apiClient.GetTimeSeries(symbol, timeframe)
	if err != nil {
		return nil, fmt.Errorf("failed to get time series: %w", err)
	}

	// Calculate technical indicators
	indicators, err := a.calculateTechnicalIndicators(symbol, timeSeries)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate indicators: %w", err)
	}

	// Generate recommendation
	recommendation, score, reasons := a.generateRecommendation(*stock, *indicators)
	riskLevel := a.calculateRiskLevel(*indicators, *stock)

	return &models.StockAnalysis{
		Stock:               *stock,
		TechnicalIndicators: *indicators,
		Recommendation:      recommendation,
		Score:               score,
		Reasons:             reasons,
		RiskLevel:           riskLevel,
	}, nil
}

func (a *Analyzer) calculateTechnicalIndicators(symbol string, timeSeries map[string]models.Stock) (*models.TechnicalIndicators, error) {
	if len(timeSeries) < 50 {
		return nil, fmt.Errorf("insufficient data for technical analysis (need at least 50 days)")
	}

	// Convert map to sorted slice
	dates := make([]string, 0, len(timeSeries))
	for date := range timeSeries {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	prices := make([]float64, len(dates))
	for i, date := range dates {
		prices[i] = timeSeries[date].Price
	}

	indicators := &models.TechnicalIndicators{
		Symbol: symbol,
	}

	// Calculate Simple Moving Averages
	if len(prices) >= 20 {
		indicators.SMA20 = a.calculateSMA(prices, 20)
	}
	if len(prices) >= 50 {
		indicators.SMA50 = a.calculateSMA(prices, 50)
	}

	// Calculate Exponential Moving Averages
	if len(prices) >= 12 {
		indicators.EMA12 = a.calculateEMA(prices, 12)
	}
	if len(prices) >= 26 {
		indicators.EMA26 = a.calculateEMA(prices, 26)
	}

	// Calculate MACD
	if indicators.EMA12 > 0 && indicators.EMA26 > 0 {
		indicators.MACD = indicators.EMA12 - indicators.EMA26
		// Simplified MACD signal (9-day EMA of MACD)
		indicators.MACDSignal = indicators.MACD * 0.9 // Approximation
	}

	// Calculate RSI
	if len(prices) >= 14 {
		indicators.RSI = a.calculateRSI(prices, 14)
	}

	// Calculate Volatility (standard deviation of returns)
	indicators.Volatility = a.calculateVolatility(prices)

	// Calculate Bollinger Bands
	if indicators.SMA20 > 0 {
		stdDev := a.calculateStandardDeviation(prices[len(prices)-20:])
		indicators.BollingerUpper = indicators.SMA20 + (2 * stdDev)
		indicators.BollingerLower = indicators.SMA20 - (2 * stdDev)
	}

	return indicators, nil
}

func (a *Analyzer) calculateSMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	sum := 0.0
	start := len(prices) - period
	for i := start; i < len(prices); i++ {
		sum += prices[i]
	}
	return sum / float64(period)
}

func (a *Analyzer) calculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	multiplier := 2.0 / (float64(period) + 1.0)
	ema := prices[0]

	for i := 1; i < len(prices); i++ {
		ema = (prices[i] * multiplier) + (ema * (1 - multiplier))
	}

	return ema
}

func (a *Analyzer) calculateRSI(prices []float64, period int) float64 {
	if len(prices) <= period {
		return 50 // Neutral RSI
	}

	gains := make([]float64, 0, len(prices)-1)
	losses := make([]float64, 0, len(prices)-1)

	for i := 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains = append(gains, change)
			losses = append(losses, 0)
		} else {
			gains = append(gains, 0)
			losses = append(losses, math.Abs(change))
		}
	}

	avgGain := a.calculateSMA(gains[len(gains)-period:], period)
	avgLoss := a.calculateSMA(losses[len(losses)-period:], period)

	if avgLoss == 0 {
		return 100 // No losses, maximum RSI
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))
	return rsi
}

func (a *Analyzer) calculateVolatility(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}

	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		returns[i-1] = (prices[i] - prices[i-1]) / prices[i-1]
	}

	return a.calculateStandardDeviation(returns) * math.Sqrt(252) // Annualized volatility
}

func (a *Analyzer) calculateStandardDeviation(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	sumSquaredDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(values))
	return math.Sqrt(variance)
}

func (a *Analyzer) generateRecommendation(stock models.Stock, indicators models.TechnicalIndicators) (models.Recommendation, float64, []string) {
	score := 0.0
	reasons := make([]string, 0)

	// RSI Analysis
	if indicators.RSI > 0 {
		if indicators.RSI < 30 {
			score += 2
			reasons = append(reasons, "RSI indicates oversold conditions (potential buy opportunity)")
		} else if indicators.RSI > 70 {
			score -= 2
			reasons = append(reasons, "RSI indicates overbought conditions (potential sell signal)")
		} else if indicators.RSI >= 45 && indicators.RSI <= 55 {
			score += 0.5
			reasons = append(reasons, "RSI shows neutral momentum")
		}
	}

	// Moving Average Analysis
	currentPrice := stock.Price
	if indicators.SMA20 > 0 && indicators.SMA50 > 0 {
		if currentPrice > indicators.SMA20 && indicators.SMA20 > indicators.SMA50 {
			score += 1.5
			reasons = append(reasons, "Price above both SMA20 and SMA50 (bullish trend)")
		} else if currentPrice < indicators.SMA20 && indicators.SMA20 < indicators.SMA50 {
			score -= 1.5
			reasons = append(reasons, "Price below both SMA20 and SMA50 (bearish trend)")
		}
	}

	// MACD Analysis
	if indicators.MACD != 0 && indicators.MACDSignal != 0 {
		if indicators.MACD > indicators.MACDSignal {
			score += 1
			reasons = append(reasons, "MACD above signal line (bullish momentum)")
		} else {
			score -= 1
			reasons = append(reasons, "MACD below signal line (bearish momentum)")
		}
	}

	// Bollinger Bands Analysis
	if indicators.BollingerUpper > 0 && indicators.BollingerLower > 0 {
		if currentPrice < indicators.BollingerLower {
			score += 1
			reasons = append(reasons, "Price near lower Bollinger Band (potential bounce)")
		} else if currentPrice > indicators.BollingerUpper {
			score -= 1
			reasons = append(reasons, "Price near upper Bollinger Band (potential pullback)")
		}
	}

	// Recent Performance
	if stock.ChangePerc > 5 {
		score -= 0.5
		reasons = append(reasons, "Recent strong gains may indicate short-term overvaluation")
	} else if stock.ChangePerc < -5 {
		score += 0.5
		reasons = append(reasons, "Recent decline may present buying opportunity")
	}

	// Determine recommendation based on score
	var recommendation models.Recommendation
	if score >= 3 {
		recommendation = models.StrongBuy
	} else if score >= 1 {
		recommendation = models.Buy
	} else if score >= -1 {
		recommendation = models.Hold
	} else if score >= -3 {
		recommendation = models.Sell
	} else {
		recommendation = models.StrongSell
	}

	// Normalize score to 0-100 range
	normalizedScore := math.Max(0, math.Min(100, (score+5)*10))

	return recommendation, normalizedScore, reasons
}

func (a *Analyzer) calculateRiskLevel(indicators models.TechnicalIndicators, stock models.Stock) string {
	riskScore := 0.0

	// High volatility increases risk
	if indicators.Volatility > 0.3 {
		riskScore += 2
	} else if indicators.Volatility > 0.2 {
		riskScore += 1
	}

	// Extreme RSI values indicate risk
	if indicators.RSI > 80 || indicators.RSI < 20 {
		riskScore += 1
	}

	// Large recent changes indicate risk
	if math.Abs(stock.ChangePerc) > 10 {
		riskScore += 1
	}

	if riskScore >= 3 {
		return "HIGH"
	} else if riskScore >= 1.5 {
		return "MEDIUM"
	} else {
		return "LOW"
	}
}

func (a *Analyzer) calculateOverallScore(analyses []models.StockAnalysis) float64 {
	if len(analyses) == 0 {
		return 0
	}

	sum := 0.0
	for _, analysis := range analyses {
		sum += analysis.Score
	}
	return sum / float64(len(analyses))
}

func (a *Analyzer) calculateOverallRisk(analyses []models.StockAnalysis) string {
	highRisk := 0
	mediumRisk := 0
	lowRisk := 0

	for _, analysis := range analyses {
		switch analysis.RiskLevel {
		case "HIGH":
			highRisk++
		case "MEDIUM":
			mediumRisk++
		case "LOW":
			lowRisk++
		}
	}

	if highRisk > len(analyses)/2 {
		return "HIGH"
	} else if mediumRisk+highRisk > len(analyses)/2 {
		return "MEDIUM"
	} else {
		return "LOW"
	}
}

func (a *Analyzer) generatePortfolioRecommendations(analyses []models.StockAnalysis) []string {
	recommendations := make([]string, 0)

	// Count recommendations
	buyCount := 0
	sellCount := 0
	holdCount := 0

	for _, analysis := range analyses {
		switch analysis.Recommendation {
		case models.Buy, models.StrongBuy:
			buyCount++
		case models.Sell, models.StrongSell:
			sellCount++
		case models.Hold:
			holdCount++
		}
	}

	total := len(analyses)
	if buyCount > total/2 {
		recommendations = append(recommendations, "Portfolio shows strong buy signals - consider increasing positions")
	} else if sellCount > total/2 {
		recommendations = append(recommendations, "Portfolio shows sell signals - consider reducing exposure")
	} else {
		recommendations = append(recommendations, "Mixed signals in portfolio - maintain current positions and monitor closely")
	}

	// Risk-based recommendations
	highRiskCount := 0
	for _, analysis := range analyses {
		if analysis.RiskLevel == "HIGH" {
			highRiskCount++
		}
	}

	if highRiskCount > total/3 {
		recommendations = append(recommendations, "High risk concentration detected - consider diversification")
	}

	return recommendations
}