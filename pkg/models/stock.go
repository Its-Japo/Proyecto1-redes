package models

import "time"

type Stock struct {
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Change      float64   `json:"change"`
	ChangePerc  float64   `json:"changePerc"`
	Volume      int64     `json:"volume"`
	MarketCap   int64     `json:"marketCap"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type Portfolio struct {
	Name    string   `json:"name"`
	Symbols []string `json:"symbols"`
	Stocks  []Stock  `json:"stocks"`
}

type TechnicalIndicators struct {
	Symbol          string  `json:"symbol"`
	RSI             float64 `json:"rsi"`
	SMA20           float64 `json:"sma20"`
	SMA50           float64 `json:"sma50"`
	EMA12           float64 `json:"ema12"`
	EMA26           float64 `json:"ema26"`
	MACD            float64 `json:"macd"`
	MACDSignal      float64 `json:"macdSignal"`
	Volatility      float64 `json:"volatility"`
	BollingerUpper  float64 `json:"bollingerUpper"`
	BollingerLower  float64 `json:"bollingerLower"`
}

type StockAnalysis struct {
	Stock               Stock               `json:"stock"`
	TechnicalIndicators TechnicalIndicators `json:"technicalIndicators"`
	Recommendation      Recommendation      `json:"recommendation"`
	Score               float64             `json:"score"`
	Reliability         float64             `json:"reliability"`
	Confidence          string              `json:"confidence"`
	Reasons             []string            `json:"reasons"`
	RiskLevel           string              `json:"riskLevel"`
	PriceTarget         PriceTarget         `json:"priceTarget"`
	HistoricalAccuracy  HistoricalAccuracy  `json:"historicalAccuracy"`
}

type PriceTarget struct {
	TargetPrice    float64 `json:"targetPrice"`
	LowEstimate    float64 `json:"lowEstimate"`
	HighEstimate   float64 `json:"highEstimate"`
	TimeHorizon    string  `json:"timeHorizon"`
	PredictionBasis string  `json:"predictionBasis"`
}

type HistoricalAccuracy struct {
	TotalPredictions    int     `json:"totalPredictions"`
	CorrectPredictions  int     `json:"correctPredictions"`
	AccuracyRate        float64 `json:"accuracyRate"`
	AvgPriceDeviation   float64 `json:"avgPriceDeviation"`
	BestPerformingSignal string  `json:"bestPerformingSignal"`
	WorstPerformingSignal string `json:"worstPerformingSignal"`
}

type PriceDataPoint struct {
	Date      time.Time `json:"date"`
	Price     float64   `json:"price"`
	Volume    int64     `json:"volume"`
	Change    float64   `json:"change"`
	ChangePerc float64  `json:"changePerc"`
}

type PriceHistory struct {
	Symbol     string            `json:"symbol"`
	Timeframe  string            `json:"timeframe"`
	DataPoints []PriceDataPoint  `json:"dataPoints"`
	Trends     TrendAnalysis     `json:"trends"`
	Patterns   []PatternMatch    `json:"patterns"`
}

type TrendAnalysis struct {
	ShortTerm  TrendDirection `json:"shortTerm"`  
	MediumTerm TrendDirection `json:"mediumTerm"`
	LongTerm   TrendDirection `json:"longTerm"` 
	Support    float64        `json:"support"` 
	Resistance float64        `json:"resistance"` 
	TrendStrength float64     `json:"trendStrength"`
}

type TrendDirection int
const (
	StronglyBearish TrendDirection = iota - 2
	Bearish
	Sideways
	Bullish
	StronglyBullish
)

func (t TrendDirection) String() string {
	switch t {
	case StronglyBearish:
		return "STRONGLY_BEARISH"
	case Bearish:
		return "BEARISH"
	case Sideways:
		return "SIDEWAYS"
	case Bullish:
		return "BULLISH"
	case StronglyBullish:
		return "STRONGLY_BULLISH"
	default:
		return "UNKNOWN"
	}
}

type PatternMatch struct {
	Pattern     string    `json:"pattern"`
	Confidence  float64   `json:"confidence"`
	Timeframe   string    `json:"timeframe"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	Implication string    `json:"implication"`
	Reliability float64   `json:"reliability"`
}

type Recommendation int

const (
	StrongSell Recommendation = iota - 2
	Sell
	Hold
	Buy
	StrongBuy
)

func (r Recommendation) String() string {
	switch r {
	case StrongSell:
		return "STRONG_SELL"
	case Sell:
		return "SELL"
	case Hold:
		return "HOLD"
	case Buy:
		return "BUY"
	case StrongBuy:
		return "STRONG_BUY"
	default:
		return "HOLD"
	}
}

type PortfolioAnalysis struct {
	Portfolio         Portfolio       `json:"portfolio"`
	StockAnalyses     []StockAnalysis `json:"stockAnalyses"`
	OverallScore      float64         `json:"overallScore"`
	OverallRisk       string          `json:"overallRisk"`
	Recommendations   []string        `json:"recommendations"`
	GeneratedAt       time.Time       `json:"generatedAt"`
}

type AlphaVantageQuote struct {
	GlobalQuote struct {
		Symbol           string `json:"01. symbol"`
		Open             string `json:"02. open"`
		High             string `json:"03. high"`
		Low              string `json:"04. low"`
		Price            string `json:"05. price"`
		Volume           string `json:"06. volume"`
		LatestTradingDay string `json:"07. latest trading day"`
		PreviousClose    string `json:"08. previous close"`
		Change           string `json:"09. change"`
		ChangePercent    string `json:"10. change percent"`
	} `json:"Global Quote"`
}

type AlphaVantageTimeSeries struct {
	MetaData   map[string]string `json:"Meta Data"`
	TimeSeries map[string]struct {
		Open   string `json:"1. open"`
		High   string `json:"2. high"`
		Low    string `json:"3. low"`
		Close  string `json:"4. close"`
		Volume string `json:"5. volume"`
	} `json:"Time Series (Daily)"`
}

type Config struct {
	Server ServerConfig `yaml:"server"`
	APIs   APIConfig    `yaml:"apis"`
	Claude ClaudeConfig `yaml:"claude"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type APIConfig struct {
	AlphaVantage AlphaVantageConfig `yaml:"alphaVantage"`
}

type AlphaVantageConfig struct {
	APIKey  string `yaml:"apiKey"`
	BaseURL string `yaml:"baseURL"`
}

type ClaudeConfig struct {
	APIKey  string `yaml:"apiKey"`
	BaseURL string `yaml:"baseURL"`
	Model   string `yaml:"model"`
}
