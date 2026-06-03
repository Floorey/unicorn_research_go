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
	r.GET("/", renderIndex)
	r.GET("/sector-hedge", renderSectorHedge)
	r.POST("/ui/simulate", handleUISimulate)
	r.POST("/ui/simulate/sector-hedge", handleSectorHedgeSimulate)
	r.POST("/ui/dashboards", handleSaveDashboard)
	r.GET("/ui/dashboards/:id", handleLoadDashboard)
}

func renderIndex(c *gin.Context) {
	dashboards, _ := db.GetDashboards()
	
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Dashboards": dashboards,
		"Content":    "index",
	})
}

func renderSectorHedge(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Content": "sector_hedge",
	})
}
func handleSectorHedgeSimulate(c *gin.Context) {
	indexSymbol := strings.ToUpper(c.PostForm("index_symbol"))
	if indexSymbol == "" { indexSymbol = "QQQ" }

	duration, _ := strconv.ParseFloat(c.PostForm("duration_years"), 64)
	rFR, _ := strconv.ParseFloat(c.PostForm("risk_free_rate"), 64)
	numPaths, _ := strconv.Atoi(c.PostForm("num_paths"))

	// Fetch Index Data
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
	totalLongValue := 0.0

	for _, sym := range stockSymbols {
		if sym == "" { continue }

		qty, _ := strconv.ParseFloat(c.PostForm("qty_"+sym), 64)
		if qty == 0 { qty = 100 } // Default

		price, _ := market.GetPrice(sym)
		val := qty * price

		beta, _ := market.GetBeta(sym, indexSymbol)
		idioVol, _ := market.GetIdiosyncraticVolatility(sym, indexSymbol, beta)

		stockPositions = append(stockPositions, val)
		stockBetas = append(stockBetas, beta)
		stockVols = append(stockVols, idioVol)

		totalWeightedBeta += (val * beta)
		totalLongValue += val
	}

	// Option Params
	optionStrike, _ := strconv.ParseFloat(c.PostForm("option_strike"), 64)
	if optionStrike == 0 { optionStrike = indexPrice } // ATM

	optionExpiry, _ := strconv.ParseFloat(c.PostForm("option_expiry"), 64)
	optionQty, _ := strconv.ParseFloat(c.PostForm("option_quantity"), 64)

	// Automatic Hedge: Short the index to neutralize the total beta
	shortIndexPos, _ := strconv.ParseFloat(c.PostForm("short_index_pos"), 64)
	hedgeRatio, _ := strconv.ParseFloat(c.PostForm("hedge_ratio"), 64)
	if hedgeRatio == 0 { hedgeRatio = 100 }
	
	if shortIndexPos == 0 {
		shortIndexPos = totalWeightedBeta * (hedgeRatio / 100.0) // Apply multiplier
	}

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
	c.HTML(http.StatusOK, "sector_hedge_results.html", resp)
}

func parseFormFloat(c *gin.Context, key string) float64 {
	val, _ := strconv.ParseFloat(c.PostForm(key), 64)
	return val
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

	dashboards, _ := db.GetDashboards()
	c.HTML(http.StatusOK, "dashboard_list.html", gin.H{"Dashboards": dashboards})
}

func handleLoadDashboard(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	dash, err := db.GetDashboardByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}
	c.HTML(http.StatusOK, "simulator_form.html", gin.H{
		"Dash": dash,
	})
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

	if req.VolATech == 0 { req.VolATech, _ = market.GetVolatility("QQQ") }
	if req.VolBEnergy == 0 { req.VolBEnergy, _ = market.GetVolatility("XLE") }
	if req.VolCBonds == 0 { req.VolCBonds, _ = market.GetVolatility("TLT") }
	if req.VolDCrypto == 0 { req.VolDCrypto, _ = market.GetVolatility("BTC") }
	if req.VolEGold == 0 { req.VolEGold, _ = market.GetVolatility("GLD") }

	resp := simulation.Run(req)
	go func() {
		db.SaveSimulation(req, resp)
	}()

	c.HTML(http.StatusOK, "results.html", gin.H{
		"Metrics":      resp.Metrics,
		"InputSummary": resp.InputSummary,
		"ChartData":    resp.ChartData,
	})
}
