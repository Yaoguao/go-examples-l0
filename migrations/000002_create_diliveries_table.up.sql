CREATE TABLE IF NOT EXISTS deliveries(
     order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
     name VARCHAR(255) NOT NULL,
     phone VARCHAR(50) NOT NULL,
     zip VARCHAR(50),
     city VARCHAR(255) NOT NULL,
     address TEXT NOT NULL,
     region VARCHAR(255),
     email VARCHAR(255)
)