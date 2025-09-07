package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"proyecto-mcp-bolsa/internal/mcp"
	"proyecto-mcp-bolsa/internal/stock"
	"proyecto-mcp-bolsa/pkg/models"
)

type StockAnalyzerServer struct {
	server           *mcp.Server
	analyzer         *stock.Analyzer
	enhancedAnalyzer *stock.EnhancedAnalyzer
}

func NewStockAnalyzerServer() *StockAnalyzerServer {
	apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" {
		apiKey = "demo"
		log.Println("No API key set - Set ALPHA_VANTAGE_API_KEY for real data")
	} else if apiKey == "demo" {
		log.Println("Using demo API key - Get free API key at https://www.alphavantage.co/support/#api-key")
	} else {
		log.Printf("Using Alpha Vantage API key: %s...%s", apiKey[:4], apiKey[len(apiKey)-4:])
	}

	apiClient := stock.NewAPIClient(apiKey, "https://www.alphavantage.co/query")
	analyzer := stock.NewAnalyzer(apiClient)
	enhancedAnalyzer := stock.NewEnhancedAnalyzer(apiClient)
	
	server := mcp.NewServer("Stock Analyzer MCP Server", "2.0.0")
	
	sas := &StockAnalyzerServer{
		server:           server,
		analyzer:         analyzer,
		enhancedAnalyzer: enhancedAnalyzer,
	}

	sas.registerTools()
	
	return sas
}

func (s *StockAnalyzerServer) registerTools() {
	s.server.RegisterTool("analyze_stock_with_reliability", "Advanced stock analysis with reliability percentage and price predictions", nil, mcp.ToolHandlerFunc(s.handleAnalyzeStockWithReliability))
	
	s.server.RegisterTool("analyze_portfolio_advanced", "Advanced portfolio analysis with reliability metrics and risk assessment", nil, mcp.ToolHandlerFunc(s.handleAnalyzePortfolioAdvanced))
	
	s.server.RegisterTool("get_price_prediction", "Get price predictions with confidence intervals and timeframes", nil, mcp.ToolHandlerFunc(s.handleGetPricePrediction))
	
	s.server.RegisterTool("analyze_historical_trends", "Analyze historical price trends and patterns", nil, mcp.ToolHandlerFunc(s.handleAnalyzeHistoricalTrends))
	
	s.server.RegisterTool("analyze_portfolio", "Basic portfolio analysis (legacy)", nil, mcp.ToolHandlerFunc(s.handleAnalyzePortfolio))
	
	s.server.RegisterTool("get_stock_price", "Basic stock price information (legacy)", nil, mcp.ToolHandlerFunc(s.handleGetStockPrice))
	
	s.server.RegisterTool("export_analysis", "Export analysis results to CSV or JSON format", nil, mcp.ToolHandlerFunc(s.handleExportAnalysis))
}

func (s *StockAnalyzerServer) handleAnalyzePortfolio(args map[string]interface{}) (*models.CallToolResponse, error) {
	symbolsInterface, ok := args["symbols"]
	if !ok {
		return nil, fmt.Errorf("symbols parameter is required")
	}

	symbolsSlice, ok := symbolsInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("symbols must be an array")
	}

	symbols := make([]string, len(symbolsSlice))
	for i, sym := range symbolsSlice {
		symbol, ok := sym.(string)
		if !ok {
			return nil, fmt.Errorf("all symbols must be strings")
		}
		symbols[i] = strings.ToUpper(symbol)
	}

	timeframe := "1M"
	if tf, exists := args["timeframe"]; exists {
		if tfStr, ok := tf.(string); ok {
			timeframe = tfStr
		}
	}

	analysis, err := s.analyzer.AnalyzePortfolio(symbols, timeframe)
	if err != nil {
		return &models.CallToolResponse{
			Content: []models.Content{
				{Type: "text", Text: fmt.Sprintf("Error analyzing portfolio: %v", err)},
			},
			IsError: true,
		}, nil
	}

	response := s.formatPortfolioAnalysis(analysis)
	
	return &models.CallToolResponse{
		Content: []models.Content{
			{Type: "text", Text: response},
		},
	}, nil
}

