package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lukasenderle/unicorn_research_go/internal/db"
	"github.com/lukasenderle/unicorn_research_go/internal/market"
	"github.com/lukasenderle/unicorn_research_go/internal/models"
	"github.com/lukasenderle/unicorn_research_go/internal/simulation"
)

func RegisterUIHandlers(r *gin.Engine) {
	r.GET("/", renderDashboard)
	r.GET("/macro", renderMacro)
	r.GET("/sector-hedge", renderSectorHedge)
	r.GET("/derivatives", renderDerivatives)
	r.GET("/derivate-lab", renderDerivatives)
	r.GET("/swap", renderDerivatives)
	r.GET("/future", renderDerivatives)
	r.GET("/sap", renderDerivatives)
	r.GET("/quant-lab", renderQuantLab)
	r.GET("/backtest", renderBacktestLab)
	r.GET("/correlation", renderCorrelationLab)
	r.GET("/stress", renderStressLab)
	r.GET("/watchlist", renderWatchlistLab)
	r.POST("/ui/simulate/custom", handleCustomSimulation)
	r.POST("/ui/stress", handleStressTest)
	r.POST("/ui/correlation", handleGenerateCorrelation)
	r.POST("/ui/backtest", handleBacktest)
	r.POST("/ui/optimize", handleOptimize)
	r.POST("/ui/simulate", handleUISimulate)
	r.POST("/ui/simulate/sector-hedge", handleSectorHedgeSimulate)
	r.POST("/ui/dashboards", handleSaveDashboard)
	r.GET("/ui/dashboards/:id", handleLoadDashboard)
	r.POST("/ui/sector-hedge/dashboards", handleSaveSectorHedgeDashboard)
	r.GET("/ui/sector-hedge/dashboards/:id", handleLoadSectorHedgeDashboard)
	r.POST("/ui/price/future", handlePriceFuture)
	r.POST("/ui/price/swap", handlePriceSwap)
}

func renderQuantLab(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Content": "quant_lab",
	})
}

func handleOptimize(c *gin.Context) {
	riskFreeRate, _ := strconv.ParseFloat(c.PostForm("risk_free_rate"), 64)
	if riskFreeRate == 0 {
		riskFreeRate = 4.25
	}

	iterations, _ := strconv.Atoi(c.PostForm("iterations"))
	if iterations == 0 {
		iterations = 5000
	}

	assets := c.PostFormArray("assets")
	if len(assets) < 2 {
		c.String(http.StatusBadRequest, "Please select at least 2 assets for optimization.")
		return
	}

	result, err := simulation.OptimizePortfolio(assets, riskFreeRate, iterations)
	if err != nil {
		c.String(http.StatusInternalServerError, "Optimization error")
		return
	}

	c.HTML(http.StatusOK, "quant_lab_results.html", result)
}

func renderDerivatives(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Content": "derivatives",
	})
}

func handlePriceFuture(c *gin.Context) {
	spot, _ := strconv.ParseFloat(c.PostForm("spot_price"), 64)
	rate, _ := strconv.ParseFloat(c.PostForm("risk_free_rate"), 64)
	div, _ := strconv.ParseFloat(c.PostForm("dividend_yield"), 64)
	time, _ := strconv.ParseFloat(c.PostForm("time_to_expiry"), 64)

	req := models.FuturePricingRequest{
		SpotPrice:     spot,
		RiskFreeRate:  rate,
		DividendYield: div,
		TimeToExpiry:  time,
	}

	resp := simulation.PriceFuture(req)
	c.HTML(http.StatusOK, "derivatives_results.html", gin.H{
		"Type":   "Future",
		"Result": resp,
	})
}

func handlePriceSwap(c *gin.Context) {
	notional, _ := strconv.ParseFloat(c.PostForm("notional"), 64)
	fixed, _ := strconv.ParseFloat(c.PostForm("fixed_rate"), 64)
	floating, _ := strconv.ParseFloat(c.PostForm("floating_rate"), 64)
	tenor, _ := strconv.ParseFloat(c.PostForm("tenor_years"), 64)
	freq, _ := strconv.Atoi(c.PostForm("frequency"))

	req := models.SwapPricingRequest{
		Notional:     notional,
		FixedRate:    fixed,
		FloatingRate: floating,
		TenorYears:   tenor,
		Frequency:    freq,
	}

	resp := simulation.PriceSwap(req)
	c.HTML(http.StatusOK, "derivatives_results.html", gin.H{
		"Type":   "Swap",
		"Result": resp,
	})
}

