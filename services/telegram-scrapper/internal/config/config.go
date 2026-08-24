package config

import (
	"fmt"
	"os"

	tgconf "github.com/artymka/jobparser/internal/telegram/config"
	"github.com/joho/godotenv"
)

type PostgresConfig struct {
	DBName   string
	User     string
	Password string
	SSLMode  string
}

type Config struct {
	TgConfig    tgconf.Config
	PgConfig    PostgresConfig
	KafkaBroker string
	PostgresDB  string
}

func New(envFileName string) (*Config, error) {
	if err := godotenv.Load(envFileName); err != nil {
		return nil, err
	}

	config := Config{}
	var err error

	// config.Phone = os.Getenv("PHONE")
	// if config.Phone == "" {
	// 	return nil, fmt.Errorf("No phone number")
	// }
	// config.AppID, err = strconv.Atoi(os.Getenv("APP_ID"))
	// if err != nil {
	// 	return nil, fmt.Errorf("wrong app_id: %w", err)
	// }
	// config.AppHash = os.Getenv("APP_HASH")
	// if config.AppHash == "" {
	// 	return nil, fmt.Errorf("No app_hash")
	// }
	// config.ProxyAddr = os.Getenv("PROXY_ADDR")
	// if config.ProxyAddr == "" {
	// 	return nil, fmt.Errorf("No proxy addres")
	// }
	// proxySecretHex := os.Getenv("PROXY_SECRET")
	// if proxySecretHex == "" {
	// 	return nil, fmt.Errorf("No proxy secret")
	// }
	// config.ProxySecret, err = hex.DecodeString(proxySecretHex)
	// if err != nil {
	// 	return nil, fmt.Errorf("hex: %w", err)
	// }
	// config.SessionPath = os.Getenv("SESSION_PATH")
	// if config.SessionPath == "" {
	// 	return nil, fmt.Errorf("No SESSION_PATH")
	// }

	tgConfig, err := tgconf.New(envFileName)
	if err != nil {
		return nil, err
	}
	config.TgConfig = *tgConfig

	config.PgConfig = PostgresConfig{}
	config.PgConfig.DBName = os.Getenv("POSTGRES_DB")
	if config.PgConfig.DBName == "" {
		return nil, fmt.Errorf("No POSTGRES_DB")
	}
	config.PgConfig.User = os.Getenv("POSTGRES_USER")
	if config.PgConfig.User == "" {
		return nil, fmt.Errorf("No POSTGRES_USER")
	}
	config.PgConfig.Password = os.Getenv("POSTGRES_PASSWORD")
	if config.PgConfig.Password == "" {
		return nil, fmt.Errorf("No POSTGRES_PASSWORD")
	}

	config.KafkaBroker = os.Getenv("KAFKA_BROKER")
	if config.KafkaBroker == "" {
		return nil, fmt.Errorf("No KAFKA_BROKER")
	}

	return &config, nil
}