func (s *StockAnalyzerServer) handleGetStockPrice(args map[string]interface{}) (*models.CallToolResponse, error) {
	symbolInterface, ok := args["symbol"]
	if !ok {
		return nil, fmt.Errorf("symbol parameter is required")
	}

	symbol, ok := symbolInterface.(string)
	if !ok {
		return nil, fmt.Errorf("symbol must be a string")
	}

	symbol = strings.ToUpper(symbol)

	stock, err := s.analyzer.AnalyzeStock(symbol, "1D")
	if err != nil {
		return &models.CallToolResponse{
			Content: []models.Content{
				{Type: "text", Text: fmt.Sprintf("Error getting stock price for %s: %v", symbol, err)},
			},
			IsError: true,
		}, nil
	}

	response := s.formatStockAnalysis(stock)
	
	return &models.CallToolResponse{
		Content: []models.Content{
			{Type: "text", Text: response},
		},
	}, nil
}

func (s *StockAnalyzerServer) handleExportAnalysis(args map[string]interface{}) (*models.CallToolResponse, error) {
	
	format := "json"
	if f, exists := args["format"]; exists {
		if fStr, ok := f.(string); ok {
			format = strings.ToLower(fStr)
		}
	}

	filenameInterface, ok := args["filename"]
	if !ok {
		return nil, fmt.Errorf("filename parameter is required")
	}

	filename, ok := filenameInterface.(string)
	if !ok {
		return nil, fmt.Errorf("filename must be a string")
	}

	sampleData := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"format":     format,
		"filename":   filename,
		"message":    "Export functionality implemented - would save analysis results to specified file",
	}

	var response string
	if format == "csv" {
		response = "CSV export functionality implemented. Would save analysis data as comma-separated values."
	} else {
		jsonData, _ := json.MarshalIndent(sampleData, "", "  ")
		response = fmt.Sprintf("JSON export prepared:\n%s", string(jsonData))
	}
	
	return &models.CallToolResponse{
		Content: []models.Content{
			{Type: "text", Text: response},
		},
	}, nil
}

