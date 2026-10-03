package repository

import (
	"context"
	"errors"
	"fmt"

	"Be-Crustea/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FAQRepository interface {
	GetAll(ctx context.Context) ([]model.FAQ, error)
	GetByID(ctx context.Context, id int64) (*model.FAQ, error)
	Create(ctx context.Context, faq *model.FAQ) error
	Update(ctx context.Context, faq *model.FAQ) error
	Delete(ctx context.Context, id int64) error
}

type postgresFAQRepository struct {
	db *pgxpool.Pool
}

func NewFAQRepository(db *pgxpool.Pool) FAQRepository {
	return &postgresFAQRepository{
		db: db,
	}
}

func (r *postgresFAQRepository) GetAll(
	ctx context.Context,
) ([]model.FAQ, error) {
	const query = `
		SELECT
			id,
			category,
			question_idn,
			question_en,
			answer_idn,
			answer_en,
			created_at,
			updated_at
		FROM faqs
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query faqs: %w", err)
	}
	defer rows.Close()

	faqs := make([]model.FAQ, 0)

	for rows.Next() {
		var faq model.FAQ

		err := rows.Scan(
			&faq.ID,
			&faq.Category,
			&faq.QuestionIDN,
			&faq.QuestionEN,
			&faq.AnswerIDN,
			&faq.AnswerEN,
			&faq.CreatedAt,
			&faq.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan faq: %w", err)
		}

		faqs = append(faqs, faq)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate faq rows: %w", err)
	}

	return faqs, nil
}

func (r *postgresFAQRepository) GetByID(
	ctx context.Context,
	id int64,
) (*model.FAQ, error) {
	const query = `
		SELECT
			id,
			category,
			question_idn,
			question_en,
			answer_idn,
			answer_en,
			created_at,
			updated_at
		FROM faqs
		WHERE id = $1
	`

	var faq model.FAQ

	err := r.db.QueryRow(ctx, query, id).Scan(
		&faq.ID,
		&faq.Category,
		&faq.QuestionIDN,
		&faq.QuestionEN,
		&faq.AnswerIDN,
		&faq.AnswerEN,
		&faq.CreatedAt,
		&faq.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}

		return nil, fmt.Errorf("failed to get faq by id: %w", err)
	}

	return &faq, nil
}

func (r *postgresFAQRepository) Create(
	ctx context.Context,
	faq *model.FAQ,
) error {
	const query = `
		INSERT INTO faqs (
			category,
			question_idn,
			question_en,
			answer_idn,
			answer_en,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
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
		faq.Category,
		faq.QuestionIDN,
		faq.QuestionEN,
		faq.AnswerIDN,
		faq.AnswerEN,
	).Scan(
		&faq.ID,
		&faq.CreatedAt,
		&faq.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create faq: %w", err)
	}

	return nil
}

func (r *postgresFAQRepository) Update(
	ctx context.Context,
	faq *model.FAQ,
) error {
	const query = `
		UPDATE faqs
		SET
			category = $1,
			question_idn = $2,
			question_en = $3,
			answer_idn = $4,
			answer_en = $5,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		faq.Category,
		faq.QuestionIDN,
		faq.QuestionEN,
		faq.AnswerIDN,
		faq.AnswerEN,
		faq.ID,
	).Scan(&faq.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}

		return fmt.Errorf("failed to update faq: %w", err)
	}

	return nil
}

func (r *postgresFAQRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	const query = `
		DELETE FROM faqs
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete faq: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
