package main

import (
	"log"
	"os"

	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/config"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/database"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/handler"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/jobs"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/middleware"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/repository"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/service"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/telegram"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.json"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize database
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	logger.Info("Connected to database")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	nodeRepo := repository.NewNodeRepository(db)
	planRepo := repository.NewPlanRepository(db)
	labelRepo := repository.NewLabelRepository(db)
	usageRepo := repository.NewUsageRepository(db)
	uuidRepo := repository.NewUUIDRepository(db)
	onlineRepo := repository.NewOnlineUserRepository(db)

	// Initialize services
	authService := service.NewAuthService(&cfg.Auth, userRepo, db)
	accountingService := service.NewAccountingService(userRepo, nodeRepo, planRepo, usageRepo, uuidRepo, logger)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userRepo, nodeRepo, planRepo, accountingService, authService)
	adminHandler := handler.NewAdminHandler(userRepo, nodeRepo, planRepo, labelRepo, uuidRepo, authService)
	nodeHandler := handler.NewNodeHandler(nodeRepo, userRepo, planRepo, uuidRepo, onlineRepo, accountingService, logger)

	// Initialize Telegram bot
	telegramBot, err := telegram.NewBot(&cfg.Telegram, userRepo, logger)
	if err != nil {
		logger.Error("Failed to initialize Telegram bot", zap.Error(err))
	} else if telegramBot != nil {
		go telegramBot.Start()
		logger.Info("Telegram bot started")
	}

	// Initialize background jobs
	jobScheduler := jobs.NewJobScheduler(db, accountingService, userRepo, usageRepo, telegramBot, logger)
	jobScheduler.Start()

	// Initialize Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS middleware
	corsOrigins := cfg.Server.GetCORSOrigins()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Forwarded-For", "X-Real-IP"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Trusted proxy middleware (handles X-Forwarded-For from localhost)
	r.Use(middleware.TrustedProxyMiddleware(cfg.Server.TrustedProxies))

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth endpoints (public)
	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
	}

	// User endpoints (authenticated)
	userGroup := r.Group("/api/v1/me")
	userGroup.Use(middleware.AuthMiddleware(authService))
	{
		userGroup.GET("", userHandler.GetMe)
		userGroup.GET("/plan", userHandler.GetMyPlan)
		userGroup.GET("/nodes", userHandler.GetMyNodes)
		userGroup.GET("/usage", userHandler.GetMyUsage)
		userGroup.GET("/usage/history", userHandler.GetMyUsageHistory)
		userGroup.POST("/telegram/link", userHandler.GenerateTelegramLink)
	}

	// Admin endpoints (authenticated + admin role)
	adminGroup := r.Group("/api/v1/admin")
	adminGroup.Use(middleware.AuthMiddleware(authService))
	adminGroup.Use(middleware.AdminMiddleware())
	{
		// Users
		adminGroup.POST("/users", adminHandler.CreateUser)
		adminGroup.GET("/users", adminHandler.ListUsers)
		adminGroup.GET("/users/:id", adminHandler.GetUser)
		adminGroup.PUT("/users/:id", adminHandler.UpdateUser)
		adminGroup.DELETE("/users/:id", adminHandler.DeleteUser)

		// Nodes
		adminGroup.POST("/nodes", adminHandler.CreateNode)
		adminGroup.GET("/nodes", adminHandler.ListNodes)

		// Plans
		adminGroup.POST("/plans", adminHandler.CreatePlan)
		adminGroup.GET("/plans", adminHandler.ListPlans)

		// Labels
		adminGroup.POST("/labels", adminHandler.CreateLabel)
		adminGroup.GET("/labels", adminHandler.ListLabels)
	}

	// Node protocol endpoints (Xboard-compatible)
	// V1 API (UniProxy)
	nodeV1 := r.Group("/api/v1/server/UniProxy")
	nodeV1.Use(middleware.NodeAuthMiddleware(&cfg.Node, nodeRepo))
	{
		nodeV1.GET("/config", nodeHandler.GetConfig)
		nodeV1.GET("/user", nodeHandler.GetUsers)
		nodeV1.POST("/push", nodeHandler.PushTraffic)
		nodeV1.POST("/alive", nodeHandler.PushAlive)
		nodeV1.GET("/alivelist", nodeHandler.GetAliveList)
		nodeV1.POST("/status", nodeHandler.PushStatus)
	}

	// V2 API (same implementation)
	nodeV2 := r.Group("/api/v2/server")
	nodeV2.Use(middleware.NodeAuthMiddleware(&cfg.Node, nodeRepo))
	{
		nodeV2.GET("/config", nodeHandler.GetConfig)
		nodeV2.GET("/user", nodeHandler.GetUsers)
		nodeV2.POST("/push", nodeHandler.PushTraffic)
		nodeV2.POST("/alive", nodeHandler.PushAlive)
		nodeV2.GET("/alivelist", nodeHandler.GetAliveList)
		nodeV2.POST("/status", nodeHandler.PushStatus)
	}

	// Start server
	addr := cfg.Server.GetAddress()
	logger.Info("Starting server", zap.String("address", addr))

	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
