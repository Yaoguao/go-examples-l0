package kafka

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log"
	"strings"
)

const (
	sessionTimeOut = 7000 // ms
	noTimeout      = -1
)

type MessageHandler interface {
	HandleMessage(message []byte, offset kafka.Offset) error
}

type Consumer struct {
	consumer       *kafka.Consumer
	handler        MessageHandler
	stop           bool
	consumerNumber int
}

func NewConsumer(address []string, topic, consumerGroup string, handler MessageHandler) (*Consumer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(address, ","),
		"group.id":           consumerGroup,
		"session.timeout.ms": sessionTimeOut,
		"auto.offset.reset":  "earliest",
	}
	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, fmt.Errorf("error with new consumer: %w", err)
	}
	if err = c.Subscribe(topic, nil); err != nil {
		return nil, err
	}
	return &Consumer{
		consumer: c,
		handler:  handler,
		stop:     false,
	}, nil
}

func (c *Consumer) Start() {
	for !c.stop {
		kafkaMsg, err := c.consumer.ReadMessage(noTimeout)
		if err != nil {
			log.Printf("error reading message: %v", err)
			continue
		}

		if kafkaMsg == nil {
			continue
		}

		if err := c.handler.HandleMessage(kafkaMsg.Value, kafkaMsg.TopicPartition.Offset); err != nil {
			log.Printf("handler error: %v", err)

			continue
		}
	}
}

func (c *Consumer) Stop() error {
	c.stop = true
	if _, err := c.consumer.Commit(); err != nil {
		return err
	}
	log.Print("Commited offset")
	return c.consumer.Close()
}
