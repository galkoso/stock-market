package main

import (
	"log"

	"stock-market/backend/internal/config"
	"stock-market/backend/internal/finnhub"
	"stock-market/backend/internal/handler"
	"stock-market/backend/internal/middleware"
	"stock-market/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	client := finnhub.NewClient(cfg.FinnhubAPIKey)
	streamHub := finnhub.NewWSHub(cfg.FinnhubAPIKey)
	defer streamHub.Close()

	stockService := service.NewStockService(client)
	stockHandler := handler.NewStockHandler(stockService)
	streamHandler := handler.NewStockStreamHandler(stockService, streamHub)

	router := gin.Default()
	router.Use(middleware.CORS())

	api := router.Group("/api")
	{
		api.GET("/stocks/search", stockHandler.Search)
	}

	router.GET("/ws/stocks", streamHandler.Stream)

	log.Printf("backend listening on http://localhost:%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