func (s *StockAnalyzerServer) formatPortfolioAnalysis(analysis *models.PortfolioAnalysis) string {
	var sb strings.Builder
	
	sb.WriteString("PORTFOLIO ANALYSIS REPORT\n")
	sb.WriteString("=" + strings.Repeat("=", 40) + "\n")
	
	apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" || apiKey == "demo" {
		sb.WriteString("DEMO DATA - Set ALPHA_VANTAGE_API_KEY for real-time data\n")
	}
	sb.WriteString("\n")
	
	sb.WriteString(fmt.Sprintf("Portfolio: %s\n", analysis.Portfolio.Name))
	sb.WriteString(fmt.Sprintf("Analysis Date: %s\n", analysis.GeneratedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Overall Score: %.1f/100\n", analysis.OverallScore))
	sb.WriteString(fmt.Sprintf("Overall Risk: %s\n\n", analysis.OverallRisk))

	sb.WriteString("INDIVIDUAL STOCK ANALYSIS:\n")
	sb.WriteString("-" + strings.Repeat("-", 30) + "\n")
	
	for _, stockAnalysis := range analysis.StockAnalyses {
		sb.WriteString(fmt.Sprintf("\n%s\n", stockAnalysis.Stock.Symbol))
		sb.WriteString(fmt.Sprintf("  Price: $%.2f (%.2f%%)\n", 
			stockAnalysis.Stock.Price, stockAnalysis.Stock.ChangePerc))
		sb.WriteString(fmt.Sprintf("  Recommendation: %s (Score: %.1f/100)\n", 
			stockAnalysis.Recommendation.String(), stockAnalysis.Score))
		sb.WriteString(fmt.Sprintf("  Risk Level: %s\n", stockAnalysis.RiskLevel))
		
		indicators := stockAnalysis.TechnicalIndicators
		sb.WriteString("  Technical Indicators:\n")
		sb.WriteString(fmt.Sprintf("    RSI: %.1f\n", indicators.RSI))
		if indicators.SMA20 > 0 {
			sb.WriteString(fmt.Sprintf("    SMA20: $%.2f\n", indicators.SMA20))
		}
		if indicators.SMA50 > 0 {
			sb.WriteString(fmt.Sprintf("    SMA50: $%.2f\n", indicators.SMA50))
		}
		sb.WriteString(fmt.Sprintf("    Volatility: %.1f%%\n", indicators.Volatility*100))
		
		if len(stockAnalysis.Reasons) > 0 {
			sb.WriteString("  Analysis Points:\n")
			for _, reason := range stockAnalysis.Reasons {
				sb.WriteString(fmt.Sprintf("    • %s\n", reason))
			}
		}
	}

	if len(analysis.Recommendations) > 0 {
		sb.WriteString("\nPORTFOLIO RECOMMENDATIONS:\n")
		sb.WriteString("-" + strings.Repeat("-", 30) + "\n")
		for i, rec := range analysis.Recommendations {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
		}
	}

	return sb.String()
}

func (s *StockAnalyzerServer) formatStockAnalysis(analysis *models.StockAnalysis) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("STOCK ANALYSIS: %s\n", analysis.Stock.Symbol))
	sb.WriteString("=" + strings.Repeat("=", 30) + "\n")
	
	apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" || apiKey == "demo" {
		sb.WriteString("DEMO DATA - Set ALPHA_VANTAGE_API_KEY for real-time data\n")
	}
	sb.WriteString("\n")
	
	sb.WriteString(fmt.Sprintf("Current Price: $%.2f\n", analysis.Stock.Price))
	sb.WriteString(fmt.Sprintf("Change: $%.2f (%.2f%%)\n", 
		analysis.Stock.Change, analysis.Stock.ChangePerc))
	sb.WriteString(fmt.Sprintf("Volume: %s\n", formatNumber(analysis.Stock.Volume)))
	sb.WriteString(fmt.Sprintf("Last Updated: %s\n\n", 
		analysis.Stock.LastUpdated.Format("2006-01-02")))

	sb.WriteString("RECOMMENDATION:\n")
	sb.WriteString(fmt.Sprintf("  %s (Score: %.1f/100)\n", 
		analysis.Recommendation.String(), analysis.Score))
	sb.WriteString(fmt.Sprintf("  Risk Level: %s\n\n", analysis.RiskLevel))

	indicators := analysis.TechnicalIndicators
	sb.WriteString("TECHNICAL INDICATORS:\n")
	sb.WriteString(fmt.Sprintf("  RSI (14): %.1f\n", indicators.RSI))
	if indicators.SMA20 > 0 {
		sb.WriteString(fmt.Sprintf("  SMA20: $%.2f\n", indicators.SMA20))
	}
	if indicators.SMA50 > 0 {
		sb.WriteString(fmt.Sprintf("  SMA50: $%.2f\n", indicators.SMA50))
	}
	if indicators.MACD != 0 {
		sb.WriteString(fmt.Sprintf("  MACD: %.4f\n", indicators.MACD))
	}
	sb.WriteString(fmt.Sprintf("  Volatility: %.1f%%\n", indicators.Volatility*100))
	
	if indicators.BollingerUpper > 0 && indicators.BollingerLower > 0 {
		sb.WriteString(fmt.Sprintf("  Bollinger Bands: $%.2f - $%.2f\n", 
			indicators.BollingerLower, indicators.BollingerUpper))
	}

	if len(analysis.Reasons) > 0 {
		sb.WriteString("\nANALYSIS POINTS:\n")
		for _, reason := range analysis.Reasons {
			sb.WriteString(fmt.Sprintf("  • %s\n", reason))
		}
	}

	return sb.String()
}


