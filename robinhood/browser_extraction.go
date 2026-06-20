package robinhood

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Decode the token extracted from local storage
func decodeJWT(encodedToken string) (*[]byte, error) {
	token := strings.Split(encodedToken, ".")[1]
	padding := len(token) % 4
	token += strings.Repeat("=", 4-padding)
	decodedToken, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	return &decodedToken, nil
}

// ReturnJWTExpiration returns the the JWT expiration
func ReturnJWTExpiration(encodedToken string) (*time.Time, error) {
	bytesJWT, err := decodeJWT(encodedToken)
	if err != nil {
		return nil, err
	}
	var exp struct {
		Exp int64 `json:"exp"`
	}
	err = json.Unmarshal(*bytesJWT, &exp)
	if err != nil {
		return nil, err
	}
	expDate := time.Unix(exp.Exp, 0)
	return &expDate, nil
}

// ValidateToken returns False on invalid token or on error
//
// e.g. expired, malformed-token, etc.
//
// Uses https://api.robinhood.com/accounts/ as the endpoint
func (rh *RobinhoodClient) ValidateToken(token string) (bool, error) {
	expiration, err := ReturnJWTExpiration(token)
	if err != nil {
		return false, err
	}
	if expiration.Compare(time.Now().UTC()) < 0 {
		return false, fmt.Errorf("token is expired")
	}
	const apiAcc = "/accounts/"
	request, err := rh.buildGetRequest(nil, apiAcc, nil)
	if err != nil {
		return false, err
	}
	request.Header.Add("Authorization", "Bearer "+token)
	response, err := rh.doGetRequest(request)
	if err != nil {
		return false, err
	}
	if response.StatusCode > 300 {
		return false, fmt.Errorf("reponse returned %d", response.StatusCode)
	}
	return true, nil
}
