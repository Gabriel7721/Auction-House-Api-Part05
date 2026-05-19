package main

import (
	"main/config"
	"main/handlers"
	"main/middleware"
	"main/repository"
	"main/services"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	config.LoadConfig()
	config.ConnectDatabase()
	config.AutoMigrate()
	config.SeedCategories()

	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	userRepo := repository.NewUserRepository(config.DB)
	categoryRepo := repository.NewCategoryRepository(config.DB)
	productRepo := repository.NewProductRepository(config.DB)
	auctionRepo := repository.NewAuctionRepository(config.DB)
	bidRepo := repository.NewBidRepository(config.DB)
	watchlistRepo := repository.NewWatchlistRepository(config.DB)
	notificationRepo := repository.NewNotificationRepository(config.DB)
	devicePushTokenRepo := repository.NewDevicePushTokenRepository(config.DB)

	authService := services.NewAuthService(userRepo)
	categoryService := services.NewCategoryService(categoryRepo)
	auctionService := services.NewAuctionService(
		config.DB,
		auctionRepo,
		productRepo,
		bidRepo,
		watchlistRepo,
		categoryRepo,
		notificationRepo,
	)
	notificationService := services.NewNotificationService(notificationRepo)
	pushTokenService := services.NewPushTokenService(devicePushTokenRepo)
	pushNotificationService := services.NewPushNotificationService(
		devicePushTokenRepo)

	authHandler := handlers.NewAuthHandler(authService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	auctionHandler := handlers.NewAuctionHandler(auctionService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	pushTokenHandler := handlers.NewPushTokenHandler(
		pushTokenService,
		pushNotificationService,
	)

	api := r.Group("/api")
	{
		api.GET("/health", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{
				"message": "auction house api is running",
			})
		})
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", middleware.AuthMiddleware(), authHandler.Logout)
			auth.GET("/me", middleware.AuthMiddleware(), authHandler.Me)
		}

		api.GET("/categories", categoryHandler.ListCategories)

		auctions := api.Group("/auctions")
		{
			auctions.GET("", auctionHandler.ListAuctions)
			auctions.GET("/:id", auctionHandler.GetAuctionDetail)
			auctions.GET("/:id/bids", auctionHandler.GetAuctionBids)

			auctions.POST("", middleware.AuthMiddleware(), auctionHandler.CreateAuction)
			auctions.PUT("/:id", middleware.AuthMiddleware(), auctionHandler.UpdateAuction)
			auctions.PATCH("/:id/cancel", middleware.AuthMiddleware(), auctionHandler.CancelAuction)

			auctions.POST("/:id/bids", middleware.AuthMiddleware(), auctionHandler.PlaceBid)

			auctions.POST("/:id/watch", middleware.AuthMiddleware(), auctionHandler.WatchAuction)
			auctions.DELETE("/:id/watch", middleware.AuthMiddleware(), auctionHandler.UnwatchAuction)
		}

		me := api.Group("/me")
		me.Use(middleware.AuthMiddleware())
		{
			me.GET("/watchlist", auctionHandler.GetMyWatchlist)
			me.GET("/auctions", auctionHandler.GetMyAuctions)
			me.GET("/bids", auctionHandler.GetMyBids)

			me.GET("/notifications", notificationHandler.GetMyNotifications)
			me.GET("/notifications/unread-count", notificationHandler.GetUnreadCount)
			me.PATCH("/notifications/:id/read", notificationHandler.MarkAsRead)
			me.PATCH("/notifications/read-all", notificationHandler.MarkAllAsRead)

			me.POST("/push-token", pushTokenHandler.RegisterPushToken)
			me.DELETE("/push-token", pushTokenHandler.UnregisterPushToken)
			me.POST("/push-token/test", pushTokenHandler.SendTestPushNotification)

		}
	}

	r.Run(config.Port)
}
