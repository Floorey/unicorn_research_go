package simulation

import (
	"log"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/lukasenderle/unicorn_research_go/internal/models"
)

func Run(req models.SimulationRequest) models.SimulationResponse {
	startTime := time.Now()
	
	totalInvestment := req.PosATech + req.PosBEnergy + req.PosCBonds + req.PosDCrypto + req.PosEGold
	if totalInvestment == 0 {
		return models.SimulationResponse{Status: "error"}
	}

	// Macro drifts
	baseRate := req.InterestRate / 100.0
	neutralOil := 70.0
	oilFactor := (req.OilPrice - neutralOil) / neutralOil

	driftBonds := baseRate
	driftTech := 0.08 - (baseRate * 0.5) - (oilFactor * 0.02)
	driftEnergy := 0.05 + (oilFactor * 0.15)
	driftCrypto := 0.15 - (baseRate * 1.0) // Highly sensitive to rates
	driftGold := 0.02 + (oilFactor * 0.05) // Hedge factor

	// Weights
	wTech := req.PosATech / totalInvestment
	wEnergy := req.PosBEnergy / totalInvestment
	wBonds := req.PosCBonds / totalInvestment
	wCrypto := req.PosDCrypto / totalInvestment
	wGold := req.PosEGold / totalInvestment

	steps := req.DurationYears * 12
	dt := 1.0 / 12.0
	numPaths := req.NumPaths
	if numPaths < 1 {
		numPaths = 1000
	}

	// Correlation Matrix 5x5: Tech, Energy, Bonds, Crypto, Gold
	correlations := [][]float64{
		{1.0, 0.3, -0.2, 0.5, -0.1},  // Tech
		{0.3, 1.0, -0.1, 0.1, 0.2},   // Energy
		{-0.2, -0.1, 1.0, -0.3, 0.4},  // Bonds
		{0.5, 0.1, -0.3, 1.0, -0.2},  // Crypto
		{-0.1, 0.2, 0.4, -0.2, 1.0},  // Gold
	}
	L := cholesky(correlations)

	portfolioEndValues := make([]float64, numPaths)
	visualPaths := make([][]float64, 100)
	var visualMutex sync.Mutex

	var wg sync.WaitGroup
	workers := 8
	pathsPerWorker := numPaths / workers

	log.Printf("[SIM] Starting simulation with %d paths across %d workers", numPaths, workers)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		startIdx := w * pathsPerWorker
		endIdx := (w + 1) * pathsPerWorker
		if w == workers-1 {
			endIdx = numPaths
		}

		go func(s, e int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(s)))

			for i := s; i < e; i++ {
				pTech, pEnergy, pBonds, pCrypto, pGold := 1.0, 1.0, 1.0, 1.0, 1.0
				
				isVisual := i < 100
				var currentPath []float64
				if isVisual {
					currentPath = make([]float64, steps+1)
					currentPath[0] = totalInvestment
				}

				for j := 1; j <= steps; j++ {
					z := []float64{r.NormFloat64(), r.NormFloat64(), r.NormFloat64(), r.NormFloat64(), r.NormFloat64()}
					
					// Apply Cholesky e = L * z
					e := make([]float64, 5)
					for row := 0; row < 5; row++ {
						for col := 0; col <= row; col++ {
							e[row] += L[row][col] * z[col]
						}
					}

					pTech *= math.Exp((driftTech-0.5*req.VolATech*req.VolATech)*dt + req.VolATech*math.Sqrt(dt)*e[0])
					pEnergy *= math.Exp((driftEnergy-0.5*req.VolBEnergy*req.VolBEnergy)*dt + req.VolBEnergy*math.Sqrt(dt)*e[1])
					pBonds *= math.Exp((driftBonds-0.5*req.VolCBonds*req.VolCBonds)*dt + req.VolCBonds*math.Sqrt(dt)*e[2])
					pCrypto *= math.Exp((driftCrypto-0.5*req.VolDCrypto*req.VolDCrypto)*dt + req.VolDCrypto*math.Sqrt(dt)*e[3])
					pGold *= math.Exp((driftGold-0.5*req.VolEGold*req.VolEGold)*dt + req.VolEGold*math.Sqrt(dt)*e[4])

					if isVisual {
						currentPath[j] = (pTech*wTech + pEnergy*wEnergy + pBonds*wBonds + pCrypto*wCrypto + pGold*wGold) * totalInvestment
					}
				}

				if isVisual {
					visualMutex.Lock()
					visualPaths[i] = currentPath
					visualMutex.Unlock()
				}
				portfolioEndValues[i] = (pTech*wTech + pEnergy*wEnergy + pBonds*wBonds + pCrypto*wCrypto + pGold*wGold) * totalInvestment
			}
		}(startIdx, endIdx)
	}
	wg.Wait()

	sort.Float64s(portfolioEndValues)
	
	// VaR 99%
	varIdx := int(float64(numPaths) * 0.01)
	if varIdx < 0 { varIdx = 0 }
	varValue := portfolioEndValues[varIdx]
	var99Percent := (1.0 - (varValue / totalInvestment)) * 100.0

	// Expected Shortfall (CVaR) - Average of all values below VaR
	esSum := 0.0
	for k := 0; k <= varIdx; k++ {
		esSum += portfolioEndValues[k]
	}
	esValue := esSum / float64(varIdx+1)
	es99Percent := (1.0 - (esValue / totalInvestment)) * 100.0

	sum := 0.0
	for _, v := range portfolioEndValues {
		sum += v
	}
	meanVal := sum / float64(numPaths)

	log.Printf("[SIM] Completed in %v. Mean: %.2f, VaR99: %.2f%%, ES99: %.2f%%", 
		time.Since(startTime), meanVal, var99Percent, es99Percent)

	return models.SimulationResponse{
		Status: "success",
		InputSummary: models.InputSummary{
			TotalInvestment: totalInvestment,
			CalculatedDrifts: models.CalculatedDrifts{
				Tech:   math.Round(driftTech*10000) / 100,
				Energy: math.Round(driftEnergy*10000) / 100,
				Bonds:  math.Round(driftBonds*10000) / 100,
				Crypto: math.Round(driftCrypto*10000) / 100,
				Gold:   math.Round(driftGold*10000) / 100,
			},
		},
		Metrics: models.Metrics{
			Var99Percent:      math.Round(var99Percent*100) / 100,
			ExpectedShortfall: math.Round(es99Percent*100) / 100,
			ExpectedMeanEur:   math.Round(meanVal*100) / 100,
		},
		ChartData: visualPaths,
	}
}

// Generic Cholesky decomposition for NxN matrix
func cholesky(matrix [][]float64) [][]float64 {
	n := len(matrix)
	L := make([][]float64, n)
	for i := range L {
		L[i] = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		for j := 0; j <= i; j++ {
			sum := 0.0
			for k := 0; k < j; k++ {
				sum += L[i][k] * L[j][k]
			}

			if i == j {
				L[i][j] = math.Sqrt(matrix[i][i] - sum)
			} else {
				L[i][j] = (1.0 / L[j][j] * (matrix[i][j] - sum))
			}
		}
	}
	return L
}
