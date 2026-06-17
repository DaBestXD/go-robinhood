// Package robinhood
package robinhood

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const APIFuturesProducts = "/arsenal/v1/futures/products/"

// StockQuote from endpoint /quotes/{SYMBOL}
//
// Does not require authentication
type StockQuote struct {
	AskPrice      float64 `json:"ask_price,string"`
	AskSize       int32   `json:"ask_size"`
	BidSize       int32   `json:"bid_size"`
	BidPrice      float64 `json:"bid_price,string"`
	Symbol        string  `json:"symbol"`
	InstrumentID  string  `json:"instrument_id"`
	InstrumentURL string  `json:"instrument"`
	TradingHalted bool    `json:"trading_halted"`
}

type StockQuotes struct {
	Results []*StockQuote `json:"results"`
}

func (rh *RobinhoodClient) BuildRequestWithMultipleSymbols(endpoint string, symbols ...string) (*http.Request, error) {
	baseURL, err := url.Parse(rh.BaseURL + endpoint)
	if err != nil {
		return nil, err
	}
	normalizedSymbols := make([]string, 0, len(symbols))
	params := url.Values{}
	for _, symbol := range symbols {
		normalizedSymbols = append(normalizedSymbols, strings.ToUpper(symbol))
	}
	params.Add("symbols", strings.Join(normalizedSymbols, ","))
	baseURL.RawQuery = params.Encode()
	request, err := http.NewRequest(http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return nil, err
	}
	return request, err
}

// BuildRequestWithSingleSymbol create a request pointer
func (rh *RobinhoodClient) BuildRequestWithSingleSymbol(endpoint string, symbol string) (*http.Request, error) {
	requestURL, err := url.Parse(rh.BaseURL + endpoint + strings.ToUpper(symbol) + "/")
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, err
	}
	return request, err
}

// GetStockQuote returns a StockQuote struct
//
// Uses /quotes/{symbol} endpoint
//
// Example: "SPY" or "spy"
func (rh *RobinhoodClient) GetStockQuote(symbol string) (*StockQuote, error) {
	request, err := rh.BuildRequestWithSingleSymbol("/quotes/", symbol)
	if err != nil {
		return nil, err
	}
	response, err := rh.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(response.Body)
	defer response.Body.Close() //nolint:errcheck
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status %s: %s", response.Status, string(body))
	}
	if err != nil {
		return nil, err
	}
	var stockQuote StockQuote
	err = json.Unmarshal(body, &stockQuote)
	if err != nil {
		return nil, err
	}
	return &stockQuote, nil
}

// GetStockQuotes returns a pointer to a StockQuotes struct
//
// Uses /quotes/?symbols=..., endpoint
//
// Example: "SPY", "QQQ" or "spy", "Qqq"
//
// If invalid symbol returns nil for that symbol
func (rh *RobinhoodClient) GetStockQuotes(symbols ...string) (*StockQuotes, error) {
	request, err := rh.BuildRequestWithMultipleSymbols("/quotes/", symbols...)
	if err != nil {
		return nil, err
	}
	response, err := rh.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(response.Body)
	defer response.Body.Close() //nolint:errcheck
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status %s: %s", response.Status, string(body))
	}
	if err != nil {
		return nil, err
	}
	var stockQuotes StockQuotes
	err = json.Unmarshal(body, &stockQuotes)
	if err != nil {
		return nil, err
	}
	return &stockQuotes, nil
}

func (rh *RobinhoodClient) GetFutureQuote(symbol string) {
}

func (rh *RobinhoodClient) GetAllFutureProducts() {
	response, err := rh.HTTPClient.Get(rh.BaseURL + APIFuturesProducts)
	if err != nil {
		return
	}
	body, err := io.ReadAll(response.Body)
	fmt.Printf("%v", response.Request.URL)
	defer response.Body.Close() //nolint:errcheck
	if err != nil {
		return
	}
	fmt.Print(string(body))
}
