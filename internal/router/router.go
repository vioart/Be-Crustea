package router

import (
	"net/http"

	"Be-Crustea/internal/handler"
	"Be-Crustea/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Auth        *handler.AuthHandler
	News        *handler.NewsHandler
	Product     *handler.ProductHandler
	FAQ         *handler.FAQHandler
	Partnership *handler.PartnershipHandler
}

func Setup(h Handlers, jwtSecret string) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORS())

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
		auth.GET("/me", middleware.AuthRequired(jwtSecret), h.Auth.Me)
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
	// News
	// --------------------------------------------

	admin.POST("/news", h.News.Create)
	admin.PUT("/news/:id", h.News.Update)
	admin.DELETE("/news/:id", h.News.Delete)

	// --------------------------------------------
	// Products
	// --------------------------------------------

	admin.POST("/products", h.Product.Create)
	admin.PUT("/products/:id", h.Product.Update)
	admin.DELETE("/products/:id", h.Product.Delete)

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

	admin.POST(
		"/partnership-projects",
		h.Partnership.Create,
	)

	admin.PUT(
		"/partnership-projects/:id",
		h.Partnership.Update,
	)

	admin.DELETE(
		"/partnership-projects/:id",
		h.Partnership.Delete,
	)

	return r
}
