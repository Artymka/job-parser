package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artymka/jobparser/internal/kafka/consumer"
	"github.com/artymka/jobparser/internal/kafka/producer"
	"github.com/artymka/jobparser/services/classifier/internal/config"
	"github.com/artymka/jobparser/services/classifier/internal/repository/postgres"
	"github.com/openai/openai-go/v3"
)

func main() {
	intCh := make(chan os.Signal, 1)
	signal.Notify(intCh, syscall.SIGINT, syscall.SIGTERM)

	config, err := config.New(".env")
	if err != nil {
		panic(err)
	}

	storage, err := postgres.NewStorage(config.PgConfig.DBName, config.PgConfig.User, config.PgConfig.Password)
	if err != nil {
		panic(err)
	}
	defer storage.Close()

	// получение тем и создание продюсеров для каждой темы
	themes, err := storage.GetThemes()
	if err != nil {
		panic(err)
	}
	producers := make(map[int]*producer.Producer)
	for _, theme := range themes {
		producers[theme.ID] = producer.NewProducer(config.KafkaBroker, fmt.Sprintf("theme_%d", theme.ID))
		defer producers[theme.ID].Close()
	}

	consumer := consumer.NewConsumer(config.KafkaBroker, "raw-messages")
	defer consumer.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	aiClient := openai.NewClient(config)

	fmt.Println("consumer started")

	go func() {
		for {
			waitCh := make(chan struct{})

			go func() {
				defer close(waitCh)
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()

				msg, err := consumer.Consume(ctx)
				if err != nil {
					panic(err)
				}

				if msg != "" {
					resp, err := aiClient.Request(msg)
					if err != nil {
						panic(err)
					}

					fmt.Printf("============ New message of class %s ============\n", resp)
					fmt.Println(msg)
				}
			}()

			<-ticker.C
			<-waitCh
		}
	}()

	<-intCh
}
