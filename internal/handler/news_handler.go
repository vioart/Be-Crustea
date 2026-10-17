package handler

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"Be-Crustea/internal/middleware"
	"Be-Crustea/internal/model"
	"Be-Crustea/internal/service"
	"Be-Crustea/internal/storage"

	"github.com/gin-gonic/gin"
)

type NewsHandler struct {
	newsService service.NewsService
	storage     *storage.LocalStorage
}

func NewNewsHandler(
	newsService service.NewsService,
	storage *storage.LocalStorage,
) *NewsHandler {
	return &NewsHandler{
		newsService: newsService,
		storage:     storage,
	}
}

type createNewsRequest struct {
	CategoryID int64
	Slug       string
	TitleIDN   string
	TitleEN    string
	ContentIDN string
	ContentEN  string
	Status     model.NewsStatus
	Thumbnail  *string
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

type importContentImageRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type movedStorageFile struct {
	OldPath string
	NewPath string
}

func (h *NewsHandler) moveTempFilesFromContent(
	content string,
) (string, []movedStorageFile, error) {

	if h.storage == nil {
		return content, nil, fmt.Errorf("storage is not configured")
	}

	// Cari URL/path yang mengarah ke /uploads/temp/...
	re := regexp.MustCompile(
		`(?:https?://[^"'<>[:space:]]+)?/uploads/temp/[^"'<>[:space:]]+`,
	)

	matches := re.FindAllString(content, -1)

	if len(matches) == 0 {
		return content, nil, nil
	}

	// Hindari memproses URL yang sama lebih dari sekali
	uniquePaths := make(map[string]struct{})

	var movedFiles []movedStorageFile
	updatedContent := content

	for _, originalURL := range matches {

		if _, exists := uniquePaths[originalURL]; exists {
			continue
		}

		uniquePaths[originalURL] = struct{}{}

		// LocalStorage bekerja menggunakan path storage,
		sourcePath := originalURL

		if parsedURL, err := url.Parse(originalURL); err == nil {
			if parsedURL.IsAbs() {
				sourcePath = parsedURL.Path
			}
		}

		if !strings.HasPrefix(sourcePath, "/uploads/temp/") {
			continue
		}

		// Ambil nama file dari URL temp.
		cleanSourcePath := sourcePath

		if index := strings.Index(cleanSourcePath, "?"); index >= 0 {
			cleanSourcePath = cleanSourcePath[:index]
		}

		filename := strings.TrimPrefix(
			cleanSourcePath,
			"/uploads/temp/",
		)

		if strings.TrimSpace(filename) == "" {
			continue
		}

		// ----------------------------------------------------
		// CEK APAKAH FILE SUDAH DIPINDAHKAN KE PERMANENT
		// ----------------------------------------------------
		permanentPath := "/uploads/news/content/" + filename

		permanentURL := permanentPath

		permanentFilePath := filepath.Join(
			h.storage.BasePath,
			"news",
			"content",
			filename,
		)

		if _, err := os.Stat(permanentFilePath); err == nil {
			updatedContent = strings.ReplaceAll(
				updatedContent,
				originalURL,
				permanentURL,
			)

			continue
		}

		// ----------------------------------------------------
		// FILE BELUM PERMANENT → PINDAHKAN DARI TEMP
		// ----------------------------------------------------

		newPath, err := h.storage.Move(
			sourcePath,
			"news/content",
		)

		if err != nil {
			for i := len(movedFiles) - 1; i >= 0; i-- {
				_ = h.storage.Delete(movedFiles[i].NewPath)
			}

			return content, nil, fmt.Errorf(
				"failed to move content image: %w",
				err,
			)
		}

		movedFiles = append(
			movedFiles,
			movedStorageFile{
				OldPath: sourcePath,
				NewPath: newPath,
			},
		)

		// Ganti URL lama dengan URL permanent.
		updatedContent = strings.ReplaceAll(
			updatedContent,
			originalURL,
			newPath,
		)
	}

	return updatedContent, movedFiles, nil
}

// moveTempThumbnail memindahkan thumbnail dari temp
// ke news/thumbnails.
func (h *NewsHandler) moveTempThumbnail(
	thumbnail *string,
) (*string, *movedStorageFile, error) {
	if thumbnail == nil || strings.TrimSpace(*thumbnail) == "" {
		return thumbnail, nil, nil
	}

	if h.storage == nil {
		return nil, nil, fmt.Errorf("storage is not configured")
	}

	originalURL := strings.TrimSpace(*thumbnail)

	sourcePath := originalURL

	if parsedURL, err := url.Parse(originalURL); err == nil {
		if parsedURL.IsAbs() {
			sourcePath = parsedURL.Path
		}
	}

	// Thumbnail yang bukan berasal dari temp
	// tidak perlu dipindahkan.
	if !strings.HasPrefix(sourcePath, "/uploads/temp/") {
		return thumbnail, nil, nil
	}

	newPath, err := h.storage.Move(
		sourcePath,
		"news/thumbnails",
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to move thumbnail: %w",
			err,
		)
	}

	return &newPath, &movedStorageFile{
		OldPath: sourcePath,
		NewPath: newPath,
	}, nil
}

