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
	ErrFAQNotFound = errors.New("faq not found")
	ErrInvalidFAQ  = errors.New("invalid faq data")
)

type FAQService interface {
	GetAll(ctx context.Context) ([]model.FAQ, error)
	GetByID(ctx context.Context, id int64) (*model.FAQ, error)
	Create(ctx context.Context, faq *model.FAQ) error
	Update(ctx context.Context, faq *model.FAQ) error
	Delete(ctx context.Context, id int64) error
}

type faqService struct {
	faqRepository repository.FAQRepository
}

func NewFAQService(
	faqRepository repository.FAQRepository,
) FAQService {
	return &faqService{
		faqRepository: faqRepository,
	}
}

func (s *faqService) GetAll(
	ctx context.Context,
) ([]model.FAQ, error) {
	faqs, err := s.faqRepository.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get faqs: %w", err)
	}

	return faqs, nil
}

func (s *faqService) GetByID(
	ctx context.Context,
	id int64,
) (*model.FAQ, error) {
	if id <= 0 {
		return nil, ErrFAQNotFound
	}

	faq, err := s.faqRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFAQNotFound
		}

		return nil, fmt.Errorf("failed to get faq: %w", err)
	}

	return faq, nil
}

func (s *faqService) Create(
	ctx context.Context,
	faq *model.FAQ,
) error {
	if faq == nil {
		return ErrInvalidFAQ
	}

	if err := validateFAQ(faq); err != nil {
		return err
	}

	normalizeFAQ(faq)

	if err := s.faqRepository.Create(ctx, faq); err != nil {
		return fmt.Errorf("failed to create faq: %w", err)
	}

	return nil
}

func (s *faqService) Update(
	ctx context.Context,
	faq *model.FAQ,
) error {
	if faq == nil || faq.ID <= 0 {
		return ErrInvalidFAQ
	}

	if err := validateFAQ(faq); err != nil {
		return err
	}

	normalizeFAQ(faq)

	if err := s.faqRepository.Update(ctx, faq); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrFAQNotFound
		}

		return fmt.Errorf("failed to update faq: %w", err)
	}

	return nil
}

func (s *faqService) Delete(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return ErrFAQNotFound
	}

	if err := s.faqRepository.Delete(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrFAQNotFound
		}

		return fmt.Errorf("failed to delete faq: %w", err)
	}

	return nil
}

func validateFAQ(faq *model.FAQ) error {
	switch faq.Category {
	case model.FAQCategoryProduct,
		model.FAQCategoryService,
		model.FAQCategoryPartnership:
		// valid
	default:
		return fmt.Errorf("%w: invalid category", ErrInvalidFAQ)
	}

	if strings.TrimSpace(faq.QuestionIDN) == "" {
		return fmt.Errorf("%w: question_idn is required", ErrInvalidFAQ)
	}

	if strings.TrimSpace(faq.QuestionEN) == "" {
		return fmt.Errorf("%w: question_en is required", ErrInvalidFAQ)
	}

	if strings.TrimSpace(faq.AnswerIDN) == "" {
		return fmt.Errorf("%w: answer_idn is required", ErrInvalidFAQ)
	}

	if strings.TrimSpace(faq.AnswerEN) == "" {
		return fmt.Errorf("%w: answer_en is required", ErrInvalidFAQ)
	}

	return nil
}

func normalizeFAQ(faq *model.FAQ) {
	faq.QuestionIDN = strings.TrimSpace(faq.QuestionIDN)
	faq.QuestionEN = strings.TrimSpace(faq.QuestionEN)
	faq.AnswerIDN = strings.TrimSpace(faq.AnswerIDN)
	faq.AnswerEN = strings.TrimSpace(faq.AnswerEN)
}
