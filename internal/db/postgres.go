package db

import (
	"database/sql"
	"log/slog"

	"github.com/1saswata/chess-broadcast-engine/db/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

func InitDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	return db, err
}

func RunDBMigration(migrationURL string, db *sql.DB) error {
	sourceDriver, err := iofs.New(migrations.FS, migrationURL)
	if err != nil {
		return err
	}
	targetDriver, err := postgres.WithInstance(db, &postgres.Config{})
	m, err := migrate.NewWithInstance("iofs", sourceDriver,
		"postgres", targetDriver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	slog.Info("Database migrations applied successfully")
	return nil
}
