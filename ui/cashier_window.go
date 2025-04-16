package ui

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hendrisulistya/cashier-app/db"
	"github.com/hendrisulistya/cashier-app/types"
)

type CashierWindow struct {
	window    fyne.Window
	database  *sql.DB
	cartItems []types.CartItem
}

func NewCashierWindow(window fyne.Window, database *sql.DB) *CashierWindow {
	return &CashierWindow{
		window:    window,
		database:  database,
		cartItems: make([]types.CartItem, 0),
	}
}

func (c *CashierWindow) Load() error {
	products, err := db.GetProducts(c.database)
	if err != nil {
		return fmt.Errorf("could not fetch products: %v", err)
	}

	content := c.createCashierContent(products)
	c.window.SetContent(content)
	return nil
}

func (c *CashierWindow) createCashierContent(products []types.Product) fyne.CanvasObject {
	// Back button
	backButton := widget.NewButton("Back to Menu", func() {
		mainWindow := NewMainWindow(c.window, c.database)
		if err := mainWindow.Load(); err != nil {
			log.Printf("Error returning to main menu: %v", err)
		}
	})

	// Header
	header := container.NewHBox(
		backButton,
		widget.NewLabel("Cashier System"),
	)

	// Cart display
	cartDisplay := widget.NewMultiLineEntry()
	cartDisplay.Disable()
	totalLabel := widget.NewLabel("Total: Rp0.00")

	updateCart := func() {
		var total float64
		cartText := ""
		for _, item := range c.cartItems {
			subtotal := item.Product.Price * float64(item.Quantity)
			cartText += fmt.Sprintf("%s x%d: Rp%.2f\n",
				item.Product.Name, item.Quantity, subtotal)
			total += subtotal
		}
		cartDisplay.SetText(cartText)
		totalLabel.SetText(fmt.Sprintf("Total: Rp%.2f", total))
	}

	// Product list
	productList := container.NewVBox()
	for _, product := range products {
		prod := product // Create a new variable to avoid closure issues
		button := widget.NewButton(
			fmt.Sprintf("%s - Rp%.2f (Stock: %d)", prod.Name, prod.Price, prod.Stock),
			func() {
				// Check stock before adding
				currentQty := 0
				for _, item := range c.cartItems {
					if item.Product.Name == prod.Name {
						currentQty = item.Quantity
						break
					}
				}

				if currentQty >= prod.Stock {
					dialog := widget.NewLabel(fmt.Sprintf("Not enough stock for %s", prod.Name))
					popup := widget.NewModalPopUp(dialog, c.window.Canvas())
					popup.Show()
					return
				}

				// Add item to cart
				found := false
				for i, item := range c.cartItems {
					if item.Product.Name == prod.Name {
						c.cartItems[i].Quantity++
						found = true
						break
					}
				}
				if !found {
					c.cartItems = append(c.cartItems, types.CartItem{
						Product:  prod,
						Quantity: 1,
					})
				}
				updateCart()
			},
		)
		productList.Add(button)
	}

	// Cart buttons
	clearButton := widget.NewButton("Clear Cart", func() {
		c.cartItems = []types.CartItem{}
		updateCart()
	})

	checkoutButton := widget.NewButton("Checkout", func() {
		if len(c.cartItems) == 0 {
			return
		}

		var subtotal float64
		for _, item := range c.cartItems {
			subtotal += item.Product.Price * float64(item.Quantity)
		}

		settings, err := db.GetSettings(c.database)
		if err != nil {
			dialog.ShowError(fmt.Errorf("error loading settings: %v", err), c.window)
			return
		}

		taxAmount := subtotal * (settings.TaxPercentage / 100)
		total := subtotal + taxAmount

		c.showCheckoutDialog(subtotal, taxAmount, total)
	})

	// Layout setup
	productSection := container.NewVBox(
		widget.NewLabel("Products"),
		productList,
	)

	cartSection := container.NewVBox(
		widget.NewLabel("Shopping Cart"),
		cartDisplay,
		totalLabel,
		container.NewHBox(clearButton, checkoutButton),
	)

	// Main content split
	split := container.NewHSplit(productSection, cartSection)
	split.SetOffset(0.5)

	// Main layout
	content := container.NewBorder(header, nil, nil, nil, split)

	return content
}

