package robinhood

import (
	"encoding/json"
	"fmt"
	"io"
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

// GetAllFutureProducts TODO: Add better doc string about any jankyness, etc.
func (rh *RobinhoodClient) GetAllFutureProducts() (*FutureProducts, error) {
	request, err := rh.buildGetRequest(APIFuturesProducts, nil)
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

// GetFutureProduct only supports ContractID not future Symbol
func (rh *RobinhoodClient) GetFutureProduct(contractID string) {
}

// GetFutureQuote only supports ContractID not future Symbol
func (rh *RobinhoodClient) GetFutureQuote(contractID string) error {
	builtURL := APIFuturesQuotes + contractID + "/"
	request, err := rh.buildGetRequest(builtURL, nil)
	if err != nil {
		return err
	}
	response, err := rh.doGetRequest(request)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	fmt.Printf("%d\n", response.StatusCode)
	defer response.Body.Close() //nolint:errcheck
	fmt.Printf("%s\n\n", string(body))
	return nil
}

func (rh *RobinhoodClient) GetFutureQuotes(contractIDs ...string) error {
	request, err := rh.buildGetRequest(
		APIFuturesQuotes,
		&map[string]string{"ids": strings.Join(contractIDs, ",")},
	)
	if err != nil {
		return err
	}
	response, err := rh.doGetRequest(request)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	fmt.Printf("%d\n", response.StatusCode)
	defer response.Body.Close() //nolint:errcheck
	fmt.Print(string(body), "\n\n")
	return nil
}
