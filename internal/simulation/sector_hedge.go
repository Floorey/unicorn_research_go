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
	if numPaths < 1 { numPaths = 1000 }

	steps := int(req.DurationYears * 12) 
	dt := 1.0 / 12.0
	rFR := req.RiskFreeRate / 100.0
	driftIndex := rFR

	numVisual := 50
	if numPaths < numVisual { numVisual = numPaths }
	pathData := make([][]float64, numVisual)
	var pathMutex sync.Mutex

	finalValues := make([]float64, numPaths)
	maxDrawdowns := make([]float64, numPaths)
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
			if workerID == workers-1 { endIdx = numPaths }

			for i := startIdx; i < endIdx; i++ {
				indexPrice := 100.0
				stockPrices := make([]float64, len(req.StockPositions))
				for s := range stockPrices { stockPrices[s] = 100.0 }

				initialPortfolioValue := 0.0
				for _, pos := range req.StockPositions { initialPortfolioValue += pos }
				
				optionPrice := BlackScholes(100.0, req.OptionStrike, req.OptionExpiry, rFR, req.IndexVol)
				initialOptionValue := optionPrice * req.OptionQuantity
				initialShortValue := req.ShortIndexPos
				
				initialNetValue := initialPortfolioValue + initialOptionValue - initialShortValue
				finalValues[i] = initialNetValue // Default in case steps = 0

				var currentPath []float64
				isVisual := i < numVisual
				if isVisual {
					currentPath = make([]float64, steps+1)
					currentPath[0] = initialNetValue
				}

				maxDD := 0.0
				peak := initialNetValue

				for t := 1; t <= steps; t++ {
					zIndex := rng.NormFloat64()
					indexPrice *= math.Exp((driftIndex-0.5*req.IndexVol*req.IndexVol)*dt + req.IndexVol*math.Sqrt(dt)*zIndex)

					netStockValue := 0.0
					for s := range stockPrices {
						zStock := rng.NormFloat64()
						beta := req.StockBetas[s]
						volIdio := req.StockVols[s]
						stockPrices[s] *= math.Exp(beta*req.IndexVol*math.Sqrt(dt)*zIndex + volIdio*math.Sqrt(dt)*zStock)
						netStockValue += (stockPrices[s] / 100.0) * req.StockPositions[s]
					}

					timeRemaining := req.OptionExpiry - (float64(t) * dt)
					if timeRemaining < 0 { timeRemaining = 0 }
					optVal := BlackScholes(indexPrice, req.OptionStrike, timeRemaining, rFR, req.IndexVol) * req.OptionQuantity
					shortVal := (indexPrice / 100.0) * req.ShortIndexPos

					currentNetValue := netStockValue + optVal - shortVal
					if isVisual { currentPath[t] = currentNetValue }

					if currentNetValue > peak { peak = currentNetValue }
					dd := (peak - currentNetValue) / peak
					if dd > maxDD { maxDD = dd }
					
					if t == steps {
						finalValues[i] = currentNetValue
					}
				}

				maxDrawdowns[i] = maxDD

				if isVisual {
					pathMutex.Lock()
					pathData[i] = currentPath
					pathMutex.Unlock()
				}

				if finalValues[i] > initialNetValue {
					winMutex.Lock()
					totalWinCount++
					winMutex.Unlock()
				}
			}
		}(w)
	}
	wg.Wait()

	sumFinal := 0.0
	sumDD := 0.0
	for j := 0; j < numPaths; j++ {
		sumFinal += finalValues[j]
		sumDD += maxDrawdowns[j]
	}
	meanFinal := sumFinal / float64(numPaths)
	meanDD := sumDD / float64(numPaths)

	return models.SectorHedgeResponse{
		Status:   "success",
		NumPaths: numPaths,
		Metrics: models.HedgeMetrics{
			FinalValueMean:     math.Round(meanFinal*100) / 100,
			WinRate:            math.Round((float64(totalWinCount)/float64(numPaths)*100)*100) / 100,
			MaxDrawdown:        math.Round(meanDD*10000) / 100,
			AlphaContribution:  0.0,
			HedgeEffectiveness: 0.0,
		},
		ChartData: pathData,
	}
}
