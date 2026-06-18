package robinhood

const APIOptionsInstruments = "/options/instruments"

type OptionMetaData struct {
	ID             string  `json:"chain_id"`
	Symbol         string  `json:"chain_symbol"`
	ExpirationDate string  `json:"expiration_date"`
	StrikePrice    float64 `json:"strike_price,string"`
	Type           string  `json:"type"`
}

func (rh *RobinhoodClient) GetOptionMetadata(symbol string) (any, error) {
	return nil, nil
}
