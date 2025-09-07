package stock

import (
	"fmt"
	"math"
	"sort"
	"time"

	"proyecto-mcp-bolsa/pkg/models"
)

type EnhancedAnalyzer struct {
	apiClient       *APIClient
	historicalData  map[string]models.PriceHistory
	predictionCache map[string]models.StockAnalysis
}

func NewEnhancedAnalyzer(apiClient *APIClient) *EnhancedAnalyzer {
	return &EnhancedAnalyzer{
		apiClient:       apiClient,
		historicalData:  make(map[string]models.PriceHistory),
		predictionCache: make(map[string]models.StockAnalysis),
	}
}

func (e *EnhancedAnalyzer) AnalyzeStockWithReliability(symbol, timeframe string) (*models.StockAnalysis, error) {
	stock, err := e.apiClient.GetQuote(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock quote: %w", err)
	}

	priceHistory, err := e.buildPriceHistory(symbol, timeframe)
	if err != nil {
		return nil, fmt.Errorf("failed to build price history: %w", err)
	}

	indicators := e.calculateEnhancedIndicators(priceHistory)

	trends := e.analyzeTrends(priceHistory)

	patterns := e.detectPatterns(priceHistory)

	recommendation, score, reliability, confidence, reasons := e.generateReliableRecommendation(*stock, indicators, trends, patterns)

	priceTarget := e.calculatePriceTarget(*stock, trends, patterns, timeframe)

	historicalAccuracy := e.calculateHistoricalAccuracy(symbol)

	riskLevel := e.calculateAdvancedRiskLevel(indicators, trends, *stock)

	return &models.StockAnalysis{
		Stock:               *stock,
		TechnicalIndicators: indicators,
		Recommendation:      recommendation,
		Score:               score,
		Reliability:         reliability,
		Confidence:          confidence,
		Reasons:             reasons,
		RiskLevel:           riskLevel,
		PriceTarget:         priceTarget,
		HistoricalAccuracy:  historicalAccuracy,
	}, nil
}

func (e *EnhancedAnalyzer) buildPriceHistory(symbol, timeframe string) (models.PriceHistory, error) {
	if cached, exists := e.historicalData[symbol]; exists {
		if time.Since(cached.DataPoints[0].Date) < 1*time.Hour {
			return cached, nil
		}
	}

	timeSeries, err := e.apiClient.GetTimeSeries(symbol, timeframe)
	if err != nil {
		return models.PriceHistory{}, err
	}

	dataPoints := make([]models.PriceDataPoint, 0, len(timeSeries))
	for date, stockData := range timeSeries {
		parsedDate, _ := time.Parse("2006-01-02", date)
		dataPoints = append(dataPoints, models.PriceDataPoint{
			Date:       parsedDate,
			Price:      stockData.Price,
			Volume:     stockData.Volume,
			Change:     stockData.Change,
			ChangePerc: stockData.ChangePerc,
		})
	}

	sort.Slice(dataPoints, func(i, j int) bool {
		return dataPoints[i].Date.After(dataPoints[j].Date)
	})

	priceHistory := models.PriceHistory{
		Symbol:     symbol,
		Timeframe:  timeframe,
		DataPoints: dataPoints,
	}

	e.historicalData[symbol] = priceHistory

	return priceHistory, nil
}

func (e *EnhancedAnalyzer) analyzeTrends(history models.PriceHistory) models.TrendAnalysis {
	if len(history.DataPoints) < 50 {
		return models.TrendAnalysis{}
	}

	prices := make([]float64, len(history.DataPoints))
	for i, point := range history.DataPoints {
		prices[i] = point.Price
	}

	shortTrend := e.calculateTrendDirection(prices[:5])
	mediumTrend := e.calculateTrendDirection(prices[:20])
	longTrend := e.calculateTrendDirection(prices[:50])

	support, resistance := e.calculateSupportResistance(prices)

	trendStrength := e.calculateTrendStrength(prices)

	return models.TrendAnalysis{
		ShortTerm:     shortTrend,
		MediumTerm:    mediumTrend,
		LongTerm:      longTrend,
		Support:       support,
		Resistance:    resistance,
		TrendStrength: trendStrength,
	}
}

