package simulation

import (
	"github.com/lukasenderle/unicorn_research_go/internal/models"
	"math"
)

// BlackScholes calculates the price of a European Call option
func BlackScholes(S, K, T, r, sigma float64) float64 {
	if T <= 0 {
		return math.Max(0, S-K)
	}
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	d2 := d1 - sigma*math.Sqrt(T)

	return S*cumulativeNormal(d1) - K*math.Exp(-r*T)*cumulativeNormal(d2)
}

// CalculateGreeks returns Delta, Gamma, Theta, Vega, and Rho for a European Call
func CalculateGreeks(S, K, T, r, sigma float64) models.OptionGreeks {
	if T <= 0 {
		delta := 0.0
		if S > K {
			delta = 1.0
		}
		return models.OptionGreeks{Delta: delta}
	}

	sqrtT := math.Sqrt(T)
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * sqrtT)
	d2 := d1 - sigma*sqrtT
	phiD1 := normalPDF(d1)

	delta := cumulativeNormal(d1)
	gamma := phiD1 / (S * sigma * sqrtT)
	vega := S * sqrtT * phiD1 / 100.0 // Per 1% vol change

	// Theta (Annualized, then divided by 365 for daily)
	theta := (-(S*phiD1*sigma)/(2*sqrtT) - r*K*math.Exp(-r*T)*cumulativeNormal(d2)) / 365.0

	rho := (K * T * math.Exp(-r*T) * cumulativeNormal(d2)) / 100.0 // Per 1% rate change

	return models.OptionGreeks{
		Delta: delta,
		Gamma: gamma,
		Theta: theta,
		Vega:  vega,
		Rho:   rho,
	}
}

func normalPDF(x float64) float64 {
	return math.Exp(-0.5*x*x) / math.Sqrt(2*math.Pi)
}

func cumulativeNormal(x float64) float64 {
	const (
		a1 = 0.319381530
		a2 = -0.356563782
		a3 = 1.781477937
		a4 = -1.821255978
		a5 = 1.330274429
	)
	L := math.Abs(x)
	K := 1.0 / (1.0 + 0.2316419*L)
	d := 0.3989422804 * math.Exp(-L*L/2.0)
	prob := 1.0 - d*(a1*K+a2*K*K+a3*math.Pow(K, 3)+a4*math.Pow(K, 4)+a5*math.Pow(K, 5))

	if x < 0 {
		return 1.0 - prob
	}
	return prob
}
