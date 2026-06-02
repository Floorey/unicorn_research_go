package simulation

import (
	"math"
)

// BlackScholes calculates the price of a European Call option
// S: current price, K: strike, T: time to expiry (years), r: risk-free rate, sigma: volatility
func BlackScholes(S, K, T, r, sigma float64) float64 {
	if T <= 0 {
		return math.Max(0, S-K)
	}
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	d2 := d1 - sigma*math.Sqrt(T)

	return S*cumulativeNormal(d1) - K*math.Exp(-r*T)*cumulativeNormal(d2)
}

// cumulativeNormal approximation (Abramowitz & Stegun)
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