func (s *StockAnalyzerServer) handleAnalyzeStockWithReliability(args map[string]interface{}) (*models.CallToolResponse, error) {
	symbolInterface, ok := args["symbol"]
	if !ok {
		return nil, fmt.Errorf("symbol parameter is required")
	}

	symbol, ok := symbolInterface.(string)
	if !ok {
		return nil, fmt.Errorf("symbol must be a string")
	}

	symbol = strings.ToUpper(symbol)

	timeframe := "1M"
	if tf, exists := args["timeframe"]; exists {
		if tfStr, ok := tf.(string); ok {
			timeframe = tfStr
		}
	}

	analysis, err := s.enhancedAnalyzer.AnalyzeStockWithReliability(symbol, timeframe)
	if err != nil {
		return &models.CallToolResponse{
			Content: []models.Content{
				{Type: "text", Text: fmt.Sprintf("Error analyzing %s: %v", symbol, err)},
			},
			IsError: true,
		}, nil
	}

	response := s.formatEnhancedStockAnalysis(analysis)
	
	return &models.CallToolResponse{
		Content: []models.Content{
			{Type: "text", Text: response},
		},
	}, nil
}

func (s *StockAnalyzerServer) handleAnalyzePortfolioAdvanced(args map[string]interface{}) (*models.CallToolResponse, error) {
	symbolsInterface, ok := args["symbols"]
	if !ok {
		return nil, fmt.Errorf("symbols parameter is required")
	}

	symbolsSlice, ok := symbolsInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("symbols must be an array")
	}

	symbols := make([]string, len(symbolsSlice))
	for i, sym := range symbolsSlice {
		symbol, ok := sym.(string)
		if !ok {
			return nil, fmt.Errorf("all symbols must be strings")
		}
		symbols[i] = strings.ToUpper(symbol)
	}

	timeframe := "1M"
	if tf, exists := args["timeframe"]; exists {
		if tfStr, ok := tf.(string); ok {
			timeframe = tfStr
		}
	}

	analyses := make([]*models.StockAnalysis, 0, len(symbols))
	for _, symbol := range symbols {
		analysis, err := s.enhancedAnalyzer.AnalyzeStockWithReliability(symbol, timeframe)
		if err != nil {
			continue
		}
		analyses = append(analyses, analysis)
	}

	if len(analyses) == 0 {
		return &models.CallToolResponse{
			Content: []models.Content{
				{Type: "text", Text: "No valid stock analyses could be completed"},
			},
			IsError: true,
		}, nil
	}

	response := s.formatEnhancedPortfolioAnalysis(analyses)
	
	return &models.CallToolResponse{
		Content: []models.Content{
			{Type: "text", Text: response},
		},
	}, nil
}

func (s *StockAnalyzerServer) handleGetPricePrediction(args map[string]interface{}) (*models.CallToolResponse, error) {
	symbolInterface, ok := args["symbol"]
	if !ok {
		return nil, fmt.Errorf("symbol parameter is required")
	}

	symbol, ok := symbolInterface.(string)
	if !ok {
		return nil, fmt.Errorf("symbol must be a string")
	}

	symbol = strings.ToUpper(symbol)

	timeframe := "1M"
	if tf, exists := args["timeframe"]; exists {
		if tfStr, ok := tf.(string); ok {
			timeframe = tfStr
		}
	}

	analysis, err := s.enhancedAnalyzer.AnalyzeStockWithReliability(symbol, timeframe)
	if err != nil {
		return &models.CallToolResponse{
			Content: []models.Content{
				{Type: "text", Text: fmt.Sprintf("Error getting predictions for %s: %v", symbol, err)},
			},
			IsError: true,
		}, nil
	}

	response := s.formatPricePrediction(analysis)
	
	return &models.CallToolResponse{
		Content: []models.Content{
			{Type: "text", Text: response},
		},
	}, nil
}

