package main

import (
	"log"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/hendrisulistya/cashier-app/config"
	"github.com/hendrisulistya/cashier-app/db"
	"github.com/hendrisulistya/cashier-app/ui"
)

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(ui.NewCustomTheme())
	
	window := myApp.NewWindow("Cashier App")
	window.Resize(fyne.NewSize(1024, 768))

	// Create login page
	loginPage := ui.NewLoginPage(window, func(username, password string) bool {
		if username == "admin" && password == "admin" {
			// Load database configuration from .env
			dbConfig := config.LoadConfig()

			// Run migrations
			if err := db.RunMigrations(dbConfig); err != nil {
				log.Printf("Migration error: %v", err)
				return false
			}

			// Connect to database
			database, err := db.NewConnection(dbConfig)
			if err != nil {
				log.Printf("Database connection error: %v", err)
				return false
			}

			// Create and load main window
			mainWindow := ui.NewMainWindow(window, database)
			if err := mainWindow.Load(); err != nil {
				log.Printf("Error loading main window: %v", err)
				return false
			}
			
			return true
		}
		return false
	})

	// Set initial content to login page
	window.SetContent(loginPage.Load())
	window.ShowAndRun()
}
