package main

import (
	"log"
	"os"

	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/hendrisulistya/cashier-app/config"
	"github.com/hendrisulistya/cashier-app/db"
	"github.com/hendrisulistya/cashier-app/ui"
)

func main() {
	// Set up logging to file
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening log file: %v\n", err)
		return
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Println("Starting application...")

	myApp := app.New()
	log.Println("Created new Fyne app")

	myApp.Settings().SetTheme(ui.NewCustomTheme())
	log.Println("Set custom theme")

	window := myApp.NewWindow("Cashier App")
	log.Println("Created new window")

	window.Resize(fyne.NewSize(1024, 768))
	log.Println("Resized window")

	// Create login page
	loginPage := ui.NewLoginPage(window, func(username, password string) bool {
		log.Printf("Login attempt with username: %s", username)
		if username == "admin" && password == "admin" {
			log.Println("Login credentials correct, loading database config...")
			dbConfig := config.LoadConfig()

			log.Println("Running migrations...")
			if err := db.RunMigrations(dbConfig); err != nil {
				log.Printf("Migration error: %v", err)
				return false
			}

			log.Println("Connecting to database...")
			database, err := db.NewConnection(dbConfig)
			if err != nil {
				log.Printf("Database connection error: %v", err)
				return false
			}

			log.Println("Creating main window...")
			mainWindow := ui.NewMainWindow(window, database)
			if err := mainWindow.Load(); err != nil {
				log.Printf("Error loading main window: %v", err)
				return false
			}

			log.Println("Login successful")
			return true
		}
		log.Println("Login failed - invalid credentials")
		return false
	})

	log.Println("Setting initial content to login page")
	window.SetContent(loginPage.Load())

	log.Println("Starting main event loop")
	window.ShowAndRun()
}
