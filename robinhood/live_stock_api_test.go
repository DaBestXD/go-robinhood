package robinhood_test

import (
	"testing"

	"meow-meow-hood-port/robinhood"
)

func TestGetStockQuotes(t *testing.T) {
	rh := robinhood.NewRobinhoodClient(robinhood.NewFirefox())
	validSymbols := []string{"QQQ", "SPY"}
	quotes, err := rh.GetStockQuotes(validSymbols...)
	if err != nil {
		t.Fatalf("expected no error got %v", err)
	}
	for i, v := range quotes.Results {
		if v == nil {
			t.Fatalf("exepected quote for %v got %v", validSymbols[i], err)
		}
		if v.Symbol != validSymbols[i] {
			t.Fatalf("validSymbols %s does not match %s", validSymbols[i], v.Symbol)
		}
	}
	badSymbols := []string{"NOWORKY", "MEOWMEOWCAT"}
	quotes, err = rh.GetStockQuotes(badSymbols...)
	if err == nil {
		t.Fatalf("expected error for %s, got nil", badSymbols)
	}
	if quotes != nil {
		t.Fatalf("expected nil for quotes for %v, got %+v", badSymbols, quotes)
	}
	mixedSymbols := []string{"NOWORKY", "SPY"}
	quotes, err = rh.GetStockQuotes(mixedSymbols...)
	if err != nil {
		t.Fatalf("expected no error for %s, got %v", mixedSymbols, err)
	}
	if quotes == nil {
		t.Fatalf("expected quotes for %v, got nil", mixedSymbols)
	}
	for i, v := range quotes.Results {
		if quotes.Results[0] != nil {
			t.Fatalf("expected nil for %v got %v", quotes.Results[i], v)
		}
	}
}

func TestGetStockQuote(t *testing.T) {
	rh := robinhood.NewRobinhoodClient(robinhood.NewFirefox())

	validSymbol := "TSLA"
	quote, err := rh.GetStockQuote(validSymbol)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if quote == nil {
		t.Fatalf("expected quote for %s, got nil", validSymbol)
	}

	if quote.Symbol != validSymbol {
		t.Fatalf("expected symbol %s, got %s", validSymbol, quote.Symbol)
	}

	badSymbol := "BADSYMBOL"
	quote, err = rh.GetStockQuote(badSymbol)
	if err == nil {
		t.Fatalf("expected error for %s, got nil", badSymbol)
	}

	if quote != nil {
		t.Fatalf("expected nil quote for %s, got %+v", badSymbol, quote)
	}
}