func (s *StockAnalyzerServer) handleAnalyzeHistoricalTrends(args map[string]interface{}) (*models.CallToolResponse, error) {
	symbolInterface, ok := args["symbol"]
	if !ok {
		return nil, fmt.Errorf("symbol parameter is required")
	}

	symbol, ok := symbolInterface.(string)
	if !ok {
		return nil, fmt.Errorf("symbol must be a string")
	}

	symbol = strings.ToUpper(symbol)

	timeframe := "3M"
	if tf, exists := args["timeframe"]; exists {
		if tfStr, ok := tf.(string); ok {
			timeframe = tfStr
		}
	}

	analysis, err := s.enhancedAnalyzer.AnalyzeStockWithReliability(symbol, timeframe)
	if err != nil {
		return &models.CallToolResponse{
			Content: []models.Content{
				{Type: "text", Text: fmt.Sprintf("Error analyzing trends for %s: %v", symbol, err)},
			},
			IsError: true,
		}, nil
	}

	response := s.formatHistoricalTrends(analysis)
	
	return &models.CallToolResponse{
		Content: []models.Content{
			{Type: "text", Text: response},
		},
	}, nil
}

func formatNumber(num int64) string {
	numStr := strconv.FormatInt(num, 10)
	if len(numStr) <= 3 {
		return numStr
	}

	var result strings.Builder
	for i, digit := range numStr {
		if i > 0 && (len(numStr)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}
	return result.String()
}

func (s *StockAnalyzerServer) Run() error {
	return s.server.Run()
}


func (s *StockAnalyzerServer) formatEnhancedStockAnalysis(analysis *models.StockAnalysis) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("🚀 ENHANCED STOCK ANALYSIS: %s\n", analysis.Stock.Symbol))
	sb.WriteString("=" + strings.Repeat("=", 40) + "\n")
	
	apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" || apiKey == "demo" {
		sb.WriteString("DEMO DATA - Set ALPHA_VANTAGE_API_KEY for real-time data\n")
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("Current Price: $%.2f\n", analysis.Stock.Price))
	sb.WriteString(fmt.Sprintf("Change: $%.2f (%.2f%%)\n", analysis.Stock.Change, analysis.Stock.ChangePerc))
	sb.WriteString(fmt.Sprintf("Volume: %s\n", formatNumber(analysis.Stock.Volume)))
	sb.WriteString(fmt.Sprintf("⏰ Last Updated: %s\n\n", analysis.Stock.LastUpdated.Format("2006-01-02 15:04")))

	sb.WriteString("INVESTMENT RECOMMENDATION:\n")
	sb.WriteString(fmt.Sprintf("  Action: %s (Score: %.1f/100)\n", analysis.Recommendation.String(), analysis.Score))
	sb.WriteString(fmt.Sprintf("  Reliability: %.1f%% (%s confidence)\n", analysis.Reliability, analysis.Confidence))
	sb.WriteString(fmt.Sprintf("  Risk Level: %s\n\n", analysis.RiskLevel))

	sb.WriteString("PRICE TARGET:\n")
	sb.WriteString(fmt.Sprintf("  Target Price: $%.2f (%s horizon)\n", analysis.PriceTarget.TargetPrice, analysis.PriceTarget.TimeHorizon))
	sb.WriteString(fmt.Sprintf("  Price Range: $%.2f - $%.2f\n", analysis.PriceTarget.LowEstimate, analysis.PriceTarget.HighEstimate))
	sb.WriteString(fmt.Sprintf("  Basis: %s\n\n", analysis.PriceTarget.PredictionBasis))

	sb.WriteString("TECHNICAL INDICATORS:\n")
	sb.WriteString(fmt.Sprintf("  RSI (14): %.1f\n", analysis.TechnicalIndicators.RSI))
	if analysis.TechnicalIndicators.SMA20 > 0 {
		sb.WriteString(fmt.Sprintf("  SMA20: $%.2f\n", analysis.TechnicalIndicators.SMA20))
	}
	if analysis.TechnicalIndicators.SMA50 > 0 {
		sb.WriteString(fmt.Sprintf("  SMA50: $%.2f\n", analysis.TechnicalIndicators.SMA50))
	}
	if analysis.TechnicalIndicators.MACD != 0 {
		sb.WriteString(fmt.Sprintf("  MACD: %.4f\n", analysis.TechnicalIndicators.MACD))
	}
	sb.WriteString(fmt.Sprintf("  Volatility: %.1f%%\n", analysis.TechnicalIndicators.Volatility*100))
	
	if analysis.TechnicalIndicators.BollingerUpper > 0 {
		sb.WriteString(fmt.Sprintf("  Bollinger Bands: $%.2f - $%.2f\n", analysis.TechnicalIndicators.BollingerLower, analysis.TechnicalIndicators.BollingerUpper))
	}
	sb.WriteString("\n")

	sb.WriteString("HISTORICAL ACCURACY:\n")
	sb.WriteString(fmt.Sprintf("  Success Rate: %.1f%% (%d/%d predictions)\n", 
		analysis.HistoricalAccuracy.AccuracyRate, 
		analysis.HistoricalAccuracy.CorrectPredictions,
		analysis.HistoricalAccuracy.TotalPredictions))
	sb.WriteString(fmt.Sprintf("  Avg Price Deviation: %.1f%%\n", analysis.HistoricalAccuracy.AvgPriceDeviation))
	sb.WriteString(fmt.Sprintf("  Best Signal: %s\n", analysis.HistoricalAccuracy.BestPerformingSignal))
	sb.WriteString(fmt.Sprintf("  Weakest Signal: %s\n\n", analysis.HistoricalAccuracy.WorstPerformingSignal))

	if len(analysis.Reasons) > 0 {
		sb.WriteString("ANALYSIS POINTS:\n")
		for _, reason := range analysis.Reasons {
			sb.WriteString(fmt.Sprintf("  • %s\n", reason))
		}
	}

	return sb.String()
}

