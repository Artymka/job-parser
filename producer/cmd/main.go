package main

import (
	"fmt"
	"time"

	"github.com/artymka/jobparser-producer/internal/config"
	"github.com/artymka/jobparser-producer/internal/kafka"
	"github.com/artymka/jobparser-producer/internal/telegram"
)

func main() {
	config, err := config.New(".env")
	if err != nil {
		panic(err)
	}

	producer := kafka.NewProducer()
	defer producer.Close()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	fmt.Println("producer started")

	for {
		<-ticker.C
		messages := telegram.GetLastMessages(config)
		fmt.Printf("First message of slice: %s\n", messages[0])
		producer.ProduceN(messages)
	}
}
