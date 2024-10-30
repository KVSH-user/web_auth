package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"web_auth/internal/models"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type Auth struct {
	log          *slog.Logger
	usrSaver     UserSaver
	userProvider UserProvider
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
	BlockUserByID(ctx context.Context, userID int64) error
}

type UserProvider interface {
	ProvideUser(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, userID int64) (*models.User, error)
	ListUsers(ctx context.Context, limit, offset int) ([]models.User, error)
}

func New(log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
) *Auth {
	return &Auth{
		usrSaver:     userSaver,
		userProvider: userProvider,
		log:          log,
	}
}

func (a *Auth) RegisterNewUser(ctx context.Context, password, email string,
) (userID int64, err error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(slog.String("op", op))

	log.Info("register new user")

	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passwordHashed)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			log.Warn("user already exists", err)
			return 0, ErrUserExists
		}
		log.Error("failed to save user", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user successfully register")

	return id, nil
}

func (a *Auth) Login(ctx context.Context, email, password string,
) error {
	const op = "auth.Login"

	log := a.log.With(slog.String("op", op))

	log.Info("login attempt")

	user, err := a.userProvider.ProvideUser(ctx, email)
	if err != nil {
		log.Error("failed to provide user", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	if user.ID == 0 {
		log.Error("user not found", ErrUserNotFound)
		return fmt.Errorf("%s: %w", op, ErrUserNotFound)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHashed), []byte(password)); err != nil {
		a.log.Warn("invalid credentials", ErrInvalidCredentials)
		return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("user logged in success")

	return nil
}

func (a *Auth) BlockUser(ctx context.Context, userID int64) error {
	const op = "auth.BlockUser"

	log := a.log.With(slog.String("op", op))
	log.Info("block user attempt", slog.Int64("userID", userID))

	err := a.usrSaver.BlockUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Warn("user not found for blocking", slog.Int64("userID", userID))
			return ErrUserNotFound
		}
		log.Error("failed to block user", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user blocked successfully", slog.Int64("userID", userID))
	return nil
}

func (a *Auth) GetUser(ctx context.Context, userID int64) (*models.User, error) {
	const op = "auth.GetUser"

	log := a.log.With(slog.String("op", op))
	log.Info("get user attempt", slog.Int64("userID", userID))

	user, err := a.userProvider.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Warn("user not found", slog.Int64("userID", userID))
			return nil, ErrUserNotFound
		}
		log.Error("failed to get user", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user retrieved successfully", slog.Int64("userID", userID))
	return user, nil
}

func (a *Auth) ListUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	const op = "auth.ListUsers"

	log := a.log.With(slog.String("op", op))
	log.Info("list users attempt")

	users, err := a.userProvider.ListUsers(ctx, limit, offset)
	if err != nil {
		log.Error("failed to list users", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("users listed successfully", slog.Int("count", len(users)))
	return users, nil
}
