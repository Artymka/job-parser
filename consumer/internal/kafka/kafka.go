package kafka

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer() *Consumer {
	// подключение к kafka
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		panic(fmt.Errorf("no KAFKA_BROKER"))
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		Topic:       "messages",
		Partition:   0,
		StartOffset: kafka.FirstOffset,
	})

	consumer := Consumer{
		reader: reader,
	}
	return &consumer
}

func (c *Consumer) Close() {
	c.reader.Close()
}

func (c *Consumer) Consume(ctx context.Context) (string, error) {
	// Создаем контекст с коротким таймаутом
	select {
	case <-ctx.Done():
		return "", nil
	default:
	}

	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return "", nil
		}
		if errors.Is(err, io.EOF) {
			return "", nil
		}
		// Если ошибка "no new messages", тоже выходим
		if strings.Contains(err.Error(), "no new messages") {
			return "", nil
		}

		// иначе возвращаем ошибку
		return "", err
	}

	return string(msg.Value), nil
}
