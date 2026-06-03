package api

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lukasenderle/unicorn_research_go/internal/db"
	"github.com/lukasenderle/unicorn_research_go/internal/market"
	"github.com/lukasenderle/unicorn_research_go/internal/models"
	"github.com/lukasenderle/unicorn_research_go/internal/simulation"
)

func RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/simulate", runSimulation)
		v1.POST("/simulate/sector-hedge", runSectorHedgeAPI)
		v1.GET("/history", getHistory)
		v1.GET("/history/export", exportHistoryCSV)

		// Asset Management
		v1.GET("/assets", getAssets)
		v1.POST("/assets", addAsset)

		// Portfolio Management
		v1.GET("/portfolio", getPortfolio)
		v1.POST("/portfolio", addPortfolioAsset)
		v1.DELETE("/portfolio/:id", removePortfolioAsset)

		// Risk Management
		v1.GET("/risk-report", getRiskReport)
		v1.GET("/risk-report/export", exportRiskReportCSV)

		// Dashboard Management
		v1.GET("/dashboards", listDashboards)
		v1.POST("/dashboards", saveDashboard)
		v1.GET("/dashboards/:id", getDashboard)

		// Derivatives Pricing
		v1.POST("/derivatives/future", runFuturePricingAPI)
		v1.POST("/derivatives/swap", runSwapPricingAPI)
	}
}

func runFuturePricingAPI(c *gin.Context) {
	var req models.FuturePricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := simulation.PriceFuture(req)
	c.JSON(http.StatusOK, resp)
}

func runSwapPricingAPI(c *gin.Context) {
	var req models.SwapPricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := simulation.PriceSwap(req)
	c.JSON(http.StatusOK, resp)
}

func runSectorHedgeAPI(c *gin.Context) {
	var req models.SectorHedgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := simulation.RunSectorHedge(req)
	go func() {
		db.SaveSectorHedge(req, resp)
	}()
	c.JSON(http.StatusOK, resp)
}

func listDashboards(c *gin.Context) {
	dashboards, err := db.GetDashboards()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch dashboards"})
		return
	}
	c.JSON(http.StatusOK, dashboards)
}

func getDashboard(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	dash, err := db.GetDashboardByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dashboard not found"})
		return
	}
	c.JSON(http.StatusOK, dash)
}

func saveDashboard(c *gin.Context) {
	var dash models.Dashboard
	if err := c.ShouldBindJSON(&dash); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dashboard data"})
		return
	}

	dash.CreatedAt = time.Now()

	if err := db.SaveDashboard(dash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save dashboard"})
		return
	}

	c.JSON(http.StatusCreated, dash)
}

