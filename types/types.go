package types

// Product represents a store item
type Product struct {
	ID    int
	Name  string
	Price float64
	Stock int
}

// CartItem represents an item in the shopping cart
type CartItem struct {
	Product  Product
	Quantity int
}
