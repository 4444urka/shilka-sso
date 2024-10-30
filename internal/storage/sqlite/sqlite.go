// Package sqlite Слой работы с базой данных
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"shilka-sso/internal/domain/models"
	"shilka-sso/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {
	const operation = "storage.sqlite.New"

	// Подключение к бд
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &Storage{db: db}, nil
}

// SaveUser Сохрарняет пользователя в бд
func (s *Storage) SaveUser(ctx context.Context, username string, passwordHash []byte) (int64, error) {
	const operation = "storage.sqlite.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO users(username, pass_hash) VALUES (?, ?)")

	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	res, err := stmt.ExecContext(ctx, username, passwordHash)

	if err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", operation, storage.ErrUserExists)
		}

		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", operation, err)
	}

	return id, nil
}

// GetUser Получает информацию о пользователе по username.
func (s *Storage) GetUser(ctx context.Context, username string) (models.User, error) {
	const operation = "storage.sqlite.GetUser"

	stmt, err := s.db.Prepare("SELECT id, username, pass_hash FROM users WHERE username = ?")

	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", operation, err)
	}

	row := stmt.QueryRowContext(ctx, username)

	var user models.User
	err = row.Scan(&user.Id, &user.Username, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", operation, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", operation, err)
	}

	return user, nil
}

// IsAdmin Определяет является ли пользователь админом.
func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const operation = "storage.sqlite.IsAdmin"

	stmt, err := s.db.Prepare("SELECT is_admin FROM users WHERE id = ?")

	if err != nil {
		return false, fmt.Errorf("%s: %w", operation, err)
	}

	row := stmt.QueryRowContext(ctx, userID)

	var isAdmin bool

	err = row.Scan(&isAdmin)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", operation, storage.ErrUserNotFound)
		}
		return false, fmt.Errorf("%s: %w", operation, err)
	}

	return isAdmin, nil
}

// GetApp Взвращает приложение по фйди из бд
func (s *Storage) GetApp(ctx context.Context, appID int) (models.App, error) {
	const operation = "storage.sqlite.GetApp"

	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")

	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", operation, err)
	}

	row := stmt.QueryRowContext(ctx, appID)

	var app models.App
	err = row.Scan(&app.Id, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", operation, storage.ErrAppNotFound)
		}
		return models.App{}, fmt.Errorf("%s: %w", operation, err)
	}
	return app, nil
}