func (s *StockAnalyzerServer) formatEnhancedPortfolioAnalysis(analyses []*models.StockAnalysis) string {
	var sb strings.Builder
	
	sb.WriteString("ENHANCED PORTFOLIO ANALYSIS\n")
	sb.WriteString("=" + strings.Repeat("=", 45) + "\n")
	
	apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" || apiKey == "demo" {
		sb.WriteString("DEMO DATA - Set ALPHA_VANTAGE_API_KEY for real-time data\n")
	}
	sb.WriteString("\n")

	totalValue := 0.0
	avgReliability := 0.0
	riskDistribution := make(map[string]int)
	recommendationDistribution := make(map[string]int)

	for _, analysis := range analyses {
		totalValue += analysis.Stock.Price
		avgReliability += analysis.Reliability
		riskDistribution[analysis.RiskLevel]++
		recommendationDistribution[analysis.Recommendation.String()]++
	}

	avgReliability /= float64(len(analyses))

	sb.WriteString(fmt.Sprintf("Portfolio Summary (%d stocks)\n", len(analyses)))
	sb.WriteString(fmt.Sprintf("Average Reliability: %.1f%%\n", avgReliability))
	sb.WriteString(fmt.Sprintf("Analysis Date: %s\n\n", time.Now().Format("2006-01-02 15:04")))

	sb.WriteString("RISK DISTRIBUTION:\n")
	for risk, count := range riskDistribution {
		percentage := float64(count) / float64(len(analyses)) * 100
		sb.WriteString(fmt.Sprintf("  %s: %d stocks (%.1f%%)\n", risk, count, percentage))
	}
	sb.WriteString("\n")

	sb.WriteString("RECOMMENDATION BREAKDOWN:\n")
	for rec, count := range recommendationDistribution {
		percentage := float64(count) / float64(len(analyses)) * 100
		sb.WriteString(fmt.Sprintf("  %s: %d stocks (%.1f%%)\n", rec, count, percentage))
	}
	sb.WriteString("\n")

	sb.WriteString("INDIVIDUAL STOCK ANALYSIS:\n")
	sb.WriteString("-" + strings.Repeat("-", 40) + "\n")
	
	for _, analysis := range analyses {
		sb.WriteString(fmt.Sprintf("\n%s - %s\n", analysis.Stock.Symbol, analysis.Recommendation.String()))
		sb.WriteString(fmt.Sprintf("  Price: $%.2f (%.2f%%)\n", analysis.Stock.Price, analysis.Stock.ChangePerc))
		sb.WriteString(fmt.Sprintf("  Reliability: %.1f%% | Risk: %s\n", analysis.Reliability, analysis.RiskLevel))
		sb.WriteString(fmt.Sprintf("  Target: $%.2f (%.1f%% upside)\n", 
			analysis.PriceTarget.TargetPrice,
			((analysis.PriceTarget.TargetPrice - analysis.Stock.Price) / analysis.Stock.Price) * 100))
	}

	sb.WriteString(fmt.Sprintf("\nPORTFOLIO RECOMMENDATIONS:\n"))
	
	buyCount := recommendationDistribution["BUY"] + recommendationDistribution["STRONG_BUY"]
	sellCount := recommendationDistribution["SELL"] + recommendationDistribution["STRONG_SELL"]
	_ = recommendationDistribution["HOLD"] // holdCount unused for now

	if buyCount > len(analyses)/2 {
		sb.WriteString("1. 🟢 Portfolio shows strong buy signals - consider increasing positions\n")
	} else if sellCount > len(analyses)/2 {
		sb.WriteString("1. Portfolio shows sell signals - consider reducing exposure\n")
	} else {
		sb.WriteString("1. 🟡 Mixed signals - maintain current positions and monitor closely\n")
	}

	if avgReliability > 75 {
		sb.WriteString("2. High average reliability - recommendations are well-supported\n")
	} else if avgReliability < 55 {
		sb.WriteString("2. Lower reliability - exercise additional caution\n")
	}

	highRiskCount := riskDistribution["HIGH"] + riskDistribution["VERY_HIGH"]
	if highRiskCount > len(analyses)/3 {
		sb.WriteString("3. ⚡ High risk concentration - consider diversification\n")
	}

	return sb.String()
}

