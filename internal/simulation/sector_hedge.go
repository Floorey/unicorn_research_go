package simulation

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/lukasenderle/unicorn_research_go/internal/models"
)

func RunSectorHedge(req models.SectorHedgeRequest) models.SectorHedgeResponse {
	numPaths := req.NumPaths
	if numPaths < 1 {
		numPaths = 1000
	}

	steps := int(req.DurationYears * 12) // Monthly steps
	dt := 1.0 / 12.0
	rFR := req.RiskFreeRate / 100.0

	// Pre-calculate drift for index (assume risk-free for simplicity in neutral hedge)
	driftIndex := rFR

	pathData := make([][]float64, 50) // Store 50 paths for visualization
	var pathMutex sync.Mutex

	finalValues := make([]float64, numPaths)
	totalWinCount := 0
	var winMutex sync.Mutex

	var wg sync.WaitGroup
	workers := 8
	pathsPerWorker := numPaths / workers

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			startIdx := workerID * pathsPerWorker
			endIdx := (workerID + 1) * pathsPerWorker
			if workerID == workers-1 {
				endIdx = numPaths
			}

			for i := startIdx; i < endIdx; i++ {
				// Initial State
				indexPrice := 100.0
				stockPrices := make([]float64, len(req.StockPositions))
				for s := range stockPrices {
					stockPrices[s] = 100.0 // Normalize to 100 for path calculation
				}

				initialPortfolioValue := 0.0
				for _, pos := range req.StockPositions {
					initialPortfolioValue += pos
				}
				
				// Option initial value
				optionPrice := BlackScholes(100.0, req.OptionStrike, req.OptionExpiry, rFR, req.IndexVol)
				initialOptionValue := optionPrice * req.OptionQuantity
				
				// Index short initial value (liability)
				initialShortValue := req.ShortIndexPos
				
				initialNetValue := initialPortfolioValue + initialOptionValue - initialShortValue

				var currentPath []float64
				isVisual := i < 50
				if isVisual {
					currentPath = make([]float64, steps+1)
					currentPath[0] = initialNetValue
				}

				maxDD := 0.0
				peak := initialNetValue

				for t := 1; t <= steps; t++ {
					// 1. Simulate Index Move
					zIndex := rng.NormFloat64()
					indexPrice *= math.Exp((driftIndex-0.5*req.IndexVol*req.IndexVol)*dt + req.IndexVol*math.Sqrt(dt)*zIndex)

					// 2. Simulate Stock Moves (Factor Model)
					netStockValue := 0.0
					for s := range stockPrices {
						zStock := rng.NormFloat64() // Idiosyncratic component
						beta := req.StockBetas[s]
						volIdio := req.StockVols[s]
						
						// R_i = Beta * R_Index + Vol_Idio * Z_Idio
						// Simplified path: Price = Price_old * exp(Beta * log(Index_new/Index_old) + idiosyncratic_part)
						stockPrices[s] *= math.Exp(beta*req.IndexVol*math.Sqrt(dt)*zIndex + volIdio*math.Sqrt(dt)*zStock)
						
						// Calculate actual EUR value
						val := (stockPrices[s] / 100.0) * req.StockPositions[s]
						netStockValue += val
					}

					// 3. Update Option Value
					timeRemaining := req.OptionExpiry - (float64(t) * dt)
					if timeRemaining < 0 { timeRemaining = 0 }
					
					// Option is on the Index for this simulation (or could be stock basket)
					// Let's assume it's on the Index as a sector call
					optVal := BlackScholes(indexPrice, req.OptionStrike, timeRemaining, rFR, req.IndexVol) * req.OptionQuantity

					// 4. Update Short Position
					shortVal := (indexPrice / 100.0) * req.ShortIndexPos

					currentNetValue := netStockValue + optVal - shortVal
					
					if isVisual {
						currentPath[t] = currentNetValue
					}

					if currentNetValue > peak {
						peak = currentNetValue
					}
					dd := (peak - currentNetValue) / peak
					if dd > maxDD {
						maxDD = dd
					}
				}

				if isVisual {
					pathMutex.Lock()
					pathData[i] = currentPath
					pathMutex.Unlock()
				}

				finalValues[i] = currentPath[steps]
				if finalValues[i] > initialNetValue {
					winMutex.Lock()
					totalWinCount++
					winMutex.Unlock()
				}
			}
		}(w)
	}
	wg.Wait()

	// Calculate Metrics
	sumFinal := 0.0
	for _, v := range finalValues {
		sumFinal += v
	}
	meanFinal := sumFinal / float64(numPaths)

	return models.SectorHedgeResponse{
		Status: "success",
		Metrics: models.HedgeMetrics{
			FinalValueMean:     math.Round(meanFinal*100) / 100,
			WinRate:            (float64(totalWinCount) / float64(numPaths)) * 100.0,
			MaxDrawdown:        0.0, // Should calculate avg max drawdown
			AlphaContribution:  0.0, // Placeholder
			HedgeEffectiveness: 0.0, // Placeholder
		},
		ChartData: pathData,
	}
}
