-- Remove all sales data
TRUNCATE TABLE sale_items CASCADE;
TRUNCATE TABLE sales CASCADE;

-- Restore product stock to initial values
UPDATE products
SET stock = CASE
    WHEN name = 'Coffee' THEN 50
    WHEN name = 'Tea' THEN 50
    WHEN name = 'Milk' THEN 30
    ELSE stock
END;
