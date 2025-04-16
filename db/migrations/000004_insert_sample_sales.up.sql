-- First, insert some sales
INSERT INTO sales (created_at, total_amount) VALUES
    (CURRENT_TIMESTAMP, 45000),  -- Coffee (2x) + Tea
    (CURRENT_TIMESTAMP, 27000),  -- Tea (2x) + Milk
    (CURRENT_TIMESTAMP, 39000),  -- Coffee (2x) + Milk
    (CURRENT_TIMESTAMP, 30000);  -- Coffee + Tea (2x)

-- Then, insert the corresponding sale items
INSERT INTO sale_items (sale_id, product_id, quantity, price_at_sale, created_at) VALUES
    -- Sale 1: Coffee (2x) + Tea
    (1, (SELECT id FROM products WHERE name = 'Coffee'), 2, 15000, CURRENT_TIMESTAMP),
    (1, (SELECT id FROM products WHERE name = 'Tea'), 1, 10000, CURRENT_TIMESTAMP),

    -- Sale 2: Tea (2x) + Milk
    (2, (SELECT id FROM products WHERE name = 'Tea'), 2, 10000, CURRENT_TIMESTAMP),
    (2, (SELECT id FROM products WHERE name = 'Milk'), 1, 12000, CURRENT_TIMESTAMP),

    -- Sale 3: Coffee (2x) + Milk
    (3, (SELECT id FROM products WHERE name = 'Coffee'), 2, 15000, CURRENT_TIMESTAMP),
    (3, (SELECT id FROM products WHERE name = 'Milk'), 1, 12000, CURRENT_TIMESTAMP),

    -- Sale 4: Coffee + Tea (2x)
    (4, (SELECT id FROM products WHERE name = 'Coffee'), 1, 15000, CURRENT_TIMESTAMP),
    (4, (SELECT id FROM products WHERE name = 'Tea'), 2, 10000, CURRENT_TIMESTAMP);

-- Update product stock based on sales
UPDATE products
SET stock = CASE
    WHEN name = 'Coffee' THEN stock - 5  -- Total Coffee sold: 5
    WHEN name = 'Tea' THEN stock - 5     -- Total Tea sold: 5
    WHEN name = 'Milk' THEN stock - 2    -- Total Milk sold: 2
    ELSE stock
END;
