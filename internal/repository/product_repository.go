package repository

import (
	"context"
	"errors"
	"fmt"

	"Be-Crustea/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository interface {
	GetAll(ctx context.Context, limit, offset int) ([]model.Product, error)
	GetByID(ctx context.Context, id int64) (*model.Product, error)
	GetBySlug(ctx context.Context, slug string) (*model.Product, error)

	GetImages(ctx context.Context, productID int64) ([]model.ProductImage, error)
	GetExcellences(ctx context.Context, productID int64) ([]model.ProductExcellence, error)
	GetSpecifications(ctx context.Context, productID int64) ([]model.ProductSpecification, error)

	Create(ctx context.Context, product *model.Product) error
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id int64) error
}

type postgresProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) ProductRepository {
	return &postgresProductRepository{
		db: db,
	}
}

func (r *postgresProductRepository) GetAll(
	ctx context.Context,
	limit, offset int,
) ([]model.Product, error) {
	const query = `
		SELECT
			id,
			category_id,
			slug,
			name_idn,
			name_en,
			description_idn,
			description_en,
			created_at,
			updated_at
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	products := make([]model.Product, 0)

	for rows.Next() {
		var product model.Product

		err := rows.Scan(
			&product.ID,
			&product.CategoryID,
			&product.Slug,
			&product.NameIDN,
			&product.NameEN,
			&product.DescriptionIDN,
			&product.DescriptionEN,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product rows: %w", err)
	}

	return products, nil
}

func (r *postgresProductRepository) GetByID(
	ctx context.Context,
	id int64,
) (*model.Product, error) {
	const query = `
		SELECT
			id,
			category_id,
			slug,
			name_idn,
			name_en,
			description_idn,
			description_en,
			created_at,
			updated_at
		FROM products
		WHERE id = $1
	`

	var product model.Product

	err := r.db.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.CategoryID,
		&product.Slug,
		&product.NameIDN,
		&product.NameEN,
		&product.DescriptionIDN,
		&product.DescriptionEN,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}

		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}

	return &product, nil
}

func (r *postgresProductRepository) GetBySlug(
	ctx context.Context,
	slug string,
) (*model.Product, error) {
	const query = `
		SELECT
			id,
			category_id,
			slug,
			name_idn,
			name_en,
			description_idn,
			description_en,
			created_at,
			updated_at
		FROM products
		WHERE slug = $1
	`

	var product model.Product

	err := r.db.QueryRow(ctx, query, slug).Scan(
		&product.ID,
		&product.CategoryID,
		&product.Slug,
		&product.NameIDN,
		&product.NameEN,
		&product.DescriptionIDN,
		&product.DescriptionEN,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}

		return nil, fmt.Errorf("failed to get product by slug: %w", err)
	}

	return &product, nil
}

func (r *postgresProductRepository) GetImages(
	ctx context.Context,
	productID int64,
) ([]model.ProductImage, error) {
	const query = `
		SELECT
			id,
			product_id,
			file_url,
			created_at,
			updated_at
		FROM product_images
		WHERE product_id = $1
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to query product images: %w", err)
	}
	defer rows.Close()

	images := make([]model.ProductImage, 0)

	for rows.Next() {
		var image model.ProductImage

		err := rows.Scan(
			&image.ID,
			&image.ProductID,
			&image.FileURL,
			&image.CreatedAt,
			&image.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product image: %w", err)
		}

		images = append(images, image)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product image rows: %w", err)
	}

	return images, nil
}

func (r *postgresProductRepository) GetExcellences(
	ctx context.Context,
	productID int64,
) ([]model.ProductExcellence, error) {
	const query = `
		SELECT
			id,
			product_id,
			title_idn,
			title_en,
			created_at,
			updated_at
		FROM product_excellences
		WHERE product_id = $1
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to query product excellences: %w", err)
	}
	defer rows.Close()

	excellences := make([]model.ProductExcellence, 0)

	for rows.Next() {
		var excellence model.ProductExcellence

		err := rows.Scan(
			&excellence.ID,
			&excellence.ProductID,
			&excellence.TitleIDN,
			&excellence.TitleEN,
			&excellence.CreatedAt,
			&excellence.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product excellence: %w", err)
		}

		excellences = append(excellences, excellence)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product excellence rows: %w", err)
	}

	return excellences, nil
}

func (r *postgresProductRepository) GetSpecifications(
	ctx context.Context,
	productID int64,
) ([]model.ProductSpecification, error) {
	const query = `
		SELECT
			id,
			product_id,
			name_idn,
			name_en,
			value,
			created_at,
			updated_at
		FROM product_specifications
		WHERE product_id = $1
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to query product specifications: %w", err)
	}
	defer rows.Close()

	specifications := make([]model.ProductSpecification, 0)

	for rows.Next() {
		var specification model.ProductSpecification

		err := rows.Scan(
			&specification.ID,
			&specification.ProductID,
			&specification.NameIDN,
			&specification.NameEN,
			&specification.Value,
			&specification.CreatedAt,
			&specification.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product specification: %w", err)
		}

		specifications = append(specifications, specification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product specification rows: %w", err)
	}

	return specifications, nil
}

func (r *postgresProductRepository) Create(
	ctx context.Context,
	product *model.Product,
) error {
	const query = `
		INSERT INTO products (
			category_id,
			slug,
			name_idn,
			name_en,
			description_idn,
			description_en,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			CURRENT_TIMESTAMP,
			CURRENT_TIMESTAMP
		)
		RETURNING
			id,
			created_at,
			updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		product.CategoryID,
		product.Slug,
		product.NameIDN,
		product.NameEN,
		product.DescriptionIDN,
		product.DescriptionEN,
	).Scan(
		&product.ID,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

func (r *postgresProductRepository) Update(
	ctx context.Context,
	product *model.Product,
) error {
	const query = `
		UPDATE products
		SET
			category_id = $1,
			slug = $2,
			name_idn = $3,
			name_en = $4,
			description_idn = $5,
			description_en = $6,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		product.CategoryID,
		product.Slug,
		product.NameIDN,
		product.NameEN,
		product.DescriptionIDN,
		product.DescriptionEN,
		product.ID,
	).Scan(&product.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}

		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

func (r *postgresProductRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	const query = `
		DELETE FROM products
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
