package stock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"proyecto-mcp-bolsa/pkg/models"
)

type APIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewAPIClient(apiKey, baseURL string) *APIClient {
	return &APIClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *APIClient) GetQuote(symbol string) (*models.Stock, error) {
	if c.apiKey == "" || c.apiKey == "demo" {
		return nil, fmt.Errorf("API key required for stock quote: %s", symbol)
	}

	params := url.Values{
		"function": {"GLOBAL_QUOTE"},
		"symbol":   {symbol},
		"apikey":   {c.apiKey},
	}

	resp, err := c.makeRequest(params)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote for %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var demoCheck map[string]interface{}
	if err := json.Unmarshal(body, &demoCheck); err == nil {
		if info, exists := demoCheck["Information"]; exists {
			if strings.Contains(fmt.Sprint(info), "demo") {
				return nil, fmt.Errorf("demo API key not supported for production use")
			}
		}
	}

	var quote models.AlphaVantageQuote
	if err := json.Unmarshal(body, &quote); err != nil {
		return nil, fmt.Errorf("failed to parse quote response: %w", err)
	}

	if quote.GlobalQuote.Symbol == "" {
		return nil, fmt.Errorf("no data returned for symbol: %s", symbol)
	}

	return c.convertToStock(quote.GlobalQuote)
}

func (c *APIClient) GetTimeSeries(symbol string, interval string) (map[string]models.Stock, error) {
	if c.apiKey == "" || c.apiKey == "demo" {
		return nil, fmt.Errorf("API key required for time series data: %s", symbol)
	}

	params := url.Values{
		"function": {"TIME_SERIES_DAILY"},
		"symbol":   {symbol},
		"apikey":   {c.apiKey},
	}

	resp, err := c.makeRequest(params)
	if err != nil {
		return nil, fmt.Errorf("failed to get time series for %s: %w", symbol, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var demoCheck map[string]interface{}
	if err := json.Unmarshal(body, &demoCheck); err == nil {
		if info, exists := demoCheck["Information"]; exists {
			if strings.Contains(fmt.Sprint(info), "demo") {
				return nil, fmt.Errorf("demo API key not supported for production use")
			}
		}
	}

	var timeSeries models.AlphaVantageTimeSeries
	if err := json.Unmarshal(body, &timeSeries); err != nil {
		return nil, fmt.Errorf("failed to parse time series response: %w", err)
	}

	result := make(map[string]models.Stock)
	for date, data := range timeSeries.TimeSeries {
		stock, err := c.convertTimeSeriesData(symbol, date, data)
		if err != nil {
			continue
		}
		result[date] = *stock
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid time series data returned for symbol: %s", symbol)
	}

	return result, nil
}

func (c *APIClient) makeRequest(params url.Values) (*http.Response, error) {
	fullURL := fmt.Sprintf("%s?%s", c.baseURL, params.Encode())
	
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "MCP Stock Analyzer/1.0")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	return resp, nil
}

func (c *APIClient) convertToStock(quote struct {
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
}) (*models.Stock, error) {
	price, err := strconv.ParseFloat(quote.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid price: %s", quote.Price)
	}

	change, err := strconv.ParseFloat(quote.Change, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid change: %s", quote.Change)
	}

	changePercStr := strings.TrimSuffix(quote.ChangePercent, "%")
	changePerc, err := strconv.ParseFloat(changePercStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid change percent: %s", quote.ChangePercent)
	}

	volume, err := strconv.ParseInt(quote.Volume, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid volume: %s", quote.Volume)
	}

	lastUpdated, err := time.Parse("2006-01-02", quote.LatestTradingDay)
	if err != nil {
		return nil, fmt.Errorf("invalid date: %s", quote.LatestTradingDay)
	}

	return &models.Stock{
		Symbol:      quote.Symbol,
		Name:        quote.Symbol,
		Price:       price,
		Change:      change,
		ChangePerc:  changePerc,
		Volume:      volume,
		LastUpdated: lastUpdated,
	}, nil
}

func (c *APIClient) convertTimeSeriesData(symbol, date string, data struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"5. volume"`
}) (*models.Stock, error) {
	price, err := strconv.ParseFloat(data.Close, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid close price: %s", data.Close)
	}

	open, err := strconv.ParseFloat(data.Open, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid open price: %s", data.Open)
	}

	volume, err := strconv.ParseInt(data.Volume, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid volume: %s", data.Volume)
	}

	lastUpdated, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date: %s", date)
	}

	change := price - open
	changePerc := (change / open) * 100

	return &models.Stock{
		Symbol:      symbol,
		Name:        symbol,
		Price:       price,
		Change:      change,
		ChangePerc:  changePerc,
		Volume:      volume,
		LastUpdated: lastUpdated,
	}, nil
}