func renderDashboard(c *gin.Context) {
	macroDashboards, _ := db.GetDashboards()
	hedgeDashboards, _ := db.GetSectorHedgeDashboards()

	c.HTML(http.StatusOK, "layout.html", gin.H{
		"MacroDashboards": macroDashboards,
		"HedgeDashboards": hedgeDashboards,
		"Content":         "dashboard",
	})
}

func renderMacro(c *gin.Context) {
	dashboards, _ := db.GetDashboards()
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Dashboards": dashboards,
		"Content":    "macro",
	})
}

func renderSectorHedge(c *gin.Context) {
	dashboards, _ := db.GetSectorHedgeDashboards()
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Dashboards": dashboards,
		"Content":    "sector_hedge",
	})
}

func handleSectorHedgeSimulate(c *gin.Context) {
	indexSymbol := strings.ToUpper(c.PostForm("index_symbol"))
	if indexSymbol == "" {
		indexSymbol = "QQQ"
	}

	duration, _ := strconv.ParseFloat(c.PostForm("duration_years"), 64)
	rFR, _ := strconv.ParseFloat(c.PostForm("risk_free_rate"), 64)
	numPaths, _ := strconv.Atoi(c.PostForm("num_paths"))

	indexPrice, _ := market.GetPrice(indexSymbol)
	indexVol, _ := market.GetVolatility(indexSymbol)

	stockSymbols := []string{
		strings.ToUpper(c.PostForm("symbol_1")),
		strings.ToUpper(c.PostForm("symbol_2")),
		strings.ToUpper(c.PostForm("symbol_3")),
	}

	stockPositions := make([]float64, 0)
	stockBetas := make([]float64, 0)
	stockVols := make([]float64, 0)

	totalWeightedBeta := 0.0
	for _, sym := range stockSymbols {
		if sym == "" {
			continue
		}
		qty, _ := strconv.ParseFloat(c.PostForm("qty_"+sym), 64)
		if qty == 0 {
			qty = 100
		}
		price, _ := market.GetPrice(sym)
		val := qty * price
		beta, _ := market.GetBeta(sym, indexSymbol)
		idioVol, _ := market.GetIdiosyncraticVolatility(sym, indexSymbol, beta)
		stockPositions = append(stockPositions, val)
		stockBetas = append(stockBetas, beta)
		stockVols = append(stockVols, idioVol)
		totalWeightedBeta += (val * beta)
	}

	optionStrike, _ := strconv.ParseFloat(c.PostForm("option_strike"), 64)
	if optionStrike == 0 {
		optionStrike = indexPrice
	}
	optionExpiry, _ := strconv.ParseFloat(c.PostForm("option_expiry"), 64)
	optionQty, _ := strconv.ParseFloat(c.PostForm("option_quantity"), 64)
	hedgeRatio, _ := strconv.ParseFloat(c.PostForm("hedge_ratio"), 64)
	if hedgeRatio == 0 {
		hedgeRatio = 100
	}
	shortIndexPos := totalWeightedBeta * (hedgeRatio / 100.0)

	req := models.SectorHedgeRequest{
		IndexSymbol:    indexSymbol,
		IndexVol:       indexVol,
		StockSymbols:   stockSymbols,
		StockPositions: stockPositions,
		StockBetas:     stockBetas,
		StockVols:      stockVols,
		OptionStrike:   optionStrike,
		OptionExpiry:   optionExpiry,
		OptionQuantity: optionQty,
		ShortIndexPos:  shortIndexPos,
		DurationYears:  duration,
		RiskFreeRate:   rFR,
		NumPaths:       numPaths,
	}

	resp := simulation.RunSectorHedge(req)
	go func() { db.SaveSectorHedge(req, resp) }()
	c.HTML(http.StatusOK, "sector_hedge_results.html", resp)
}

