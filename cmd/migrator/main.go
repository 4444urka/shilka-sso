// В этом файле описываются миграции для базы данных, для их применения нужно запустить
// go run ./cmd/migrator/main.go --storage-path=<Путь до бд> --migrations-path=<Пусть до миграций>
package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var storagePath, migrationsPath, migrationsTable string

	flag.StringVar(&storagePath, "storage-path", "", "Path to a directory containing migration files")
	flag.StringVar(&migrationsPath, "migrations-path", "", "Path to a directory containing migration files")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "Name of migrations table")
	flag.Parse()

	if storagePath == "" {
		panic("storage-path is required")
	}

	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", storagePath, migrationsTable),
	)

	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No migrations found")

			return
		}

		panic(err)
	}
	fmt.Println("Successfully migrated migrations")
}
