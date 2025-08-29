CREATE TABLE IF NOT EXISTS orders(
     order_uid VARCHAR(255) PRIMARY KEY,
     track_number VARCHAR(255),
     entry VARCHAR(50),
     locale VARCHAR(10),
     internal_signature VARCHAR(255),
     customer_id VARCHAR(255),
     delivery_service VARCHAR(100),
     shardkey VARCHAR(10),
     sm_id INT,
     date_created TIMESTAMPTZ,
     oof_shard VARCHAR(10)
)