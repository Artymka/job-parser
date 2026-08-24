package client

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/artymka/jobparser/internal/telegram/config"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"github.com/mdp/qrterminal/v3"
)

type Client struct {
	TgClient *telegram.Client
}

type CallbackFunc func(c *Client) error

func NewClient(config *config.Config) (*Client, error) {
	resolver, err := dcs.MTProxy(config.ProxyAddr, config.ProxySecret, dcs.MTProxyOptions{})
	if err != nil {
		panic(fmt.Errorf("cannot use proxy: %w", err))
	}

	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Resolver: resolver,
	})

	if err != nil {
		return nil, err
	}
	res := Client{
		TgClient: client,
	}
	return &res, nil
}

func (c *Client) Run(callback CallbackFunc) error {
	dispatcher := tg.NewUpdateDispatcher()
	loggedIn := qrlogin.OnLoginToken(&dispatcher)

	// устанавливем таймаут на подключение к telegram user api
	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer pingCancel()
	connected := make(chan struct{})

	go func() {
		select {
		case <-connected:
			return
		case <-pingCtx.Done():
			fmt.Println("connection timeout passed")
			runCancel()
			panic("connection timeout passed")
		}
	}()

	// запускаем клиент
	return c.TgClient.Run(runCtx, func(ctx context.Context) error {
		close(connected)

		status, err := c.TgClient.Auth().Status(ctx)
		if err != nil {
			return err
		}

		if !status.Authorized {
			show := func(ctx context.Context, token qrlogin.Token) error {
				qrterminal.Generate(token.URL(), qrterminal.L, os.Stderr)
				return nil
			}

			if _, err := c.TgClient.QR().Auth(runCtx, loggedIn, show); err != nil {
				return err
			}
		}
		fmt.Println("authorized successfuly")

		return callback(c)
	})
}

// package telegram

// import (
// 	"context"
// 	"fmt"
// 	"os"

// 	"github.com/gotd/contrib/bg"
// 	"github.com/gotd/td/session"
// 	"github.com/gotd/td/telegram"
// 	"github.com/gotd/td/telegram/auth/qrlogin"
// 	"github.com/gotd/td/telegram/dcs"
// 	"github.com/gotd/td/tg"
// 	"github.com/mdp/qrterminal/v3"
// )

// type TelegramConfig struct {
// 	Phone       string
// 	AppID       int
// 	AppHash     string
// 	ProxyAddr   string
// 	ProxySecret []byte
// 	SessionPath string
// }

// type Client struct {
// 	client *telegram.Client
// 	stop   bg.StopFunc
// }

// func NewClient(config *TelegramConfig) (*Client, error) {
// 	ctx := context.TODO()

// 	resolver, err := dcs.MTProxy(config.ProxyAddr, config.ProxySecret, dcs.MTProxyOptions{})
// 	if err != nil {
// 		return nil, fmt.Errorf("cannot use proxy: %w", err)
// 	}

// 	dispatcher := tg.NewUpdateDispatcher()
// 	loggedIn := qrlogin.OnLoginToken(&dispatcher)
// 	client := telegram.NewClient(config.AppID, config.AppHash, telegram.Options{
// 		SessionStorage: &session.FileStorage{Path: config.SessionPath},
// 		Resolver:       resolver,
// 	})

// 	stop, err := bg.Connect(client)
// 	if err != nil {
// 		return nil, fmt.Errorf("cannot connect to tg api: %w", err)
// 	}

// 	status, err := client.Auth().Status(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("auth status: %w", err)
// 	}
// 	if status.Authorized {
// 		fmt.Println("status authorized")
// 	} else {
// 		fmt.Println("create new session")
// 		show := func(ctx context.Context, token qrlogin.Token) error {
// 			qrterminal.Generate(token.URL(), qrterminal.L, os.Stderr)
// 			return nil
// 		}

// 		if _, err := client.QR().Auth(ctx, loggedIn, show); err != nil {
// 			return nil, err
// 		}
// 	}

// 	// if _, err := os.Stat(config.SessionPath); errors.Is(err, os.ErrNotExist) {
// 	// 	show := func(ctx context.Context, token qrlogin.Token) error {
// 	// 		qrterminal.Generate(token.URL(), qrterminal.L, os.Stderr)
// 	// 		return nil
// 	// 	}

// 	// 	if _, err := client.QR().Auth(ctx, loggedIn, show); err != nil {
// 	// 		return nil, err
// 	// 	}
// 	// }

// 	res := Client{
// 		client: client,
// 		stop:   stop,
// 	}
// 	return &res, nil
// }

// func (c *Client) Stop() error {
// 	return c.stop()
// }

// func (c *Client)