func (c *CashierWindow) generateInvoice(invoiceNumber string, payment, change float64) string {
	settings, err := db.GetSettings(c.database)
	if err != nil {
		log.Printf("Error getting settings: %v", err)
		return ""
	}

	var subtotal float64
	invoice := "\n=================================\n"
	invoice += fmt.Sprintf("           %s          \n", settings.StoreName)
	invoice += "=================================\n"
	invoice += fmt.Sprintf("Invoice: %s\n", invoiceNumber)
	invoice += fmt.Sprintf("Date: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	invoice += fmt.Sprintf("Address: %s\n", settings.StoreAddress)
	invoice += fmt.Sprintf("Phone: %s\n", settings.StorePhone)
	invoice += "---------------------------------\n"
	invoice += "Items:\n"

	for _, item := range c.cartItems {
		itemTotal := item.Product.Price * float64(item.Quantity)
		invoice += fmt.Sprintf("%-20s x%d\n", item.Product.Name, item.Quantity)
		invoice += fmt.Sprintf("    @Rp%-14.2f Rp%.2f\n", item.Product.Price, itemTotal)
		subtotal += itemTotal
	}

	taxAmount := subtotal * (settings.TaxPercentage / 100)
	total := subtotal + taxAmount

	invoice += "---------------------------------\n"
	invoice += fmt.Sprintf("Subtotal:       Rp%.2f\n", subtotal)
	invoice += fmt.Sprintf("Tax (%.1f%%):     Rp%.2f\n", settings.TaxPercentage, taxAmount)
	invoice += fmt.Sprintf("Total:          Rp%.2f\n", total)
	invoice += fmt.Sprintf("Payment:        Rp%.2f\n", payment)
	invoice += fmt.Sprintf("Change:         Rp%.2f\n", change)
	invoice += "=================================\n"
	invoice += "          Thank You!             \n"
	invoice += "=================================\n"

	return invoice
}

func (c *CashierWindow) printInvoice(invoice string) {
	// TODO: Add printer configuration from settings
	// Example structure:
	// printerConfig, err := db.GetPrinterSettings(c.database)
	// if err != nil {
	//     dialog.ShowError(fmt.Errorf("failed to get printer settings: %v", err), c.window)
	//     return
	// }

	// TODO: Implement actual printer integration
	// Common options include:
	// 1. ESC/POS for thermal receipt printers
	// 2. CUPS for Unix-like systems
	// 3. Windows Printer API for Windows systems
	// Example:
	// err := printer.Print(printerConfig, invoice)

	// Temporary solution: save to file
	filename := fmt.Sprintf("invoice_%s.txt", time.Now().Format("20060102150405"))
	err := os.WriteFile(filename, []byte(invoice), 0644)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to save invoice: %v", err), c.window)
		return
	}

	dialog.ShowInformation("Success",
		fmt.Sprintf("Invoice saved to %s", filename),
		c.window,
	)
}

