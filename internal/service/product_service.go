package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"Be-Crustea/internal/model"
	"Be-Crustea/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidProduct  = errors.New("invalid product data")
)

const (
	productsCacheKey = "products"
	productsCacheTTL = 5 * time.Minute
)

const (
	ProductCategoryTechnology = "technology"
	ProductCategoryKrasty     = "krasty"
	ProductCategoryTrading    = "trading"
)

type ProductDetail struct {
	Product        model.Product                `json:"product"`
	Images         []model.ProductImage         `json:"images"`
	Excellences    []model.ProductExcellence    `json:"excellences"`
	Specifications []model.ProductSpecification `json:"specifications"`
}

type ProductService interface {
	GetAll(ctx context.Context, limit, offset int) ([]model.Product, error)
	GetByCategory(ctx context.Context, category string, limit, offset int) ([]model.Product, error)
	GetByID(ctx context.Context, id int64) (*ProductDetail, error)
	GetBySlug(ctx context.Context, slug string) (*ProductDetail, error)

	Create(ctx context.Context, product *model.Product) error
	CreateWithDetails(
		ctx context.Context,
		product *model.Product,
		images []model.ProductImage,
		excellences []model.ProductExcellence,
		specifications []model.ProductSpecification,
	) error
	UpdateBySlug(ctx context.Context, slug string, product *model.Product) error
	UpdateWithDetails(
		ctx context.Context,
		slug string,
		product *model.Product,
		images []model.ProductImage,
		deleteImages []string,
		excellences []model.ProductExcellence,
		specifications []model.ProductSpecification,
	) error
	DeleteBySlug(ctx context.Context, slug string) error
}

type productService struct {
	productRepository repository.ProductRepository
	redis             *redis.Client
}

func NewProductService(
	productRepository repository.ProductRepository,
	redisClient *redis.Client,
) ProductService {
	return &productService{
		productRepository: productRepository,
		redis:             redisClient,
	}
}

// normalizePagination memastikan pagination valid.
func normalizePagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return limit, offset
}

// GetAll mengambil semua produk dengan pagination dan Redis cache.
func (s *productService) GetAll(
	ctx context.Context,
	limit, offset int,
) ([]model.Product, error) {
	limit, offset = normalizePagination(limit, offset)
	cacheKey := fmt.Sprintf("%s:all:%d:%d", productsCacheKey, limit, offset)

	if s.redis != nil {
		cachedData, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var products []model.Product
			if json.Unmarshal([]byte(cachedData), &products) == nil {
				return products, nil
			}
		}
	}

	products, err := s.productRepository.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	s.cacheProducts(ctx, cacheKey, products)

	return products, nil
}

// GetByCategory mengambil produk berdasarkan kategori.
func (s *productService) GetByCategory(
	ctx context.Context,
	category string,
	limit, offset int,
) ([]model.Product, error) {
	category = strings.ToLower(strings.TrimSpace(category))

	if !isValidProductCategory(category) {
		return nil, fmt.Errorf(
			"%w: category must be technology, krasty, or trading",
			ErrInvalidProduct,
		)
	}

	limit, offset = normalizePagination(limit, offset)
	cacheKey := fmt.Sprintf(
		"%s:category:%s:%d:%d",
		productsCacheKey,
		category,
		limit,
		offset,
	)

	if s.redis != nil {
		cachedData, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var products []model.Product
			if json.Unmarshal([]byte(cachedData), &products) == nil {
				return products, nil
			}
		}
	}

	products, err := s.productRepository.GetByCategory(
		ctx,
		category,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get products by category %s: %w",
			category,
			err,
		)
	}

	s.cacheProducts(ctx, cacheKey, products)

	return products, nil
}

// cacheProducts menyimpan daftar produk ke Redis.
// Kegagalan Redis tidak menggagalkan request API.
func (s *productService) cacheProducts(
	ctx context.Context,
	cacheKey string,
	products []model.Product,
) {
	if s.redis == nil {
		return
	}

	data, err := json.Marshal(products)
	if err != nil {
		return
	}

	_ = s.redis.Set(ctx, cacheKey, data, productsCacheTTL).Err()
}

// GetByID mengambil produk beserta detail terkait berdasarkan ID.
func (s *productService) GetByID(
	ctx context.Context,
	id int64,
) (*ProductDetail, error) {
	if id <= 0 {
		return nil, ErrProductNotFound
	}

	product, err := s.productRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}

		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return s.getProductDetail(ctx, product)
}

// GetBySlug mengambil produk beserta detail terkait berdasarkan slug.
func (s *productService) GetBySlug(
	ctx context.Context,
	slug string,
) (*ProductDetail, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, ErrProductNotFound
	}

	product, err := s.productRepository.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}

		return nil, fmt.Errorf("failed to get product by slug: %w", err)
	}

	return s.getProductDetail(ctx, product)
}

