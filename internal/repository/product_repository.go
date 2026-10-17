package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"Be-Crustea/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository interface {
	GetAll(ctx context.Context, limit, offset int) ([]model.Product, error)
	GetByCategory(ctx context.Context, category string, limit, offset int) ([]model.Product, error)
	GetByID(ctx context.Context, id int64) (*model.Product, error)
	GetBySlug(ctx context.Context, slug string) (*model.Product, error)
	GetImages(ctx context.Context, productID int64) ([]model.ProductImage, error)
	GetExcellences(ctx context.Context, productID int64) ([]model.ProductExcellence, error)
	GetSpecifications(ctx context.Context, productID int64) ([]model.ProductSpecification, error)
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

type postgresProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) ProductRepository {
	return &postgresProductRepository{db: db}
}

const productColumns = `
	id,
	category,
	slug,
	name_idn,
	name_en,
	description_idn,
	description_en,
	created_at,
	updated_at
`

func scanProduct(row pgx.Row) (*model.Product, error) {
	var product model.Product

	err := row.Scan(
		&product.ID,
		&product.Category,
		&product.Slug,
		&product.NameIDN,
		&product.NameEN,
		&product.DescriptionIDN,
		&product.DescriptionEN,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *postgresProductRepository) loadProductDetails(
	ctx context.Context,
	product *model.Product,
) error {
	images, err := r.GetImages(ctx, product.ID)
	if err != nil {
		return fmt.Errorf(
			"failed to load images for product %d: %w",
			product.ID,
			err,
		)
	}

	excellences, err := r.GetExcellences(ctx, product.ID)
	if err != nil {
		return fmt.Errorf(
			"failed to load excellences for product %d: %w",
			product.ID,
			err,
		)
	}

	specifications, err := r.GetSpecifications(ctx, product.ID)
	if err != nil {
		return fmt.Errorf(
			"failed to load specifications for product %d: %w",
			product.ID,
			err,
		)
	}

	product.Images = images
	product.Excellences = excellences
	product.Specifications = specifications

	return nil
}

func (r *postgresProductRepository) GetAll(
	ctx context.Context,
	limit, offset int,
) ([]model.Product, error) {
	query := `SELECT ` + productColumns + `
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}

	products := make([]model.Product, 0)

	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		products = append(products, *product)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("failed to iterate product rows: %w", err)
	}

	rows.Close()

	for i := range products {
		if err := r.loadProductDetails(ctx, &products[i]); err != nil {
			return nil, err
		}
	}

	return products, nil
}

func (r *postgresProductRepository) GetByCategory(
	ctx context.Context,
	category string,
	limit, offset int,
) ([]model.Product, error) {
	query := `SELECT ` + productColumns + `
		FROM products
		WHERE category = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, category, limit, offset)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query products by category: %w",
			err,
		)
	}

	products := make([]model.Product, 0)

	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf(
				"failed to scan product: %w",
				err,
			)
		}

		products = append(products, *product)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf(
			"failed to iterate product rows: %w",
			err,
		)
	}

	rows.Close()

	for i := range products {
		if err := r.loadProductDetails(ctx, &products[i]); err != nil {
			return nil, err
		}
	}

	return products, nil
}