func handleUISimulate(c *gin.Context) {
	interestRate, _ := strconv.ParseFloat(c.PostForm("interest_rate"), 64)
	oilPrice, _ := strconv.ParseFloat(c.PostForm("oil_price"), 64)
	posATech, _ := strconv.ParseFloat(c.PostForm("pos_a_tech"), 64)
	posBEnergy, _ := strconv.ParseFloat(c.PostForm("pos_b_energy"), 64)
	posCBonds, _ := strconv.ParseFloat(c.PostForm("pos_c_bonds"), 64)
	posDCrypto, _ := strconv.ParseFloat(c.PostForm("pos_d_crypto"), 64)
	posEGold, _ := strconv.ParseFloat(c.PostForm("pos_e_gold"), 64)
	durationYears, _ := strconv.Atoi(c.PostForm("duration_years"))
	numPaths, _ := strconv.Atoi(c.PostForm("num_paths"))

	req := models.SimulationRequest{
		InterestRate:  interestRate,
		OilPrice:      oilPrice,
		PosATech:      posATech,
		PosBEnergy:    posBEnergy,
		PosCBonds:     posCBonds,
		PosDCrypto:    posDCrypto,
		PosEGold:      posEGold,
		DurationYears: durationYears,
		NumPaths:      numPaths,
	}

	if req.VolATech == 0 {
		req.VolATech, _ = market.GetVolatility("QQQ")
	}
	if req.VolBEnergy == 0 {
		req.VolBEnergy, _ = market.GetVolatility("XLE")
	}
	if req.VolCBonds == 0 {
		req.VolCBonds, _ = market.GetVolatility("TLT")
	}
	if req.VolDCrypto == 0 {
		req.VolDCrypto, _ = market.GetVolatility("BTC")
	}
	if req.VolEGold == 0 {
		req.VolEGold, _ = market.GetVolatility("GLD")
	}

	resp := simulation.Run(req)
	go func() { db.SaveSimulation(req, resp) }()

	c.HTML(http.StatusOK, "results.html", gin.H{
		"Metrics":      resp.Metrics,
		"InputSummary": resp.InputSummary,
		"ChartData":    resp.ChartData,
	})
}

func handleSaveDashboard(c *gin.Context) {
	interestRate, _ := strconv.ParseFloat(c.PostForm("interest_rate"), 64)
	oilPrice, _ := strconv.ParseFloat(c.PostForm("oil_price"), 64)
	posATech, _ := strconv.ParseFloat(c.PostForm("pos_a_tech"), 64)
	posBEnergy, _ := strconv.ParseFloat(c.PostForm("pos_b_energy"), 64)
	posCBonds, _ := strconv.ParseFloat(c.PostForm("pos_c_bonds"), 64)
	posDCrypto, _ := strconv.ParseFloat(c.PostForm("pos_d_crypto"), 64)
	posEGold, _ := strconv.ParseFloat(c.PostForm("pos_e_gold"), 64)
	durationYears, _ := strconv.Atoi(c.PostForm("duration_years"))
	numPaths, _ := strconv.Atoi(c.PostForm("num_paths"))
	name := c.PostForm("dashboard_name")

	dash := models.Dashboard{
		Name:          name,
		InterestRate:  interestRate,
		OilPrice:      oilPrice,
		PosATech:      posATech,
		PosBEnergy:    posBEnergy,
		PosCBonds:     posCBonds,
		PosDCrypto:    posDCrypto,
		PosEGold:      posEGold,
		DurationYears: durationYears,
		NumPaths:      numPaths,
		CreatedAt:     time.Now(),
	}

	if err := db.SaveDashboard(dash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save dashboard"})
		return
	}
	c.HTML(http.StatusOK, "dashboard_list.html", gin.H{"Dashboards": []models.Dashboard{dash}})
}

func handleLoadDashboard(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	dash, _ := db.GetDashboardByID(uint(id))
	c.HTML(http.StatusOK, "simulator_form.html", gin.H{"Dash": dash})
}

func handleSaveSectorHedgeDashboard(c *gin.Context) {
	name := c.PostForm("lab_name")
	index := c.PostForm("index_symbol")
	dash := models.SectorHedgeDashboard{Name: name, IndexSymbol: index, CreatedAt: time.Now()}
	db.SaveSectorHedgeDashboard(dash)
	c.String(http.StatusOK, "Saved")
}

func handleLoadSectorHedgeDashboard(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	dash, _ := db.GetSectorHedgeDashboardByID(uint(id))
	c.HTML(http.StatusOK, "sector_hedge_form.html", gin.H{"Dash": dash})
}

func renderBacktestLab(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Content": "backtest_lab",
	})
}