// getProductDetail mengambil gambar, keunggulan, dan spesifikasi produk.
func (s *productService) getProductDetail(
	ctx context.Context,
	product *model.Product,
) (*ProductDetail, error) {
	if product == nil {
		return nil, ErrProductNotFound
	}

	return &ProductDetail{
		Product:        *product,
		Images:         product.Images,
		Excellences:    product.Excellences,
		Specifications: product.Specifications,
	}, nil
}

// Create membuat produk tanpa detail tambahan.
func (s *productService) Create(
	ctx context.Context,
	product *model.Product,
) error {
	if product == nil {
		return ErrInvalidProduct
	}

	normalizeProduct(product)

	if err := validateProduct(product); err != nil {
		return err
	}

	if err := s.productRepository.Create(ctx, product); err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	_ = s.invalidateProductCache(ctx)

	return nil
}

// CreateWithDetails menyimpan produk beserta gambar,
// keunggulan, dan spesifikasi bilingual.
func (s *productService) CreateWithDetails(
	ctx context.Context,
	product *model.Product,
	images []model.ProductImage,
	excellences []model.ProductExcellence,
	specifications []model.ProductSpecification,
) error {
	if product == nil {
		return ErrInvalidProduct
	}

	normalizeProduct(product)

	if err := validateProduct(product); err != nil {
		return err
	}

	// Validasi gambar.
	for i := range images {
		images[i].FileURL = strings.TrimSpace(images[i].FileURL)

		if images[i].FileURL == "" {
			return fmt.Errorf(
				"%w: image file_url is required",
				ErrInvalidProduct,
			)
		}
	}

	// Validasi keunggulan produk.
	for i := range excellences {
		excellences[i].TitleIDN = strings.TrimSpace(
			excellences[i].TitleIDN,
		)
		excellences[i].TitleEN = strings.TrimSpace(
			excellences[i].TitleEN,
		)

		if excellences[i].TitleIDN == "" ||
			excellences[i].TitleEN == "" {
			return fmt.Errorf(
				"%w: excellence title_idn and title_en are required",
				ErrInvalidProduct,
			)
		}
	}

	// Validasi spesifikasi bilingual.
	for i := range specifications {
		specifications[i].NameIDN = strings.TrimSpace(
			specifications[i].NameIDN,
		)
		specifications[i].NameEN = strings.TrimSpace(
			specifications[i].NameEN,
		)
		specifications[i].ValueIDN = strings.TrimSpace(
			specifications[i].ValueIDN,
		)
		specifications[i].ValueEN = strings.TrimSpace(
			specifications[i].ValueEN,
		)

		if specifications[i].NameIDN == "" ||
			specifications[i].NameEN == "" ||
			specifications[i].ValueIDN == "" ||
			specifications[i].ValueEN == "" {
			return fmt.Errorf(
				"%w: specification name_idn, name_en, value_idn, and value_en are required",
				ErrInvalidProduct,
			)
		}
	}

	if err := s.productRepository.CreateWithDetails(
		ctx,
		product,
		images,
		excellences,
		specifications,
	); err != nil {
		return fmt.Errorf(
			"failed to create product with details: %w",
			err,
		)
	}

	_ = s.invalidateProductCache(ctx)

	return nil
}

// Update memperbarui data utama produk berdasarkan Slug.
func (s *productService) UpdateBySlug(
	ctx context.Context,
	slug string,
	product *model.Product,
) error {
	slug = strings.TrimSpace(slug)

	if slug == "" || product == nil {
		return ErrInvalidProduct
	}

	normalizeProduct(product)

	if err := validateProduct(product); err != nil {
		return err
	}

	if err := s.productRepository.UpdateBySlug(ctx, slug, product); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProductNotFound
		}

		return fmt.Errorf("failed to update product: %w", err)
	}

	_ = s.invalidateProductCache(ctx)

	return nil
}

