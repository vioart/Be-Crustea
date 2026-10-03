package handler

import (
	"errors"
	"net/http"
	"strings"

	"Be-Crustea/internal/middleware"
	"Be-Crustea/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type registerRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	user, err := h.authService.Register(
		c.Request.Context(),
		req.Name,
		req.Email,
		req.Password,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"message": "email already registered",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to register user",
			})
		}

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"data":    user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	token, user, err := h.authService.Login(
		c.Request.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid email or password",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to login",
			})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"data": gin.H{
			"token": token,
			"user":  user,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "invalid user id",
		})
		return
	}

	user, err := h.authService.GetCurrentUser(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "user not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}
