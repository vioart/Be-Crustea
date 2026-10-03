package repository

import (
	"context"
	"errors"
	"fmt"

	"Be-Crustea/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PartnershipRepository interface {
	GetAll(ctx context.Context) ([]model.PartnershipProject, error)
	GetByID(ctx context.Context, id int64) (*model.PartnershipProject, error)
	GetPartners(ctx context.Context, projectID int64) ([]model.PartnershipProjectPartner, error)
	Create(ctx context.Context, project *model.PartnershipProject) error
	Update(ctx context.Context, project *model.PartnershipProject) error
	Delete(ctx context.Context, id int64) error
}

type postgresPartnershipRepository struct {
	db *pgxpool.Pool
}

func NewPartnershipRepository(db *pgxpool.Pool) PartnershipRepository {
	return &postgresPartnershipRepository{
		db: db,
	}
}

func (r *postgresPartnershipRepository) GetAll(
	ctx context.Context,
) ([]model.PartnershipProject, error) {
	const query = `
		SELECT
			id,
			name_idn,
			name_en,
			description_idn,
			description_en,
			image,
			created_at,
			updated_at
		FROM partnership_projects
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query partnership projects: %w", err)
	}
	defer rows.Close()

	projects := make([]model.PartnershipProject, 0)

	for rows.Next() {
		var project model.PartnershipProject

		err := rows.Scan(
			&project.ID,
			&project.NameIDN,
			&project.NameEN,
			&project.DescriptionIDN,
			&project.DescriptionEN,
			&project.Image,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan partnership project: %w", err)
		}

		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate partnership project rows: %w", err)
	}

	return projects, nil
}

func (r *postgresPartnershipRepository) GetByID(
	ctx context.Context,
	id int64,
) (*model.PartnershipProject, error) {
	const query = `
		SELECT
			id,
			name_idn,
			name_en,
			description_idn,
			description_en,
			image,
			created_at,
			updated_at
		FROM partnership_projects
		WHERE id = $1
	`

	var project model.PartnershipProject

	err := r.db.QueryRow(ctx, query, id).Scan(
		&project.ID,
		&project.NameIDN,
		&project.NameEN,
		&project.DescriptionIDN,
		&project.DescriptionEN,
		&project.Image,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}

		return nil, fmt.Errorf("failed to get partnership project by id: %w", err)
	}

	return &project, nil
}

func (r *postgresPartnershipRepository) GetPartners(
	ctx context.Context,
	projectID int64,
) ([]model.PartnershipProjectPartner, error) {
	const query = `
		SELECT
			id,
			partnership_project_id,
			logo,
			created_at,
			updated_at
		FROM partnership_project_partners
		WHERE partnership_project_id = $1
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query partnership project partners: %w", err)
	}
	defer rows.Close()

	partners := make([]model.PartnershipProjectPartner, 0)

	for rows.Next() {
		var partner model.PartnershipProjectPartner

		err := rows.Scan(
			&partner.ID,
			&partner.PartnershipProjectID,
			&partner.Logo,
			&partner.CreatedAt,
			&partner.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan partnership project partner: %w", err)
		}

		partners = append(partners, partner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate partnership project partner rows: %w", err)
	}

	return partners, nil
}

func (r *postgresPartnershipRepository) Create(
	ctx context.Context,
	project *model.PartnershipProject,
) error {
	const query = `
		INSERT INTO partnership_projects (
			name_idn,
			name_en,
			description_idn,
			description_en,
			image,
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
		project.NameIDN,
		project.NameEN,
		project.DescriptionIDN,
		project.DescriptionEN,
		project.Image,
	).Scan(
		&project.ID,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create partnership project: %w", err)
	}

	return nil
}

func (r *postgresPartnershipRepository) Update(
	ctx context.Context,
	project *model.PartnershipProject,
) error {
	const query = `
		UPDATE partnership_projects
		SET
			name_idn = $1,
			name_en = $2,
			description_idn = $3,
			description_en = $4,
			image = $5,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		project.NameIDN,
		project.NameEN,
		project.DescriptionIDN,
		project.DescriptionEN,
		project.Image,
		project.ID,
	).Scan(&project.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}

		return fmt.Errorf("failed to update partnership project: %w", err)
	}

	return nil
}

func (r *postgresPartnershipRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	const query = `
		DELETE FROM partnership_projects
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete partnership project: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
