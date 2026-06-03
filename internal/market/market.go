package market

import (
	"log"
	"math"
	"os"
	"strings"

	api "github.com/MarketDataApp/sdk-go"
)

func init() {
	token := os.Getenv("MARKETDATA_TOKEN")
	if token != "" {
		log.Println("[MARKET] MarketData API token found.")
	}
}

func GetPrice(symbol string) (float64, error) {
	token := os.Getenv("MARKETDATA_TOKEN")
	if token == "" {
		// Fallback to YF if no token
		price, err := GetRealPrice(symbol)
		if err == nil && price > 0 {
			return price, nil
		}
		return 100.0, nil
	}

	// MarketDataApp uses different endpoints for indices
	if strings.HasPrefix(symbol, "^") {
		// For indices like ^GSPC, MarketDataApp might need a different format or use IndexQuote
		// But let's try StockQuote first as many indices are supported there
	}

	quotes, err := api.StockQuote().Symbol(symbol).Get()
	if err != nil {
		// Fallback to YF
		price, _ := GetRealPrice(symbol)
		if price > 0 {
			return price, nil
		}
		return 100.0, err
	}
	if len(quotes) > 0 {
		return quotes[0].Last, nil
	}

	// Final fallback
	price, _ := GetRealPrice(symbol)
	if price > 0 {
		return price, nil
	}
	return 100.0, nil
}

func GetVolatility(symbol string) (float64, error) {
	token := os.Getenv("MARKETDATA_TOKEN")
	if token == "" {
		vol, err := GetRealVolatility(symbol)
		if err == nil && vol > 0 {
			return vol, nil
		}
		return 0.18, nil
	}

	candles, err := api.StockCandles().Symbol(symbol).Resolution("D").Countback(252).Get()
	if err != nil || len(candles) < 2 {
		vol, _ := GetRealVolatility(symbol)
		if vol > 0 {
			return vol, nil
		}
		return 0.18, nil
	}

	returns := make([]float64, len(candles)-1)
	for i := 1; i < len(candles); i++ {
		returns[i-1] = math.Log(candles[i].Close / candles[i-1].Close)
	}

	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	varianceSum := 0.0
	for _, r := range returns {
		diff := r - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(len(returns)-1)
	return math.Sqrt(variance) * math.Sqrt(252), nil
}

func GetBeta(stockSymbol, indexSymbol string) (float64, error) {
	token := os.Getenv("MARKETDATA_TOKEN")
	if token == "" {
		beta, err := GetRealBeta(stockSymbol, indexSymbol)
		if err == nil {
			return beta, nil
		}
		return 1.0, nil
	}

	stockCandles, err := api.StockCandles().Symbol(stockSymbol).Resolution("D").Countback(252).Get()
	if err != nil {
		beta, _ := GetRealBeta(stockSymbol, indexSymbol)
		return beta, nil
	}
	indexCandles, err := api.StockCandles().Symbol(indexSymbol).Resolution("D").Countback(252).Get()
	if err != nil {
		beta, _ := GetRealBeta(stockSymbol, indexSymbol)
		return beta, nil
	}

	minLen := len(stockCandles)
	if len(indexCandles) < minLen {
		minLen = len(indexCandles)
	}
	if minLen < 2 {
		return 1.0, nil
	}

	stockReturns := make([]float64, 0)
	indexReturns := make([]float64, 0)
	for i := 1; i < minLen; i++ {
		stockReturns = append(stockReturns, math.Log(stockCandles[i].Close/stockCandles[i-1].Close))
		indexReturns = append(indexReturns, math.Log(indexCandles[i].Close/indexCandles[i-1].Close))
	}

	sumI, sumS := 0.0, 0.0
	for i := range indexReturns {
		sumI += indexReturns[i]
		sumS += stockReturns[i]
	}
	meanI, meanS := sumI/float64(len(indexReturns)), sumS/float64(len(stockReturns))

	covariance, variance := 0.0, 0.0
	for i := range indexReturns {
		covariance += (indexReturns[i] - meanI) * (stockReturns[i] - meanS)
		variance += (indexReturns[i] - meanI) * (indexReturns[i] - meanI)
	}

	if variance == 0 {
		return 1.0, nil
	}
	return covariance / variance, nil
}

func GetIdiosyncraticVolatility(stockSymbol, indexSymbol string, beta float64) (float64, error) {
	stockVol, _ := GetVolatility(stockSymbol)
	indexVol, _ := GetVolatility(indexSymbol)
	idioVar := (stockVol * stockVol) - (beta * beta * indexVol * indexVol)
	if idioVar < 0 {
		idioVar = 0.01
	}
	return math.Sqrt(idioVar), nil
}