func (h *NewsHandler) GetAll(c *gin.Context) {
	news, err := h.newsService.GetAll(c.Request.Context())
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
	// Ambil user ID dari JWT middleware.
	authorID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "user authentication is required",
		})
		return
	}

	// Ambil data text dari multipart/form-data.
	categoryIDValue := strings.TrimSpace(c.PostForm("category_id"))
	slug := strings.TrimSpace(c.PostForm("slug"))
	titleIDN := strings.TrimSpace(c.PostForm("title_idn"))
	titleEN := strings.TrimSpace(c.PostForm("title_en"))
	contentIDN := c.PostForm("content_idn")
	contentEN := c.PostForm("content_en")
	statusValue := strings.TrimSpace(c.PostForm("status"))

	if categoryIDValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "category_id is required",
		})
		return
	}

	categoryID, err := strconv.ParseInt(categoryIDValue, 10, 64)
	if err != nil || categoryID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid category_id",
		})
		return
	}

	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "slug is required",
		})
		return
	}

	if titleIDN == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "title_idn is required",
		})
		return
	}

	if titleEN == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "title_en is required",
		})
		return
	}

	if contentIDN == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "content_idn is required",
		})
		return
	}

	if contentEN == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "content_en is required",
		})
		return
	}

	status := model.NewsStatus(statusValue)

	if status == "" {
		status = model.NewsStatusDraft
	}

	if status != model.NewsStatusDraft &&
		status != model.NewsStatusPublished {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid status",
		})
		return
	}

	if h.storage == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "storage is not configured",
		})
		return
	}

	// ============================================================
	// THUMBNAIL
	// ============================================================

	var thumbnail *string

	file, err := c.FormFile("thumbnail")

	if err == nil && file != nil {
		const maxThumbnailSize = 10 * 1024 * 1024

		if file.Size > maxThumbnailSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "thumbnail file size must not exceed 10 MB",
			})
			return
		}

		// Thumbnail sekarang disimpan ke temp terlebih dahulu.
		savedPath, err := h.storage.Save(
			file,
			"temp",
		)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to save thumbnail",
				"error":   err.Error(),
			})
			return
		}

		thumbnail = &savedPath
	} else if err != nil && !errors.Is(err, http.ErrMissingFile) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid thumbnail",
			"error":   err.Error(),
		})
		return
	}

	// ============================================================
	// MOVE CONTENT IMAGE DARI TEMP → NEWS/CONTENT
	// ============================================================

	updatedContentIDN, movedContentIDN, err :=
		h.moveTempFilesFromContent(contentIDN)

	if err != nil {
		// Thumbnail masih berada di temp.
		if thumbnail != nil {
			_ = h.storage.Delete(*thumbnail)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to process content images",
			"error":   err.Error(),
		})
		return
	}

	updatedContentEN, movedContentEN, err :=
		h.moveTempFilesFromContent(contentEN)

	if err != nil {
		// Hapus thumbnail temp.
		if thumbnail != nil {
			_ = h.storage.Delete(*thumbnail)
		}

		// Rollback content IDN yang sudah dipindahkan.
		for i := len(movedContentIDN) - 1; i >= 0; i-- {
			_ = h.storage.Delete(movedContentIDN[i].NewPath)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to process content images",
			"error":   err.Error(),
		})
		return
	}

	// ============================================================
	// MOVE THUMBNAIL DARI TEMP → NEWS/THUMBNAILS
	// ============================================================

	movedThumbnail, movedThumbnailFile, err :=
		h.moveTempThumbnail(thumbnail)

	if err != nil {
		// Rollback content IDN.
		for i := len(movedContentIDN) - 1; i >= 0; i-- {
			_ = h.storage.Delete(movedContentIDN[i].NewPath)
		}

		// Rollback content EN.
		for i := len(movedContentEN) - 1; i >= 0; i-- {
			_ = h.storage.Delete(movedContentEN[i].NewPath)
		}

		// Thumbnail masih mungkin berada di temp.
		if thumbnail != nil {
			_ = h.storage.Delete(*thumbnail)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to process thumbnail",
			"error":   err.Error(),
		})
		return
	}

	// ============================================================
	// BUAT MODEL NEWS
	// ============================================================

	news := &model.News{
		CategoryID: categoryID,
		AuthorID:   authorID,
		Slug:       slug,
		Thumbnail:  movedThumbnail,
		TitleIDN:   titleIDN,
		TitleEN:    titleEN,
		ContentIDN: updatedContentIDN,
		ContentEN:  updatedContentEN,
		Status:     status,
	}

	// ============================================================
	// SIMPAN KE DATABASE
	// ============================================================

	if err := h.newsService.Create(
		c.Request.Context(),
		news,
	); err != nil {

		// Database gagal.
		// Hapus semua file permanent yang baru saja dipindahkan.

		for i := len(movedContentIDN) - 1; i >= 0; i-- {
			_ = h.storage.Delete(movedContentIDN[i].NewPath)
		}

		for i := len(movedContentEN) - 1; i >= 0; i-- {
			_ = h.storage.Delete(movedContentEN[i].NewPath)
		}

		if movedThumbnailFile != nil {
			_ = h.storage.Delete(movedThumbnailFile.NewPath)
		}

		switch {
		case errors.Is(err, service.ErrInvalidNews):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})

		case errors.Is(err, service.ErrNewsSlugExists):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "slug already exists",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to create news",
			})
		}

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "news created successfully",
		"data":    news,
	})
}

