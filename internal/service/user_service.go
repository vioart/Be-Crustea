package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"Be-Crustea/internal/model"
	"Be-Crustea/internal/repository"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrInvalidUser            = errors.New("invalid user data")
	ErrEmailExists            = errors.New("email already exists")
	ErrInvalidCurrentPassword = errors.New("current password is incorrect")
	ErrPasswordUnchanged      = errors.New("new password must differ from current password")
)

type UserService interface {
	GetAll(ctx context.Context) ([]model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	ChangePassword(
		ctx context.Context,
		id int64,
		currentPassword string,
		newPassword string,
	) error
	Delete(ctx context.Context, id int64) error
}

type userService struct {
	userRepository repository.UserRepository
}

func NewUserService(
	userRepository repository.UserRepository,
) UserService {
	return &userService{
		userRepository: userRepository,
	}
}

func (s *userService) GetAll(
	ctx context.Context,
) ([]model.User, error) {
	users, err := s.userRepository.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get users: %w",
			err,
		)
	}

	return users, nil
}

func (s *userService) GetByID(
	ctx context.Context,
	id int64,
) (*model.User, error) {
	if id <= 0 {
		return nil, ErrUserNotFound
	}

	user, err := s.userRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf(
			"failed to get user: %w",
			err,
		)
	}

	return user, nil
}

func (s *userService) Create(
	ctx context.Context,
	user *model.User,
) error {
	if user == nil {
		return ErrInvalidUser
	}

	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.ToLower(
		strings.TrimSpace(user.Email),
	)

	if user.Name == "" {
		return fmt.Errorf(
			"%w: name is required",
			ErrInvalidUser,
		)
	}

	if user.Email == "" {
		return fmt.Errorf(
			"%w: email is required",
			ErrInvalidUser,
		)
	}

	if strings.TrimSpace(user.Password) == "" {
		return fmt.Errorf(
			"%w: password is required",
			ErrInvalidUser,
		)
	}

	if len(user.Password) < 8 {
		return fmt.Errorf(
			"%w: password must be at least 8 characters",
			ErrInvalidUser,
		)
	}

	// Periksa apakah email sudah digunakan.
	existingUser, err := s.userRepository.GetByEmail(
		ctx,
		user.Email,
	)

	if err == nil && existingUser != nil {
		return ErrEmailExists
	}

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf(
			"failed to check user email: %w",
			err,
		)
	}

	// Hash password sebelum disimpan.
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(user.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to hash password: %w",
			err,
		)
	}

	user.Password = string(hashedPassword)

	if err := s.userRepository.Create(ctx, user); err != nil {
		return fmt.Errorf(
			"failed to create user: %w",
			err,
		)
	}

	return nil
}

func (s *userService) Update(
	ctx context.Context,
	user *model.User,
) error {
	if user == nil || user.ID <= 0 {
		return ErrInvalidUser
	}

	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.ToLower(
		strings.TrimSpace(user.Email),
	)

	if user.Name == "" {
		return fmt.Errorf(
			"%w: name is required",
			ErrInvalidUser,
		)
	}

	if user.Email == "" {
		return fmt.Errorf(
			"%w: email is required",
			ErrInvalidUser,
		)
	}

	// Pastikan pengguna memang ada.
	existingUser, err := s.userRepository.GetByID(
		ctx,
		user.ID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}

		return fmt.Errorf(
			"failed to get existing user: %w",
			err,
		)
	}

	// Periksa apakah email digunakan pengguna lain.
	emailUser, err := s.userRepository.GetByEmail(
		ctx,
		user.Email,
	)

	if err == nil && emailUser != nil {
		if emailUser.ID != user.ID {
			return ErrEmailExists
		}
	}

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf(
			"failed to check user email: %w",
			err,
		)
	}

	// Jika password kosong, pertahankan hash password lama.
	if strings.TrimSpace(user.Password) == "" {
		user.Password = existingUser.Password
	} else {
		if len(user.Password) < 8 {
			return fmt.Errorf(
				"%w: password must be at least 8 characters",
				ErrInvalidUser,
			)
		}

		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(user.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to hash password: %w",
				err,
			)
		}

		user.Password = string(hashedPassword)
	}

	// Pertahankan avatar lama jika tidak ada avatar baru.
	if user.Avatar == nil {
		user.Avatar = existingUser.Avatar
	}

	if err := s.userRepository.Update(ctx, user); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}

		return fmt.Errorf(
			"failed to update user: %w",
			err,
		)
	}

	return nil
}

// ChangePassword mengganti password pengguna yang sedang login.
// Password lama harus valid sebelum password baru disimpan.
func (s *userService) ChangePassword(
	ctx context.Context,
	id int64,
	currentPassword string,
	newPassword string,
) error {
	if id <= 0 {
		return ErrUserNotFound
	}

	if strings.TrimSpace(currentPassword) == "" ||
		strings.TrimSpace(newPassword) == "" {
		return fmt.Errorf(
			"%w: current and new password are required",
			ErrInvalidUser,
		)
	}

	if len(newPassword) < 8 {
		return fmt.Errorf(
			"%w: new password must be at least 8 characters",
			ErrInvalidUser,
		)
	}

	// Ambil pengguna beserta hash password saat ini.
	user, err := s.userRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}

		return fmt.Errorf(
			"failed to get user: %w",
			err,
		)
	}

	// Verifikasi password lama.
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(currentPassword),
	); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCurrentPassword
		}

		return fmt.Errorf(
			"failed to verify current password: %w",
			err,
		)
	}

	// Pastikan password baru berbeda dari password lama.
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(newPassword),
	); err == nil {
		return ErrPasswordUnchanged
	} else if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return fmt.Errorf(
			"failed to verify new password: %w",
			err,
		)
	}

	// Hash password baru.
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(newPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to hash new password: %w",
			err,
		)
	}

	// Perbarui hanya kolom password.
	if err := s.userRepository.UpdatePassword(
		ctx,
		id,
		string(hashedPassword),
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}

		return fmt.Errorf(
			"failed to update password: %w",
			err,
		)
	}

	return nil
}

func (s *userService) Delete(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return ErrUserNotFound
	}

	if err := s.userRepository.Delete(ctx, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}

		return fmt.Errorf(
			"failed to delete user: %w",
			err,
		)
	}

	return nil
}
