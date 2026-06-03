package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lukasenderle/unicorn_research_go/internal/api"
	"github.com/lukasenderle/unicorn_research_go/internal/db"
	"html/template"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	// Initialize Database
	db.InitDB("./simulations.db")

	r := gin.Default()

	// Add pprof routes
	r.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
	r.POST("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))

	// Add Template Functions
	r.SetFuncMap(template.FuncMap{
		"seq": func(start, end int) []int {
			var res []int
			for i := start; i <= end; i++ {
				res = append(res, i)
			}
			return res
		},
		"multiply": func(a, b float64) float64 { return a * b },
	})

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
