package telegram

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/artymka/jobparser-producer/internal/config"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"github.com/mdp/qrterminal/v3"
)

func GetLastMessages(config *config.Config) []string {
	// load config
	// godotenv.Load()
	// phone := os.Getenv("PHONE")
	// if phone == "" {
	// 	panic("No phone number")
	// }
	// appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	// if err != nil {
	// 	panic(fmt.Errorf("wrong app_id: %w", err))
	// }
	// appHash := os.Getenv("APP_HASH")
	// if appHash == "" {
	// 	panic("No app_hash")
	// }

	// proxyAddr := os.Getenv("PROXY_ADDR")
	// if proxyAddr == "" {
	// 	panic("No proxy addres")
	// }
	// proxySecretHex := os.Getenv("PROXY_SECRET")
	// if proxySecretHex == "" {
	// 	panic("No proxy secret")
	// }
	// proxySecret, err := hex.DecodeString(proxySecretHex)
	// if err != nil {
	// 	panic(fmt.Errorf("hex: %w", err))
	// }

	// resolver, err := dcs.MTProxy(proxyAddr, []byte(proxySecret), dcs.MTProxyOptions{})
	resolver, err := dcs.MTProxy(config.ProxyAddr, config.ProxySecret, dcs.MTProxyOptions{})
	if err != nil {
		panic(fmt.Errorf("cannot use proxy: %w", err))
	}
	// fmt.Println(resolver)

	dispatcher := tg.NewUpdateDispatcher()
	loggedIn := qrlogin.OnLoginToken(&dispatcher)
	client := telegram.NewClient(config.AppID, config.AppHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: config.SessionPath},
		Resolver:       resolver,
	})

	// codePrompt := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	// 	fmt.Print("Enter code: ")
	// 	scanner := bufio.NewScanner(os.Stdin)
	// 	if !scanner.Scan() {
	// 		return "", scanner.Err()
	// 	}
	// 	return strings.TrimSpace(scanner.Text()), nil
	// }

	res := make([]string, 0)

	// ВАЖНО: используем client.Run() для запуска клиента
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// Теперь запускаем аутентификацию внутри клиента
		// flow := auth.NewFlow(
		// 	auth.CodeOnly(phone, auth.CodeAuthenticatorFunc(codePrompt)),
		// 	auth.SendCodeOptions{},
		// )
		// if err := client.Auth().IfNecessary(ctx, flow); err != nil {
		// 	return err
		// }
		// if err := flow.Run(ctx, client.Auth()); err != nil {
		// 	return err
		// }
		if _, err := os.Stat(config.SessionPath); errors.Is(err, os.ErrNotExist) {
			show := func(ctx context.Context, token qrlogin.Token) error {
				qrterminal.Generate(token.URL(), qrterminal.L, os.Stderr)
				return nil
			}

			if _, err := client.QR().Auth(ctx, loggedIn, show); err != nil {
				return err
			}
		}

		fmt.Println("Authentication successful!")

		// Здесь можно делать запросы к API
		// Например, получить информацию о пользователе
		api := client.API()
		user, err := client.Self(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("Logged in as: %s\n", user.FirstName)

		// 1. Получаем список диалогов (чатов, каналов, групп)
		dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetPeer: &tg.InputPeerEmpty{},
			Limit:      100, // Можно увеличить при необходимости
		})
		if err != nil {
			return fmt.Errorf("failed to get dialogs: %w", err)
		}

		// допиши логику получения сообщений здесь

		// Приводим результат к нужному типу
		var dialogsSlice *tg.MessagesDialogsSlice
		// var chats []tg.ChatClass
		// var messages []tg.MessageClass

		switch d := dialogs.(type) {
		case *tg.MessagesDialogs:
			// Обычный MessagesDialogs
			dialogsSlice = &tg.MessagesDialogsSlice{
				Dialogs:  d.Dialogs,
				Messages: d.Messages,
				Chats:    d.Chats,
				Users:    d.Users,
			}
		case *tg.MessagesDialogsSlice:
			// Уже Slice
			dialogsSlice = d
		case *tg.MessagesDialogsNotModified:
			// Если ничего не изменилось с прошлого запроса
			fmt.Println("No new dialogs since last check")
			return nil
		default:
			return fmt.Errorf("unexpected dialogs type: %T", dialogs)
		}

		// Проверяем, что dialogsSlice не nil
		if dialogsSlice == nil {
			return fmt.Errorf("failed to get dialogs slice")
		}

		fmt.Println()
		fmt.Println("=== Channels and their last messages ===")
		fmt.Println()

		// Создаем карту для быстрого доступа к чатам по ID
		chatsMap := make(map[int64]tg.ChatClass)
		for _, chat := range dialogsSlice.Chats {
			switch c := chat.(type) {
			case *tg.Channel:
				chatsMap[c.ID] = c
			case *tg.Chat:
				chatsMap[c.ID] = c
			case *tg.ChatEmpty:
				// Пустой чат, пропускаем
				continue
			case *tg.ChatForbidden:
				// Запрещенный чат
				continue
			case *tg.ChannelForbidden:
				// Запрещенный канал
				continue
			default:
				fmt.Printf("Unknown chat type: %T\n", chat)
			}
		}

		// 2. Проходим по всем диалогам
		channelCounter := 0
		for _, dialog := range dialogsSlice.Dialogs {
			// Проверяем тип пира
			switch peer := dialog.GetPeer().(type) {
			case *tg.PeerChannel:
				// Это канал, получаем его ID
				channelID := peer.ChannelID

				// Находим канал в карте
				chat, exists := chatsMap[channelID]
				if !exists {
					continue
				}

				// Проверяем, что это действительно канал
				channel, ok := chat.(*tg.Channel)
				if !ok {
					continue
				}

				// Пропускаем, если канал не является публичным или мы не подписаны
				if channel.Left || channel.Restricted {
					continue
				}

				channelCounter++

				// Получаем сообщения из канала
				messagesReq := &tg.MessagesGetHistoryRequest{
					Peer: &tg.InputPeerChannel{
						ChannelID:  channel.ID,
						AccessHash: channel.AccessHash,
					},
					Limit: 1, // Получаем только последнее сообщение
				}

				messages, err := api.MessagesGetHistory(ctx, messagesReq)
				if err != nil {
					fmt.Printf("Failed to get messages for channel %s: %v\n", channel.Title, err)
					continue
				}

				// Извлекаем сообщения из ответа
				var msgText string
				// var msgDate time.Time
				switch msgs := messages.(type) {
				case *tg.MessagesChannelMessages:
					if len(msgs.Messages) > 0 {
						// Берем последнее сообщение (оно будет первым в списке)
						lastMsg := msgs.Messages[0]
						switch m := lastMsg.(type) {
						case *tg.Message:
							if m.Message != "" {
								msgText = m.Message
								// } else if m.Media != nil {
								// 	// Пытаемся определить тип медиа
								// 	switch m.Media.(type) {
								// 	case *tg.MessageMediaPhoto:
								// 		msgText = "[Photo]"
								// 	case *tg.MessageMediaDocument:
								// 		msgText = "[Document]"
								// 	case *tg.MessageMediaGeo:
								// 		msgText = "[Location]"
								// 	case *tg.MessageMediaContact:
								// 		msgText = "[Contact]"
								// 	case *tg.MessageMediaVenue:
								// 		msgText = "[Venue]"
								// 	case *tg.MessageMediaGame:
								// 		msgText = "[Game]"
								// 	case *tg.MessageMediaInvoice:
								// 		msgText = "[Invoice]"
								// 	default:
								// 		msgText = "[Media content]"
								// 	}
							} else {
								msgText = "[Empty message]"
							}
							// msgDate = time.Unix(int64(m.Date), 0)
						default:
							msgText = "[Unsupported message type]"
						}
					} else {
						msgText = "[No messages]"
					}
				case *tg.MessagesMessages:
					if len(msgs.Messages) > 0 {
						lastMsg := msgs.Messages[0]
						switch m := lastMsg.(type) {
						case *tg.Message:
							if m.Message != "" {
								msgText = m.Message
							} else {
								msgText = "[Empty message]"
							}
							// msgDate = time.Unix(int64(m.Date), 0)
						default:
							msgText = "[Unsupported message type]"
						}
					} else {
						msgText = "[No messages]"
					}
				case *tg.MessagesMessagesSlice:
					if len(msgs.Messages) > 0 {
						lastMsg := msgs.Messages[0]
						switch m := lastMsg.(type) {
						case *tg.Message:
							if m.Message != "" {
								msgText = m.Message
							} else {
								msgText = "[Empty message]"
							}
							// msgDate = time.Unix(int64(m.Date), 0)
						default:
							msgText = "[Unsupported message type]"
						}
					} else {
						msgText = "[No messages]"
					}
				default:
					msgText = "[Unexpected response type]"
				}

				// Выводим информацию о канале и последнем сообщении
				res = append(res, msgText)
				// fmt.Printf("Channel #%d: %s\n", channelCounter, channel.Title)
				// if channel.Username != "" {
				// 	fmt.Printf("  Username: @%s\n", channel.Username)
				// }
				// fmt.Printf("  Last message date: %s\n", msgDate.Format("2006-01-02 15:04:05"))
				// fmt.Printf("  Last message: %s\n", msgText)
				// fmt.Println(strings.Repeat("-", 60))

			case *tg.PeerChat:
				// Это обычный чат/группа, пропускаем
				continue
			case *tg.PeerUser:
				// Это личный диалог, пропускаем
				continue
			default:
				fmt.Printf("Unknown peer type: %T\n", peer)
			}
		}

		if channelCounter == 0 {
			fmt.Println("No channels found in your dialogs")
		}

		fmt.Println("=== End of channel messages ===")
		fmt.Println()

		return nil
	}); err != nil {
		panic(err)
	}

	return res
}
