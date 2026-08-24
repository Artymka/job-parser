package lib

import (
	"context"
	"fmt"

	"github.com/artymka/jobparser/internal/telegram/client"
	m "github.com/artymka/jobparser/internal/telegram/messages"
	"github.com/artymka/jobparser/services/telegram-scrapper/internal/models"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

func ScrapChannel(c *client.Client, channel models.Channel) ([]m.MessageData, error) {
	const op = "scrap_channel"
	// user, err := c.TgClient.Self(context.TODO())
	// if err != nil {
	// 	return err
	// }
	// selfPeer := user.AsInputPeer()

	api := c.TgClient.API()

	// с помощью manager просто получать InputPeers
	manager := peers.Options{}.Build(api)
	channelPeer, err := manager.Resolve(context.TODO(), channel.Username)
	if err != nil {
		return nil, fmt.Errorf("%s: cannot resolve contact - %w", op, err)
	}

	inputChannelPeer, ok := channelPeer.InputPeer().(*tg.InputPeerChannel)
	if !ok {
		return nil, fmt.Errorf("%s: manager does not work", op)
	}

	// получаем сообщения из канала
	rawMessages, err := api.MessagesGetHistory(context.TODO(), &tg.MessagesGetHistoryRequest{
		Peer:  inputChannelPeer,
		Limit: 2,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: get chat history: %w", op, err)
	}

	// преобразуем результат в []tg.MessageClass
	messages, err := m.ExtractMessages(rawMessages)
	if err != nil {
		return nil, err
	}

	resMessages := make([]m.MessageData, 0, len(messages))
	for _, message := range messages {
		msgData, err := m.ExtractMessageData(message, channel.Username)
		if err != nil {
			fmt.Printf("message with error: %v\n", err)
			continue
		}
		resMessages = append(resMessages, msgData)
	}

	return resMessages, nil

	// // получаем необходимую для работы информацию, а именно:
	// // текст - для классификации
	// // id и username канала (просто дублированный) - для последующей пересылки
	// idsToForward := make([]int, 0)
	// for _, message := range messages {
	// 	msgData, err := m.ExtractMessageData(message, username)
	// 	if err != nil {
	// 		fmt.Printf("message with error: %v\n", err)
	// 		continue
	// 	}
	// 	idsToForward = append(idsToForward, msgData.ID)
	// }

	// // randomID нужны телеграмму, чтобы он мог откинуть дубликаты
	// // т.е. если повторяем запрос по пересылке сообщений через telegram user api,
	// // мы должны указывать одни и те же random ids
	// randomIds := make([]int64, len(idsToForward))
	// for i := 0; i < len(randomIds); i++ {
	// 	randomIds[i] = rand.Int63()
	// }

	// // пересылаем сообщения в избранное
	// _, err = api.MessagesForwardMessages(context.TODO(), &tg.MessagesForwardMessagesRequest{
	// 	FromPeer: inputChannelPeer,
	// 	ToPeer:   selfPeer,
	// 	ID:       idsToForward,
	// 	RandomID: randomIds,
	// })
	// if err != nil {
	// 	fmt.Printf("could not forward a message: %v\n", err)
	// }

	// return nil
}
