CREATE TABLE IF NOT EXISTS settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(50) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS invoices (
    id SERIAL PRIMARY KEY,
    sale_id INTEGER REFERENCES sales(id) ON DELETE CASCADE,
    invoice_number VARCHAR(20) UNIQUE NOT NULL,
    store_name VARCHAR(100) NOT NULL,
    store_address TEXT,
    store_phone VARCHAR(20),
    tax_percentage DECIMAL(5,2),
    tax_amount DECIMAL(10,2),
    subtotal DECIMAL(10,2),
    total_amount DECIMAL(10,2),
    payment_amount DECIMAL(10,2),
    change_amount DECIMAL(10,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert default settings
INSERT INTO settings (key, value) VALUES
    ('store_name', 'My Store'),
    ('store_address', 'Store Address'),
    ('store_phone', '123-456-789'),
    ('tax_percentage', '10'),
    ('invoice_prefix', 'INV'),
    ('last_invoice_number', '0');

-- Add printer settings
INSERT INTO settings (key, value) VALUES
    ('printer_name', ''),
    ('printer_port', ''),
    ('paper_width', '80'),
    ('print_mode', 'file'); -- 'file', 'thermal', 'network'
