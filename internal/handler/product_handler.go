package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"Be-Crustea/internal/model"
	"Be-Crustea/internal/service"
	"Be-Crustea/internal/storage"

	"github.com/gin-gonic/gin"
)

const (
	maxProductRequestSize = 25 << 20 // 25 MiB total request
	maxProductImageSize   = 2 << 20  // 2 MiB per image
	maxProductImageCount  = 10
)

type ProductHandler struct {
	productService service.ProductService
	productStorage *storage.LocalStorage
}

func NewProductHandler(
	productService service.ProductService,
	productStorage *storage.LocalStorage,
) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		productStorage: productStorage,
	}
}

// Request untuk field teks. images dikirim sebagai file multipart dengan field "images".
type createProductRequest struct {
	Category       string `form:"category" binding:"required"`
	Slug           string `form:"slug" binding:"required"`
	NameIDN        string `form:"name_idn" binding:"required"`
	NameEN         string `form:"name_en" binding:"required"`
	DescriptionIDN string `form:"description_idn" binding:"required"`
	DescriptionEN  string `form:"description_en" binding:"required"`
	Excellences    string `form:"excellences"`
	Specifications string `form:"specifications"`
}

type updateProductRequest struct {
	Category       string `form:"category" binding:"required"`
	Slug           string `form:"slug" binding:"required"`
	NameIDN        string `form:"name_idn" binding:"required"`
	NameEN         string `form:"name_en" binding:"required"`
	DescriptionIDN string `form:"description_idn" binding:"required"`
	DescriptionEN  string `form:"description_en" binding:"required"`
	Excellences    string `form:"excellences"`
	Specifications string `form:"specifications"`
}

type productExcellenceRequest struct {
	TitleIDN string `json:"title_idn" binding:"required"`
	TitleEN  string `json:"title_en" binding:"required"`
}

type productSpecificationRequest struct {
	NameIDN  string `json:"name_idn" binding:"required"`
	NameEN   string `json:"name_en" binding:"required"`
	ValueIDN string `json:"value_idn" binding:"required"`
	ValueEN  string `json:"value_en" binding:"required"`
}

// getPagination mengambil parameter limit dan offset.
func getPagination(c *gin.Context) (int, int, bool) {
	limit := 10
	offset := 0

	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid limit"})
			return 0, 0, false
		}
		limit = parsed
	}

	if value := c.Query("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid offset"})
			return 0, 0, false
		}
		offset = parsed
	}

	return limit, offset, true
}

func (h *ProductHandler) GetAll(c *gin.Context) {
	limit, offset, ok := getPagination(c)
	if !ok {
		return
	}

	products, err := h.productService.GetAll(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": products})
}

func (h *ProductHandler) getProductsByCategory(c *gin.Context, category string) {
	limit, offset, ok := getPagination(c)
	if !ok {
		return
	}

	products, err := h.productService.GetByCategory(
		c.Request.Context(), category, limit, offset,
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidProduct) {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get products",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": products})
}

func (h *ProductHandler) GetTechnology(c *gin.Context) {
	h.getProductsByCategory(c, service.ProductCategoryTechnology)
}

func (h *ProductHandler) GetKrasty(c *gin.Context) {
	h.getProductsByCategory(c, service.ProductCategoryKrasty)
}

func (h *ProductHandler) GetTrading(c *gin.Context) {
	h.getProductsByCategory(c, service.ProductCategoryTrading)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid product id"})
		return
	}

	product, err := h.productService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": product})
}

func (h *ProductHandler) GetBySlug(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid product slug"})
		return
	}

	product, err := h.productService.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": product})
}