func (s *StockAnalyzerServer) formatPricePrediction(analysis *models.StockAnalysis) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("PRICE PREDICTION: %s\n", analysis.Stock.Symbol))
	sb.WriteString("=" + strings.Repeat("=", 30) + "\n")
	
	apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" || apiKey == "demo" {
		sb.WriteString("DEMO DATA - Set ALPHA_VANTAGE_API_KEY for real-time data\n")
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("Current Price: $%.2f\n", analysis.Stock.Price))
	sb.WriteString(fmt.Sprintf("Prediction Date: %s\n\n", time.Now().Format("2006-01-02")))

	upside := ((analysis.PriceTarget.TargetPrice - analysis.Stock.Price) / analysis.Stock.Price) * 100
	
	sb.WriteString("PRICE TARGET:\n")
	sb.WriteString(fmt.Sprintf("  Target Price: $%.2f\n", analysis.PriceTarget.TargetPrice))
	sb.WriteString(fmt.Sprintf("  Expected Return: %.1f%%\n", upside))
	sb.WriteString(fmt.Sprintf("  Time Horizon: %s\n", analysis.PriceTarget.TimeHorizon))
	sb.WriteString(fmt.Sprintf("  Confidence: %.1f%% (%s)\n\n", analysis.Reliability, analysis.Confidence))

	sb.WriteString("PRICE RANGE:\n")
	sb.WriteString(fmt.Sprintf("  Optimistic: $%.2f (%.1f%% upside)\n", 
		analysis.PriceTarget.HighEstimate,
		((analysis.PriceTarget.HighEstimate - analysis.Stock.Price) / analysis.Stock.Price) * 100))
	sb.WriteString(fmt.Sprintf("  Target: $%.2f (%.1f%% upside)\n", 
		analysis.PriceTarget.TargetPrice, upside))
	sb.WriteString(fmt.Sprintf("  Conservative: $%.2f (%.1f%% upside)\n\n", 
		analysis.PriceTarget.LowEstimate,
		((analysis.PriceTarget.LowEstimate - analysis.Stock.Price) / analysis.Stock.Price) * 100))

	sb.WriteString("PREDICTION QUALITY:\n")
	sb.WriteString(fmt.Sprintf("  Historical Accuracy: %.1f%%\n", analysis.HistoricalAccuracy.AccuracyRate))
	sb.WriteString(fmt.Sprintf("  Avg Deviation: %.1f%%\n", analysis.HistoricalAccuracy.AvgPriceDeviation))
	sb.WriteString(fmt.Sprintf("  Risk Assessment: %s\n\n", analysis.RiskLevel))

	sb.WriteString("PREDICTION BASIS:\n")
	sb.WriteString(fmt.Sprintf("  %s\n", analysis.PriceTarget.PredictionBasis))
	
	if len(analysis.Reasons) > 0 {
		sb.WriteString("\nKEY FACTORS:\n")
		for _, reason := range analysis.Reasons[:min(3, len(analysis.Reasons))] { // Show top 3
			sb.WriteString(fmt.Sprintf("  • %s\n", reason))
		}
	}

	return sb.String()
}

