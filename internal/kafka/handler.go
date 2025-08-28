package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log/slog"
	"wb-examples-l0/internal/models"
	"wb-examples-l0/internal/validator"
)

type OrderSaver interface {
	SaveOrder(order *models.Order) error
}

type OrderHandler struct {
	log        *slog.Logger
	orderSaver OrderSaver
}

func NewOrderHandler(logger *slog.Logger, orderSaver OrderSaver) *OrderHandler {
	return &OrderHandler{
		log:        logger,
		orderSaver: orderSaver,
	}
}

func (h *OrderHandler) HandleMessage(message []byte, offset kafka.Offset) error {
	var order models.Order

	if err := json.Unmarshal(message, &order); err != nil {
		h.log.Error("json unmarshal failed", "error", err, "offset", offset)
		return fmt.Errorf("json unmarshal failed: %w", err)
	}

	v := validator.New()
	models.ValidateOrder(v, &order)
	if !v.Valid() {
		h.log.Error("order validation failed", "errors", v.Errors, "order_uid", order.OrderUID)
		return fmt.Errorf("order validation failed: %v", v.Errors)
	}

	if err := h.orderSaver.SaveOrder(&order); err != nil {
		h.log.Error("failed to save order", "error", err, "order_uid", order.OrderUID)
		return fmt.Errorf("failed to save order: %w", err)
	}

	h.log.Info("order processed successfully",
		"order_uid", order.OrderUID,
		"offset", offset,
		"items_count", len(order.Items))

	return nil
}