// UploadContentImage mengupload gambar yang digunakan di dalam
// Tiptap content editor ke temporary storage.
func (h *NewsHandler) UploadContentImage(c *gin.Context) {
	if h.storage == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "storage is not configured",
		})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "image is required",
		})
		return
	}

	if file == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "image is required",
		})
		return
	}

	// Batas ukuran gambar content.
	const maxContentImageSize = 5 * 1024 * 1024

	if file.Size > maxContentImageSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "content image file size must not exceed 5 MB",
		})
		return
	}

	// Simpan ke temporary storage terlebih dahulu.
	url, err := h.storage.Save(
		file,
		"temp",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to save content image",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}

func (h *NewsHandler) ImportContentImage(c *gin.Context) {
	if h.storage == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "storage is not configured",
		})
		return
	}

	var req importContentImageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	imageURL := strings.TrimSpace(req.URL)

	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid image URL",
		})
		return
	}

	// Hanya izinkan HTTP dan HTTPS.
	if parsedURL.Scheme != "http" &&
		parsedURL.Scheme != "https" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "only http and https URLs are allowed",
		})
		return
	}

	if parsedURL.Hostname() == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid image URL",
		})
		return
	}

	// Cegah SSRF ke IP private/internal.
	ips, err := net.LookupIP(parsedURL.Hostname())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to resolve image host",
		})
		return
	}

	for _, ip := range ips {
		if ip.IsLoopback() ||
			ip.IsPrivate() ||
			ip.IsLinkLocalUnicast() ||
			ip.IsUnspecified() {

			c.JSON(http.StatusBadRequest, gin.H{
				"message": "image host is not allowed",
			})
			return
		}
	}

	const maxContentImageSize = 5 * 1024 * 1024

	client := &http.Client{
		CheckRedirect: func(
			req *http.Request,
			via []*http.Request,
		) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}

			if req.URL.Scheme != "http" &&
				req.URL.Scheme != "https" {
				return fmt.Errorf(
					"redirect scheme is not allowed",
				)
			}

			redirectIPs, err := net.LookupIP(
				req.URL.Hostname(),
			)
			if err != nil {
				return err
			}

			for _, ip := range redirectIPs {
				if ip.IsLoopback() ||
					ip.IsPrivate() ||
					ip.IsLinkLocalUnicast() ||
					ip.IsUnspecified() {

					return fmt.Errorf(
						"redirect host is not allowed",
					)
				}
			}

			return nil
		},
	}

	request, err := http.NewRequestWithContext(
		c.Request.Context(),
		http.MethodGet,
		imageURL,
		nil,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to create image request",
		})
		return
	}

	request.Header.Set(
		"User-Agent",
		"Be-Crustea/1.0",
	)

	request.Header.Set(
		"Accept",
		"image/*",
	)

	response, err := client.Do(request)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"message": "failed to download image",
			"error":   err.Error(),
		})
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 ||
		response.StatusCode >= 300 {

		c.JSON(http.StatusBadGateway, gin.H{
			"message": "external image returned an invalid status",
		})
		return
	}

	contentType := strings.ToLower(
		strings.TrimSpace(
			strings.Split(
				response.Header.Get("Content-Type"),
				";",
			)[0],
		),
	)

	var extension string

	switch contentType {
	case "image/jpeg":
		extension = ".jpg"

	case "image/png":
		extension = ".png"

	case "image/webp":
		extension = ".webp"

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "external file is not a supported image",
		})
		return
	}

	if response.ContentLength > maxContentImageSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "content image file size must not exceed 5 MB",
		})
		return
	}

	body, err := io.ReadAll(
		io.LimitReader(
			response.Body,
			maxContentImageSize+1,
		),
	)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"message": "failed to read image",
		})
		return
	}

	if len(body) > maxContentImageSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "content image file size must not exceed 5 MB",
		})
		return
	}

	filename := "content-image" + extension

	savedURL, err := h.storage.SaveBytes(
		body,
		filename,
		"temp",
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to save content image",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": savedURL,
	})
}

