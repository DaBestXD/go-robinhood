package robinhood_test

import (
	"testing"

	"meow-meow-hood-port/robinhood"
)

func TestExtractToken(t *testing.T) {
	firefox := robinhood.NewFirefox()
	_, err := firefox.ExtractToken()
	if err != nil {
		t.Fatalf("expected no error got %v", err)
	}
	// if token
}
