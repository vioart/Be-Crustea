package main

import (
	"log"
	"time"

	"Be-Crustea/internal/cache"
	"Be-Crustea/internal/config"
	"Be-Crustea/internal/database"
	"Be-Crustea/internal/handler"
	"Be-Crustea/internal/repository"
	"Be-Crustea/internal/router"
	"Be-Crustea/internal/service"
	"Be-Crustea/internal/storage"
)

func main() {
	// CONFIG
	cfg := config.Load()

	// DATABASE
	db, err := database.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// REDIS
	redisClient, err := cache.NewRedisClient(
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.RedisDB,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	// LOCAL STORAGE
	localStorage := storage.NewLocalStorage(
		"./storage",
		"/uploads",
	)

	newsStorage := storage.NewLocalStorage(
		"./storage",
		"/uploads",
	)

	productStorage := storage.NewLocalStorage(
		"./storage",
		"/uploads",
	)

	// TEMP STORAGE CLEANUP: khusus berita
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for {
			if err := newsStorage.CleanupTempFiles(24 * time.Hour); err != nil {
				log.Printf(
					"failed to cleanup temp storage: %v",
					err,
				)
			}

			<-ticker.C
		}
	}()

	// REPOSITORIES
	userRepository := repository.NewUserRepository(db)
	newsRepository := repository.NewNewsRepository(db)
	newsCategoryRepository :=
		repository.NewNewsCategoryRepository(db)
	productRepository := repository.NewProductRepository(db)
	faqRepository := repository.NewFAQRepository(db)
	partnershipRepository :=
		repository.NewPartnershipRepository(db)

	// SERVICES
	authService := service.NewAuthService(
		userRepository,
		cfg.JWTSecret,
	)

	userService := service.NewUserService(userRepository)
	newsService := service.NewNewsService(newsRepository)

	newsCategoryService :=
		service.NewNewsCategoryService(newsCategoryRepository)

	productService := service.NewProductService(
		productRepository,
		redisClient,
	)

	faqService := service.NewFAQService(faqRepository)

	partnershipService :=
		service.NewPartnershipService(partnershipRepository)

	// HANDLERS
	authHandler := handler.NewAuthHandler(authService)

	userHandler := handler.NewUserHandler(
		userService,
		localStorage,
	)

	newsHandler := handler.NewNewsHandler(
		newsService,
		newsStorage,
	)

	newsCategoryHandler :=
		handler.NewNewsCategoryHandler(newsCategoryService)

	productHandler := handler.NewProductHandler(
		productService,
		productStorage,
	)

	faqHandler := handler.NewFAQHandler(faqService)

	partnershipHandler :=
		handler.NewPartnershipHandler(partnershipService)

	// ROUTER
	r := router.Setup(
		router.Handlers{
			Auth:         authHandler,
			User:         userHandler,
			News:         newsHandler,
			NewsCategory: newsCategoryHandler,
			Product:      productHandler,
			FAQ:          faqHandler,
			Partnership:  partnershipHandler,
		},
		cfg.JWTSecret,
	)

	// SERVER
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