// Create menerima teks, file gambar, keunggulan, dan spesifikasi dalam satu multipart request.
func (h *ProductHandler) Create(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxProductRequestSize)

	if err := c.Request.ParseMultipartForm(maxProductRequestSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid multipart request",
			"error":   err.Error(),
		})
		return
	}

	var req createProductRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	excellences, specifications, err := parseProductDetails(req.Excellences, req.Specifications)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	files := c.Request.MultipartForm.File["images"]
	if len(files) > maxProductImageCount {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("maximum %d images are allowed", maxProductImageCount),
		})
		return
	}

	product := &model.Product{
		Category:       strings.TrimSpace(req.Category),
		Slug:           strings.TrimSpace(req.Slug),
		NameIDN:        strings.TrimSpace(req.NameIDN),
		NameEN:         strings.TrimSpace(req.NameEN),
		DescriptionIDN: strings.TrimSpace(req.DescriptionIDN),
		DescriptionEN:  strings.TrimSpace(req.DescriptionEN),
	}

	savedFiles := make([]string, 0, len(files))
	rollbackFiles := func() {
		for _, path := range savedFiles {
			_ = h.productStorage.Delete(path)
		}
	}

	images := make([]model.ProductImage, 0, len(files))
	for _, file := range files {
		if file.Size <= 0 || file.Size > maxProductImageSize {
			rollbackFiles()
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("each image must be between 1 byte and %d MiB", maxProductImageSize/(1<<20)),
			})
			return
		}

		if err := validateProductImage(file); err != nil {
			rollbackFiles()
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		if h.productStorage == nil {
			rollbackFiles()
			c.JSON(http.StatusInternalServerError, gin.H{"message": "product storage is not configured"})
			return
		}

		fileURL, err := h.productStorage.Save(file, "products")
		if err != nil {
			rollbackFiles()
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to save product image",
				"error":   err.Error(),
			})
			return
		}

		savedFiles = append(savedFiles, fileURL)
		images = append(images, model.ProductImage{FileURL: fileURL})
	}

	if err := h.productService.CreateWithDetails(
		c.Request.Context(),
		product,
		images,
		excellences,
		specifications,
	); err != nil {
		rollbackFiles()
		if errors.Is(err, service.ErrInvalidProduct) {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create product",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "product created successfully",
		"data": gin.H{
			"product":        product,
			"images":         images,
			"excellences":    excellences,
			"specifications": specifications,
		},
	})
}

// UpdateBySlug memperbarui produk berdasarkan slug dari URL.
func (h *ProductHandler) UpdateBySlug(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid product slug",
		})
		return
	}

	// Request update menggunakan multipart/form-data.
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		maxProductRequestSize,
	)

	if err := c.Request.ParseMultipartForm(maxProductRequestSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid multipart request",
			"error":   err.Error(),
		})
		return
	}

	var req updateProductRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
			"error":   err.Error(),
		})
		return
	}

	excellences, specifications, err := parseProductDetails(
		req.Excellences,
		req.Specifications,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// URL gambar lama yang dipilih untuk dihapus.
	deleteImages, err := parseDeleteImages(c.PostForm("delete_images"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// Gambar baru yang diunggah.
	files := c.Request.MultipartForm.File["images"]
	if len(files) > maxProductImageCount {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf(
				"maximum %d new images are allowed",
				maxProductImageCount,
			),
		})
		return
	}

	// Pastikan produk lama ditemukan sebelum menyimpan gambar baru.
	existing, err := h.productService.GetBySlug(
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
			"error":   err.Error(),
		})
		return
	}

	// Pastikan URL yang akan dihapus memang milik produk ini.
	existingImageURLs := make(map[string]bool, len(existing.Images))
	for _, image := range existing.Images {
		existingImageURLs[image.FileURL] = true
	}

	for _, imageURL := range deleteImages {
		if !existingImageURLs[imageURL] {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "delete_images contains an image that does not belong to this product",
			})
			return
		}
	}

	if h.productStorage == nil && len(files) > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "product storage is not configured",
		})
		return
	}

	// Simpan gambar baru terlebih dahulu.
	savedFiles := make([]string, 0, len(files))
	rollbackFiles := func() {
		for _, path := range savedFiles {
			_ = h.productStorage.Delete(path)
		}
	}

	newImages := make([]model.ProductImage, 0, len(files))
	for _, file := range files {
		if file.Size <= 0 || file.Size > maxProductImageSize {
			rollbackFiles()
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf(
					"each image must be between 1 byte and %d MiB",
					maxProductImageSize/(1<<20),
				),
			})
			return
		}

		if err := validateProductImage(file); err != nil {
			rollbackFiles()
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		fileURL, err := h.productStorage.Save(file, "products")
		if err != nil {
			rollbackFiles()
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to save product image",
				"error":   err.Error(),
			})
			return
		}

		savedFiles = append(savedFiles, fileURL)
		newImages = append(newImages, model.ProductImage{
			FileURL: fileURL,
		})
	}

	product := &model.Product{
		Category:       strings.TrimSpace(req.Category),
		Slug:           strings.TrimSpace(req.Slug),
		NameIDN:        strings.TrimSpace(req.NameIDN),
		NameEN:         strings.TrimSpace(req.NameEN),
		DescriptionIDN: strings.TrimSpace(req.DescriptionIDN),
		DescriptionEN:  strings.TrimSpace(req.DescriptionEN),
	}

	// Method ini harus memperbarui data utama dan relasinya
	// (gambar, keunggulan, serta spesifikasi) di database.
	err = h.productService.UpdateWithDetails(
		c.Request.Context(),
		slug,
		product,
		newImages,
		deleteImages,
		excellences,
		specifications,
	)
	if err != nil {
		rollbackFiles()

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
				"error":   err.Error(),
			})
		}
		return
	}

	// Hapus file lama dari storage hanya setelah perubahan database berhasil.
	for _, imageURL := range deleteImages {
		if err := h.productStorage.Delete(imageURL); err != nil {
			// Data database sudah diperbarui; catat kegagalan ini
			// agar file yang tersisa dapat dibersihkan.
			c.Error(fmt.Errorf(
				"failed to delete old product image %q: %w",
				imageURL,
				err,
			))
		}
	}

	updated, err := h.productService.GetBySlug(
		c.Request.Context(),
		product.Slug,
	)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "product updated successfully",
			"data":    product,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "product updated successfully",
		"data":    updated,
	})
}

