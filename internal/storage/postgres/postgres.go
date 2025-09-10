package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"wb-examples-l0/internal/config"
	"wb-examples-l0/internal/models"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(cfg *config.Config) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("postgres", cfg.Storage.Postgres.Dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	db.SetMaxOpenConns(cfg.Storage.Postgres.MaxOpenConns)

	db.SetMaxIdleConns(cfg.Storage.Postgres.MaxIdleConns)

	duration, err := time.ParseDuration(cfg.Storage.Postgres.MaxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

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
