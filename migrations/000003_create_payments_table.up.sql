CREATE TABLE IF NOT EXISTS payments(
   order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
   request_id VARCHAR(255),
   currency VARCHAR(10) NOT NULL,
   provider VARCHAR(100) NOT NULL,
   amount INT NOT NULL,
   payment_dt BIGINT NOT NULL,
   bank VARCHAR(100),
   delivery_cost INT,
   goods_total INT NOT NULL,
   custom_fee INT
)