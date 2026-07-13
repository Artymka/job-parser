package main

import (
	"fmt"
	"time"

	"github.com/artymka/jobparser/internal/kafka/producer"
	"github.com/artymka/jobparser/internal/telegram"
	"github.com/artymka/jobparser/services/telegram-scrapper/internal/config"
)

func main() {
	config, err := config.New(".env")
	if err != nil {
		panic(err)
	}
	tgConfig := telegram.TelegramConfig{
		Phone:       config.Phone,
		AppID:       config.AppID,
		AppHash:     config.AppHash,
		ProxyAddr:   config.ProxyAddr,
		ProxySecret: config.ProxySecret,
		SessionPath: config.SessionPath,
	}

	producer := producer.NewProducer(config.KafkaBroker)
	defer producer.Close()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	fmt.Println("producer started")

	for {
		<-ticker.C
		messages := telegram.GetLastMessages(&tgConfig)
		fmt.Printf("First message of slice: %s\n", messages[0])
		producer.ProduceN(messages)
	}
}