func getHistory(c *gin.Context) {
	records, err := db.GetSimulations(50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}

func exportHistoryCSV(c *gin.Context) {
	records, err := db.GetSimulations(1000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=simulation_history.csv")
	c.Header("Content-Type", "text/csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	writer.Write([]string{"ID", "Timestamp", "Total Investment", "VaR 99%", "ES 99%", "Mean (EUR)", "Sharpe Ratio", "Sortino Ratio", "Max Drawdown %"})

	for _, r := range records {
		row := []string{
			fmt.Sprintf("%d", r.ID),
			r.CreatedAt.Format(time.RFC3339),
			fmt.Sprintf("%.2f", r.TotalInvestment),
			fmt.Sprintf("%.2f", r.Var99),
			fmt.Sprintf("%.2f", r.ES99),
			fmt.Sprintf("%.2f", r.MeanEur),
			fmt.Sprintf("%.2f", r.SharpeRatio),
			fmt.Sprintf("%.2f", r.SortinoRatio),
			fmt.Sprintf("%.2f", r.MaxDrawdown),
		}
		writer.Write(row)
	}
}

func getAssets(c *gin.Context) {
	var assets []models.Asset
	if err := db.DB.Find(&assets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assets)
}

func addAsset(c *gin.Context) {
	var asset models.Asset
	if err := c.ShouldBindJSON(&asset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.DB.Create(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, asset)
}

func getPortfolio(c *gin.Context) {
	var portfolio []models.PortfolioAsset
	if err := db.DB.Preload("Asset").Find(&portfolio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, portfolio)
}

func addPortfolioAsset(c *gin.Context) {
	var req struct {
		Symbol   string  `json:"symbol" binding:"required"`
		Quantity float64 `json:"quantity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol and quantity required"})
		return
	}

	var asset models.Asset
	if err := db.DB.Where(models.Asset{Symbol: req.Symbol}).FirstOrCreate(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	price, _ := market.GetPrice(req.Symbol)
	pAsset := models.PortfolioAsset{
		AssetID:       asset.ID,
		Quantity:      req.Quantity,
		PurchasePrice: price,
	}

	if err := db.DB.Create(&pAsset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, pAsset)
}

func removePortfolioAsset(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := db.DB.Delete(&models.PortfolioAsset{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Asset removed from portfolio"})
}

func generateRiskReport() (models.PortfolioRiskReport, error) {
	var portfolio []models.PortfolioAsset
	if err := db.DB.Preload("Asset").Find(&portfolio).Error; err != nil {
		return models.PortfolioRiskReport{}, err
	}

	if len(portfolio) == 0 {
		return models.PortfolioRiskReport{}, fmt.Errorf("portfolio is empty")
	}

	req := models.SimulationRequest{
		DurationYears: 1,
		NumPaths:      1000,
		InterestRate:  4.0,
		OilPrice:      80.0,
	}

	var totalValue float64
	assetRisks := make([]models.AssetRisk, 0)
	for _, pAsset := range portfolio {
		price, _ := market.GetPrice(pAsset.Asset.Symbol)
		val := pAsset.Quantity * price
		totalValue += val
		vol, _ := market.GetVolatility(pAsset.Asset.Symbol)

		switch pAsset.Asset.Type {
		case "tech", "stock":
			req.PosATech += val
			req.VolATech = vol
		case "energy":
			req.PosBEnergy += val
			req.VolBEnergy = vol
		case "bond":
			req.PosCBonds += val
			req.VolCBonds = vol
		case "crypto":
			req.PosDCrypto += val
			req.VolDCrypto = vol
		case "gold", "commodity":
			req.PosEGold += val
			req.VolEGold = vol
		default:
			req.PosATech += val
			req.VolATech = vol
		}

		assetRisks = append(assetRisks, models.AssetRisk{
			Symbol:     pAsset.Asset.Symbol,
			Value:      val,
			Volatility: vol,
		})
	}

	for i := range assetRisks {
		assetRisks[i].Weight = (assetRisks[i].Value / totalValue) * 100
		assetRisks[i].Contribution = (assetRisks[i].Weight * assetRisks[i].Volatility)
	}

	resp := simulation.Run(req)
	return models.PortfolioRiskReport{
		TotalValue:       totalValue,
		PortfolioMetrics: resp.Metrics,
		AssetBreakdown:   assetRisks,
		Timestamp:        time.Now(),
	}, nil
}

func getRiskReport(c *gin.Context) {
	report, err := generateRiskReport()
	if err != nil {
		if err.Error() == "portfolio is empty" {
			c.JSON(http.StatusOK, gin.H{"message": "Portfolio is empty"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

func exportRiskReportCSV(c *gin.Context) {
	report, err := generateRiskReport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=portfolio_risk_report.csv")
	c.Header("Content-Type", "text/csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	writer.Write([]string{"Portfolio Risk Report", report.Timestamp.Format(time.RFC3339)})
	writer.Write([]string{"Total Portfolio Value", fmt.Sprintf("%.2f", report.TotalValue)})
	writer.Write([]string{""})
	writer.Write([]string{"Portfolio Metrics"})
	writer.Write([]string{"Metric", "Value"})
	writer.Write([]string{"VaR 99% (%)", fmt.Sprintf("%.2f", report.PortfolioMetrics.Var99Percent)})
	writer.Write([]string{"Expected Shortfall (%)", fmt.Sprintf("%.2f", report.PortfolioMetrics.ExpectedShortfall)})
	writer.Write([]string{"Expected Mean (EUR)", fmt.Sprintf("%.2f", report.PortfolioMetrics.ExpectedMeanEur)})
	writer.Write([]string{"Sharpe Ratio", fmt.Sprintf("%.2f", report.PortfolioMetrics.SharpeRatio)})
	writer.Write([]string{"Sortino Ratio", fmt.Sprintf("%.2f", report.PortfolioMetrics.SortinoRatio)})
	writer.Write([]string{"Max Drawdown (%)", fmt.Sprintf("%.2f", report.PortfolioMetrics.MaxDrawdown)})
	writer.Write([]string{""})

	writer.Write([]string{"Asset Breakdown"})
	writer.Write([]string{"Symbol", "Value", "Weight %", "Volatility", "Risk Contribution"})
	for _, a := range report.AssetBreakdown {
		writer.Write([]string{
			a.Symbol,
			fmt.Sprintf("%.2f", a.Value),
			fmt.Sprintf("%.2f", a.Weight),
			fmt.Sprintf("%.2f", a.Volatility),
			fmt.Sprintf("%.2f", a.Contribution),
		})
	}
}

func runSimulation(c *gin.Context) {
	var req models.SimulationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
	go func() {
		db.SaveSimulation(req, resp)
	}()
	c.JSON(http.StatusOK, resp)
}
