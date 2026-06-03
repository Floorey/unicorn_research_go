package db

import (
	"encoding/json"
	"log"

	"github.com/lukasenderle/unicorn_research_go/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dataSourceName string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Auto migrate models
	err = DB.AutoMigrate(&models.Asset{}, &models.PortfolioAsset{}, &models.SimulationResult{}, &models.Dashboard{})
	if err != nil {
		log.Fatalf("Error during migration: %v", err)
	}
	log.Println("Database initialized and migrated successfully")
}

func GetDashboards() ([]models.Dashboard, error) {
	var dashboards []models.Dashboard
	err := DB.Order("created_at desc").Find(&dashboards).Error
	return dashboards, err
}

func GetDashboardByID(id uint) (models.Dashboard, error) {
	var dash models.Dashboard
	err := DB.Where("id = ?", id).First(&dash).Error
	return dash, err
}

func SaveDashboard(dash models.Dashboard) error {
	return DB.Save(&dash).Error
}

func SaveSimulation(req models.SimulationRequest, resp models.SimulationResponse) error {
	inputJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	record := models.SimulationResult{
		InputJSON:       string(inputJSON),
		TotalInvestment: resp.InputSummary.TotalInvestment,
		Var99:           resp.Metrics.Var99Percent,
		ES99:            resp.Metrics.ExpectedShortfall,
		MeanEur:         resp.Metrics.ExpectedMeanEur,
		SharpeRatio:     resp.Metrics.SharpeRatio,
		SortinoRatio:    resp.Metrics.SortinoRatio,
		MaxDrawdown:     resp.Metrics.MaxDrawdown,
	}

	return DB.Create(&record).Error
}

func GetSimulations(limit int) ([]models.SimulationResult, error) {
	var results []models.SimulationResult
	err := DB.Order("created_at desc").Limit(limit).Find(&results).Error
	return results, err
}
