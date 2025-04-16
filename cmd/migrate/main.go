package main

import (
	"fmt"
	"os"

	"github.com/hendrisulistya/cashier-app/config"
	"github.com/hendrisulistya/cashier-app/db"
)

func main() {
	// Load database configuration from .env
	dbConfig := config.LoadConfig()

	// Log the start of the migration process
	fmt.Println("Starting migration process...")
	fmt.Printf("Database configuration: Host=%s, Port=%d, User=%s, DBName=%s\n",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.DBName)

	// Run migrations
	if err := db.RunMigrations(dbConfig); err != nil {
		fmt.Printf("Migration failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Migration completed successfully")
}
