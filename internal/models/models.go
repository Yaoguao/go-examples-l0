package models

import (
	"errors"
	"fmt"
	"time"
	"wb-examples-l0/internal/validator"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Order struct {
	OrderUID          string    `json:"order_uid"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery"`
	Payment           Payment   `json:"payment"`
	Items             []Item    `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

type Delivery struct {
	OrderUID string `json:"-"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Zip      string `json:"zip"`
	City     string `json:"city"`
	Address  string `json:"address"`
	Region   string `json:"region"`
	Email    string `json:"email"`
}

type Payment struct {
	OrderUID     string `json:"-"`
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ID          int    `json:"-"`
	OrderUID    string `json:"-"`
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func ValidateOrder(v *validator.Validator, order *Order) {
	v.Check(order.OrderUID != "", "order_uid", "must be provided")

	v.Check(order.TrackNumber != "", "track_number", "must be provided")

	v.Check(order.Entry != "", "entry", "must be provided")

	//	???
	// v.Check(validatorp.PermittedValue(order.Locale, "en", "ru", "es", "fr", "de", "it"), "locale", "must be a valid locale")

	v.Check(len(order.InternalSignature) <= 255, "internal_signature", "must not be more than 255 bytes long")

	v.Check(order.CustomerID != "", "customer_id", "must be provided")

	v.Check(order.DeliveryService != "", "delivery_service", "must be provided")

	v.Check(order.Shardkey != "", "shardkey", "must be provided")

	v.Check(order.SmID >= 0, "sm_id", "must be non-negative")

	v.Check(!order.DateCreated.IsZero(), "date_created", "must be provided")

	v.Check(order.OofShard != "", "oof_shard", "must be provided")

	ValidateDelivery(v, &order.Delivery)

	ValidatePayment(v, &order.Payment, order.OrderUID)

	v.Check(len(order.Items) > 0, "items", "must contain at least one item")
	for i, item := range order.Items {
		ValidateItem(v, &item, i)
	}
}

func ValidateDelivery(v *validator.Validator, delivery *Delivery) {
	v.Check(delivery.Name != "", "delivery.name", "must be provided")

	v.Check(delivery.Phone != "", "delivery.phone", "must be provided")

	v.Check(delivery.Zip != "", "delivery.zip", "must be provided")

	v.Check(delivery.City != "", "delivery.city", "must be provided")

	v.Check(delivery.Address != "", "delivery.address", "must be provided")

	v.Check(delivery.Region != "", "delivery.region", "must be provided")

	v.Check(delivery.Email != "", "delivery.email", "must be provided")
	v.Check(validator.Matches(delivery.Email, validator.EmailRX), "delivery.email", "must be a valid email address")
}

func ValidatePayment(v *validator.Validator, payment *Payment, orderUID string) {
	v.Check(payment.Transaction == orderUID, "payment.transaction", "must match order_uid")

	v.Check(payment.Currency != "", "payment.currency", "must be provided")

	v.Check(payment.Provider != "", "payment.provider", "must be provided")

	v.Check(payment.Amount > 0, "payment.amount", "must be positive")

	v.Check(payment.Bank != "", "payment.bank", "must be provided")

	v.Check(payment.DeliveryCost >= 0, "payment.delivery_cost", "must be non-negative")
	v.Check(payment.GoodsTotal >= 0, "payment.goods_total", "must be non-negative")
	v.Check(payment.CustomFee >= 0, "payment.custom_fee", "must be non-negative")

	//v.Check(payment.Amount == payment.DeliveryCost+payment.GoodsTotal-payment.CustomFee,
	//	"payment.amount", "must equal delivery_cost + goods_total - custom_fee")
}

func ValidateItem(v *validator.Validator, item *Item, index int) {
	prefix := func(field string) string {
		return fmt.Sprintf("items[%d].%s", index, field)
	}

	v.Check(item.ChrtID > 0, prefix("chrt_id"), "must be positive")

	v.Check(item.TrackNumber != "", prefix("track_number"), "must be provided")

	v.Check(item.Price >= 0, prefix("price"), "must be non-negative")

	v.Check(item.Rid != "", prefix("rid"), "must be provided")

	v.Check(item.Name != "", prefix("name"), "must be provided")

	v.Check(item.Size != "", prefix("size"), "must be provided")

	v.Check(item.TotalPrice >= 0, prefix("total_price"), "must be non-negative")

	v.Check(item.NmID > 0, prefix("nm_id"), "must be positive")

	v.Check(item.Brand != "", prefix("brand"), "must be provided")

	v.Check(item.Status >= 0, prefix("status"), "must be non-negative")

	//if item.Sale > 0 {
	//	expectedPrice := item.Price * (100 - item.Sale) / 100
	//	v.Check(item.TotalPrice == expectedPrice, prefix("total_price"),
	//		fmt.Sprintf("must equal price * (100 - sale) / 100 = %d", expectedPrice))
	//} else {
	//	v.Check(item.TotalPrice == item.Price, prefix("total_price"), "must equal price when sale is 0")
	//}
}
