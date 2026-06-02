package simulation

import (
	"testing"
	"github.com/lukasenderle/unicorn_research_go/internal/models"
)

func TestRun(t *testing.T) {
	req := models.SimulationRequest{
		InterestRate:  4.25,
		OilPrice:      82.50,
		PosATech:      307.70,
		PosBEnergy:    100.00,
		PosCBonds:     200.00,
		VolATech:      0.18,
		VolBEnergy:    0.18,
		VolCBonds:     0.05,
		DurationYears: 5,
		NumPaths:      100,
	}

	resp := Run(req)

	if resp.Status != "success" {
		t.Errorf("Expected status success, got %s", resp.Status)
	}

	expectedTotal := 607.70
	if resp.InputSummary.TotalInvestment != expectedTotal {
		t.Errorf("Expected total investment %f, got %f", expectedTotal, resp.InputSummary.TotalInvestment)
	}

	if len(resp.ChartData) != 100 {
		t.Errorf("Expected 100 visual paths, got %d", len(resp.ChartData))
	}
}
