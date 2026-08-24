package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/artymka/jobparser/internal/kafka/producer"
	"github.com/artymka/jobparser/internal/telegram/client"
	"github.com/artymka/jobparser/services/telegram-scrapper/internal/config"
	"github.com/artymka/jobparser/services/telegram-scrapper/internal/lib"
	"github.com/artymka/jobparser/services/telegram-scrapper/internal/models"
	"github.com/artymka/jobparser/services/telegram-scrapper/internal/repository/postgres"
)

type ChannelsRepo interface {
	GetChannels() ([]models.Channel, error)
	SaveLastMessageIDs(channels []models.Channel) error
}

func main() {
	// config, err := config.New(".env")
	// if err != nil {
	// 	panic(err)
	// }
	// tgConfig := telegram.TelegramConfig{
	// 	Phone:       config.Phone,
	// 	AppID:       config.AppID,
	// 	AppHash:     config.AppHash,
	// 	ProxyAddr:   config.ProxyAddr,
	// 	ProxySecret: config.ProxySecret,
	// 	SessionPath: config.SessionPath,
	// }

	// producer := producer.NewProducer(config.KafkaBroker)
	// defer producer.Close()

	// ticker := time.NewTicker(time.Minute)
	// defer ticker.Stop()

	// fmt.Println("producer started")

	// for {
	// 	<-ticker.C
	// 	messages := telegram.GetLastMessages(&tgConfig)
	// 	fmt.Printf("First message of slice: %s\n", messages[0])
	// 	producer.ProduceN(messages)
	// }

	// config
	config, err := config.New(".env")
	if err != nil {
		panic(err)
	}

	// telegram user api
	tc, err := client.NewClient(&config.TgConfig)
	if err != nil {
		panic(err)
	}

	// postgres
	storage, err := postgres.NewStorage(config.PgConfig.DBName, config.PgConfig.User, config.PgConfig.Password)
	if err != nil {
		panic(err)
	}
	defer storage.Close()

	// kafka
	kafkaProd := producer.NewProducer(config.KafkaBroker)
	defer kafkaProd.Close()

	// получение каналов
	var repo ChannelsRepo = storage
	channels, err := repo.GetChannels()
	if err != nil {
		panic(err)
	}

	if err := tc.Run(func(c *client.Client) error {
		for {
			for channelI, channel := range channels {
				messages, err := lib.ScrapChannel(c, channel)
				if err != nil {
					fmt.Printf("error while scrapping channel: %v\n", err)
					continue
				}

				lastID := messages[len(messages)-1].ID
				convertedMsgs := make([][]byte, len(messages))
				for i, msg := range messages {
					convertedMsgs[i], err = json.Marshal(msg)
					if err != nil {
						fmt.Printf("error while marshalling message: %v\n", err)
						continue
					}
				}

				err = kafkaProd.ProduceN(convertedMsgs)
				if err != nil {
					return err
				}
				channels[channelI].LastMessageID = lastID
			}
			err = repo.SaveLastMessageIDs(channels)
			if err != nil {
				return err
			}

			time.Sleep(10 * time.Second)
		}
	}); err != nil {
		panic(err)
	}
}