func (h *NewsHandler) UpdateBySlug(c *gin.Context) {
	// ============================================================
	// AMBIL SLUG LAMA DARI URL
	// ============================================================

	oldSlug := strings.TrimSpace(c.Param("slug"))

	if oldSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid news slug",
		})
		return
	}

	// ============================================================
	// CARI NEWS BERDASARKAN SLUG LAMA
	// ============================================================

	existingNews, err := h.newsService.GetBySlug(
		c.Request.Context(),
		oldSlug,
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

	// ============================================================
	// AMBIL DATA FORM
	// ============================================================

	categoryIDValue := strings.TrimSpace(
		c.PostForm("category_id"),
	)

	slug := strings.TrimSpace(
		c.PostForm("slug"),
	)

	titleIDN := strings.TrimSpace(
		c.PostForm("title_idn"),
	)

	titleEN := strings.TrimSpace(
		c.PostForm("title_en"),
	)

	contentIDN := c.PostForm("content_idn")
	contentEN := c.PostForm("content_en")

	statusValue := strings.TrimSpace(
		c.PostForm("status"),
	)

	// ============================================================
	// VALIDASI CATEGORY
	// ============================================================

	if categoryIDValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "category_id is required",
		})
		return
	}

	categoryID, err := strconv.ParseInt(
		categoryIDValue,
		10,
		64,
	)

	if err != nil || categoryID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid category_id",
		})
		return
	}

	// ============================================================
	// VALIDASI FIELD
	// ============================================================

	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "slug is required",
		})
		return
	}

	if titleIDN == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "title_idn is required",
		})
		return
	}

	if titleEN == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "title_en is required",
		})
		return
	}

	if contentIDN == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "content_idn is required",
		})
		return
	}

	if contentEN == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "content_en is required",
		})
		return
	}

	// ============================================================
	// STATUS
	// ============================================================

	status := model.NewsStatus(statusValue)

	if status == "" {
		status = model.NewsStatusDraft
	}

	if status != model.NewsStatusDraft &&
		status != model.NewsStatusPublished {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid status",
		})
		return
	}

	// ============================================================
	// STORAGE CHECK
	// ============================================================

	if h.storage == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "storage is not configured",
		})
		return
	}

	// ============================================================
	// THUMBNAIL
	// ============================================================

	// Default: gunakan thumbnail lama.
	thumbnail := existingNews.Thumbnail

	var newThumbnailFile *movedStorageFile

	file, err := c.FormFile("thumbnail")

	if err == nil && file != nil {
		const maxThumbnailSize = 10 * 1024 * 1024

		if file.Size > maxThumbnailSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "thumbnail file size must not exceed 10 MB",
			})
			return
		}

		// Simpan thumbnail baru ke temp.
		tempThumbnail, err := h.storage.Save(
			file,
			"temp",
		)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to save thumbnail",
				"error":   err.Error(),
			})
			return
		}

		// Pindahkan dari temp → news/thumbnails.
		movedThumbnail, movedFile, err :=
			h.moveTempThumbnail(&tempThumbnail)

		if err != nil {
			_ = h.storage.Delete(tempThumbnail)

			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to process thumbnail",
				"error":   err.Error(),
			})
			return
		}

		thumbnail = movedThumbnail
		newThumbnailFile = movedFile

	} else if err != nil &&
		!errors.Is(err, http.ErrMissingFile) {

		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid thumbnail",
			"error":   err.Error(),
		})
		return
	}

	// ============================================================
	// MOVE CONTENT IMAGE
	// TEMP → NEWS/CONTENT
	// ============================================================

	updatedContentIDN, movedContentIDN, err :=
		h.moveTempFilesFromContent(contentIDN)

	if err != nil {
		if newThumbnailFile != nil {
			_ = h.storage.Delete(
				newThumbnailFile.NewPath,
			)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to process content images",
			"error":   err.Error(),
		})
		return
	}

	updatedContentEN, movedContentEN, err :=
		h.moveTempFilesFromContent(contentEN)

	if err != nil {
		// Rollback IDN.
		for i := len(movedContentIDN) - 1; i >= 0; i-- {
			_ = h.storage.Delete(
				movedContentIDN[i].NewPath,
			)
		}

		// Rollback thumbnail.
		if newThumbnailFile != nil {
			_ = h.storage.Delete(
				newThumbnailFile.NewPath,
			)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to process content images",
			"error":   err.Error(),
		})
		return
	}

	// ============================================================
	// BUAT MODEL NEWS
	// ============================================================

	news := &model.News{
		ID:         existingNews.ID,
		CategoryID: categoryID,

		// Author tetap menggunakan author lama.
		AuthorID: existingNews.AuthorID,

		Slug:      slug,
		Thumbnail: thumbnail,

		TitleIDN: titleIDN,
		TitleEN:  titleEN,

		ContentIDN: updatedContentIDN,
		ContentEN:  updatedContentEN,

		Status: status,
	}

	// ============================================================
	// UPDATE DATABASE
	// ============================================================

	if err := h.newsService.Update(
		c.Request.Context(),
		news,
	); err != nil {

		// Rollback content IDN.
		for i := len(movedContentIDN) - 1; i >= 0; i-- {
			_ = h.storage.Delete(
				movedContentIDN[i].NewPath,
			)
		}

		// Rollback content EN.
		for i := len(movedContentEN) - 1; i >= 0; i-- {
			_ = h.storage.Delete(
				movedContentEN[i].NewPath,
			)
		}

		// Rollback thumbnail baru.
		if newThumbnailFile != nil {
			_ = h.storage.Delete(
				newThumbnailFile.NewPath,
			)
		}

		switch {
		case errors.Is(err, service.ErrInvalidNews):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})

		case errors.Is(err, service.ErrNewsNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"message": "news not found",
			})

		case errors.Is(err, service.ErrNewsSlugExists):
			c.JSON(http.StatusConflict, gin.H{
				"message": "slug already exists",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to update news",
			})
		}

		return
	}

	// ============================================================
	// HAPUS THUMBNAIL LAMA
	// ============================================================

	if newThumbnailFile != nil &&
		existingNews.Thumbnail != nil &&
		strings.TrimSpace(*existingNews.Thumbnail) != "" {

		oldThumbnail := *existingNews.Thumbnail

		if parsedURL, err := url.Parse(oldThumbnail); err == nil {
			if parsedURL.IsAbs() {
				oldThumbnail = parsedURL.Path
			}
		}

		// Hanya hapus file lokal /uploads/...
		if strings.HasPrefix(oldThumbnail, "/uploads/") {
			_ = h.storage.Delete(oldThumbnail)
		}
	}

	// ============================================================
	// RESPONSE
	// ============================================================

	c.JSON(http.StatusOK, gin.H{
		"message": "news updated successfully",
		"data":    news,
	})
}

