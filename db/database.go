package db

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/hendrisulistya/cashier-app/config"
	"github.com/hendrisulistya/cashier-app/types"
	_ "github.com/lib/pq"
)

type Settings struct {
	StoreName     string
	StoreAddress  string
	StorePhone    string
	TaxPercentage float64
	InvoicePrefix string
}

func NewConnection(config *config.DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GetProducts(db *sql.DB) ([]types.Product, error) {
	rows, err := db.Query("SELECT id, name, price, stock FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []types.Product
	for rows.Next() {
		var p types.Product
		err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func SaveSale(db *sql.DB, cartItems []types.CartItem) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Calculate total
	var total float64
	for _, item := range cartItems {
		total += item.Product.Price * float64(item.Quantity)
	}

	// Insert sale
	var saleID int
	err = tx.QueryRow("INSERT INTO sales (total_amount) VALUES ($1) RETURNING id", total).Scan(&saleID)
	if err != nil {
		return 0, err
	}

	// Insert sale items
	for _, item := range cartItems {
		_, err = tx.Exec(`
			INSERT INTO sale_items (sale_id, product_id, quantity, price_at_sale)
			SELECT $1, id, $2, $3
			FROM products WHERE name = $4`,
			saleID, item.Quantity, item.Product.Price, item.Product.Name)
		if err != nil {
			return 0, err
		}

		// Update stock
		_, err = tx.Exec(`
			UPDATE products
			SET stock = stock - $1
			WHERE name = $2`,
			item.Quantity, item.Product.Name)
		if err != nil {
			return 0, err
		}
	}

	return saleID, tx.Commit()
}

func AddProduct(db *sql.DB, product types.Product) error {
	_, err := db.Exec(`
		INSERT INTO products (name, price, stock)
		VALUES ($1, $2, $3)`,
		product.Name, product.Price, product.Stock)
	return err
}

func UpdateProduct(db *sql.DB, product types.Product) error {
	_, err := db.Exec(`
		UPDATE products
		SET name = $1, price = $2, stock = $3
		WHERE id = $4`,
		product.Name, product.Price, product.Stock, product.ID)
	return err
}

func DeleteProduct(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM products WHERE id = $1", id)
	return err
}

func GetSettings(db *sql.DB) (Settings, error) {
	settings := Settings{}
	rows, err := db.Query("SELECT key, value FROM settings")
	if err != nil {
		return settings, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return settings, err
		}

		switch key {
		case "store_name":
			settings.StoreName = value
		case "store_address":
			settings.StoreAddress = value
		case "store_phone":
			settings.StorePhone = value
		case "tax_percentage":
			settings.TaxPercentage, _ = strconv.ParseFloat(value, 64)
		case "invoice_prefix":
			settings.InvoicePrefix = value
		}
	}
	return settings, nil
}

func GenerateInvoiceNumber(db *sql.DB) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var prefix string
	var lastNum int
	err = tx.QueryRow("SELECT value FROM settings WHERE key = 'invoice_prefix'").Scan(&prefix)
	if err != nil {
		return "", err
	}

	err = tx.QueryRow("SELECT CAST(value AS INTEGER) FROM settings WHERE key = 'last_invoice_number'").Scan(&lastNum)
	if err != nil {
		return "", err
	}

	newNum := lastNum + 1
	_, err = tx.Exec("UPDATE settings SET value = $1 WHERE key = 'last_invoice_number'", strconv.Itoa(newNum))
	if err != nil {
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%06d", prefix, newNum), nil
}

func SaveInvoice(db *sql.DB, saleID int, invoiceNumber string, payment float64, change float64) error {
	settings, err := GetSettings(db)
	if err != nil {
		return err
	}

	var subtotal float64
	err = db.QueryRow("SELECT total_amount FROM sales WHERE id = $1", saleID).Scan(&subtotal)
	if err != nil {
		return err
	}

	taxAmount := subtotal * (settings.TaxPercentage / 100)
	totalAmount := subtotal + taxAmount

	_, err = db.Exec(`
		INSERT INTO invoices (
			sale_id, invoice_number, store_name, store_address, store_phone,
			tax_percentage, tax_amount, subtotal, total_amount, payment_amount, change_amount
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		saleID, invoiceNumber, settings.StoreName, settings.StoreAddress, settings.StorePhone,
		settings.TaxPercentage, taxAmount, subtotal, totalAmount, payment, change)

	return err
}

func UpdateSettings(db *sql.DB, settings Settings) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updates := map[string]string{
		"store_name":     settings.StoreName,
		"store_address":  settings.StoreAddress,
		"store_phone":    settings.StorePhone,
		"tax_percentage": fmt.Sprintf("%.2f", settings.TaxPercentage),
	}

	for key, value := range updates {
		_, err := tx.Exec("UPDATE settings SET value = $1, updated_at = CURRENT_TIMESTAMP WHERE key = $2",
			value, key)
		if err != nil {
			return fmt.Errorf("error updating %s: %v", key, err)
		}
	}

	return tx.Commit()
}

func ResetInvoiceNumber(db *sql.DB) error {
	_, err := db.Exec("UPDATE settings SET value = '0', updated_at = CURRENT_TIMESTAMP WHERE key = 'last_invoice_number'")
	return err
}