func (e *EnhancedAnalyzer) calculateTrendDirection(prices []float64) models.TrendDirection {
	if len(prices) < 2 {
		return models.Sideways
	}

	n := float64(len(prices))
	var sumX, sumY, sumXY, sumX2 float64

	for i, price := range prices {
		x := float64(i)
		sumX += x
		sumY += price
		sumXY += x * price
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	
	startPrice := prices[len(prices)-1]
	endPrice := prices[0]
	percentChange := ((endPrice - startPrice) / startPrice) * 100

	if slope > 0.5 && percentChange > 5 {
		return models.StronglyBullish
	} else if slope > 0.1 && percentChange > 1 {
		return models.Bullish
	} else if slope < -0.5 && percentChange < -5 {
		return models.StronglyBearish
	} else if slope < -0.1 && percentChange < -1 {
		return models.Bearish
	} else {
		return models.Sideways
	}
}

func (e *EnhancedAnalyzer) calculateSupportResistance(prices []float64) (float64, float64) {
	if len(prices) < 10 {
		return 0, 0
	}

	minPrices := make([]float64, 0)
	maxPrices := make([]float64, 0)

	for i := 1; i < len(prices)-1; i++ {
		if prices[i] < prices[i-1] && prices[i] < prices[i+1] {
			minPrices = append(minPrices, prices[i])
		}
		if prices[i] > prices[i-1] && prices[i] > prices[i+1] {
			maxPrices = append(maxPrices, prices[i])
		}
	}

	var support float64
	if len(minPrices) > 0 {
		sum := 0.0
		for _, price := range minPrices {
			sum += price
		}
		support = sum / float64(len(minPrices))
	}

	var resistance float64
	if len(maxPrices) > 0 {
		sum := 0.0
		for _, price := range maxPrices {
			sum += price
		}
		resistance = sum / float64(len(maxPrices))
	}

	return support, resistance
}

func (e *EnhancedAnalyzer) calculateTrendStrength(prices []float64) float64 {
	if len(prices) < 10 {
		return 50.0
	}

	n := float64(len(prices))
	var sumX, sumY, sumXY, sumX2, sumY2 float64

	for i, price := range prices {
		x := float64(i)
		sumX += x
		sumY += price
		sumXY += x * price
		sumX2 += x * x
		sumY2 += price * price
	}

	correlation := (n*sumXY - sumX*sumY) / math.Sqrt((n*sumX2-sumX*sumX)*(n*sumY2-sumY*sumY))
	rSquared := correlation * correlation

	return rSquared * 100
}

func (e *EnhancedAnalyzer) detectPatterns(history models.PriceHistory) []models.PatternMatch {
	patterns := make([]models.PatternMatch, 0)

	if len(history.DataPoints) < 20 {
		return patterns
	}

	prices := make([]float64, len(history.DataPoints))
	for i, point := range history.DataPoints {
		prices[i] = point.Price
	}

	patterns = append(patterns, e.detectHeadAndShoulders(prices, history.DataPoints)...)
	patterns = append(patterns, e.detectDoubleTop(prices, history.DataPoints)...)
	patterns = append(patterns, e.detectDoubleBottom(prices, history.DataPoints)...)
	patterns = append(patterns, e.detectTriangle(prices, history.DataPoints)...)

	return patterns
}

func (e *EnhancedAnalyzer) detectHeadAndShoulders(prices []float64, dataPoints []models.PriceDataPoint) []models.PatternMatch {
	patterns := make([]models.PatternMatch, 0)
	
	if len(prices) < 15 {
		return patterns
	}

	for i := 5; i < len(prices)-10; i++ {
		leftShoulder := prices[i-5:i]
		head := prices[i:i+5]
		rightShoulder := prices[i+5:i+10]

		leftPeak := e.findMaxInSlice(leftShoulder)
		headPeak := e.findMaxInSlice(head)
		rightPeak := e.findMaxInSlice(rightShoulder)

		if headPeak > leftPeak && headPeak > rightPeak {
			shoulderDiff := math.Abs(leftPeak-rightPeak) / leftPeak
			if shoulderDiff < 0.05 {
				patterns = append(patterns, models.PatternMatch{
					Pattern:     "HEAD_AND_SHOULDERS",
					Confidence:  70.0,
					Timeframe:   "15D",
					StartDate:   dataPoints[i-5].Date,
					EndDate:     dataPoints[i+10].Date,
					Implication: "BEARISH",
					Reliability: 68.0,
				})
				break
			}
		}
	}

	return patterns
}

func (e *EnhancedAnalyzer) detectDoubleTop(prices []float64, dataPoints []models.PriceDataPoint) []models.PatternMatch {
	patterns := make([]models.PatternMatch, 0)
	
	if len(prices) < 10 {
		return patterns
	}

	for i := 5; i < len(prices)-5; i++ {
		leftPeak := e.findMaxInSlice(prices[i-5:i])
		rightPeak := e.findMaxInSlice(prices[i:i+5])
		valley := e.findMinInSlice(prices[i-2:i+2])

		peakDiff := math.Abs(leftPeak-rightPeak) / leftPeak
		if peakDiff < 0.03 && valley < leftPeak*0.95 {
			patterns = append(patterns, models.PatternMatch{
				Pattern:     "DOUBLE_TOP",
				Confidence:  65.0,
				Timeframe:   "10D",
				StartDate:   dataPoints[i-5].Date,
				EndDate:     dataPoints[i+5].Date,
				Implication: "BEARISH",
				Reliability: 72.0,
			})
			break
		}
	}

	return patterns
}

func (e *EnhancedAnalyzer) detectDoubleBottom(prices []float64, dataPoints []models.PriceDataPoint) []models.PatternMatch {
	patterns := make([]models.PatternMatch, 0)
	
	if len(prices) < 10 {
		return patterns
	}

	for i := 5; i < len(prices)-5; i++ {
		leftBottom := e.findMinInSlice(prices[i-5:i])
		rightBottom := e.findMinInSlice(prices[i:i+5])
		peak := e.findMaxInSlice(prices[i-2:i+2])

		bottomDiff := math.Abs(leftBottom-rightBottom) / leftBottom
		if bottomDiff < 0.03 && peak > leftBottom*1.05 {
			patterns = append(patterns, models.PatternMatch{
				Pattern:     "DOUBLE_BOTTOM",
				Confidence:  65.0,
				Timeframe:   "10D",
				StartDate:   dataPoints[i-5].Date,
				EndDate:     dataPoints[i+5].Date,
				Implication: "BULLISH",
				Reliability: 74.0,
			})
			break
		}
	}

	return patterns
}

func (e *EnhancedAnalyzer) detectTriangle(prices []float64, dataPoints []models.PriceDataPoint) []models.PatternMatch {
	patterns := make([]models.PatternMatch, 0)
	
	if len(prices) < 15 {
		return patterns
	}

	highs := make([]float64, 0)
	lows := make([]float64, 0)

	for i := 1; i < len(prices)-1; i++ {
		if prices[i] > prices[i-1] && prices[i] > prices[i+1] {
			highs = append(highs, prices[i])
		}
		if prices[i] < prices[i-1] && prices[i] < prices[i+1] {
			lows = append(lows, prices[i])
		}
	}

	if len(highs) >= 3 && len(lows) >= 3 {
		highsSlope := e.calculateSlope(highs)
		lowsSlope := e.calculateSlope(lows)

		if highsSlope < -0.1 && lowsSlope > 0.1 {
			patterns = append(patterns, models.PatternMatch{
				Pattern:     "SYMMETRICAL_TRIANGLE",
				Confidence:  60.0,
				Timeframe:   "15D",
				StartDate:   dataPoints[15].Date,
				EndDate:     dataPoints[0].Date,
				Implication: "NEUTRAL",
				Reliability: 58.0,
			})
		}
	}

	return patterns
}

func (e *EnhancedAnalyzer) findMaxInSlice(slice []float64) float64 {
	if len(slice) == 0 {
		return 0
	}
	max := slice[0]
	for _, val := range slice {
		if val > max {
			max = val
		}
	}
	return max
}

func (e *EnhancedAnalyzer) findMinInSlice(slice []float64) float64 {
	if len(slice) == 0 {
		return 0
	}
	min := slice[0]
	for _, val := range slice {
		if val < min {
			min = val
		}
	}
	return min
}

func (e *EnhancedAnalyzer) calculateSlope(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	n := float64(len(values))
	var sumX, sumY, sumXY, sumX2 float64

	for i, val := range values {
		x := float64(i)
		sumX += x
		sumY += val
		sumXY += x * val
		sumX2 += x * x
	}

	return (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
}