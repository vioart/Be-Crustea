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

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

type createProductRequest struct {
	CategoryID     int64  `json:"category_id" binding:"required"`
	Slug           string `json:"slug" binding:"required"`
	NameIDN        string `json:"name_idn" binding:"required"`
	NameEN         string `json:"name_en" binding:"required"`
	DescriptionIDN string `json:"description_idn" binding:"required"`
	DescriptionEN  string `json:"description_en" binding:"required"`
}

type updateProductRequest struct {
	CategoryID     int64  `json:"category_id" binding:"required"`
	Slug           string `json:"slug" binding:"required"`
	NameIDN        string `json:"name_idn" binding:"required"`
	NameEN         string `json:"name_en" binding:"required"`
	DescriptionIDN string `json:"description_idn" binding:"required"`
	DescriptionEN  string `json:"description_en" binding:"required"`
}

func (h *ProductHandler) GetAll(c *gin.Context) {
	limit := 10
	offset := 0

	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid limit",
			})
			return
		}

		limit = parsed
	}

	if value := c.Query("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid offset",
			})
			return
		}

		offset = parsed
	}

	products, err := h.productService.GetAll(
		c.Request.Context(),
		limit,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get products",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
	})
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid product id",
		})
		return
	}

	product, err := h.productService.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "product not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get product",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": product,
	})
}

func (h *ProductHandler) GetBySlug(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))

	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid product slug",
		})
		return
	}

	product, err := h.productService.GetBySlug(
		c.Request.Context(),
		slug,
	)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "product not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get product",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": product,
	})
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req createProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	product := &model.Product{
		CategoryID:     req.CategoryID,
		Slug:           strings.TrimSpace(req.Slug),
		NameIDN:        strings.TrimSpace(req.NameIDN),
		NameEN:         strings.TrimSpace(req.NameEN),
		DescriptionIDN: req.DescriptionIDN,
		DescriptionEN:  req.DescriptionEN,
	}

	if err := h.productService.Create(
		c.Request.Context(),
		product,
	); err != nil {
		if errors.Is(err, service.ErrInvalidProduct) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create product",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "product created successfully",
		"data":    product,
	})
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid product id",
		})
		return
	}

	var req updateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	product := &model.Product{
		ID:             id,
		CategoryID:     req.CategoryID,
		Slug:           strings.TrimSpace(req.Slug),
		NameIDN:        strings.TrimSpace(req.NameIDN),
		NameEN:         strings.TrimSpace(req.NameEN),
		DescriptionIDN: req.DescriptionIDN,
		DescriptionEN:  req.DescriptionEN,
	}

	if err := h.productService.Update(
		c.Request.Context(),
		product,
	); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidProduct):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
		case errors.Is(err, service.ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"message": "product not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to update product",
			})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "product updated successfully",
		"data":    product,
	})
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid product id",
		})
		return
	}

	if err := h.productService.Delete(
		c.Request.Context(),
		id,
	); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "product not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to delete product",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "product deleted successfully",
	})
}
