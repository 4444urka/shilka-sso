// Package auth - Сервисный слой
package auth

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"shilka-sso/internal/domain/models"
	"shilka-sso/internal/lib/jwt"
	"shilka-sso/internal/lib/logger/sl"
	"shilka-sso/internal/storage"
	"time"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

// TODO: Добавить методы для смены пароля и роли

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		username string,
		passwordHash []byte,
	) (userID int64, err error)
}

type UserProvider interface {
	GetUser(ctx context.Context, username string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	GetApp(ctx context.Context, appID int) (models.App, error)
}

// Ошибки сервисного слоя
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppId       = errors.New("invalid app id")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidUserId      = errors.New("invalid user id")
)

// New возвращает новый объект Auth сервиса
func New(
	log *slog.Logger,
	userProvider UserProvider,
	userSaver UserSaver,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

// Методы сервиса Auth

// Login проверяет существует ли пользователь с указанными данными в бд
// Если не существует выдаёт ошибку
// Если пароль не правильный выдаёт ошибку
func (a *Auth) Login(
	ctx context.Context,
	username string,
	password string,
	appID int,
) (string, error) {
	const operator = "auth.Login"

	log := a.log.With(
		slog.String("operator", operator),
		slog.String("username", username),
	)

	log.Info("Trying to login user")

	user, err := a.userProvider.GetUser(ctx, username)

	// Идентефикация пользователя
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Error("GetUser not found with given username", sl.Err(err))

			return "", fmt.Errorf("%s: %w", operator, ErrInvalidCredentials)
		}

		a.log.Error("Failed to get user", err)

		return "", fmt.Errorf("%s: %w", operator, err)
	}

	// Авторизация
	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		a.log.Error("Invalid password", sl.Err(err))

		return "", fmt.Errorf("%s: %w", operator, ErrInvalidCredentials)
	}

	// Проверяем приложение в которое пользователь пытается зайти
	app, err := a.appProvider.GetApp(ctx, appID)

	if err != nil {
		return "", fmt.Errorf("%s: %w", operator, err)
	}

	log.Info("Successfully logged in")

	//	Создание токена
	token, err := jwt.NewToken(user, app, a.tokenTTL)

	if err != nil {
		a.log.Error("Failed to create token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", operator, ErrInvalidCredentials)
	}

	return token, nil
}

// Register создаёт пользователя с указанными данными
// Если пользователь с данным ником уже существует выдаёт ошибку
func (a *Auth) Register(
	ctx context.Context,
	username string,
	password string,
) (int64, error) {
	const operator = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("operator", operator),
		slog.String("username", username),
	)

	log.Info("Registering user")

	// Хэширование пароля
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Error("Failed to hash password", err)

		return 0, fmt.Errorf("%s: %w", operator, err)
	}

	id, err := a.userSaver.SaveUser(ctx, username, passwordHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Error("GetUser already exists", sl.Err(err))
			return 0, fmt.Errorf("%s: %w", operator, ErrUserExists)
		}
		log.Error("Failed to save user", err)
		return 0, fmt.Errorf("%s: %w", operator, err)
	}

	log.Info("Successfully registered user")

	return id, nil
}

// IsAdmin возвращает true если пользователь с указанным ID админ и false в обратном случае
// Возвращает ошибку если пользователя с указанным ID не существует
func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const operator = "auth.IsAdmin"

	log := a.log.With(
		slog.String("operator", operator),
		slog.String("userID", fmt.Sprint(userID)),
	)

	log.Info("Checking if user is admin")

	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("GetUser not found with given username", sl.Err(err))
			return false, fmt.Errorf("%s: %w", operator, ErrInvalidUserId)
		}

		log.Error("Failed to check if user is admin", sl.Err(err))
		return false, fmt.Errorf("%s: %w", operator, err)
	}

	log.Info("checked if user is admin", slog.Bool("IsAdmin", isAdmin))

	return isAdmin, nil
}