func (c *CashierWindow) showCheckoutDialog(subtotal, taxAmount, total float64) {
	// Get theme colors
	bgColor := theme.BackgroundColor()
	textColor := theme.ForegroundColor()

	// Create styled labels with theme-aware text
	titleStyle := fyne.TextStyle{Bold: true}
	amountStyle := fyne.TextStyle{Bold: true, Monospace: true}

	// Create styled text with theme colors
	createThemedLabel := func(text string, align fyne.TextAlign, style fyne.TextStyle) *canvas.Text {
		label := canvas.NewText(text, textColor)
		label.Alignment = align
		label.TextStyle = style
		return label
	}

	// Summary section with themed background
	summaryBg := canvas.NewRectangle(bgColor)
	summaryContent := container.NewVBox(
		container.NewGridWithColumns(2,
			createThemedLabel("Subtotal:", fyne.TextAlignLeading, titleStyle),
			createThemedLabel(fmt.Sprintf("Rp%.2f", subtotal), fyne.TextAlignTrailing, amountStyle),
		),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			createThemedLabel("Tax Amount:", fyne.TextAlignLeading, titleStyle),
			createThemedLabel(fmt.Sprintf("Rp%.2f", taxAmount), fyne.TextAlignTrailing, amountStyle),
		),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			createThemedLabel("Total Amount:", fyne.TextAlignLeading, titleStyle),
			createThemedLabel(fmt.Sprintf("Rp%.2f", total), fyne.TextAlignTrailing, amountStyle),
		),
	)

	summaryCard := container.NewMax(
		summaryBg,
		container.NewPadded(
			container.NewVBox(
				createThemedLabel("Transaction Summary", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				summaryContent,
			),
		),
	)

	// Payment section
	paymentEntry := widget.NewEntry()
	paymentEntry.SetPlaceHolder("Enter payment amount")

	changeLabel := createThemedLabel("Change: Rp0.00", fyne.TextAlignTrailing, amountStyle)

	// Calculate change in real-time
	paymentEntry.OnChanged = func(value string) {
		payment, err := strconv.ParseFloat(value, 64)
		if err != nil {
			changeLabel.Text = "Change: Invalid amount"
			changeLabel.Refresh()
			return
		}
		change := payment - total
		if change < 0 {
			changeLabel.Text = "Insufficient payment"
			changeLabel.Color = theme.ErrorColor()
			changeLabel.Refresh()
			return
		}
		changeLabel.Text = fmt.Sprintf("Change: Rp%.2f", change)
		changeLabel.Color = textColor
		changeLabel.Refresh()
	}

	paymentBg := canvas.NewRectangle(bgColor)
	paymentContent := container.NewVBox(
		createThemedLabel("Payment Details", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		createThemedLabel("Payment Amount:", fyne.TextAlignLeading, titleStyle),
		paymentEntry,
		widget.NewSeparator(),
		changeLabel,
	)

	paymentCard := container.NewMax(
		paymentBg,
		container.NewPadded(paymentContent),
	)

	// Cart items summary with theme colors
	cartBg := canvas.NewRectangle(bgColor)
	cartText := "Items:\n"
	for _, item := range c.cartItems {
		cartText += fmt.Sprintf("- %s x%d (Rp%.2f)\n",
			item.Product.Name,
			item.Quantity,
			item.Product.Price*float64(item.Quantity))
	}

	cartContent := container.NewVBox(
		createThemedLabel("Cart Summary", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewTextGridFromString(cartText),
	)

	cartCard := container.NewMax(
		cartBg,
		container.NewPadded(cartContent),
	)

	// Buttons with theme-aware styling
	processBtn := widget.NewButton("Process Payment", func() {
		payment, err := strconv.ParseFloat(paymentEntry.Text, 64)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid payment amount"), c.window)
			return
		}

		if payment < total {
			dialog.ShowError(fmt.Errorf("insufficient payment"), c.window)
			return
		}

		c.processTransaction(payment, total)
	})
	processBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {})
	cancelBtn.Importance = widget.DangerImportance

	buttons := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		processBtn,
	)

	// Main content with theme-aware background
	mainBg := canvas.NewRectangle(bgColor)
	content := container.NewVBox(
		summaryCard,
		widget.NewSeparator(),
		cartCard,
		widget.NewSeparator(),
		paymentCard,
		widget.NewSeparator(),
		buttons,
	)

	mainContainer := container.NewMax(
		mainBg,
		container.NewPadded(content),
	)

	// Show dialog with proper size
	dialog := dialog.NewCustom("Checkout", "",
		mainContainer,
		c.window,
	)

	// Set a minimum size for the dialog
	dialog.Resize(fyne.NewSize(400, 600))
	dialog.Show()
}

func (c *CashierWindow) processTransaction(payment, total float64) {
	change := payment - total

	// Process the sale
	saleID, err := db.SaveSale(c.database, c.cartItems)
	if err != nil {
		dialog.ShowError(fmt.Errorf("error processing sale: %v", err), c.window)
		return
	}

	// Generate invoice number
	invoiceNumber, err := db.GenerateInvoiceNumber(c.database)
	if err != nil {
		dialog.ShowError(fmt.Errorf("error generating invoice number: %v", err), c.window)
		return
	}

	// Save invoice
	err = db.SaveInvoice(c.database, saleID, invoiceNumber, payment, change)
	if err != nil {
		dialog.ShowError(fmt.Errorf("error saving invoice: %v", err), c.window)
		return
	}

	// Generate and show invoice
	invoice := c.generateInvoice(invoiceNumber, payment, change)

	// Show invoice dialog with print option
	printBtn := widget.NewButton("Print Invoice", func() {
		c.printInvoice(invoice)
	})

	// Get theme colors
	bgColor := theme.BackgroundColor()
	textColor := theme.ForegroundColor()

	// Create themed invoice display
	invoiceDisplay := widget.NewTextGridFromString(invoice)
	invoiceDisplay.SetStyleRange(0, 0, len(invoice), 0,
		&widget.CustomTextGridStyle{
			TextStyle: fyne.TextStyle{Monospace: true},
			FGColor:   textColor,
		},
	)

	invoiceContainer := container.NewMax(
		canvas.NewRectangle(bgColor),
		container.NewPadded(
			container.NewVBox(
				invoiceDisplay,
				printBtn,
			),
		),
	)

	dialog := dialog.NewCustom("Invoice", "Close",
		invoiceContainer,
		c.window,
	)

	// Set a minimum size for better readability
	dialog.Resize(fyne.NewSize(400, 600))
	dialog.Show()

	// Clear cart and refresh
	c.cartItems = []types.CartItem{}
	c.Load()
}
