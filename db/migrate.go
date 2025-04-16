package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hendrisulistya/cashier-app/config"
)

func RunMigrations(config *config.DBConfig) error {
	fmt.Println("Starting database migration...")
	fmt.Printf("Database configuration: Host=%s, Port=%d, User=%s, DBName=%s\n",
		config.Host, config.Port, config.User, config.DBName)

	migrationURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
	)

	// First reset the database schema
	db, err := sql.Open("postgres", migrationURL)
	if err != nil {
		return fmt.Errorf("failed to open db connection: %v", err)
	}
	defer db.Close()

	fmt.Println("Cleaning existing schema...")
	_, err = db.Exec(`
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
		GRANT ALL ON SCHEMA public TO cashier_user;
		GRANT ALL ON SCHEMA public TO public;

		CREATE TABLE IF NOT EXISTS schema_migrations (
			version bigint NOT NULL,
			dirty boolean NOT NULL,
			CONSTRAINT schema_migrations_pkey PRIMARY KEY (version)
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to reset schema: %v", err)
	}

	// Now create a new migration instance
	fmt.Println("Initializing migration...")
	m, err := migrate.New(
		"file://db/migrations",
		migrationURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}
	defer m.Close()

	// Run migrations
	fmt.Println("Applying migrations...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	// Get current version after migration
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current version: %v", err)
	}
	fmt.Printf("Current migration version: %d, Dirty: %v\n", version, dirty)

	if err == migrate.ErrNoChange {
		fmt.Println("No migration needed - database is up to date")
	} else {
		fmt.Println("Successfully migrated database")
	}

	return nil
}
