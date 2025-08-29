CREATE TABLE IF NOT EXISTS items(
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id BIGINT NOT NULL,
    track_number VARCHAR(255),
    price INT NOT NULL,
    rid VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    sale INT,
    size VARCHAR(50),
    total_price INT NOT NULL,
    nm_id BIGINT,
    brand VARCHAR(255),
    status INT NOT NULL
)