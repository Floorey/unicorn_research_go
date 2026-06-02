package api

import (
	"net/http"
	"strconv"
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
	indexVol, _ := strconv.ParseFloat(c.PostForm("index_vol"), 64)
	optionStrike, _ := strconv.ParseFloat(c.PostForm("option_strike"), 64)
	optionExpiry, _ := strconv.ParseFloat(c.PostForm("option_expiry"), 64)
	optionQty, _ := strconv.ParseFloat(c.PostForm("option_quantity"), 64)
	shortIndexPos, _ := strconv.ParseFloat(c.PostForm("short_index_pos"), 64)
	duration, _ := strconv.ParseFloat(c.PostForm("duration_years"), 64)
	rFR, _ := strconv.ParseFloat(c.PostForm("risk_free_rate"), 64)
	numPaths, _ := strconv.Atoi(c.PostForm("num_paths"))

	stockPositions := []float64{
		parseFormFloat(c, "pos_1"),
		parseFormFloat(c, "pos_2"),
		parseFormFloat(c, "pos_3"),
	}
	stockBetas := []float64{
		parseFormFloat(c, "beta_1"),
		parseFormFloat(c, "beta_2"),
		parseFormFloat(c, "beta_3"),
	}
	stockVols := []float64{
		parseFormFloat(c, "vol_1"),
		parseFormFloat(c, "vol_2"),
		parseFormFloat(c, "vol_3"),
	}

	req := models.SectorHedgeRequest{
		IndexVol:       indexVol / 100.0,
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
