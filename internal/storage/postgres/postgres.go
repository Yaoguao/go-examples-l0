package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"wb-examples-l0/internal/models"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(dsn string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	//	MIGRATION (сделаю пока тут, возможно сделаю отдельным инструментом)
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS orders(
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
        )`,

		`CREATE TABLE IF NOT EXISTS deliveries(
            order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
            name VARCHAR(255) NOT NULL,
            phone VARCHAR(50) NOT NULL,
            zip VARCHAR(50),
            city VARCHAR(255) NOT NULL,
            address TEXT NOT NULL,
            region VARCHAR(255),
            email VARCHAR(255)
        )`,

		`CREATE TABLE IF NOT EXISTS payments(
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
        )`,

		`CREATE TABLE IF NOT EXISTS items(
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
        )`,

		`CREATE INDEX IF NOT EXISTS idx_items_order_uid ON items(order_uid)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_track_number ON orders(track_number)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_date_created ON orders(date_created)`,
	}

	for i, migration := range migrations {
		_, err = db.Exec(migration)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to execute migration %d: %w", op, i+1, err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		db,
	}, nil
}

//	TODO - добавить контекст в запросах

func (s *Storage) SaveOrder(order *models.Order) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
        INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, 
                          customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	_, err = tx.Exec(`
        INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("insert delivery: %w", err)
	}

	_, err = tx.Exec(`
        INSERT INTO payments (order_uid, request_id, currency, provider, amount, 
                             payment_dt, bank, delivery_cost, goods_total, custom_fee)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `, order.OrderUID, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(`
            INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, 
                              sale, size, total_price, nm_id, brand, status)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        `, order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("insert item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *Storage) GetOrderByUID(orderUID string) (*models.Order, error) {
	var order models.Order

	err := s.db.QueryRow(`
        SELECT order_uid, track_number, entry, locale, internal_signature, 
               customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders WHERE order_uid = $1
    `, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
	)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	err = s.db.QueryRow(`
        SELECT name, phone, zip, city, address, region, email
        FROM deliveries WHERE order_uid = $1
    `, orderUID).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City,
		&order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("get delivery: %w", err)
	}

	err = s.db.QueryRow(`
        SELECT request_id, currency, provider, amount, payment_dt, 
               bank, delivery_cost, goods_total, custom_fee
        FROM payments WHERE order_uid = $1
    `, orderUID).Scan(
		&order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount,
		&order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
		&order.Payment.CustomFee,
	)
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}
	order.Payment.Transaction = orderUID

	rows, err := s.db.Query(`
        SELECT chrt_id, track_number, price, rid, name, sale, size, 
               total_price, nm_id, brand, status
        FROM items WHERE order_uid = $1
    `, orderUID)
	if err != nil {
		return nil, fmt.Errorf("get items: %w", err)
	}
	defer rows.Close()

	order.Items = make([]models.Item, 0)
	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size,
			&item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &order, nil
}

func (s *Storage) OrderExists(orderUID string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM orders WHERE order_uid = $1)
    `, orderUID).Scan(&exists)
	return exists, err
}

func (s *Storage) GetAllLimitOrderUIDs(limit int) ([]string, error) {
	var query string
	var args []interface{}

	if limit > 0 {
		query = "SELECT order_uid FROM orders ORDER BY date_created DESC LIMIT $1"
		args = []interface{}{limit}
	} else {
		query = "SELECT order_uid FROM orders ORDER BY date_created DESC"
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query order UIDs: %w", err)
	}
	defer rows.Close()

	var uids []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("scan order UID: %w", err)
		}
		uids = append(uids, uid)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return uids, nil
}

// TODO - попытаться оптимизировать запрос
//func (s *Storage) GetAllLimitOrder(limit int) ([]models.Order, error) {
//	uids, err := s.getAllLimitOrderUIDs(limit)
//	if err != nil {
//		return nil, fmt.Errorf("get order UIDs: %w", err)
//	}
//
//	var orders []models.Order
//	for _, uid := range uids {
//		order, err := s.GetOrderByUID(uid)
//		if err != nil {
//			log.Printf("error load order")
//			continue
//		}
//		orders = append(orders, *order)
//	}
//
//	return orders, nil
//}
