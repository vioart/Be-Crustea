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
	ErrNewsCategoryNotFound = errors.New("news category not found")
	ErrInvalidNewsCategory  = errors.New("invalid news category data")
)

type NewsCategoryService interface {
	GetAll(ctx context.Context) ([]model.NewsCategory, error)
	GetByID(ctx context.Context, id int64) (*model.NewsCategory, error)
	Create(ctx context.Context, category *model.NewsCategory) error
	Update(ctx context.Context, category *model.NewsCategory) error
	Delete(ctx context.Context, id int64) error
}

type newsCategoryService struct {
	repository repository.NewsCategoryRepository
}

func NewNewsCategoryService(
	repository repository.NewsCategoryRepository,
) NewsCategoryService {
	return &newsCategoryService{
		repository: repository,
	}
}

// GetAll mengambil seluruh kategori berita.
func (s *newsCategoryService) GetAll(
	ctx context.Context,
) ([]model.NewsCategory, error) {
	categories, err := s.repository.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get news categories: %w",
			err,
		)
	}

	return categories, nil
}

// GetByID mengambil kategori berdasarkan ID.
func (s *newsCategoryService) GetByID(
	ctx context.Context,
	id int64,
) (*model.NewsCategory, error) {
	if id <= 0 {
		return nil, ErrNewsCategoryNotFound
	}

	category, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNewsCategoryNotFound
		}

		return nil, fmt.Errorf(
			"failed to get news category: %w",
			err,
		)
	}

	return category, nil
}

// Create membuat kategori berita baru.
func (s *newsCategoryService) Create(
	ctx context.Context,
	category *model.NewsCategory,
) error {
	if category == nil {
		return fmt.Errorf(
			"%w: category cannot be nil",
			ErrInvalidNewsCategory,
		)
	}

	normalizeNewsCategory(category)

	if err := validateNewsCategory(category); err != nil {
		return err
	}

	if err := s.repository.Create(ctx, category); err != nil {
		return fmt.Errorf(
			"failed to create news category: %w",
			err,
		)
	}

	return nil
}

// Update memperbarui kategori berita berdasarkan ID.
func (s *newsCategoryService) Update(
	ctx context.Context,
	category *model.NewsCategory,
) error {
	if category == nil {
		return fmt.Errorf(
			"%w: category cannot be nil",
			ErrInvalidNewsCategory,
		)
	}

	if category.ID <= 0 {
		return fmt.Errorf(
			"%w: invalid category id",
			ErrInvalidNewsCategory,
		)
	}

	normalizeNewsCategory(category)

	if err := validateNewsCategory(category); err != nil {
		return err
	}

	if err := s.repository.Update(ctx, category); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNewsCategoryNotFound
		}

		return fmt.Errorf(
			"failed to update news category: %w",
			err,
		)
	}

	return nil
}

// Delete menghapus kategori berita berdasarkan ID.
func (s *newsCategoryService) Delete(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return ErrNewsCategoryNotFound
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNewsCategoryNotFound
		}

		return fmt.Errorf(
			"failed to delete news category: %w",
			err,
		)
	}

	return nil
}

// normalizeNewsCategory membersihkan spasi pada input.
func normalizeNewsCategory(
	category *model.NewsCategory,
) {
	category.Slug = strings.TrimSpace(category.Slug)
	category.NameIDN = strings.TrimSpace(category.NameIDN)
	category.NameEN = strings.TrimSpace(category.NameEN)
}

// validateNewsCategory memvalidasi field wajib.
func validateNewsCategory(
	category *model.NewsCategory,
) error {
	if category.Slug == "" {
		return fmt.Errorf(
			"%w: slug is required",
			ErrInvalidNewsCategory,
		)
	}

	if category.NameIDN == "" {
		return fmt.Errorf(
			"%w: name_idn is required",
			ErrInvalidNewsCategory,
		)
	}

	if category.NameEN == "" {
		return fmt.Errorf(
			"%w: name_en is required",
			ErrInvalidNewsCategory,
		)
	}

	return nil
}
