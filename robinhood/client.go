package robinhood

import (
	"net/http"
	"time"
)

func NewRobinhoodClient() *RobinhoodClient {
	return &RobinhoodClient{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		BaseURL: "https://api.robinhood.com",
	}
}
