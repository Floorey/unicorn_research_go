package market

import (
	"log"
	"math"
	"os"

	api "github.com/MarketDataApp/sdk-go"
)

func init() {
	token := os.Getenv("MARKETDATA_TOKEN")
	if token != "" {
		// The SDK usually picks up the token from the environment variable automatically
		// but we can log that it's found.
		log.Println("[MARKET] MarketData API token found.")
	}
}

func GetPrice(symbol string) (float64, error) {
	token := os.Getenv("MARKETDATA_TOKEN")
	if token == "" {
		log.Printf("[MARKET] No API token, returning mock price for %s", symbol)
		return 100.0, nil
	}

	quotes, err := api.StockQuote().Symbol(symbol).Get()
	if err != nil {
		return 0, err
	}

	if len(quotes) > 0 {
		return quotes[0].Last, nil
	}

	return 0, nil
}

// GetVolatility fetches historical daily candles and calculates the annualized volatility
func GetVolatility(symbol string) (float64, error) {
	token := os.Getenv("MARKETDATA_TOKEN")
	if token == "" {
		log.Printf("[MARKET] No API token, returning mock volatility for %s", symbol)
		return 0.18, nil
	}

	// Fetch daily candles for the last year (roughly 252 trading days)
	candles, err := api.StockCandles().
		Symbol(symbol).
		Resolution("D").
		Countback(252).
		Get()

	if err != nil {
		log.Printf("[MARKET] Error fetching candles for %s: %v", symbol, err)
		return 0.18, nil
	}

	if len(candles) < 2 {
		log.Printf("[MARKET] Not enough data for %s, returning default", symbol)
		return 0.18, nil
	}

	// Calculate log returns
	returns := make([]float64, len(candles)-1)
	for i := 1; i < len(candles); i++ {
		returns[i-1] = math.Log(candles[i].Close / candles[i-1].Close)
	}

	// Calculate mean return
	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	// Calculate variance
	varianceSum := 0.0
	for _, r := range returns {
		diff := r - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(len(returns)-1)

	// Annualize volatility: daily_std * sqrt(252)
	dailyVol := math.Sqrt(variance)
	annualVol := dailyVol * math.Sqrt(252)

	log.Printf("[MARKET] Calculated annual volatility for %s: %.4f", symbol, annualVol)
	return annualVol, nil
}
