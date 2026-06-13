package main

import (
	"context"
	"log"
	"time"

	"stock-market/backend/internal/auth"
	"stock-market/backend/internal/cache"
	"stock-market/backend/internal/config"
	"stock-market/backend/internal/finnhub"
	"stock-market/backend/internal/forex"
	"stock-market/backend/internal/handler"
	"stock-market/backend/internal/middleware"
	mongopkg "stock-market/backend/internal/mongo"
	notificationshub "stock-market/backend/internal/notifications"
	finnhubprovider "stock-market/backend/internal/provider/finnhub"
	"stock-market/backend/internal/repositories"
	"stock-market/backend/internal/scheduler"
	"stock-market/backend/internal/service"
	"stock-market/backend/internal/services"
	"stock-market/backend/internal/telegram"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mongoDB, err := mongopkg.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("mongo connection error: %v", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := mongoDB.Close(shutdownCtx); err != nil {
			log.Printf("mongo disconnect error: %v", err)
		}
	}()

	redisCache, err := cache.Connect(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis connection error: %v", err)
	}
	defer func() {
		if err := redisCache.Close(); err != nil {
			log.Printf("redis disconnect error: %v", err)
		}
	}()

	finnhubClient := finnhub.NewClient(cfg.FinnhubAPIKey)
	marketProvider := finnhubprovider.New(finnhubClient, redisCache)
	forexClient := forex.NewClient(redisCache)
	streamHub := finnhub.NewWSHub(cfg.FinnhubAPIKey)
	defer streamHub.Close()

	stockService := service.NewStockService(finnhubClient)
	stockHandler := handler.NewStockHandler(stockService)
	streamHandler := handler.NewStockStreamHandler(stockService, streamHub)

	watchlistRepo := repositories.NewWatchlistRepository(mongoDB.Watchlist)
	alertsRepo := repositories.NewAlertsRepository(mongoDB.Alerts)
	notificationsRepo := repositories.NewNotificationsRepository(mongoDB.Notifications)
	portfolioRepo := repositories.NewPortfolioRepository(mongoDB.Portfolio)

	marketService := services.NewMarketService(marketProvider)
	watchlistService := services.NewWatchlistService(watchlistRepo)
	portfolioService := services.NewPortfolioService(portfolioRepo, marketProvider, forexClient)
	alertsService := services.NewAlertsService(alertsRepo)
	notificationsService := services.NewNotificationsService(notificationsRepo)
	notificationHub := notificationshub.NewHub()
	telegramNotifier := telegram.NewNotifier(telegram.Config{
		BotToken: cfg.Telegram.BotToken,
		ChatID:   cfg.Telegram.ChatID,
	})
	if telegramNotifier.Enabled() {
		log.Println("telegram notifier enabled")
	} else {
		log.Println("telegram notifier disabled (set TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID to enable)")
	}
	alertEngine := services.NewAlertEngine(marketProvider, alertsRepo, notificationsRepo, notificationHub, telegramNotifier)

	marketHandler := handler.NewMarketHandler(marketService, watchlistService)
	watchlistHandler := handler.NewWatchlistHandler(watchlistService)
	portfolioHandler := handler.NewPortfolioHandler(portfolioService)
	alertsHandler := handler.NewAlertsHandler(alertsService, alertEngine)
	notificationsHandler := handler.NewNotificationsHandler(notificationsService, notificationHub, telegramNotifier)

	authService := auth.NewService(cfg.Auth, mongoDB.Users)
	authHandler := auth.NewHandler(authService, cfg.Auth)
	authMiddleware := auth.Authenticate(cfg.Auth)

	schedulerCtx, schedulerCancel := context.WithCancel(context.Background())
	defer schedulerCancel()
	scheduler.New(marketProvider, watchlistRepo, alertsRepo, notificationsRepo, alertEngine).Start(schedulerCtx)

	router := gin.Default()
	router.Use(middleware.CORS())

	authRoutes := router.Group("/api/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
		authRoutes.GET("/refresh", authHandler.Refresh)
		authRoutes.POST("/logout", authHandler.Logout)
	}

	api := router.Group("/api")
	api.Use(authMiddleware)
	{
		api.GET("/stocks/search", stockHandler.Search)
		api.GET("/stocks/quotes", stockHandler.Quotes)
		api.GET("/stocks/:symbol", marketHandler.Details)
		api.GET("/stocks/:symbol/recommendations", marketHandler.Recommendations)
		api.GET("/market/search", marketHandler.Search)
		api.GET("/earnings", marketHandler.Earnings)
		api.GET("/earnings/history", marketHandler.EarningsHistory)
		api.GET("/watchlist/earnings", marketHandler.WatchlistEarnings)
		api.GET("/stocks/:symbol/earnings-surprises", marketHandler.EarningsSurprises)
		api.GET("/news/:symbol", marketHandler.News)
		api.GET("/filings/:symbol", marketHandler.Filings)
		api.GET("/movers", marketHandler.Movers)
		api.GET("/watchlist", watchlistHandler.List)
		api.POST("/watchlist", watchlistHandler.Add)
		api.DELETE("/watchlist/:symbol", watchlistHandler.Remove)
		api.GET("/portfolio/allocation", portfolioHandler.Allocation)
		api.GET("/portfolio", portfolioHandler.List)
		api.POST("/portfolio", portfolioHandler.Add)
		api.PUT("/portfolio/:symbol", portfolioHandler.Update)
		api.DELETE("/portfolio/:symbol", portfolioHandler.Remove)
		api.GET("/alerts", alertsHandler.List)
		api.POST("/alerts", alertsHandler.Create)
		api.POST("/alerts/evaluate", alertsHandler.Evaluate)
		api.DELETE("/alerts/:id", alertsHandler.Delete)
		api.GET("/notifications", notificationsHandler.List)
		api.GET("/notifications/stream", notificationsHandler.Stream)
		api.POST("/notifications/:id/read", notificationsHandler.MarkRead)
		api.POST("/notifications/read-all", notificationsHandler.MarkAllRead)
		api.POST("/notifications/telegram/test", notificationsHandler.TestTelegram)
	}

	router.GET("/ws/stocks", authMiddleware, streamHandler.Stream)

	log.Printf("backend listening on http://localhost:%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
