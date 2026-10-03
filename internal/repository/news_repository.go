package repository

import (
	"context"
	"errors"
	"fmt"

	"Be-Crustea/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NewsRepository interface {
	GetAll(ctx context.Context, limit, offset int) ([]model.News, error)
	GetByID(ctx context.Context, id int64) (*model.News, error)
	GetBySlug(ctx context.Context, slug string) (*model.News, error)
	Create(ctx context.Context, news *model.News) error
	Update(ctx context.Context, news *model.News) error
	Delete(ctx context.Context, id int64) error
}

type postgresNewsRepository struct {
	db *pgxpool.Pool
}

func NewNewsRepository(db *pgxpool.Pool) NewsRepository {
	return &postgresNewsRepository{
		db: db,
	}
}

func (r *postgresNewsRepository) GetAll(
	ctx context.Context,
	limit, offset int,
) ([]model.News, error) {
	const query = `
		SELECT
			id,
			category_id,
			author_id,
			slug,
			thumbnail,
			title_idn,
			title_en,
			content_idn,
			content_en,
			status,
			published_at,
			created_at,
			updated_at
		FROM news
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query news: %w", err)
	}
	defer rows.Close()

	newsList := make([]model.News, 0)

	for rows.Next() {
		var news model.News

		err := rows.Scan(
			&news.ID,
			&news.CategoryID,
			&news.AuthorID,
			&news.Slug,
			&news.Thumbnail,
			&news.TitleIDN,
			&news.TitleEN,
			&news.ContentIDN,
			&news.ContentEN,
			&news.Status,
			&news.PublishedAt,
			&news.CreatedAt,
			&news.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan news: %w", err)
		}

		newsList = append(newsList, news)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate news rows: %w", err)
	}

	return newsList, nil
}

func (r *postgresNewsRepository) GetByID(
	ctx context.Context,
	id int64,
) (*model.News, error) {
	const query = `
		SELECT
			id,
			category_id,
			author_id,
			slug,
			thumbnail,
			title_idn,
			title_en,
			content_idn,
			content_en,
			status,
			published_at,
			created_at,
			updated_at
		FROM news
		WHERE id = $1
	`

	var news model.News

	err := r.db.QueryRow(ctx, query, id).Scan(
		&news.ID,
		&news.CategoryID,
		&news.AuthorID,
		&news.Slug,
		&news.Thumbnail,
		&news.TitleIDN,
		&news.TitleEN,
		&news.ContentIDN,
		&news.ContentEN,
		&news.Status,
		&news.PublishedAt,
		&news.CreatedAt,
		&news.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}

		return nil, fmt.Errorf("failed to get news by id: %w", err)
	}

	return &news, nil
}

func (r *postgresNewsRepository) GetBySlug(
	ctx context.Context,
	slug string,
) (*model.News, error) {
	const query = `
		SELECT
			id,
			category_id,
			author_id,
			slug,
			thumbnail,
			title_idn,
			title_en,
			content_idn,
			content_en,
			status,
			published_at,
			created_at,
			updated_at
		FROM news
		WHERE slug = $1
	`

	var news model.News

	err := r.db.QueryRow(ctx, query, slug).Scan(
		&news.ID,
		&news.CategoryID,
		&news.AuthorID,
		&news.Slug,
		&news.Thumbnail,
		&news.TitleIDN,
		&news.TitleEN,
		&news.ContentIDN,
		&news.ContentEN,
		&news.Status,
		&news.PublishedAt,
		&news.CreatedAt,
		&news.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}

		return nil, fmt.Errorf("failed to get news by slug: %w", err)
	}

	return &news, nil
}

func (r *postgresNewsRepository) Create(
	ctx context.Context,
	news *model.News,
) error {
	const query = `
		INSERT INTO news (
			category_id,
			author_id,
			slug,
			thumbnail,
			title_idn,
			title_en,
			content_idn,
			content_en,
			status,
			published_at,
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
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
		news.CategoryID,
		news.AuthorID,
		news.Slug,
		news.Thumbnail,
		news.TitleIDN,
		news.TitleEN,
		news.ContentIDN,
		news.ContentEN,
		news.Status,
		news.PublishedAt,
	).Scan(
		&news.ID,
		&news.CreatedAt,
		&news.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create news: %w", err)
	}

	return nil
}

func (r *postgresNewsRepository) Update(
	ctx context.Context,
	news *model.News,
) error {
	const query = `
		UPDATE news
		SET
			category_id = $1,
			author_id = $2,
			slug = $3,
			thumbnail = $4,
			title_idn = $5,
			title_en = $6,
			content_idn = $7,
			content_en = $8,
			status = $9,
			published_at = $10,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $11
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		news.CategoryID,
		news.AuthorID,
		news.Slug,
		news.Thumbnail,
		news.TitleIDN,
		news.TitleEN,
		news.ContentIDN,
		news.ContentEN,
		news.Status,
		news.PublishedAt,
		news.ID,
	).Scan(&news.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}

		return fmt.Errorf("failed to update news: %w", err)
	}

	return nil
}

func (r *postgresNewsRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	const query = `
		DELETE FROM news
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete news: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
