package ui

import (
    "database/sql"
    "fmt"
    "log"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "github.com/hendrisulistya/cashier-app/db"
    "github.com/hendrisulistya/cashier-app/types"
)

type MainWindow struct {
    window    fyne.Window
    database  *sql.DB
    cartItems []types.CartItem
}

func NewMainWindow(window fyne.Window, database *sql.DB) *MainWindow {
    return &MainWindow{
        window:    window,
        database:  database,
        cartItems: make([]types.CartItem, 0),
    }
}

func (m *MainWindow) Load() error {
    products, err := db.GetProducts(m.database)
    if err != nil {
        return fmt.Errorf("could not fetch products: %v", err)
    }

    content := m.createMainContent(products)
    m.window.SetContent(content)
    return nil
}

func (m *MainWindow) createMainContent(products []types.Product) fyne.CanvasObject {
    // Create UI components
    header := widget.NewLabel("Cashier System")
    header.TextStyle = fyne.TextStyle{Bold: true}

    // Cart display
    cartDisplay := widget.NewMultiLineEntry()
    cartDisplay.Disable()
    totalLabel := widget.NewLabel("Total: Rp0.00")

    updateCart := func() {
        var total float64
        cartText := ""
        for _, item := range m.cartItems {
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
                for _, item := range m.cartItems {
                    if item.Product.Name == prod.Name {
                        currentQty = item.Quantity
                        break
                    }
                }

                if currentQty >= prod.Stock {
                    dialog := widget.NewLabel(fmt.Sprintf("Not enough stock for %s", prod.Name))
                    popup := widget.NewModalPopUp(dialog, m.window.Canvas())
                    popup.Show()
                    return
                }

                // Add item to cart
                found := false
                for i, item := range m.cartItems {
                    if item.Product.Name == prod.Name {
                        m.cartItems[i].Quantity++
                        found = true
                        break
                    }
                }
                if !found {
                    m.cartItems = append(m.cartItems, types.CartItem{
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
        m.cartItems = []types.CartItem{}
        updateCart()
    })

    checkoutButton := widget.NewButton("Checkout", func() {
        if len(m.cartItems) == 0 {
            return
        }

        err := db.SaveSale(m.database, m.cartItems)
        if err != nil {
            dialog := widget.NewLabel(fmt.Sprintf("Error processing sale: %v", err))
            popup := widget.NewModalPopUp(dialog, m.window.Canvas())
            popup.Show()
            return
        }

        m.cartItems = []types.CartItem{}
        updateCart()

        // Refresh product list with updated stock
        products, err := db.GetProducts(m.database)
        if err != nil {
            log.Printf("Could not refresh products: %v", err)
            return
        }
        productList.Objects = nil
        for _, prod := range products {
            p := prod
            button := widget.NewButton(
                fmt.Sprintf("%s - Rp%.2f (Stock: %d)", p.Name, p.Price, p.Stock),
                func() {
                    // ... (same button logic as above)
                },
            )
            productList.Add(button)
        }
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