package api

import (
	"log"
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
	r.POST("/ui/simulate", handleUISimulate)
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
		err := db.SaveSimulation(req, resp)
		if err != nil {
			log.Printf("[DB] Error saving simulation: %v", err)
		}
	}()

	c.HTML(http.StatusOK, "results.html", gin.H{
		"Metrics":      resp.Metrics,
		"InputSummary": resp.InputSummary,
		"ChartData":    resp.ChartData,
	})
}
