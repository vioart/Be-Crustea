package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"Be-Crustea/internal/model"
	"Be-Crustea/internal/repository"

	"github.com/jackc/pgx/v5"
)

var (
	ErrPartnershipProjectNotFound = errors.New("partnership project not found")
	ErrInvalidPartnershipProject  = errors.New("invalid partnership project data")
)

type PartnershipDetail struct {
	Project  model.PartnershipProject          `json:"project"`
	Partners []model.PartnershipProjectPartner `json:"partners"`
}

type PartnershipService interface {
	GetAll(ctx context.Context) ([]model.PartnershipProject, error)
	GetByID(ctx context.Context, id int64) (*PartnershipDetail, error)
	Create(ctx context.Context, project *model.PartnershipProject) error
	Update(ctx context.Context, project *model.PartnershipProject) error
	Delete(ctx context.Context, id int64) error
}

type partnershipService struct {
	partnershipRepository repository.PartnershipRepository
}

func NewPartnershipService(
	partnershipRepository repository.PartnershipRepository,
) PartnershipService {
	return &partnershipService{
		partnershipRepository: partnershipRepository,
	}
}

func (s *partnershipService) GetAll(
	ctx context.Context,
) ([]model.PartnershipProject, error) {
	projects, err := s.partnershipRepository.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get partnership projects: %w", err)
	}

	return projects, nil
}

func (s *partnershipService) GetByID(
	ctx context.Context,
	id int64,
) (*PartnershipDetail, error) {
	if id <= 0 {
		return nil, ErrPartnershipProjectNotFound
	}

	project, err := s.partnershipRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPartnershipProjectNotFound
		}

		return nil, fmt.Errorf("failed to get partnership project: %w", err)
	}

	partners, err := s.partnershipRepository.GetPartners(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get partnership project partners: %w", err)
	}

	return &PartnershipDetail{
		Project:  *project,
		Partners: partners,
	}, nil
}

func (s *partnershipService) Create(
	ctx context.Context,
	project *model.PartnershipProject,
) error {
	if project == nil {
		return ErrInvalidPartnershipProject
	}

	if err := validatePartnershipProject(project); err != nil {
		return err
	}

	normalizePartnershipProject(project)

	if err := s.partnershipRepository.Create(ctx, project); err != nil {
		return fmt.Errorf("failed to create partnership project: %w", err)
	}

	return nil
}

func (s *partnershipService) Update(
	ctx context.Context,
	project *model.PartnershipProject,
) error {
	if project == nil || project.ID <= 0 {
		return ErrInvalidPartnershipProject
	}

	if err := validatePartnershipProject(project); err != nil {
		return err
	}

	normalizePartnershipProject(project)

	if err := s.partnershipRepository.Update(ctx, project); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPartnershipProjectNotFound
		}

		return fmt.Errorf("failed to update partnership project: %w", err)
	}

	return nil
}

func (s *partnershipService) Delete(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return ErrPartnershipProjectNotFound
	}

	if err := s.partnershipRepository.Delete(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPartnershipProjectNotFound
		}

		return fmt.Errorf("failed to delete partnership project: %w", err)
	}

	return nil
}

func validatePartnershipProject(
	project *model.PartnershipProject,
) error {
	if strings.TrimSpace(project.NameIDN) == "" {
		return fmt.Errorf(
			"%w: name_idn is required",
			ErrInvalidPartnershipProject,
		)
	}

	if strings.TrimSpace(project.NameEN) == "" {
		return fmt.Errorf(
			"%w: name_en is required",
			ErrInvalidPartnershipProject,
		)
	}

	if strings.TrimSpace(project.DescriptionIDN) == "" {
		return fmt.Errorf(
			"%w: description_idn is required",
			ErrInvalidPartnershipProject,
		)
	}

	if strings.TrimSpace(project.DescriptionEN) == "" {
		return fmt.Errorf(
			"%w: description_en is required",
			ErrInvalidPartnershipProject,
		)
	}

	return nil
}

func normalizePartnershipProject(
	project *model.PartnershipProject,
) {
	project.NameIDN = strings.TrimSpace(project.NameIDN)
	project.NameEN = strings.TrimSpace(project.NameEN)
	project.DescriptionIDN = strings.TrimSpace(project.DescriptionIDN)
	project.DescriptionEN = strings.TrimSpace(project.DescriptionEN)
}
