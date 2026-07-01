package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"github.com/joho/godotenv"
)

func main() {
	// load config
	godotenv.Load()
	phone := os.Getenv("PHONE")
	if phone == "" {
		panic("No phone number")
	}
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		panic(fmt.Errorf("wrong app_id: %w", err))
	}
	appHash := os.Getenv("APP_HASH")
	if appHash == "" {
		panic("No app_hash")
	}

	proxyAddr := os.Getenv("PROXY_ADDR")
	if proxyAddr == "" {
		panic("No proxy addres")
	}
	proxySecretHex := os.Getenv("PROXY_SECRET")
	if proxySecretHex == "" {
		panic("No proxy secret")
	}
	proxySecret, err := hex.DecodeString(proxySecretHex)
	if err != nil {
		panic(fmt.Errorf("hex: %w", err))
	}

	// resolver, err := dcs.MTProxy(proxyAddr, []byte(proxySecret), dcs.MTProxyOptions{})
	resolver, err := dcs.MTProxy(proxyAddr, proxySecret, dcs.MTProxyOptions{})
	if err != nil {
		panic(fmt.Errorf("cannot use proxy: %w", err))
	}
	// fmt.Println(resolver)

	client := telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: "session.json"},
		Resolver:       resolver,
	})

	codePrompt := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
		fmt.Print("Enter code: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return "", scanner.Err()
		}
		return strings.TrimSpace(scanner.Text()), nil
	}

	// ВАЖНО: используем client.Run() для запуска клиента
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// Теперь запускаем аутентификацию внутри клиента
		flow := auth.NewFlow(
			auth.CodeOnly(phone, auth.CodeAuthenticatorFunc(codePrompt)),
			auth.SendCodeOptions{},
		)
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}
		// if err := flow.Run(ctx, client.Auth()); err != nil {
		// 	return err
		// }

		fmt.Println("Authentication successful!")

		// Здесь можно делать запросы к API
		// Например, получить информацию о пользователе
		// api := client.API()
		user, err := client.Self(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("Logged in as: %s\n", user.FirstName)

		return nil
	}); err != nil {
		panic(err)
	}
}