func handleBacktest(c *gin.Context) {
	durationYears, _ := strconv.Atoi(c.PostForm("duration_years"))
	if durationYears == 0 {
		durationYears = 5
	}

	weights := make(map[string]float64)

	wTech, _ := strconv.ParseFloat(c.PostForm("weight_tech"), 64)
	wEnergy, _ := strconv.ParseFloat(c.PostForm("weight_energy"), 64)
	wBonds, _ := strconv.ParseFloat(c.PostForm("weight_bonds"), 64)
	wCrypto, _ := strconv.ParseFloat(c.PostForm("weight_crypto"), 64)
	wGold, _ := strconv.ParseFloat(c.PostForm("weight_gold"), 64)

	// Normalize
	total := wTech + wEnergy + wBonds + wCrypto + wGold
	if total > 0 {
		weights["Tech"] = wTech / total
		weights["Energy"] = wEnergy / total
		weights["Bonds"] = wBonds / total
		weights["Crypto"] = wCrypto / total
		weights["Gold"] = wGold / total
	} else {
		// Default evenly distributed
		weights["Tech"] = 0.2
		weights["Energy"] = 0.2
		weights["Bonds"] = 0.2
		weights["Crypto"] = 0.2
		weights["Gold"] = 0.2
	}

	req := models.BacktestRequest{
		DurationYears: durationYears,
		Weights:       weights,
	}

	resp := simulation.RunBacktest(req)

	c.HTML(http.StatusOK, "backtest_results.html", gin.H{
		"Result":  resp,
		"Weights": weights,
	})
}

func renderCorrelationLab(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Content": "correlation_lab",
	})
}

func handleGenerateCorrelation(c *gin.Context) {
	// In a more advanced version, we could take the timeframe as input
	matrix := simulation.GenerateCorrelationMatrix()
	c.HTML(http.StatusOK, "correlation_results.html", matrix)
}

func renderStressLab(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Content": "stress_lab",
	})
}

func handleStressTest(c *gin.Context) {
	scenario := c.PostForm("scenario")
	if scenario == "" {
		scenario = "2008_GFC"
	}

	weights := make(map[string]float64)

	wTech, _ := strconv.ParseFloat(c.PostForm("weight_tech"), 64)
	wEnergy, _ := strconv.ParseFloat(c.PostForm("weight_energy"), 64)
	wBonds, _ := strconv.ParseFloat(c.PostForm("weight_bonds"), 64)
	wCrypto, _ := strconv.ParseFloat(c.PostForm("weight_crypto"), 64)
	wGold, _ := strconv.ParseFloat(c.PostForm("weight_gold"), 64)

	total := wTech + wEnergy + wBonds + wCrypto + wGold
	if total > 0 {
		weights["Tech"] = wTech / total
		weights["Energy"] = wEnergy / total
		weights["Bonds"] = wBonds / total
		weights["Crypto"] = wCrypto / total
		weights["Gold"] = wGold / total
	} else {
		weights["Tech"] = 0.2
		weights["Energy"] = 0.2
		weights["Bonds"] = 0.2
		weights["Crypto"] = 0.2
		weights["Gold"] = 0.2
	}

	req := models.StressTestRequest{
		ScenarioName: scenario,
		Weights:      weights,
	}

	resp := simulation.RunStressTest(req)

	c.HTML(http.StatusOK, "stress_results.html", gin.H{
		"Result":  resp,
		"Weights": weights,
	})
}

func renderWatchlistLab(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Content": "watchlist_lab",
	})
}

func handleCustomSimulation(c *gin.Context) {
	// Parse dynamic form inputs
	symbols := []string{
		strings.ToUpper(c.PostForm("symbol_1")),
		strings.ToUpper(c.PostForm("symbol_2")),
		strings.ToUpper(c.PostForm("symbol_3")),
		strings.ToUpper(c.PostForm("symbol_4")),
		strings.ToUpper(c.PostForm("symbol_5")),
	}

	allocations := make([]float64, 5)
	for i := 1; i <= 5; i++ {
		key := "alloc_" + strconv.Itoa(i)
		val, _ := strconv.ParseFloat(c.PostForm(key), 64)
		allocations[i-1] = val
	}

	duration, _ := strconv.Atoi(c.PostForm("duration_years"))
	if duration == 0 {
		duration = 5
	}

	paths, _ := strconv.Atoi(c.PostForm("num_paths"))
	if paths == 0 {
		paths = 1000
	}

	rate, _ := strconv.ParseFloat(c.PostForm("risk_free_rate"), 64)
	if rate == 0 {
		rate = 4.5
	}

	req := models.CustomSimulationRequest{
		Symbols:       symbols,
		Allocations:   allocations,
		DurationYears: duration,
		NumPaths:      paths,
		RiskFreeRate:  rate,
	}

	resp := simulation.RunCustomSimulation(req)

	if resp.Status == "error" {
		c.String(http.StatusBadRequest, "Invalid input or unable to fetch market data. Ensure you have valid symbols and allocations.")
		return
	}

	c.HTML(http.StatusOK, "watchlist_results.html", resp)
}
