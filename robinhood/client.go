package robinhood

import (
	"fmt"
	"net/http"
	"time"
)

type RobinhoodClient struct {
	HTTPClient  *http.Client
	BaseURL     string
	AccessToken string
}

type Browser interface {
	String() string
	// TODO: implement this later
	// OpenAndClose(waitTime float64, headless bool)
	ExtractToken() (string, error)
	PathToDB() string
	PathToExecutable() string
}

func NewRobinhoodClient(browser Browser) *RobinhoodClient {
	token, err := browser.ExtractToken()
	if err != nil {
		panic(err)
	}
	return &RobinhoodClient{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		BaseURL:     "https://api.robinhood.com",
		AccessToken: token,
	}
}

// BuildGetRequest takes an endpoint e.g. "/accounts/"
//
// Params be nil or mapping of string, string e.g. "symbols : 'SPY'"
func (rh *RobinhoodClient) buildGetRequest(endpoint string, params *map[string]string) (*http.Request, error) {
	request, err := http.NewRequest(http.MethodGet, rh.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	if rh.AccessToken != "" {
		request.Header.Set("Authorization", "Bearer "+rh.AccessToken)
	}
	request.Header.Set("Accept", "application/json")
	query := request.URL.Query()
	if params != nil {
		for k, v := range *params {
			query.Add(k, v)
		}
	}
	request.URL.RawQuery = query.Encode()
	return request, nil
}

func (rh *RobinhoodClient) doGetRequest(request *http.Request) (*http.Response, error) {
	response, err := rh.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode > 300 {
		return nil, fmt.Errorf("bad status code %d", response.StatusCode)
	}
	return response, nil
}
