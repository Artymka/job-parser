package messages

import (
	"fmt"

	"github.com/gotd/td/tg"
)

type MessageData struct {
	Text            string `json:"text"`
	ID              int    `json:"id"`
	ChannelUsername string `json:"channel_username"`
}

func ExtractMessages(msgs tg.MessagesMessagesClass) ([]tg.MessageClass, error) {
	switch converted := msgs.(type) {
	case *tg.MessagesMessages:
		return converted.Messages, nil
	case *tg.MessagesMessagesSlice:
		return converted.Messages, nil
	case *tg.MessagesChannelMessages: // обычно из каналов приходит именно этот тип
		return converted.Messages, nil
	case *tg.MessagesMessagesNotModified:
		return []tg.MessageClass{}, nil
	default:
		return nil, fmt.Errorf("invalid messages type")
	}
}

func ExtractMessageData(msg tg.MessageClass, channelUsername string) (MessageData, error) {
	switch converted := msg.(type) {
	case *tg.MessageEmpty:
		return MessageData{}, fmt.Errorf("empty message")
	case *tg.Message:
		return MessageData{
			Text:            converted.Message,
			ID:              converted.ID,
			ChannelUsername: channelUsername,
		}, nil
	case *tg.MessageService:
		return MessageData{}, fmt.Errorf("service message")
	default:
		return MessageData{}, fmt.Errorf("invalid message")
	}
}
