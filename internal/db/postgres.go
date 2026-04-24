package db

import (
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

var migrationsFS embed.FS

func InitDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	return db, err
}

func RunDBMigration(migrationURL string, dbSource string) error {
	d, err := iofs.New(migrationsFS, migrationURL)
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, dbSource)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
