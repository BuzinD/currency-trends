package infrastructure

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"

	"log"
	"path/filepath"
)

func MigrateTestDb(db *sql.DB) {
	absPath, err := filepath.Abs("../../../migrations")
	if err != nil {
		log.Fatalf("failed to get absolute path: %v", err)
	}
	migrationPath := fmt.Sprintf("file://%s", absPath)

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		log.Fatalf("failed to create sqlite driver: %v", err)
	}

	// Создаём миграции
	m, err := migrate.NewWithDatabaseInstance(migrationPath, "sqlite3", driver)
	if err != nil {
		log.Fatalf("failed to initialize migrations: %v", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(migrate.ErrNoChange, err) {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	log.Println("Migrations applied successfully")
}

func ClearTestDb(db *sql.DB) {

}
