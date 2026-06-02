package db

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/lukasenderle/unicorn_research_go/internal/models"
)

var DB *sql.DB

func InitDB(dataSourceName string) {
	var err error
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Create table for simulation results
	query := `
	CREATE TABLE IF NOT EXISTS simulations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		input_json TEXT,
		total_investment REAL,
		var_99 REAL,
		es_99 REAL,
		mean_eur REAL
	);`

	_, err = DB.Exec(query)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
	log.Println("Database initialized successfully")
}

func SaveSimulation(req models.SimulationRequest, resp models.SimulationResponse) error {
	inputJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}

	query := `INSERT INTO simulations (input_json, total_investment, var_99, es_99, mean_eur) VALUES (?, ?, ?, ?, ?)`
	_, err = DB.Exec(query, string(inputJSON), resp.InputSummary.TotalInvestment, resp.Metrics.Var99Percent, resp.Metrics.ExpectedShortfall, resp.Metrics.ExpectedMeanEur)
	return err
}

type SimulationRecord struct {
	ID              int       `json:"id"`
	Timestamp       time.Time `json:"timestamp"`
	Input           string    `json:"input"`
	TotalInvestment float64   `json:"total_investment"`
	Var99           float64   `json:"var_99"`
	ES99            float64   `json:"es_99"`
	MeanEur         float64   `json:"mean_eur"`
}

func GetSimulations(limit int) ([]SimulationRecord, error) {
	rows, err := DB.Query("SELECT id, timestamp, input_json, total_investment, var_99, es_99, mean_eur FROM simulations ORDER BY timestamp DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []SimulationRecord
	for rows.Next() {
		var r SimulationRecord
		err := rows.Scan(&r.ID, &r.Timestamp, &r.Input, &r.TotalInvestment, &r.Var99, &r.ES99, &r.MeanEur)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}
