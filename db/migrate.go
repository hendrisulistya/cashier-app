package db

import (
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

	fmt.Println("Connecting to database...")
	m, err := migrate.New(
		"file://db/migrations",
		migrationURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}
	defer m.Close()

	// Get current version before migration
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current version: %v", err)
	}
	fmt.Printf("Current migration version: %d, Dirty: %v\n", version, dirty)

	// Run migrations
	fmt.Println("Applying migrations...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	// Get new version after migration
	newVersion, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get new version: %v", err)
	}
	fmt.Printf("New migration version: %d, Dirty: %v\n", newVersion, dirty)

	if err == migrate.ErrNoChange {
		fmt.Println("No migration needed - database is up to date")
	} else {
		fmt.Printf("Successfully migrated from version %d to %d\n", version, newVersion)
	}

	return nil
}
