package router

import (
	"net/http"

	"Be-Crustea/internal/handler"
	"Be-Crustea/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Auth         *handler.AuthHandler
	User         *handler.UserHandler
	News         *handler.NewsHandler
	NewsCategory *handler.NewsCategoryHandler
	Product      *handler.ProductHandler
	FAQ          *handler.FAQHandler
	Partnership  *handler.PartnershipHandler
}

func Setup(h Handlers, jwtSecret string) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORS())

	// ============================================
	// STATIC FILES
	// ============================================

	r.Static("/uploads", "./storage")

	// ============================================
	// HEALTH CHECK
	// ============================================

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Be-Crustea API is running",
		})
	})

	api := r.Group("/api")

	// ============================================
	// AUTH
	// ============================================

	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)

		auth.GET(
			"/me",
			middleware.AuthRequired(jwtSecret),
			h.Auth.Me,
		)
	}

	// ============================================
	// PUBLIC NEWS CATEGORIES
	// ============================================

	newsCategories := api.Group("/news-categories")
	{
		newsCategories.GET("", h.NewsCategory.GetAll)
		newsCategories.GET("/:id", h.NewsCategory.GetByID)
	}

	// ============================================
	// PUBLIC NEWS
	// ============================================

	news := api.Group("/news")
	{
		news.GET("", h.News.GetAll)
		news.GET("/slug/:slug", h.News.GetBySlug)
		news.GET("/:id", h.News.GetByID)
	}

	// ============================================
	// PUBLIC PRODUCTS
	// ============================================

	products := api.Group("/products")
	{
		products.GET("", h.Product.GetAll)
		products.GET("/slug/:slug", h.Product.GetBySlug)
		products.GET("/:id", h.Product.GetByID)
	}

	// ============================================
	// PUBLIC FAQS
	// ============================================

	faqs := api.Group("/faqs")
	{
		faqs.GET("", h.FAQ.GetAll)
		faqs.GET("/:id", h.FAQ.GetByID)
	}

	// ============================================
	// PUBLIC PARTNERSHIP
	// ============================================

	partnerships := api.Group("/partnership-projects")
	{
		partnerships.GET("", h.Partnership.GetAll)
		partnerships.GET("/:id", h.Partnership.GetByID)
	}

	// ============================================
	// PROTECTED ADMIN ROUTES
	// ============================================

	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(jwtSecret))

	// --------------------------------------------
	// Users
	// --------------------------------------------

	admin.GET("/users", h.User.GetAll)
	admin.GET("/users/:id", h.User.GetByID)
	admin.POST("/users", h.User.Create)
	admin.PUT("/users/:id", h.User.Update)
	admin.DELETE("/users/:id", h.User.Delete)

	// --------------------------------------------
	// News Categories
	// --------------------------------------------

	admin.GET("/news-categories", h.NewsCategory.GetAll)
	admin.GET("/news-categories/:id", h.NewsCategory.GetByID)
	admin.POST("/news-categories", h.NewsCategory.Create)
	admin.PUT("/news-categories/:id", h.NewsCategory.Update)
	admin.DELETE("/news-categories/:id", h.NewsCategory.Delete)

	// --------------------------------------------
	// News
	// --------------------------------------------

	admin.GET("/news", h.News.GetAll)
	admin.GET("/news/:slug", h.News.GetBySlug)

	admin.POST("/news", h.News.Create)
	admin.POST("/news/content-image", h.News.UploadContentImage)
	admin.POST("/news/content-image/import", h.News.ImportContentImage)

	admin.PUT("/news/:slug", h.News.UpdateBySlug)
	admin.DELETE("/news/:slug", h.News.DeleteBySlug)

	// --------------------------------------------
	// Products
	// --------------------------------------------

	// Get products by category
	admin.GET("/products/technology", h.Product.GetTechnology)
	admin.GET("/products/krasty", h.Product.GetKrasty)
	admin.GET("/products/trading", h.Product.GetTrading)

	admin.GET("/products/:slug", h.Product.GetBySlug)
	admin.POST("/products", h.Product.Create)
	admin.PUT("/products/:slug", h.Product.UpdateBySlug)
	admin.DELETE("/products/:slug", h.Product.DeleteBySlug)

	// --------------------------------------------
	// FAQs
	// --------------------------------------------

	admin.GET("/faqs", h.FAQ.GetAll)
	admin.GET("/faqs/:id", h.FAQ.GetByID)
	admin.POST("/faqs", h.FAQ.Create)
	admin.PUT("/faqs/:id", h.FAQ.Update)
	admin.DELETE("/faqs/:id", h.FAQ.Delete)

	// --------------------------------------------
	// Partnership Projects
	// --------------------------------------------

	admin.POST("/partnership-projects", h.Partnership.Create)
	admin.PUT("/partnership-projects/:id", h.Partnership.Update)
	admin.DELETE("/partnership-projects/:id", h.Partnership.Delete)

	// ============================================
	// ACCOUNT SETTINGS
	// ============================================

	account := api.Group("/account")
	account.Use(middleware.AuthRequired(jwtSecret))
	{
		account.GET("/me", h.User.GetMyAccount)
		account.PUT("/me", h.User.UpdateMyAccount)
		account.PUT("/me/password", h.User.ChangeMyPassword)
	}

	return r
}