func extractContentImagePaths(content string) []string {
	re := regexp.MustCompile(
		`(?:https?://[^"'<>[:space:]]+)?/uploads/news/content/[^"'<>[:space:]]+`,
	)

	matches := re.FindAllString(content, -1)

	if len(matches) == 0 {
		return nil
	}

	unique := make(map[string]struct{})
	paths := make([]string, 0, len(matches))

	for _, imageURL := range matches {
		path := imageURL

		if parsedURL, err := url.Parse(imageURL); err == nil {
			if parsedURL.IsAbs() {
				path = parsedURL.Path
			}
		}

		if !strings.HasPrefix(path, "/uploads/news/content/") {
			continue
		}

		if _, exists := unique[path]; exists {
			continue
		}

		unique[path] = struct{}{}
		paths = append(paths, path)
	}

	return paths
}

func (h *NewsHandler) DeleteBySlug(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))

	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid news slug",
		})
		return
	}

	// ============================================================
	// CARI NEWS
	// ============================================================

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

	// ============================================================
	// VALIDASI STORAGE
	// ============================================================

	if h.storage == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "storage is not configured",
		})
		return
	}

	// ============================================================
	// KUMPULKAN FILE YANG AKAN DIHAPUS
	// ============================================================

	filesToDelete := make([]string, 0)

	// ------------------------------------------------------------
	// THUMBNAIL
	// ------------------------------------------------------------

	if news.Thumbnail != nil &&
		strings.TrimSpace(*news.Thumbnail) != "" {

		thumbnailPath := strings.TrimSpace(*news.Thumbnail)

		if parsedURL, err := url.Parse(thumbnailPath); err == nil {
			if parsedURL.IsAbs() {
				thumbnailPath = parsedURL.Path
			}
		}

		if strings.HasPrefix(thumbnailPath, "/uploads/") {
			filesToDelete = append(filesToDelete, thumbnailPath)
		}
	}

	// ------------------------------------------------------------
	// CONTENT IDN
	// ------------------------------------------------------------

	filesToDelete = append(
		filesToDelete,
		extractContentImagePaths(news.ContentIDN)...,
	)

	// ------------------------------------------------------------
	// CONTENT EN
	// ------------------------------------------------------------

	filesToDelete = append(
		filesToDelete,
		extractContentImagePaths(news.ContentEN)...,
	)

	// ============================================================
	// HAPUS DATABASE
	// ============================================================

	if err := h.newsService.Delete(
		c.Request.Context(),
		news.ID,
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

	// ============================================================
	// HAPUS FILE STORAGE
	// ============================================================

	uniqueFiles := make(map[string]struct{})

	for _, filePath := range filesToDelete {
		if _, exists := uniqueFiles[filePath]; exists {
			continue
		}

		uniqueFiles[filePath] = struct{}{}

		if err := h.storage.Delete(filePath); err != nil {
			// Database sudah berhasil dihapus.
			// Jadi jangan menggagalkan response utama.
			fmt.Printf(
				"warning: failed to delete storage file %s: %v\n",
				filePath,
				err,
			)
		}
	}

	// ============================================================
	// RESPONSE
	// ============================================================

	c.JSON(http.StatusOK, gin.H{
		"message": "news deleted successfully",
	})
}
