package ui

import (
	"database/sql"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MainWindow struct {
	window   fyne.Window
	database *sql.DB
}

func NewMainWindow(window fyne.Window, database *sql.DB) *MainWindow {
	return &MainWindow{
		window:   window,
		database: database,
	}
}

func (m *MainWindow) Load() error {
	content := m.createMainMenu()
	m.window.SetContent(content)
	return nil
}

func (m *MainWindow) createMainMenu() fyne.CanvasObject {
	// Create header with logo
	logo := canvas.NewImageFromFile("assets/logo.svg")
	logo.SetMinSize(fyne.NewSize(200, 60))
	logo.FillMode = canvas.ImageFillOriginal

	header := container.NewCenter(
		container.NewVBox(
			logo,
			widget.NewLabelWithStyle("Point of Sale System", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		),
	)

	// Create menu grid with modern styling
	menuGrid := container.NewGridWithColumns(2,
		createMenuButton("Cashier", theme.ListIcon(), func() {
			cashierWindow := NewCashierWindow(m.window, m.database)
			if err := cashierWindow.Load(); err != nil {
				log.Printf("Error loading cashier window: %v", err)
			}
		}),

		createMenuButton("Inventory", theme.ListIcon(), func() {
			inventoryWindow := NewInventoryWindow(m.window, m.database)
			if err := inventoryWindow.Load(); err != nil {
				log.Printf("Error loading inventory window: %v", err)
				dialog.ShowError(err, m.window)
			}
		}),

		createMenuButton("Reports", theme.DocumentIcon(), func() {
			reportsWindow := NewReportWindow(m.window, m.database)
			if err := reportsWindow.Load(); err != nil {
				log.Printf("Error loading reports window: %v", err)
				dialog.ShowError(err, m.window)
			}
		}),

		createMenuButton("Settings", theme.SettingsIcon(), func() {
			settingsWindow := NewSettingsWindow(m.window, m.database)
			if err := settingsWindow.Load(); err != nil {
				log.Printf("Error loading settings window: %v", err)
				dialog.ShowError(err, m.window)
			}
		}),
	)

	// Create logout button with different style
	logoutButton := widget.NewButtonWithIcon("Logout", theme.LogoutIcon(), func() {
		loginPage := NewLoginPage(m.window, func(username, password string) bool {
			if username == "admin" && password == "admin" {
				mainWindow := NewMainWindow(m.window, m.database)
				if err := mainWindow.Load(); err != nil {
					log.Printf("Error loading main window: %v", err)
					return false
				}
				return true
			}
			return false
		})
		m.window.SetContent(loginPage.Load())
	})
	logoutButton.Importance = widget.DangerImportance

	// Helper function to create styled menu buttons
	footer := container.NewHBox(
		widget.NewLabel("© 2024 Cashier App"),
		layout.NewSpacer(),
		logoutButton,
	)

	// Main layout with padding and spacing
	content := container.NewBorder(
		header,
		footer,
		nil,
		nil,
		container.NewPadded(
			container.NewVBox(
				layout.NewSpacer(),
				menuGrid,
				layout.NewSpacer(),
			),
		),
	)

	return content
}

// Helper function to create consistent menu buttons
func createMenuButton(label string, icon fyne.Resource, action func()) *widget.Button {
	btn := widget.NewButtonWithIcon(label, icon, action)
	btn.Importance = widget.HighImportance
	btn.Resize(fyne.NewSize(200, 80))
	return btn
}
