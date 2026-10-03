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

type FAQHandler struct {
	faqService service.FAQService
}

func NewFAQHandler(faqService service.FAQService) *FAQHandler {
	return &FAQHandler{
		faqService: faqService,
	}
}

type createFAQRequest struct {
	Category    model.FAQCategory `json:"category" binding:"required"`
	QuestionIDN string            `json:"question_idn" binding:"required"`
	QuestionEN  string            `json:"question_en" binding:"required"`
	AnswerIDN   string            `json:"answer_idn" binding:"required"`
	AnswerEN    string            `json:"answer_en" binding:"required"`
}

type updateFAQRequest struct {
	Category    model.FAQCategory `json:"category" binding:"required"`
	QuestionIDN string            `json:"question_idn" binding:"required"`
	QuestionEN  string            `json:"question_en" binding:"required"`
	AnswerIDN   string            `json:"answer_idn" binding:"required"`
	AnswerEN    string            `json:"answer_en" binding:"required"`
}

func (h *FAQHandler) GetAll(c *gin.Context) {
	faqs, err := h.faqService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get faqs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": faqs,
	})
}

func (h *FAQHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid faq id",
		})
		return
	}

	faq, err := h.faqService.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		if errors.Is(err, service.ErrFAQNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "faq not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get faq",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": faq,
	})
}

func (h *FAQHandler) Create(c *gin.Context) {
	var req createFAQRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	faq := &model.FAQ{
		Category:    req.Category,
		QuestionIDN: strings.TrimSpace(req.QuestionIDN),
		QuestionEN:  strings.TrimSpace(req.QuestionEN),
		AnswerIDN:   req.AnswerIDN,
		AnswerEN:    req.AnswerEN,
	}

	if err := h.faqService.Create(
		c.Request.Context(),
		faq,
	); err != nil {
		if errors.Is(err, service.ErrInvalidFAQ) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create faq",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "faq created successfully",
		"data":    faq,
	})
}

func (h *FAQHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid faq id",
		})
		return
	}

	var req updateFAQRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	faq := &model.FAQ{
		ID:          id,
		Category:    req.Category,
		QuestionIDN: strings.TrimSpace(req.QuestionIDN),
		QuestionEN:  strings.TrimSpace(req.QuestionEN),
		AnswerIDN:   req.AnswerIDN,
		AnswerEN:    req.AnswerEN,
	}

	if err := h.faqService.Update(
		c.Request.Context(),
		faq,
	); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidFAQ):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
		case errors.Is(err, service.ErrFAQNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"message": "faq not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to update faq",
			})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "faq updated successfully",
		"data":    faq,
	})
}

func (h *FAQHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid faq id",
		})
		return
	}

	if err := h.faqService.Delete(
		c.Request.Context(),
		id,
	); err != nil {
		if errors.Is(err, service.ErrFAQNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "faq not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to delete faq",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "faq deleted successfully",
	})
}
