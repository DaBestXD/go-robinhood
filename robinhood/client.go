package robinhood

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type RobinhoodClient struct {
	HTTPClient *http.Client
	BaseURL    string
	Token      uuid.UUID
}

func NewRobinhoodClient() *RobinhoodClient {
	return &RobinhoodClient{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		BaseURL: "https://api.robinhood.com",
		Token:   loadToken(),
	}
}

func (rh *RobinhoodClient) BuildRequest() {}