// UpdateWithDetails memperbarui produk beserta gambar,
// keunggulan, dan spesifikasi berdasarkan slug lama.
func (s *productService) UpdateWithDetails(
	ctx context.Context,
	slug string,
	product *model.Product,
	images []model.ProductImage,
	deleteImages []string,
	excellences []model.ProductExcellence,
	specifications []model.ProductSpecification,
) error {
	slug = strings.TrimSpace(slug)

	if slug == "" || product == nil {
		return ErrInvalidProduct
	}

	normalizeProduct(product)

	if err := validateProduct(product); err != nil {
		return err
	}

	// Validasi gambar baru.
	for i := range images {
		images[i].FileURL = strings.TrimSpace(images[i].FileURL)
		if images[i].FileURL == "" {
			return fmt.Errorf(
				"%w: image file_url is required",
				ErrInvalidProduct,
			)
		}
	}

	// Validasi URL gambar lama yang ingin dihapus.
	for i := range deleteImages {
		deleteImages[i] = strings.TrimSpace(deleteImages[i])
		if deleteImages[i] == "" {
			return fmt.Errorf(
				"%w: deleted image URL is required",
				ErrInvalidProduct,
			)
		}
	}

	// Validasi keunggulan.
	for i := range excellences {
		excellences[i].TitleIDN = strings.TrimSpace(
			excellences[i].TitleIDN,
		)
		excellences[i].TitleEN = strings.TrimSpace(
			excellences[i].TitleEN,
		)

		if excellences[i].TitleIDN == "" ||
			excellences[i].TitleEN == "" {
			return fmt.Errorf(
				"%w: excellence title_idn and title_en are required",
				ErrInvalidProduct,
			)
		}
	}

	// Validasi spesifikasi.
	for i := range specifications {
		specifications[i].NameIDN = strings.TrimSpace(
			specifications[i].NameIDN,
		)
		specifications[i].NameEN = strings.TrimSpace(
			specifications[i].NameEN,
		)
		specifications[i].ValueIDN = strings.TrimSpace(
			specifications[i].ValueIDN,
		)
		specifications[i].ValueEN = strings.TrimSpace(
			specifications[i].ValueEN,
		)

		if specifications[i].NameIDN == "" ||
			specifications[i].NameEN == "" ||
			specifications[i].ValueIDN == "" ||
			specifications[i].ValueEN == "" {
			return fmt.Errorf(
				"%w: specification name_idn, name_en, value_idn, and value_en are required",
				ErrInvalidProduct,
			)
		}
	}

	// Simpan perubahan utama dan seluruh detail melalui repository.
	err := s.productRepository.UpdateWithDetails(
		ctx,
		slug,
		product,
		images,
		deleteImages,
		excellences,
		specifications,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProductNotFound
		}

		return fmt.Errorf(
			"failed to update product with details: %w",
			err,
		)
	}

	// Bersihkan cache setelah update berhasil.
	_ = s.invalidateProductCache(ctx)

	return nil
}

// DeleteBySlug menghapus produk berdasarkan slug.
func (s *productService) DeleteBySlug(
	ctx context.Context,
	slug string,
) error {
	slug = strings.TrimSpace(slug)

	if slug == "" {
		return ErrProductNotFound
	}

	if err := s.productRepository.DeleteBySlug(ctx, slug); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProductNotFound
		}

		return fmt.Errorf("failed to delete product: %w", err)
	}

	_ = s.invalidateProductCache(ctx)

	return nil
}

// normalizeProduct merapikan data produk sebelum validasi.
func normalizeProduct(product *model.Product) {
	product.Category = strings.ToLower(strings.TrimSpace(product.Category))
	product.Slug = strings.TrimSpace(product.Slug)
	product.NameIDN = strings.TrimSpace(product.NameIDN)
	product.NameEN = strings.TrimSpace(product.NameEN)
	product.DescriptionIDN = strings.TrimSpace(product.DescriptionIDN)
	product.DescriptionEN = strings.TrimSpace(product.DescriptionEN)
}

// isValidProductCategory memeriksa kategori yang diperbolehkan.

func isValidProductCategory(category string) bool {
	switch category {
	case ProductCategoryTechnology,
		ProductCategoryKrasty,
		ProductCategoryTrading:
		return true
	default:
		return false
	}
}

// invalidateProductCache menghapus cache daftar produk dan kategori.
func (s *productService) invalidateProductCache(ctx context.Context) error {
	if s.redis == nil {
		return nil
	}

	iter := s.redis.Scan(
		ctx,
		0,
		productsCacheKey+":*",
		100,
	).Iterator()

	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())

		if len(keys) >= 100 {
			if err := s.redis.Del(ctx, keys...).Err(); err != nil {
				return err
			}

			keys = keys[:0]
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		if err := s.redis.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	}

	return nil
}

// validateProduct memvalidasi data utama produk.
func validateProduct(product *model.Product) error {

	if !isValidProductCategory(product.Category) {
		return fmt.Errorf(
			"%w: category must be technology, krasty, or trading",
			ErrInvalidProduct,
		)
	}

	if product.Slug == "" {
		return fmt.Errorf("%w: slug is required", ErrInvalidProduct)
	}

	if product.NameIDN == "" {
		return fmt.Errorf("%w: name_idn is required", ErrInvalidProduct)
	}

	if product.NameEN == "" {
		return fmt.Errorf("%w: name_en is required", ErrInvalidProduct)
	}

	if product.DescriptionIDN == "" {
		return fmt.Errorf(
			"%w: description_idn is required",
			ErrInvalidProduct,
		)
	}

	if product.DescriptionEN == "" {
		return fmt.Errorf(
			"%w: description_en is required",
			ErrInvalidProduct,
		)
	}

	return nil
}
