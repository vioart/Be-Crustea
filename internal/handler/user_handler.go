package handler

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"Be-Crustea/internal/middleware"
	"Be-Crustea/internal/model"
	"Be-Crustea/internal/service"
	"Be-Crustea/internal/storage"

	"github.com/gin-gonic/gin"
)

const (
	maxAvatarSize = 2 * 1024 * 1024 // 2 MB
)

var allowedAvatarMIME = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

type UserHandler struct {
	userService service.UserService
	storage     *storage.LocalStorage
}

func NewUserHandler(
	userService service.UserService,
	storage *storage.LocalStorage,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		storage:     storage,
	}
}

// ======================================================
// ADMIN: USERS
// ======================================================

// GET /api/admin/users
func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.userService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}

// GET /api/admin/users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user id",
		})
		return
	}

	user, err := h.userService.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		h.respondUserError(c, err, "failed to get user")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

// POST /api/admin/users
func (h *UserHandler) Create(c *gin.Context) {
	if !strings.HasPrefix(
		c.GetHeader("Content-Type"),
		"multipart/form-data",
	) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "content type must be multipart/form-data",
		})
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	email := strings.ToLower(
		strings.TrimSpace(c.PostForm("email")),
	)
	password := c.PostForm("password")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "name is required",
		})
		return
	}

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "email is required",
		})
		return
	}

	if strings.TrimSpace(password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "password is required",
		})
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "avatar is required",
		})
		return
	}

	if err := validateAvatar(file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	avatarPath, err := h.storage.Save(file, "users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to save avatar",
		})
		return
	}

	user := model.User{
		Name:     name,
		Email:    email,
		Password: password,
		Avatar:   &avatarPath,
	}

	if err := h.userService.Create(
		c.Request.Context(),
		&user,
	); err != nil {
		// Hapus file jika penyimpanan user gagal.
		_ = h.storage.Delete(avatarPath)

		h.respondUserError(c, err, "failed to create user")
		return
	}

	user.Password = ""

	c.JSON(http.StatusCreated, gin.H{
		"message": "user created successfully",
		"data":    user,
	})
}

// PUT /api/admin/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user id",
		})
		return
	}

	existingUser, err := h.userService.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		h.respondUserError(c, err, "failed to get user")
		return
	}

	if !strings.HasPrefix(
		c.GetHeader("Content-Type"),
		"multipart/form-data",
	) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "content type must be multipart/form-data",
		})
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	email := strings.ToLower(
		strings.TrimSpace(c.PostForm("email")),
	)
	password := c.PostForm("password")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "name is required",
		})
		return
	}

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "email is required",
		})
		return
	}

	user := model.User{
		ID:       id,
		Name:     name,
		Email:    email,
		Password: password,
		Avatar:   existingUser.Avatar,
	}

	oldAvatar := avatarPath(existingUser.Avatar)
	newAvatarPath := ""

	file, fileErr := c.FormFile("avatar")
	if fileErr == nil {
		if err := validateAvatar(file); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		newAvatarPath, err = h.storage.Save(file, "users")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to save avatar",
			})
			return
		}

		user.Avatar = &newAvatarPath
	} else if !errors.Is(fileErr, http.ErrMissingFile) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid avatar upload",
		})
		return
	}

	if err := h.userService.Update(
		c.Request.Context(),
		&user,
	); err != nil {
		if newAvatarPath != "" {
			_ = h.storage.Delete(newAvatarPath)
		}

		h.respondUserError(c, err, "failed to update user")
		return
	}

	// Hapus avatar lama hanya setelah database berhasil diperbarui.
	if newAvatarPath != "" && oldAvatar != "" {
		_ = h.storage.Delete(oldAvatar)
	}

	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "user updated successfully",
		"data":    user,
	})
}

// DELETE /api/admin/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user id",
		})
		return
	}

	user, err := h.userService.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		h.respondUserError(c, err, "failed to get user")
		return
	}

	if err := h.userService.Delete(
		c.Request.Context(),
		id,
	); err != nil {
		h.respondUserError(c, err, "failed to delete user")
		return
	}

	avatar := avatarPath(user.Avatar)
	if avatar != "" {
		_ = h.storage.Delete(avatar)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user deleted successfully",
	})
}

