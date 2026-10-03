package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"Be-Crustea/internal/model"
	"Be-Crustea/internal/repository"

	"github.com/jackc/pgx/v5"
)

var (
	ErrNewsNotFound = errors.New("news not found")
	ErrInvalidNews  = errors.New("invalid news data")
)

type NewsService interface {
	GetAll(ctx context.Context, limit, offset int) ([]model.News, error)
	GetByID(ctx context.Context, id int64) (*model.News, error)
	GetBySlug(ctx context.Context, slug string) (*model.News, error)
	Create(ctx context.Context, news *model.News) error
	Update(ctx context.Context, news *model.News) error
	Delete(ctx context.Context, id int64) error
}

type newsService struct {
	newsRepository repository.NewsRepository
}

func NewNewsService(
	newsRepository repository.NewsRepository,
) NewsService {
	return &newsService{
		newsRepository: newsRepository,
	}
}

func (s *newsService) GetAll(
	ctx context.Context,
	limit, offset int,
) ([]model.News, error) {
	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	news, err := s.newsRepository.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get news: %w", err)
	}

	return news, nil
}

func (s *newsService) GetByID(
	ctx context.Context,
	id int64,
) (*model.News, error) {
	if id <= 0 {
		return nil, ErrNewsNotFound
	}

	news, err := s.newsRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNewsNotFound
		}

		return nil, fmt.Errorf("failed to get news: %w", err)
	}

	return news, nil
}

func (s *newsService) GetBySlug(
	ctx context.Context,
	slug string,
) (*model.News, error) {
	slug = strings.TrimSpace(slug)

	if slug == "" {
		return nil, ErrNewsNotFound
	}

	news, err := s.newsRepository.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNewsNotFound
		}

		return nil, fmt.Errorf("failed to get news by slug: %w", err)
	}

	return news, nil
}

func (s *newsService) Create(
	ctx context.Context,
	news *model.News,
) error {
	if news == nil {
		return ErrInvalidNews
	}

	if err := validateNews(news); err != nil {
		return err
	}

	news.Slug = strings.TrimSpace(news.Slug)

	if news.Status == "" {
		news.Status = model.NewsStatusDraft
	}

	if news.Status == model.NewsStatusPublished && news.PublishedAt == nil {
		now := time.Now()
		news.PublishedAt = &now
	}

	if news.Status == model.NewsStatusDraft {
		news.PublishedAt = nil
	}

	if err := s.newsRepository.Create(ctx, news); err != nil {
		return fmt.Errorf("failed to create news: %w", err)
	}

	return nil
}

func (s *newsService) Update(
	ctx context.Context,
	news *model.News,
) error {
	if news == nil || news.ID <= 0 {
		return ErrInvalidNews
	}

	if err := validateNews(news); err != nil {
		return err
	}

	news.Slug = strings.TrimSpace(news.Slug)

	if news.Status == model.NewsStatusPublished && news.PublishedAt == nil {
		now := time.Now()
		news.PublishedAt = &now
	}

	if news.Status == model.NewsStatusDraft {
		news.PublishedAt = nil
	}

	if err := s.newsRepository.Update(ctx, news); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNewsNotFound
		}

		return fmt.Errorf("failed to update news: %w", err)
	}

	return nil
}

func (s *newsService) Delete(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return ErrNewsNotFound
	}

	if err := s.newsRepository.Delete(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNewsNotFound
		}

		return fmt.Errorf("failed to delete news: %w", err)
	}

	return nil
}

func validateNews(news *model.News) error {
	if strings.TrimSpace(news.Slug) == "" {
		return fmt.Errorf("%w: slug is required", ErrInvalidNews)
	}

	if strings.TrimSpace(news.TitleIDN) == "" {
		return fmt.Errorf("%w: title_idn is required", ErrInvalidNews)
	}

	if strings.TrimSpace(news.TitleEN) == "" {
		return fmt.Errorf("%w: title_en is required", ErrInvalidNews)
	}

	if strings.TrimSpace(news.ContentIDN) == "" {
		return fmt.Errorf("%w: content_idn is required", ErrInvalidNews)
	}

	if strings.TrimSpace(news.ContentEN) == "" {
		return fmt.Errorf("%w: content_en is required", ErrInvalidNews)
	}

	if news.CategoryID <= 0 {
		return fmt.Errorf("%w: category_id is required", ErrInvalidNews)
	}

	if news.AuthorID <= 0 {
		return fmt.Errorf("%w: author_id is required", ErrInvalidNews)
	}

	switch news.Status {
	case "":
	case model.NewsStatusDraft:
	case model.NewsStatusPublished:
	default:
		return fmt.Errorf("%w: invalid status", ErrInvalidNews)
	}

	return nil
}