// parseDeleteImages membaca field delete_images berupa JSON array URL.
func parseDeleteImages(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{}, nil
	}

	var images []string
	if err := json.Unmarshal([]byte(value), &images); err != nil {
		return nil, fmt.Errorf(
			"field delete_images must be a valid JSON array",
		)
	}

	seen := make(map[string]bool, len(images))
	result := make([]string, 0, len(images))

	for _, imageURL := range images {
		imageURL = strings.TrimSpace(imageURL)
		if imageURL == "" {
			return nil, fmt.Errorf(
				"field delete_images must not contain empty URLs",
			)
		}

		if !seen[imageURL] {
			seen[imageURL] = true
			result = append(result, imageURL)
		}
	}

	return result, nil
}

// DeleteBySlug menghapus produk berdasarkan slug.
func (h *ProductHandler) DeleteBySlug(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid product slug",
		})
		return
	}

	ctx := c.Request.Context()

	// Ambil produk beserta daftar gambar sebelum data dihapus.
	product, err := h.productService.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "product not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get product before deletion",
			"error":   err.Error(),
		})
		return
	}

	// Hapus produk beserta data relasinya di database.
	if err := h.productService.DeleteBySlug(ctx, slug); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "product not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to delete product",
			"error":   err.Error(),
		})
		return
	}

	// Hapus file gambar fisik setelah database berhasil dihapus.
	var failedFiles []string

	if h.productStorage != nil {
		for _, image := range product.Images {
			if err := h.productStorage.Delete(image.FileURL); err != nil {
				failedFiles = append(failedFiles, image.FileURL)

				// Catat error untuk logging middleware Gin.
				_ = c.Error(fmt.Errorf(
					"failed to delete product image %q: %w",
					image.FileURL,
					err,
				))
			}
		}
	} else if len(product.Images) > 0 {
		failedFiles = append(failedFiles, "product storage is not configured")
	}

	if len(failedFiles) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "product deleted, but some image files could not be deleted",
			"warning": failedFiles,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "product deleted successfully",
	})
}

func parseProductDetails(
	excellencesJSON string,
	specificationsJSON string,
) ([]model.ProductExcellence, []model.ProductSpecification, error) {
	excellences := make([]model.ProductExcellence, 0)
	specifications := make([]model.ProductSpecification, 0)

	if strings.TrimSpace(excellencesJSON) != "" {
		if err := json.Unmarshal([]byte(excellencesJSON), &excellences); err != nil {
			return nil, nil, fmt.Errorf("field excellences must be valid JSON: %w", err)
		}
	}
	if strings.TrimSpace(specificationsJSON) != "" {
		if err := json.Unmarshal([]byte(specificationsJSON), &specifications); err != nil {
			return nil, nil, fmt.Errorf("field specifications must be valid JSON: %w", err)
		}
	}

	return excellences, specifications, nil
}

// validateProductImage memeriksa ekstensi dan signature file, bukan hanya nama file.
func validateProductImage(fileHeader *multipart.FileHeader) error {
	ext := strings.ToLower(strings.TrimSpace(fileHeader.Filename))
	switch {
	case strings.HasSuffix(ext, ".jpg"),
		strings.HasSuffix(ext, ".jpeg"),
		strings.HasSuffix(ext, ".png"),
		strings.HasSuffix(ext, ".webp"):
	default:
		return fmt.Errorf("unsupported image type: %s", fileHeader.Filename)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to read image: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("image file is empty")
	}

	contentType := http.DetectContentType(header[:n])
	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
		return nil
	default:
		return fmt.Errorf("uploaded file is not a supported image")
	}
}
