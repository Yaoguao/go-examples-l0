package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
	"wb-examples-l0/internal/config"
	"wb-examples-l0/internal/kafka"
	"wb-examples-l0/internal/lib/logger/sl"
	"wb-examples-l0/internal/models"
)

func main() {
	cfg := config.MustLoad()

	log := sl.InitLogger(cfg.Env, os.Stdout)

	log.Debug("config", cfg)

	producer, err := kafka.NewProducer(cfg.Kafka.Addresses)
	if err != nil {
		log.Error("Failed to create producer", err)
		return
	}
	defer producer.Close()

	for i := 0; ; i++ {
		message, err := generateTestOrderWithTimestamp()
		if err != nil {
			log.Error("error gen", err)
			return
		}
		producer.Produce(message, cfg.Kafka.Consumer.OrderTopic, "0", time.Now())

		log.Info("Message sent",
			"message_number", i,
			"topic", cfg.Kafka.Consumer.OrderTopic,
		)

		time.Sleep(5 * time.Second)
	}

}
func generateTestOrderWithTimestamp() (string, error) {
	randomSuffix := fmt.Sprintf("%06d", rand.Intn(1000000))
	orderUID := fmt.Sprintf("b563feb7b2b84b6%s", randomSuffix)

	order := models.Order{
		OrderUID:    orderUID,
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: models.Payment{
			Transaction:  orderUID,
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      9934930 + rand.Intn(1000),
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         fmt.Sprintf("ab4219087a764ae0b%s", randomSuffix),
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}

	jsonData, err := json.Marshal(order)
	if err != nil {
		return "", fmt.Errorf("failed to marshal order: %w", err)
	}

	return string(jsonData), nil
}
