package robinhood

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	APIIndexes     = "/indexes/"
	APIIndexQuotes = "/marketdata/indexes/values/v1/"
)

type IndexInfo struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Symbol      string `json:"symbol"`
	DisplayName string `json:"simple_name"`
	// Can either be 'active' or 'inactive'
	State            string    `json:"state"`
	TradableChainIDs *[]string `json:"tradable_chain_ids"`
}
type IndexInfos struct {
	Results []*IndexInfo `json:"results"`
}

func (rh *RobinhoodClient) GetIndexInfos(symbols ...string) (*IndexInfos, error) {
	request, err := rh.buildGetRequest(nil,
		APIIndexes,
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
	if err != nil {
		return nil, err
	}
	var indexInfos IndexInfos
	err = json.Unmarshal(body, &indexInfos)
	if err != nil {
		return nil, err
	}
	return &indexInfos, nil
}

func (rh *RobinhoodClient) GetAllIndexInfos() (*IndexInfos, error) {
	request, err := rh.buildGetRequest(nil,
		APIIndexes,
		nil,
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
	var indexInfos IndexInfos
	err = json.Unmarshal(body, &indexInfos)
	if err != nil {
		return nil, err
	}
	return &indexInfos, nil
}

type (
	IndexQuote struct {
		Price  float64 `json:"value,string"`
		Symbol string  `json:"symbol"`
		ID     string  `json:"instrument_id"`
	}
	IndexQuotes struct {
		Results []*IndexQuote `json:"data"`
	}
)

func (rh *RobinhoodClient) GetIndexQuotes(symbols ...string) (*IndexQuotes, error) {
	request, err := rh.buildGetRequest(
		nil,
		APIIndexQuotes,
		&map[string]string{"symbols": normalizeSymbols(symbols)},
	)
	if err != nil {
		return nil, err
	}
	// quotes, err := rh.doGetRequestIndexQuotes(request)
	quotes, err := doNestedGetRequest[IndexQuote](rh, request)
	if err != nil {
		return nil, err
	}
	var indexQuotes IndexQuotes
	for _, v := range quotes.Results {
		indexQuotes.Results = append(indexQuotes.Results, v.Results)
	}
	return &indexQuotes, nil
}

type InnerData[T any] struct {
	Status  string `json:"status"`
	Results *T     `json:"data"`
}

type NestedResponse[T any] struct {
	Status  string          `json:"status"`
	Results []*InnerData[T] `json:"data"`
}

// {
// 	status: Success
// 	data: [
// 		{
// 			Status : Success,
// 			data: {obj data}
// 		},
// 		{
// 			Status : Success,
// 			data: {obj data}
// 		}, ...
// 	]
// }
// Base struct is slice of pointers to the innter data

func doNestedGetRequest[T any](rh *RobinhoodClient, request *http.Request) (*NestedResponse[T], error) {
	response, err := rh.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var nested NestedResponse[T]
	defer response.Body.Close() //nolint:errcheck
	err = json.Unmarshal(body, &nested)
	if err != nil {
		return nil, err
	}
	if nested.Status != "SUCCESS" {
		return nil, fmt.Errorf("failed to return sucess")
	}
	return &nested, nil
}
