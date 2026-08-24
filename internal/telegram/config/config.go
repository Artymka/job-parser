package config

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ProxyAddr   string
	ProxySecret []byte
}

func New(envFileName string) (*Config, error) {
	if err := godotenv.Load(envFileName); err != nil {
		return nil, err
	}

	config := Config{}
	var err error

	config.ProxyAddr = os.Getenv("PROXY_ADDR")
	if config.ProxyAddr == "" {
		return nil, fmt.Errorf("No proxy addres")
	}
	proxySecretHex := os.Getenv("PROXY_SECRET")
	if proxySecretHex == "" {
		return nil, fmt.Errorf("No proxy secret")
	}
	config.ProxySecret, err = hex.DecodeString(proxySecretHex)
	if err != nil {
		return nil, fmt.Errorf("hex: %w", err)
	}

	return &config, nil
}
