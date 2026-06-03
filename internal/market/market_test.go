package market

import (
	"testing"
)

func TestGetVolatilityMock(t *testing.T) {
	// Without token, it should return mock 0.18
	vol, err := GetVolatility("AAPL")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if vol != 0.18 {
		t.Errorf("Expected mock volatility 0.18, got %f", vol)
	}
}
