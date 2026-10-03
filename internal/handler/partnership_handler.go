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

type PartnershipHandler struct {
	partnershipService service.PartnershipService
}

func NewPartnershipHandler(
	partnershipService service.PartnershipService,
) *PartnershipHandler {
	return &PartnershipHandler{
		partnershipService: partnershipService,
	}
}

type createPartnershipRequest struct {
	NameIDN        string  `json:"name_idn" binding:"required"`
	NameEN         string  `json:"name_en" binding:"required"`
	DescriptionIDN string  `json:"description_idn" binding:"required"`
	DescriptionEN  string  `json:"description_en" binding:"required"`
	Image          *string `json:"image"`
}

type updatePartnershipRequest struct {
	NameIDN        string  `json:"name_idn" binding:"required"`
	NameEN         string  `json:"name_en" binding:"required"`
	DescriptionIDN string  `json:"description_idn" binding:"required"`
	DescriptionEN  string  `json:"description_en" binding:"required"`
	Image          *string `json:"image"`
}

func (h *PartnershipHandler) GetAll(c *gin.Context) {
	projects, err := h.partnershipService.GetAll(
		c.Request.Context(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get partnership projects",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": projects,
	})
}

func (h *PartnershipHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid partnership project id",
		})
		return
	}

	project, err := h.partnershipService.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		if errors.Is(err, service.ErrPartnershipProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "partnership project not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get partnership project",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": project,
	})
}

func (h *PartnershipHandler) Create(c *gin.Context) {
	var req createPartnershipRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	project := &model.PartnershipProject{
		NameIDN:        strings.TrimSpace(req.NameIDN),
		NameEN:         strings.TrimSpace(req.NameEN),
		DescriptionIDN: req.DescriptionIDN,
		DescriptionEN:  req.DescriptionEN,
		Image:          req.Image,
	}

	if err := h.partnershipService.Create(
		c.Request.Context(),
		project,
	); err != nil {
		if errors.Is(err, service.ErrInvalidPartnershipProject) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create partnership project",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "partnership project created successfully",
		"data":    project,
	})
}

func (h *PartnershipHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid partnership project id",
		})
		return
	}

	var req updatePartnershipRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	project := &model.PartnershipProject{
		ID:             id,
		NameIDN:        strings.TrimSpace(req.NameIDN),
		NameEN:         strings.TrimSpace(req.NameEN),
		DescriptionIDN: req.DescriptionIDN,
		DescriptionEN:  req.DescriptionEN,
		Image:          req.Image,
	}

	if err := h.partnershipService.Update(
		c.Request.Context(),
		project,
	); err != nil {
		switch {
		case errors.Is(
			err,
			service.ErrInvalidPartnershipProject,
		):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})

		case errors.Is(
			err,
			service.ErrPartnershipProjectNotFound,
		):
			c.JSON(http.StatusNotFound, gin.H{
				"message": "partnership project not found",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to update partnership project",
			})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "partnership project updated successfully",
		"data":    project,
	})
}

func (h *PartnershipHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid partnership project id",
		})
		return
	}

	if err := h.partnershipService.Delete(
		c.Request.Context(),
		id,
	); err != nil {
		if errors.Is(err, service.ErrPartnershipProjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "partnership project not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to delete partnership project",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "partnership project deleted successfully",
	})
}
