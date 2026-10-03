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

type NewsHandler struct {
	newsService service.NewsService
}

func NewNewsHandler(newsService service.NewsService) *NewsHandler {
	return &NewsHandler{
		newsService: newsService,
	}
}

type createNewsRequest struct {
	CategoryID int64            `json:"category_id" binding:"required"`
	AuthorID   int64            `json:"author_id" binding:"required"`
	Slug       string           `json:"slug" binding:"required"`
	Thumbnail  *string          `json:"thumbnail"`
	TitleIDN   string           `json:"title_idn" binding:"required"`
	TitleEN    string           `json:"title_en" binding:"required"`
	ContentIDN string           `json:"content_idn" binding:"required"`
	ContentEN  string           `json:"content_en" binding:"required"`
	Status     model.NewsStatus `json:"status"`
}

type updateNewsRequest struct {
	CategoryID int64            `json:"category_id" binding:"required"`
	AuthorID   int64            `json:"author_id" binding:"required"`
	Slug       string           `json:"slug" binding:"required"`
	Thumbnail  *string          `json:"thumbnail"`
	TitleIDN   string           `json:"title_idn" binding:"required"`
	TitleEN    string           `json:"title_en" binding:"required"`
	ContentIDN string           `json:"content_idn" binding:"required"`
	ContentEN  string           `json:"content_en" binding:"required"`
	Status     model.NewsStatus `json:"status" binding:"required"`
}

func (h *NewsHandler) GetAll(c *gin.Context) {
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

	news, err := h.newsService.GetAll(
		c.Request.Context(),
		limit,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get news",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": news,
	})
}

func (h *NewsHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid news id",
		})
		return
	}

	news, err := h.newsService.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		if errors.Is(err, service.ErrNewsNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "news not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get news",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": news,
	})
}

func (h *NewsHandler) GetBySlug(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))

	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid news slug",
		})
		return
	}

	news, err := h.newsService.GetBySlug(
		c.Request.Context(),
		slug,
	)
	if err != nil {
		if errors.Is(err, service.ErrNewsNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "news not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get news",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": news,
	})
}

func (h *NewsHandler) Create(c *gin.Context) {
	var req createNewsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	news := &model.News{
		CategoryID: req.CategoryID,
		AuthorID:   req.AuthorID,
		Slug:       strings.TrimSpace(req.Slug),
		Thumbnail:  req.Thumbnail,
		TitleIDN:   strings.TrimSpace(req.TitleIDN),
		TitleEN:    strings.TrimSpace(req.TitleEN),
		ContentIDN: req.ContentIDN,
		ContentEN:  req.ContentEN,
		Status:     req.Status,
	}

	if err := h.newsService.Create(
		c.Request.Context(),
		news,
	); err != nil {
		if errors.Is(err, service.ErrInvalidNews) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create news",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "news created successfully",
		"data":    news,
	})
}

func (h *NewsHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid news id",
		})
		return
	}

	var req updateNewsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	news := &model.News{
		ID:         id,
		CategoryID: req.CategoryID,
		AuthorID:   req.AuthorID,
		Slug:       strings.TrimSpace(req.Slug),
		Thumbnail:  req.Thumbnail,
		TitleIDN:   strings.TrimSpace(req.TitleIDN),
		TitleEN:    strings.TrimSpace(req.TitleEN),
		ContentIDN: req.ContentIDN,
		ContentEN:  req.ContentEN,
		Status:     req.Status,
	}

	if err := h.newsService.Update(
		c.Request.Context(),
		news,
	); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidNews):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
		case errors.Is(err, service.ErrNewsNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"message": "news not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to update news",
			})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "news updated successfully",
		"data":    news,
	})
}

func (h *NewsHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid news id",
		})
		return
	}

	if err := h.newsService.Delete(
		c.Request.Context(),
		id,
	); err != nil {
		if errors.Is(err, service.ErrNewsNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "news not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to delete news",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "news deleted successfully",
	})
}
