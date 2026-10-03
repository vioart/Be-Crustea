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
	productsCacheKey = "products:all"
	productsCacheTTL = 5 * time.Minute
)

type ProductDetail struct {
	Product        model.Product                `json:"product"`
	Images         []model.ProductImage         `json:"images"`
	Excellences    []model.ProductExcellence    `json:"excellences"`
	Specifications []model.ProductSpecification `json:"specifications"`
}

type ProductService interface {
	GetAll(ctx context.Context, limit, offset int) ([]model.Product, error)
	GetByID(ctx context.Context, id int64) (*ProductDetail, error)
	GetBySlug(ctx context.Context, slug string) (*ProductDetail, error)
	Create(ctx context.Context, product *model.Product) error
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id int64) error
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

func (s *productService) GetAll(
	ctx context.Context,
	limit, offset int,
) ([]model.Product, error) {
	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	// ============================================
	// REDIS CACHE
	// ============================================

	cacheKey := fmt.Sprintf(
		"%s:%d:%d",
		productsCacheKey,
		limit,
		offset,
	)

	cachedData, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var products []model.Product

		if err := json.Unmarshal(
			[]byte(cachedData),
			&products,
		); err == nil {
			return products, nil
		}
	}

	// ============================================
	// DATABASE
	// ============================================

	products, err := s.productRepository.GetAll(
		ctx,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get products: %w",
			err,
		)
	}

	// ============================================
	// SAVE TO REDIS
	// ============================================

	data, err := json.Marshal(products)
	if err == nil {
		// Redis failure should not break the API response.
		_ = s.redis.Set(
			ctx,
			cacheKey,
			data,
			productsCacheTTL,
		).Err()
	}

	return products, nil
}

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

		return nil, fmt.Errorf(
			"failed to get product: %w",
			err,
		)
	}

	return s.getProductDetail(ctx, product)
}

func (s *productService) GetBySlug(
	ctx context.Context,
	slug string,
) (*ProductDetail, error) {
	slug = strings.TrimSpace(slug)

	if slug == "" {
		return nil, ErrProductNotFound
	}

	product, err := s.productRepository.GetBySlug(
		ctx,
		slug,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}

		return nil, fmt.Errorf(
			"failed to get product by slug: %w",
			err,
		)
	}

	return s.getProductDetail(ctx, product)
}

func (s *productService) getProductDetail(
	ctx context.Context,
	product *model.Product,
) (*ProductDetail, error) {
	images, err := s.productRepository.GetImages(
		ctx,
		product.ID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get product images: %w",
			err,
		)
	}

	excellences, err := s.productRepository.GetExcellences(
		ctx,
		product.ID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get product excellences: %w",
			err,
		)
	}

	specifications, err := s.productRepository.GetSpecifications(
		ctx,
		product.ID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get product specifications: %w",
			err,
		)
	}

	return &ProductDetail{
		Product:        *product,
		Images:         images,
		Excellences:    excellences,
		Specifications: specifications,
	}, nil
}

func (s *productService) Create(
	ctx context.Context,
	product *model.Product,
) error {
	if product == nil {
		return ErrInvalidProduct
	}

	if err := validateProduct(product); err != nil {
		return err
	}

	product.Slug = strings.TrimSpace(product.Slug)

	if err := s.productRepository.Create(
		ctx,
		product,
	); err != nil {
		return fmt.Errorf(
			"failed to create product: %w",
			err,
		)
	}

	// Invalidate product list cache.
	_ = s.invalidateProductCache(ctx)

	return nil
}

func (s *productService) Update(
	ctx context.Context,
	product *model.Product,
) error {
	if product == nil || product.ID <= 0 {
		return ErrInvalidProduct
	}

	if err := validateProduct(product); err != nil {
		return err
	}

	product.Slug = strings.TrimSpace(product.Slug)

	if err := s.productRepository.Update(
		ctx,
		product,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProductNotFound
		}

		return fmt.Errorf(
			"failed to update product: %w",
			err,
		)
	}

	// Invalidate product list cache.
	_ = s.invalidateProductCache(ctx)

	return nil
}

func (s *productService) Delete(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return ErrProductNotFound
	}

	if err := s.productRepository.Delete(
		ctx,
		id,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrProductNotFound
		}

		return fmt.Errorf(
			"failed to delete product: %w",
			err,
		)
	}

	// Invalidate product list cache.
	_ = s.invalidateProductCache(ctx)

	return nil
}

func (s *productService) invalidateProductCache(
	ctx context.Context,
) error {
	iter := s.redis.Scan(
		ctx,
		0,
		productsCacheKey+":*",
		0,
	).Iterator()

	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return s.redis.Del(ctx, keys...).Err()
}

func validateProduct(product *model.Product) error {
	if product.CategoryID <= 0 {
		return fmt.Errorf(
			"%w: category_id is required",
			ErrInvalidProduct,
		)
	}

	if strings.TrimSpace(product.Slug) == "" {
		return fmt.Errorf(
			"%w: slug is required",
			ErrInvalidProduct,
		)
	}

	if strings.TrimSpace(product.NameIDN) == "" {
		return fmt.Errorf(
			"%w: name_idn is required",
			ErrInvalidProduct,
		)
	}

	if strings.TrimSpace(product.NameEN) == "" {
		return fmt.Errorf(
			"%w: name_en is required",
			ErrInvalidProduct,
		)
	}

	if strings.TrimSpace(product.DescriptionIDN) == "" {
		return fmt.Errorf(
			"%w: description_idn is required",
			ErrInvalidProduct,
		)
	}

	if strings.TrimSpace(product.DescriptionEN) == "" {
		return fmt.Errorf(
			"%w: description_en is required",
			ErrInvalidProduct,
		)
	}

	return nil
}
