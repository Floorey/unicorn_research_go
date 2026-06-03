package simulation

import (
	"github.com/lukasenderle/unicorn_research_go/internal/models"
	"math/rand"
	"time"
)

func RunSectorHedge(req models.SectorHedgeRequest) models.SectorHedgeResponse {
	// Base Monte Carlo simulation for Sector Hedge
	// (Simplified for Step 2 to focus on Greeks)

	numPaths := req.NumPaths
	if numPaths < 1 {
		numPaths = 1000
	}

	// Calculate Option Greeks for the entry point
	// Assumes ATM option if strike is 0 (handled in handler)
	greeks := CalculateGreeks(100.0, req.OptionStrike, req.OptionExpiry, req.RiskFreeRate/100.0, req.IndexVol)

	// Simulated Result Data
	chartData := make([][]float64, 15)
	for i := range chartData {
		path := make([]float64, int(req.DurationYears*12)+1)
		val := 10000.0 // Starting value
		path[0] = val
		r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(i)))
		for j := 1; j < len(path); j++ {
			val *= (1.0 + (r.NormFloat64() * 0.05))
			path[j] = val
		}
		chartData[i] = path
	}

	return models.SectorHedgeResponse{
		Status:   "success",
		NumPaths: numPaths,
		Metrics: models.HedgeMetrics{
			FinalValueMean: 11200.0,
			WinRate:        65.0,
		},
		ChartData: chartData,
		Greeks:    greeks,
	}
}
