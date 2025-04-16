package ui

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/hendrisulistya/cashier-app/db"
)

type SettingsWindow struct {
	window   fyne.Window
	database *sql.DB
}

func NewSettingsWindow(window fyne.Window, database *sql.DB) *SettingsWindow {
	return &SettingsWindow{
		window:   window,
		database: database,
	}
}

func (s *SettingsWindow) Load() error {
	content := s.createSettingsContent()
	s.window.SetContent(content)
	return nil
}

func (s *SettingsWindow) createSettingsContent() fyne.CanvasObject {
	// Back button
	backButton := widget.NewButton("Back to Menu", func() {
		mainWindow := NewMainWindow(s.window, s.database)
		if err := mainWindow.Load(); err != nil {
			log.Printf("Error returning to main menu: %v", err)
		}
	})

	// Header
	header := container.NewHBox(
		backButton,
		widget.NewLabel("Store Settings"),
	)

	// Load current settings
	settings, err := db.GetSettings(s.database)
	if err != nil {
		log.Printf("Error loading settings: %v", err)
		dialog.ShowError(err, s.window)
		return container.NewVBox()
	}

	// Store Information Form
	storeNameEntry := widget.NewEntry()
	storeNameEntry.SetText(settings.StoreName)
	storeNameEntry.SetPlaceHolder("Enter store name")

	storeAddressEntry := widget.NewMultiLineEntry()
	storeAddressEntry.SetText(settings.StoreAddress)
	storeAddressEntry.SetPlaceHolder("Enter store address")

	storePhoneEntry := widget.NewEntry()
	storePhoneEntry.SetText(settings.StorePhone)
	storePhoneEntry.SetPlaceHolder("Enter store phone")

	// Tax Settings
	taxEntry := widget.NewEntry()
	taxEntry.SetText(fmt.Sprintf("%.2f", settings.TaxPercentage))
	taxEntry.SetPlaceHolder("Enter tax percentage")

	// Invoice Settings
	invoicePrefixEntry := widget.NewEntry()
	var prefix string
	err = s.database.QueryRow("SELECT value FROM settings WHERE key = 'invoice_prefix'").Scan(&prefix)
	if err != nil {
		log.Printf("Error loading invoice prefix: %v", err)
	}
	invoicePrefixEntry.SetText(prefix)
	invoicePrefixEntry.SetPlaceHolder("Enter invoice prefix")

	// Last Invoice Number (read-only)
	var lastInvoiceNum string
	err = s.database.QueryRow("SELECT value FROM settings WHERE key = 'last_invoice_number'").Scan(&lastInvoiceNum)
	if err != nil {
		log.Printf("Error loading last invoice number: %v", err)
	}
	lastInvoiceLabel := widget.NewLabel(fmt.Sprintf("Last Invoice Number: %s", lastInvoiceNum))

	// Printer Settings
	printerNameEntry := widget.NewEntry()
	var printerName string
	err = s.database.QueryRow("SELECT value FROM settings WHERE key = 'printer_name'").Scan(&printerName)
	if err != nil {
		log.Printf("Error loading printer name: %v", err)
	}
	printerNameEntry.SetText(printerName)
	printerNameEntry.SetPlaceHolder("Enter printer name")

	printerPortEntry := widget.NewEntry()
	var printerPort string
	err = s.database.QueryRow("SELECT value FROM settings WHERE key = 'printer_port'").Scan(&printerPort)
	if err != nil {
		log.Printf("Error loading printer port: %v", err)
	}
	printerPortEntry.SetText(printerPort)
	printerPortEntry.SetPlaceHolder("Enter printer port")

	paperWidthEntry := widget.NewEntry()
	var paperWidth string
	err = s.database.QueryRow("SELECT value FROM settings WHERE key = 'paper_width'").Scan(&paperWidth)
	if err != nil {
		log.Printf("Error loading paper width: %v", err)
	}
	paperWidthEntry.SetText(paperWidth)
	paperWidthEntry.SetPlaceHolder("Enter paper width in inches")

	// Save Button
	saveButton := widget.NewButton("Save Settings", func() {
		// Validate tax percentage
		taxPercentage, err := strconv.ParseFloat(taxEntry.Text, 64)
		if err != nil || taxPercentage < 0 || taxPercentage > 100 {
			dialog.ShowError(fmt.Errorf("invalid tax percentage"), s.window)
			return
		}

		// Start transaction
		tx, err := s.database.Begin()
		if err != nil {
			dialog.ShowError(fmt.Errorf("database error: %v", err), s.window)
			return
		}
		defer tx.Rollback()

		// Update settings
		updates := map[string]string{
			"store_name":     storeNameEntry.Text,
			"store_address":  storeAddressEntry.Text,
			"store_phone":    storePhoneEntry.Text,
			"tax_percentage": taxEntry.Text,
			"invoice_prefix": invoicePrefixEntry.Text,
		}

		for key, value := range updates {
			_, err := tx.Exec("UPDATE settings SET value = $1, updated_at = CURRENT_TIMESTAMP WHERE key = $2",
				value, key)
			if err != nil {
				dialog.ShowError(fmt.Errorf("error updating %s: %v", key, err), s.window)
				return
			}
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			dialog.ShowError(fmt.Errorf("error saving settings: %v", err), s.window)
			return
		}

		dialog.ShowInformation("Success", "Settings saved successfully", s.window)
	})

	// Reset Invoice Number Button
	resetInvoiceButton := widget.NewButton("Reset Invoice Number", func() {
		confirmDialog := dialog.NewConfirm(
			"Reset Invoice Number",
			"Are you sure you want to reset the invoice number to 0? This action cannot be undone.",
			func(confirm bool) {
				if confirm {
					_, err := s.database.Exec("UPDATE settings SET value = '0' WHERE key = 'last_invoice_number'")
					if err != nil {
						dialog.ShowError(fmt.Errorf("error resetting invoice number: %v", err), s.window)
						return
					}
					lastInvoiceLabel.SetText("Last Invoice Number: 0")
					dialog.ShowInformation("Success", "Invoice number has been reset", s.window)
				}
			},
			s.window,
		)
		confirmDialog.Show()
	})

	// Layout
	form := container.NewVBox(
		widget.NewCard("Store Information", "",
			container.NewVBox(
				widget.NewLabel("Store Name"),
				storeNameEntry,
				widget.NewLabel("Store Address"),
				storeAddressEntry,
				widget.NewLabel("Store Phone"),
				storePhoneEntry,
			),
		),
		widget.NewCard("Tax Settings", "",
			container.NewVBox(
				widget.NewLabel("Tax Percentage"),
				taxEntry,
			),
		),
		widget.NewCard("Invoice Settings", "",
			container.NewVBox(
				widget.NewLabel("Invoice Prefix"),
				invoicePrefixEntry,
				lastInvoiceLabel,
				resetInvoiceButton,
			),
		),
		widget.NewCard("Printer Settings", "",
			container.NewVBox(
				widget.NewLabel("Printer Name"),
				printerNameEntry,
				widget.NewLabel("Printer Port"),
				printerPortEntry,
				widget.NewLabel("Paper Width"),
				paperWidthEntry,
				// Add other printer-specific settings
			),
		),
		saveButton,
	)

	return container.NewBorder(header, nil, nil, nil, form)
}
