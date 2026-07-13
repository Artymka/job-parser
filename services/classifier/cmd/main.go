package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artymka/jobparser/internal/kafka/consumer"
	"github.com/artymka/jobparser/services/classifier/internal/config"
	"github.com/artymka/jobparser/services/classifier/internal/openai"
)

func main() {
	intCh := make(chan os.Signal, 1)
	signal.Notify(intCh, syscall.SIGINT, syscall.SIGTERM)

	config, err := config.New(".env")
	if err != nil {
		panic(err)
	}

	consumer := consumer.NewConsumer(config.KafkaBroker)
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
