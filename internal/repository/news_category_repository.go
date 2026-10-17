package repository

import (
	"context"
	"errors"
	"fmt"

	"Be-Crustea/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NewsCategoryRepository interface {
	GetAll(ctx context.Context) ([]model.NewsCategory, error)
	GetByID(ctx context.Context, id int64) (*model.NewsCategory, error)
	Create(ctx context.Context, category *model.NewsCategory) error
	Update(ctx context.Context, category *model.NewsCategory) error
	Delete(ctx context.Context, id int64) error
}

type postgresNewsCategoryRepository struct {
	db *pgxpool.Pool
}

func NewNewsCategoryRepository(
	db *pgxpool.Pool,
) NewsCategoryRepository {
	return &postgresNewsCategoryRepository{
		db: db,
	}
}

// GetAll mengambil seluruh kategori berita.
// GET /api/news-categories
func (r *postgresNewsCategoryRepository) GetAll(
	ctx context.Context,
) ([]model.NewsCategory, error) {
	const query = `
		SELECT
			id,
			slug,
			name_idn,
			name_en,
			is_active,
			created_at,
			updated_at
		FROM news_categories
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query news categories: %w",
			err,
		)
	}
	defer rows.Close()

	categories := make([]model.NewsCategory, 0)

	for rows.Next() {
		var category model.NewsCategory

		if err := rows.Scan(
			&category.ID,
			&category.Slug,
			&category.NameIDN,
			&category.NameEN,
			&category.IsActive,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"failed to scan news category: %w",
				err,
			)
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"failed to iterate news category rows: %w",
			err,
		)
	}

	return categories, nil
}

// GetByID mengambil kategori berita berdasarkan ID.
// GET /api/news-categories/:id
func (r *postgresNewsCategoryRepository) GetByID(
	ctx context.Context,
	id int64,
) (*model.NewsCategory, error) {
	const query = `
		SELECT
			id,
			slug,
			name_idn,
			name_en,
			is_active,
			created_at,
			updated_at
		FROM news_categories
		WHERE id = $1
	`

	var category model.NewsCategory

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&category.ID,
		&category.Slug,
		&category.NameIDN,
		&category.NameEN,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}

		return nil, fmt.Errorf(
			"failed to get news category: %w",
			err,
		)
	}

	return &category, nil
}

// Create membuat kategori berita baru.
func (r *postgresNewsCategoryRepository) Create(
	ctx context.Context,
	category *model.NewsCategory,
) error {
	const query = `
		INSERT INTO news_categories (
			slug,
			name_idn,
			name_en,
			is_active,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
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
		category.Slug,
		category.NameIDN,
		category.NameEN,
		category.IsActive,
	).Scan(
		&category.ID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf(
			"failed to create news category: %w",
			err,
		)
	}

	return nil
}

// Update memperbarui kategori berita berdasarkan ID.
func (r *postgresNewsCategoryRepository) Update(
	ctx context.Context,
	category *model.NewsCategory,
) error {
	const query = `
		UPDATE news_categories
		SET
			slug = $1,
			name_idn = $2,
			name_en = $3,
			is_active = $4,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING
			created_at,
			updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		category.Slug,
		category.NameIDN,
		category.NameEN,
		category.IsActive,
		category.ID,
	).Scan(
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}

		return fmt.Errorf(
			"failed to update news category: %w",
			err,
		)
	}

	return nil
}

// Delete menghapus kategori berita berdasarkan ID.
func (r *postgresNewsCategoryRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	const query = `
		DELETE FROM news_categories
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		id,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to delete news category: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