// ======================================================
// ACCOUNT SETTINGS
// ======================================================

// GET /api/account/me
func (h *UserHandler) GetMyAccount(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "invalid user id",
		})
		return
	}

	user, err := h.userService.GetByID(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		h.respondUserError(c, err, "failed to get account")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

// PUT /api/account/me
// Request: multipart/form-data
// Fields: name, email, avatar (optional)
func (h *UserHandler) UpdateMyAccount(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "invalid user id",
		})
		return
	}

	if !strings.HasPrefix(
		c.GetHeader("Content-Type"),
		"multipart/form-data",
	) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "content type must be multipart/form-data",
		})
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	email := strings.ToLower(
		strings.TrimSpace(c.PostForm("email")),
	)

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "name is required",
		})
		return
	}

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "email is required",
		})
		return
	}

	existingUser, err := h.userService.GetByID(
		c.Request.Context(),
		userID,
	)
	if err != nil {
		h.respondUserError(c, err, "failed to get account")
		return
	}

	user := model.User{
		ID:     userID,
		Name:   name,
		Email:  email,
		Avatar: existingUser.Avatar,
	}

	oldAvatar := avatarPath(existingUser.Avatar)
	newAvatarPath := ""

	file, fileErr := c.FormFile("avatar")
	if fileErr == nil {
		if err := validateAvatar(file); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		newAvatarPath, err = h.storage.Save(file, "users")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to save avatar",
			})
			return
		}

		user.Avatar = &newAvatarPath
	} else if !errors.Is(fileErr, http.ErrMissingFile) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid avatar upload",
		})
		return
	}

	// Password dikosongkan agar service mempertahankan hash lama.
	if err := h.userService.Update(
		c.Request.Context(),
		&user,
	); err != nil {
		if newAvatarPath != "" {
			_ = h.storage.Delete(newAvatarPath)
		}

		h.respondUserError(c, err, "failed to update account")
		return
	}

	if newAvatarPath != "" && oldAvatar != "" {
		_ = h.storage.Delete(oldAvatar)
	}

	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "account updated successfully",
		"data":    user,
	})
}

type changeMyPasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// PUT /api/account/me/password
func (h *UserHandler) ChangeMyPassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "invalid user id",
		})
		return
	}

	var req changeMyPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid password request",
			"error":   err.Error(),
		})
		return
	}

	err := h.userService.ChangePassword(
		c.Request.Context(),
		userID,
		req.CurrentPassword,
		req.NewPassword,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"message": "user not found",
			})

		case errors.Is(err, service.ErrInvalidCurrentPassword):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": "current password is incorrect",
			})

		case errors.Is(err, service.ErrPasswordUnchanged):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "new password must differ from current password",
			})

		case errors.Is(err, service.ErrInvalidUser):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to change password",
			})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "password updated successfully",
	})
}

// ======================================================
// HELPERS
// ======================================================

func (h *UserHandler) respondUserError(
	c *gin.Context,
	err error,
	fallbackMessage string,
) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"message": "user not found",
		})

	case errors.Is(err, service.ErrInvalidUser):
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})

	case errors.Is(err, service.ErrEmailExists):
		c.JSON(http.StatusConflict, gin.H{
			"message": "email already exists",
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fallbackMessage,
		})
	}
}

func avatarPath(avatar *string) string {
	if avatar == nil {
		return ""
	}

	return strings.TrimSpace(*avatar)
}

// ======================================================
// AVATAR VALIDATION
// ======================================================

func validateAvatar(file *multipart.FileHeader) error {
	if file == nil {
		return fmt.Errorf("avatar is required")
	}

	if file.Size <= 0 {
		return fmt.Errorf("avatar file is empty")
	}

	if file.Size > maxAvatarSize {
		return fmt.Errorf("avatar size must not exceed 2 MB")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))

	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		// Extension diperbolehkan.
	default:
		return fmt.Errorf("avatar must be JPG, JPEG, PNG, or WEBP")
	}

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open avatar")
	}
	defer src.Close()

	buffer := make([]byte, 512)
	n, err := src.Read(buffer)

	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to read avatar")
	}

	contentType := http.DetectContentType(buffer[:n])
	if !allowedAvatarMIME[contentType] {
		return fmt.Errorf("invalid avatar file type")
	}

	return nil
}
