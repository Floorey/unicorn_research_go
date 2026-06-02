package api

import (
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/lukasenderle/unicorn_research_go/internal/db"
	"github.com/lukasenderle/unicorn_research_go/internal/models"
	"github.com/lukasenderle/unicorn_research_go/internal/simulation"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/health", healthCheck)
	v1 := r.Group("/api/v1")
	{
		v1.POST("/simulate", runSimulation)
		v1.GET("/history", getHistory)
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func getHistory(c *gin.Context) {
	records, err := db.GetSimulations(50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, records)
}

func runSimulation(c *gin.Context) {
	var req models.SimulationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := simulation.Run(req)

	// Save to DB in background
	go func() {
		err := db.SaveSimulation(req, resp)
		if err != nil {
			log.Printf("[DB] Error saving simulation: %v", err)
		}
	}()

	c.JSON(http.StatusOK, resp)
}
