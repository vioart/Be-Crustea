package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"Be-Crustea/internal/model"
	"Be-Crustea/internal/service"

	"github.com/gin-gonic/gin"
)

type NewsCategoryHandler struct {
	service service.NewsCategoryService
}

func NewNewsCategoryHandler(
	service service.NewsCategoryService,
) *NewsCategoryHandler {
	return &NewsCategoryHandler{
		service: service,
	}
}

type newsCategoryRequest struct {
	Slug     string `json:"slug" binding:"required"`
	NameIDN  string `json:"name_idn" binding:"required"`
	NameEN   string `json:"name_en" binding:"required"`
	IsActive bool   `json:"is_active"`
}

// parseNewsCategoryID mengambil dan memvalidasi ID kategori dari URL.
func parseNewsCategoryID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid news category id",
		})
		return 0, false
	}

	return id, true
}

// GET /api/news-categories
// GET /api/admin/news-categories
func (h *NewsCategoryHandler) GetAll(c *gin.Context) {
	categories, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get news categories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": categories,
	})
}

// GET /api/news-categories/:id
// GET /api/admin/news-categories/:id
func (h *NewsCategoryHandler) GetByID(c *gin.Context) {
	id, ok := parseNewsCategoryID(c)
	if !ok {
		return
	}

	category, err := h.service.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNewsCategoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"message": "news category not found",
			})

		case errors.Is(err, service.ErrInvalidNewsCategory):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to get news category",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": category,
	})
}

// POST /api/admin/news-categories
func (h *NewsCategoryHandler) Create(c *gin.Context) {
	var req newsCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	category := &model.NewsCategory{
		Slug:     strings.TrimSpace(req.Slug),
		NameIDN:  strings.TrimSpace(req.NameIDN),
		NameEN:   strings.TrimSpace(req.NameEN),
		IsActive: req.IsActive,
	}

	if err := h.service.Create(
		c.Request.Context(),
		category,
	); err != nil {
		if errors.Is(err, service.ErrInvalidNewsCategory) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create news category",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "news category created successfully",
		"data":    category,
	})
}

// PUT /api/admin/news-categories/:id
func (h *NewsCategoryHandler) Update(c *gin.Context) {
	id, ok := parseNewsCategoryID(c)
	if !ok {
		return
	}

	var req newsCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	category := &model.NewsCategory{
		ID:       id,
		Slug:     strings.TrimSpace(req.Slug),
		NameIDN:  strings.TrimSpace(req.NameIDN),
		NameEN:   strings.TrimSpace(req.NameEN),
		IsActive: req.IsActive,
	}

	if err := h.service.Update(
		c.Request.Context(),
		category,
	); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidNewsCategory):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})

		case errors.Is(err, service.ErrNewsCategoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"message": "news category not found",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to update news category",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "news category updated successfully",
		"data":    category,
	})
}

// DELETE /api/admin/news-categories/:id
func (h *NewsCategoryHandler) Delete(c *gin.Context) {
	id, ok := parseNewsCategoryID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(
		c.Request.Context(),
		id,
	); err != nil {
		switch {
		case errors.Is(err, service.ErrNewsCategoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"message": "news category not found",
			})

		case errors.Is(err, service.ErrInvalidNewsCategory):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to delete news category",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "news category deleted successfully",
	})
}
