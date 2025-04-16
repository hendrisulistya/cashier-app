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
	"github.com/hendrisulistya/cashier-app/types"
)

type InventoryWindow struct {
	window   fyne.Window
	database *sql.DB
	list     *widget.List
	products []types.Product
}

func NewInventoryWindow(window fyne.Window, database *sql.DB) *InventoryWindow {
	return &InventoryWindow{
		window:   window,
		database: database,
	}
}

func (i *InventoryWindow) Load() error {
	var err error
	i.products, err = db.GetProducts(i.database)
	if err != nil {
		return fmt.Errorf("could not fetch products: %v", err)
	}

	content := i.createInventoryContent()
	i.window.SetContent(content)
	return nil
}

func (i *InventoryWindow) createInventoryContent() fyne.CanvasObject {
	// Back button
	backButton := widget.NewButton("Back to Menu", func() {
		mainWindow := NewMainWindow(i.window, i.database)
		if err := mainWindow.Load(); err != nil {
			log.Printf("Error returning to main menu: %v", err)
		}
	})

	// Header
	header := container.NewHBox(
		backButton,
		widget.NewLabel("Inventory Management"),
	)

	// Create product list
	i.list = widget.NewList(
		func() int { return len(i.products) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),                   // Product name
				widget.NewLabel(""),                   // Price
				widget.NewLabel(""),                   // Stock
				widget.NewButton("Edit", func() {}),   // Edit button placeholder
				widget.NewButton("Delete", func() {}), // Delete button placeholder
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			product := i.products[id]
			box := item.(*fyne.Container)

			// Update labels
			box.Objects[0].(*widget.Label).SetText(product.Name)
			box.Objects[1].(*widget.Label).SetText(fmt.Sprintf("Rp%.2f", product.Price))
			box.Objects[2].(*widget.Label).SetText(fmt.Sprintf("%d", product.Stock))

			// Update edit button
			box.Objects[3].(*widget.Button).OnTapped = func() {
				i.showEditDialog(product)
			}

			// Update delete button
			box.Objects[4].(*widget.Button).OnTapped = func() {
				i.showDeleteDialog(product)
			}
		},
	)

	// Add new product button
	addButton := widget.NewButton("Add New Product", func() {
		i.showAddDialog()
	})
	addButton.Importance = widget.HighImportance

	// Layout setup
	content := container.NewBorder(
		header,
		addButton,
		nil,
		nil,
		container.NewScroll(i.list),
	)

	return content
}

func (i *InventoryWindow) showAddDialog() {
	nameEntry := widget.NewEntry()
	priceEntry := widget.NewEntry()
	stockEntry := widget.NewEntry()

	items := []*widget.FormItem{
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Price", priceEntry),
		widget.NewFormItem("Stock", stockEntry),
	}

	dialog.ShowForm("Add New Product", "Add", "Cancel", items,
		func(confirm bool) {
			if !confirm {
				return
			}

			price, err := strconv.ParseFloat(priceEntry.Text, 64)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid price format"), i.window)
				return
			}

			stock, err := strconv.Atoi(stockEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid stock format"), i.window)
				return
			}

			product := types.Product{
				Name:  nameEntry.Text,
				Price: price,
				Stock: stock,
			}

			if err := db.AddProduct(i.database, product); err != nil {
				dialog.ShowError(fmt.Errorf("failed to add product: %v", err), i.window)
				return
			}

			// Refresh the product list
			i.refreshProducts()
		}, i.window)
}

func (i *InventoryWindow) showEditDialog(product types.Product) {
	nameEntry := widget.NewEntry()
	nameEntry.SetText(product.Name)

	priceEntry := widget.NewEntry()
	priceEntry.SetText(fmt.Sprintf("%.2f", product.Price))

	stockEntry := widget.NewEntry()
	stockEntry.SetText(fmt.Sprintf("%d", product.Stock))

	items := []*widget.FormItem{
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Price", priceEntry),
		widget.NewFormItem("Stock", stockEntry),
	}

	dialog.ShowForm("Edit Product", "Save", "Cancel", items,
		func(confirm bool) {
			if !confirm {
				return
			}

			price, err := strconv.ParseFloat(priceEntry.Text, 64)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid price format"), i.window)
				return
			}

			stock, err := strconv.Atoi(stockEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid stock format"), i.window)
				return
			}

			updatedProduct := types.Product{
				ID:    product.ID,
				Name:  nameEntry.Text,
				Price: price,
				Stock: stock,
			}

			if err := db.UpdateProduct(i.database, updatedProduct); err != nil {
				dialog.ShowError(fmt.Errorf("failed to update product: %v", err), i.window)
				return
			}

			// Refresh the product list
			i.refreshProducts()
		}, i.window)
}

func (i *InventoryWindow) showDeleteDialog(product types.Product) {
	dialog.ShowConfirm("Delete Product",
		fmt.Sprintf("Are you sure you want to delete %s?", product.Name),
		func(confirm bool) {
			if !confirm {
				return
			}

			if err := db.DeleteProduct(i.database, product.ID); err != nil {
				dialog.ShowError(fmt.Errorf("failed to delete product: %v", err), i.window)
				return
			}

			// Refresh the product list
			i.refreshProducts()
		}, i.window)
}

func (i *InventoryWindow) refreshProducts() {
	var err error
	i.products, err = db.GetProducts(i.database)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to refresh products: %v", err), i.window)
		return
	}
	i.list.Refresh()
}
