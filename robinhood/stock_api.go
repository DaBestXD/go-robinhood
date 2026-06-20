// Package robinhood
package robinhood

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	APIQuotes      = "/quotes/"
	APIInstrumemts = "/instruments/"
)

// StockQuote from endpoint /quotes/{SYMBOL}
//
// Does not require authentication
type StockQuote struct {
	AskPrice float32 `json:"ask_price,string"`
	AskSize  int32   `json:"ask_size"`
	BidSize  int32   `json:"bid_size"`
	BidPrice float32 `json:"bid_price,string"`
	Symbol   string  `json:"symbol"`
	// TODO: InstrumentID should probably change this but I'll figure this out when
	// I implement the robinhood watchlist to keep consistent naming
	InstrumentID  string `json:"instrument_id"`
	InstrumentURL string `json:"instrument"`
	TradingHalted bool   `json:"trading_halted"`
}

type StockQuotes struct {
	Results []*StockQuote `json:"results"`
}

// GetStockQuotes returns a pointer to a StockQuotes struct
//
// Uses /quotes/?symbols=..., endpoint
//
// Example: "SPY", "QQQ" or "spy", "Qqq"
//
// If invalid symbol returns nil for that symbol
func (rh *RobinhoodClient) GetStockQuotes(symbols ...string) (*StockQuotes, error) {
	request, err := rh.buildGetRequest(
		nil,
		APIQuotes,
		&map[string]string{
			"symbols": normalizeSymbols(symbols),
		},
	)
	if err != nil {
		return nil, err
	}
	response, err := rh.doGetRequest(request)
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

type StockInfo struct {
	Symbol        string `json:"symbol"`
	ID            string `json:"id"`
	URL           string `json:"url"`
	Splits        string `json:"splits"`
	OptionChainID string `json:"tradable_chain_id"`
	IsTradeable   bool   `json:"tradable"`
}

type StockInfos struct {
	Results []*StockInfo `json:"results"`
}

func (rh *RobinhoodClient) GetStockInfos(symbols ...string) (*StockInfos, error) {
	request, err := rh.buildGetRequest(
		nil,
		APIInstrumemts,
		&map[string]string{"symbols": normalizeSymbols(symbols)},
	)
	if err != nil {
		return nil, err
	}
	response, err := rh.doGetRequest(request)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(response.Body)
	defer response.Body.Close() //nolint:errcheck
	if err != nil {
		return nil, err
	}
	var stockInfos StockInfos
	err = json.Unmarshal(body, &stockInfos)
	if err != nil {
		return nil, err
	}
	return &stockInfos, nil
}

// Helper function to uppercase all symbols in the string slice
// and convert to a joint string by ","
func normalizeSymbols(symbols []string) string {
	for i, v := range symbols {
		symbols[i] = strings.ToUpper(v)
	}
	return strings.Join(symbols, ",")
}
