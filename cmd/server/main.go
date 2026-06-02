package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/lukasenderle/unicorn_research_go/internal/api"
	"github.com/lukasenderle/unicorn_research_go/internal/db"
)

func main() {
	// Initialize Database
	db.InitDB("./simulations.db")

	r := gin.Default()

	// Load Templates
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	// CORS configuration (Public API)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api.RegisterRoutes(r)
	api.RegisterUIHandlers(r)

	log.Println("Public Quant API starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