func (r *postgresProductRepository) GetByID(
	ctx context.Context,
	id int64,
) (*model.Product, error) {
	query := `SELECT ` + productColumns + `
		FROM products
		WHERE id = $1`

	product, err := scanProduct(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}

		return nil, fmt.Errorf(
			"failed to get product by id: %w",
			err,
		)
	}

	if err := r.loadProductDetails(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (r *postgresProductRepository) GetBySlug(
	ctx context.Context,
	slug string,
) (*model.Product, error) {
	query := `SELECT ` + productColumns + `
		FROM products
		WHERE slug = $1`

	product, err := scanProduct(r.db.QueryRow(ctx, query, slug))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}

		return nil, fmt.Errorf(
			"failed to get product by slug: %w",
			err,
		)
	}

	if err := r.loadProductDetails(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (r *postgresProductRepository) GetImages(
	ctx context.Context,
	productID int64,
) ([]model.ProductImage, error) {
	const query = `
		SELECT id, product_id, file_url, created_at, updated_at
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
		var item model.ProductImage

		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.FileURL,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product image: %w", err)
		}

		images = append(images, item)
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
		SELECT id, product_id, title_idn, title_en, created_at, updated_at
		FROM product_excellences
		WHERE product_id = $1
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to query product excellences: %w", err)
	}
	defer rows.Close()

	items := make([]model.ProductExcellence, 0)

	for rows.Next() {
		var item model.ProductExcellence

		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.TitleIDN,
			&item.TitleEN,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product excellence: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product excellence rows: %w", err)
	}

	return items, nil
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
			value_idn,
			value_en,
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

	items := make([]model.ProductSpecification, 0)

	for rows.Next() {
		var item model.ProductSpecification

		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.NameIDN,
			&item.NameEN,
			&item.ValueIDN,
			&item.ValueEN,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product specification: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product specification rows: %w", err)
	}

	return items, nil
}

func (r *postgresProductRepository) Create(
	ctx context.Context,
	product *model.Product,
) error {
	const query = `
		INSERT INTO products (
			category,
			slug,
			name_idn,
			name_en,
			description_idn,
			description_en,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		product.Category,
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

// CreateWithDetails menyimpan produk beserta gambar,
// keunggulan, dan spesifikasinya dalam satu transaksi.
func (r *postgresProductRepository) CreateWithDetails(
	ctx context.Context,
	product *model.Product,
	images []model.ProductImage,
	excellences []model.ProductExcellence,
	specifications []model.ProductSpecification,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin product transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const productQuery = `
		INSERT INTO products (
			category,
			slug,
			name_idn,
			name_en,
			description_idn,
			description_en,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(
		ctx,
		productQuery,
		product.Category,
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

	const imageQuery = `
		INSERT INTO product_images (
			product_id,
			file_url,
			created_at,
			updated_at
		)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	for _, image := range images {
		if strings.TrimSpace(image.FileURL) == "" {
			return fmt.Errorf("product image file_url cannot be empty")
		}

		if _, err := tx.Exec(ctx, imageQuery, product.ID, image.FileURL); err != nil {
			return fmt.Errorf("failed to create product image: %w", err)
		}
	}

	const excellenceQuery = `
		INSERT INTO product_excellences (
			product_id,
			title_idn,
			title_en,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	for _, item := range excellences {
		if _, err := tx.Exec(
			ctx,
			excellenceQuery,
			product.ID,
			item.TitleIDN,
			item.TitleEN,
		); err != nil {
			return fmt.Errorf("failed to create product excellence: %w", err)
		}
	}

	const specificationQuery = `
		INSERT INTO product_specifications (
			product_id,
			name_idn,
			name_en,
			value_idn,
			value_en,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	for _, item := range specifications {
		if _, err := tx.Exec(
			ctx,
			specificationQuery,
			product.ID,
			item.NameIDN,
			item.NameEN,
			item.ValueIDN,
			item.ValueEN,
		); err != nil {
			return fmt.Errorf("failed to create product specification: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit product transaction: %w", err)
	}

	return nil
}

func (r *postgresProductRepository) UpdateBySlug(
	ctx context.Context,
	slug string,
	product *model.Product,
) error {
	const query = `
        UPDATE products
        SET category = $1,
            slug = $2,
            name_idn = $3,
            name_en = $4,
            description_idn = $5,
            description_en = $6,
            updated_at = CURRENT_TIMESTAMP
        WHERE slug = $7
        RETURNING id, created_at, updated_at
    `

	err := r.db.QueryRow(
		ctx,
		query,
		product.Category,
		product.Slug,
		product.NameIDN,
		product.NameEN,
		product.DescriptionIDN,
		product.DescriptionEN,
		slug,
	).Scan(
		&product.ID,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}

		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

func (r *postgresProductRepository) UpdateWithDetails(
	ctx context.Context,
	slug string,
	product *model.Product,
	images []model.ProductImage,
	deleteImages []string,
	excellences []model.ProductExcellence,
	specifications []model.ProductSpecification,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin product update transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Update produk berdasarkan slug lama dari URL.
	const updateProductQuery = `
        UPDATE products
        SET category = $1,
            slug = $2,
            name_idn = $3,
            name_en = $4,
            description_idn = $5,
            description_en = $6,
            updated_at = CURRENT_TIMESTAMP
        WHERE slug = $7
        RETURNING id, created_at, updated_at
    `

	err = tx.QueryRow(
		ctx,
		updateProductQuery,
		product.Category,
		product.Slug,
		product.NameIDN,
		product.NameEN,
		product.DescriptionIDN,
		product.DescriptionEN,
		slug,
	).Scan(
		&product.ID,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("failed to update product: %w", err)
	}

	// Hapus hanya gambar lama yang dipilih untuk dihapus.
	const deleteImageQuery = `
        DELETE FROM product_images
        WHERE product_id = $1 AND file_url = $2
    `

	for _, fileURL := range deleteImages {
		result, err := tx.Exec(
			ctx,
			deleteImageQuery,
			product.ID,
			fileURL,
		)
		if err != nil {
			return fmt.Errorf("failed to delete product image record: %w", err)
		}

		if result.RowsAffected() == 0 {
			return fmt.Errorf("product image not found: %s", fileURL)
		}
	}

	// Simpan gambar baru tanpa menghapus gambar lama yang dipertahankan.
	const insertImageQuery = `
        INSERT INTO product_images (
            product_id,
            file_url,
            created_at,
            updated_at
        )
        VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    `

	for _, image := range images {
		fileURL := strings.TrimSpace(image.FileURL)
		if fileURL == "" {
			return fmt.Errorf("product image file_url cannot be empty")
		}

		if _, err := tx.Exec(
			ctx,
			insertImageQuery,
			product.ID,
			fileURL,
		); err != nil {
			return fmt.Errorf("failed to insert product image: %w", err)
		}
	}

	// Ganti daftar keunggulan dengan data terbaru.
	if _, err := tx.Exec(
		ctx,
		`DELETE FROM product_excellences WHERE product_id = $1`,
		product.ID,
	); err != nil {
		return fmt.Errorf("failed to delete product excellences: %w", err)
	}

	const insertExcellenceQuery = `
        INSERT INTO product_excellences (
            product_id,
            title_idn,
            title_en,
            created_at,
            updated_at
        )
        VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    `

	for _, item := range excellences {
		if _, err := tx.Exec(
			ctx,
			insertExcellenceQuery,
			product.ID,
			item.TitleIDN,
			item.TitleEN,
		); err != nil {
			return fmt.Errorf("failed to insert product excellence: %w", err)
		}
	}

	// Ganti daftar spesifikasi dengan data terbaru.
	if _, err := tx.Exec(
		ctx,
		`DELETE FROM product_specifications WHERE product_id = $1`,
		product.ID,
	); err != nil {
		return fmt.Errorf("failed to delete product specifications: %w", err)
	}

	const insertSpecificationQuery = `
        INSERT INTO product_specifications (
            product_id,
            name_idn,
            name_en,
            value_idn,
            value_en,
            created_at,
            updated_at
        )
        VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    `

	for _, item := range specifications {
		if _, err := tx.Exec(
			ctx,
			insertSpecificationQuery,
			product.ID,
			item.NameIDN,
			item.NameEN,
			item.ValueIDN,
			item.ValueEN,
		); err != nil {
			return fmt.Errorf("failed to insert product specification: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit product update transaction: %w", err)
	}

	return nil
}

func (r *postgresProductRepository) DeleteBySlug(
	ctx context.Context,
	slug string,
) error {
	const query = `DELETE FROM products WHERE slug = $1`

	result, err := r.db.Exec(ctx, query, slug)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