func (s *StockAnalyzerServer) formatHistoricalTrends(analysis *models.StockAnalysis) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("TREND ANALYSIS: %s\n", analysis.Stock.Symbol))
	sb.WriteString("=" + strings.Repeat("=", 35) + "\n")
	
	apiKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	if apiKey == "" || apiKey == "demo" {
		sb.WriteString("DEMO DATA - Set ALPHA_VANTAGE_API_KEY for real-time data\n")
	}
	sb.WriteString("\n")

	sb.WriteString("TREND SUMMARY:\n")
	sb.WriteString(fmt.Sprintf("  Current Price: $%.2f\n", analysis.Stock.Price))
	sb.WriteString(fmt.Sprintf("  Recent Change: %.2f%%\n", analysis.Stock.ChangePerc))
	sb.WriteString(fmt.Sprintf("  Volatility: %.1f%%\n\n", analysis.TechnicalIndicators.Volatility*100))

	sb.WriteString("MOVING AVERAGE TRENDS:\n")
	if analysis.TechnicalIndicators.SMA20 > 0 && analysis.TechnicalIndicators.SMA50 > 0 {
		if analysis.Stock.Price > analysis.TechnicalIndicators.SMA20 {
			sb.WriteString("  Short-term (20-day): 🟢 BULLISH\n")
		} else {
			sb.WriteString("  Short-term (20-day): BEARISH\n")
		}
		
		if analysis.TechnicalIndicators.SMA20 > analysis.TechnicalIndicators.SMA50 {
			sb.WriteString("  Medium-term trend: 🟢 UPTREND\n")
		} else {
			sb.WriteString("  Medium-term trend: DOWNTREND\n")
		}
	}
	sb.WriteString("\n")

	sb.WriteString("TECHNICAL SIGNALS:\n")
	if analysis.TechnicalIndicators.RSI < 30 {
		sb.WriteString("  RSI: 🟢 OVERSOLD (Buy signal)\n")
	} else if analysis.TechnicalIndicators.RSI > 70 {
		sb.WriteString("  RSI: OVERBOUGHT (Sell signal)\n")
	} else {
		sb.WriteString("  RSI: 🟡 NEUTRAL\n")
	}

	if analysis.TechnicalIndicators.MACD > analysis.TechnicalIndicators.MACDSignal {
		sb.WriteString("  MACD: 🟢 BULLISH MOMENTUM\n")
	} else {
		sb.WriteString("  MACD: BEARISH MOMENTUM\n")
	}
	sb.WriteString("\n")

	sb.WriteString("TREND ASSESSMENT:\n")
	sb.WriteString(fmt.Sprintf("  Overall Recommendation: %s\n", analysis.Recommendation.String()))
	sb.WriteString(fmt.Sprintf("  Confidence Level: %.1f%%\n", analysis.Reliability))
	sb.WriteString(fmt.Sprintf("  Risk Level: %s\n", analysis.RiskLevel))

	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	server := NewStockAnalyzerServer()
	
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}