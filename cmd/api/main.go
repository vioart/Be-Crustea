package main

import (
	"log"

	"Be-Crustea/internal/cache"
	"Be-Crustea/internal/config"
	"Be-Crustea/internal/database"
	"Be-Crustea/internal/handler"
	"Be-Crustea/internal/repository"
	"Be-Crustea/internal/router"
	"Be-Crustea/internal/service"
)

func main() {
	// ============================================
	// CONFIG
	// ============================================

	cfg := config.Load()

	// ============================================
	// DATABASE
	// ============================================

	db, err := database.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// ============================================
	// REDIS
	// ============================================

	redisClient, err := cache.NewRedisClient(
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.RedisDB,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	// ============================================
	// REPOSITORIES
	// ============================================

	userRepository := repository.NewUserRepository(db)
	newsRepository := repository.NewNewsRepository(db)
	productRepository := repository.NewProductRepository(db)
	faqRepository := repository.NewFAQRepository(db)
	partnershipRepository := repository.NewPartnershipRepository(db)

	// ============================================
	// SERVICES
	// ============================================

	authService := service.NewAuthService(
		userRepository,
		cfg.JWTSecret,
	)

	newsService := service.NewNewsService(
		newsRepository,
	)

	productService := service.NewProductService(
		productRepository,
		redisClient,
	)

	faqService := service.NewFAQService(
		faqRepository,
	)

	partnershipService := service.NewPartnershipService(
		partnershipRepository,
	)

	// ============================================
	// HANDLERS
	// ============================================

	authHandler := handler.NewAuthHandler(
		authService,
	)

	newsHandler := handler.NewNewsHandler(
		newsService,
	)

	productHandler := handler.NewProductHandler(
		productService,
	)

	faqHandler := handler.NewFAQHandler(
		faqService,
	)

	partnershipHandler := handler.NewPartnershipHandler(
		partnershipService,
	)

	// ============================================
	// ROUTER
	// ============================================

	r := router.Setup(
		router.Handlers{
			Auth:        authHandler,
			News:        newsHandler,
			Product:     productHandler,
			FAQ:         faqHandler,
			Partnership: partnershipHandler,
		},
		cfg.JWTSecret,
	)

	// ============================================
	// SERVER
	// ============================================

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
