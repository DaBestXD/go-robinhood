package robinhood

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	APIFuturesProducts = "/arsenal/v1/futures/products/"
	APIFuturesQuotes   = "/marketdata/futures/quotes/v1/"
)

type (
	futureTradingHours struct {
		DailyOpenTime  *time.Time
		DailyCloseTime *time.Time
		WeeklyOpen     *time.Time
		WeeklyClose    *time.Time
	}
	FutureProduct struct {
		// TODO: ID should probably change this but I'll figure this out when
		// I implement the robinhood watchlist to keep consistent naming
		ID                 string  `json:"id"`
		ActiveContractID   string  `json:"activeFuturesContractId"`
		Symbol             string  `json:"displaySymbol"`
		SymbolWithExchange string  `json:"symbol"`
		SimpleName         string  `json:"simpleName"`
		Currency           string  `json:"currency"`
		Country            string  `json:"country"`
		DeliveryType       string  `json:"delivery"`
		PirceIncrement     float64 `json:"priceIncrements,string"`
		// Time is disaplyed as hour:minutes
		SettlementTime string `json:"settlementStartTime"`
		// IDK when to use pointer to struct or just put the struct
		TradingHours *futureTradingHours `json:"tradingHoursInfo"`
	}

	FutureProducts struct {
		Results []*FutureProduct `json:"results"`
	}

	FutureQuote struct {
		ID             string  `json:"instrument_id"`
		AskPrice       float32 `json:"ask_price,string"`
		BidPrice       float32 `json:"bid_price,string"`
		LastTradePrice float32 `json:"last_trade_price,string"`
		AskSize        int32   `json:"ask_size"`
		BidSize        int32   `json:"bid_size"`
		LastTradeSize  int32   `json:"last_trade_size"`
	}

	FutureQuotes struct {
		Results []*FutureQuote `json:"data"`
	}
)

// Custom Unmarshalling for Future TradingHours
// easier to call TradingHours.DailyOpenTime then having to fight through string maps
func (f *futureTradingHours) UnmarshalJSON(data []byte) error {
	var variable struct {
		JSONTimes []struct {
			Name string    `json:"name"`
			Time time.Time `json:"time"`
		} `json:"variables"`
	}
	if err := json.Unmarshal(data, &variable); err != nil {
		return err
	}
	for _, v := range variable.JSONTimes {
		switch v.Name {
		case "daily_close_end":
			f.DailyOpenTime = &v.Time
		case "daily_close_start":
			f.DailyCloseTime = &v.Time
		case "week_start":
			f.WeeklyOpen = &v.Time
		case "week_end":
			f.WeeklyClose = &v.Time
		// Each case under this comment are for Crypto Futures
		// Crypto Futures only have a closing time of 2 minutes per day
		case "daily_close_time":
			f.DailyOpenTime = &v.Time
			temp := v.Time.Add(2 * time.Minute)
			f.DailyCloseTime = &temp
		case "sat_maint_start":
			f.WeeklyOpen = &v.Time
		case "sat_maint_end":
			f.WeeklyClose = &v.Time
		default:
			return fmt.Errorf("unknown name %s", v.Name)
		}
	}
	return nil
}

// GetAllFutureProducts returns all future products
// Warning: the ordering of the prodcuts returned is randomized
// TODO: Add better doc string about any jankness, etc.
func (rh *RobinhoodClient) GetAllFutureProducts() (*FutureProducts, error) {
	request, err := rh.buildGetRequest(
		nil,
		APIFuturesProducts,
		nil)
	if err != nil {
		return nil, err
	}
	response, err := rh.doGetRequest(request)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() //nolint:errcheck
	var results FutureProducts
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, err
	}
	return &results, nil
}

// GetFutureProductInfos only supports ContractID not future Symbol
// or ActiveContractID from FutureProduct struct
func (rh *RobinhoodClient) GetFutureProductInfos(contractID ...string) (*FutureProducts, error) {
	request, err := rh.buildGetRequest(
		nil,
		APIFuturesProducts,
		&map[string]string{"product_ids": strings.Join(contractID, ",")},
	)
	if err != nil {
		return nil, err
	}
	response, err := rh.doGetRequest(request)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() //nolint:errcheck
	var futureProducts FutureProducts
	err = json.Unmarshal(body, &futureProducts)
	if err != nil {
		return nil, err
	}
	return &futureProducts, nil
}

// Helper function for extactly just the future quote endpoint because robinhood
// probably got aids creating ts
// For invalid option ids will return nil in results
func (rh *RobinhoodClient) doGetRequestFutureQuote(request *http.Request) (*FutureQuotes, error) {
	response, err := rh.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	var results struct {
		Results []*struct {
			Status string      `json:"STATUS"`
			Inner  FutureQuote `json:"data"`
		} `json:"data"`
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, err
	}
	var futureQuotes FutureQuotes
	for _, v := range results.Results {
		// IDK how to return a nil pointer
		if v.Status != "SUCCESS" {
			futureQuotes.Results = append(futureQuotes.Results, nil)
		} else {
			futureQuotes.Results = append(futureQuotes.Results, &v.Inner)
		}
	}
	return &futureQuotes, nil
}

func (rh *RobinhoodClient) GetFutureQuotes(contractIDs ...string) (*FutureQuotes, error) {
	request, err := rh.buildGetRequest(
		nil,
		APIFuturesQuotes,
		&map[string]string{"ids": strings.Join(contractIDs, ",")},
	)
	if err != nil {
		return nil, err
	}
	quotes, err := rh.doGetRequestFutureQuote(request)
	if err != nil {
		return nil, err
	}
	return quotes, nil
}
