package main

import (
	"context"
	"fmt"
	"time"

	"github.com/artymka/jobparser-consumer/internal/kafka"
)

func main() {
	consumer := kafka.NewConsumer()
	defer consumer.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	fmt.Println("consumer started")

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
				fmt.Println("============ New message ============")
				fmt.Println(msg)
			}
		}()

		<-ticker.C
		<-waitCh
	}
}
