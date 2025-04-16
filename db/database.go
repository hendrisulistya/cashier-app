package db

import (
	"database/sql"
	"fmt"

	"github.com/hendrisulistya/cashier-app/config"
	"github.com/hendrisulistya/cashier-app/types"
	_ "github.com/lib/pq"
)

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
	rows, err := db.Query("SELECT name, price, stock FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []types.Product
	for rows.Next() {
		var p types.Product
		err := rows.Scan(&p.Name, &p.Price, &p.Stock)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func SaveSale(db *sql.DB, cartItems []types.CartItem) error {
	tx, err := db.Begin()
	if err != nil {
		return err
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
		return err
	}

	// Insert sale items
	for _, item := range cartItems {
		_, err = tx.Exec(`
			INSERT INTO sale_items (sale_id, product_id, quantity, price_at_sale)
			SELECT $1, id, $2, $3
			FROM products WHERE name = $4`,
			saleID, item.Quantity, item.Product.Price, item.Product.Name)
		if err != nil {
			return err
		}

		// Update stock
		_, err = tx.Exec(`
			UPDATE products 
			SET stock = stock - $1
			WHERE name = $2`,
			item.Quantity, item.Product.Name)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
