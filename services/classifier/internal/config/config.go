package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type PostgresConfig struct {
	DBName   string
	User     string
	Password string
	SSLMode  string
}

type Config struct {
	PgConfig    PostgresConfig
	KafkaBroker string
	AIKey       string
	AIModel     string
	AIEndpoint  string
}

func New(configPath string) (*Config, error) {
	if err := godotenv.Load(configPath); err != nil {
		return nil, err
	}

	config := Config{}

	config.KafkaBroker = os.Getenv("KAFKA_BROKER")
	if config.KafkaBroker == "" {
		return nil, fmt.Errorf("No KAFKA_BORKER")
	}
	config.AIKey = os.Getenv("AI_KEY")
	if config.AIKey == "" {
		return nil, fmt.Errorf("No AI_KEY")
	}
	config.AIModel = os.Getenv("AI_MODEL")
	if config.AIModel == "" {
		return nil, fmt.Errorf("No AI_MODEL")
	}
	config.AIEndpoint = os.Getenv("AI_ENDPOINT")
	if config.AIEndpoint == "" {
		return nil, fmt.Errorf("No AI_ENDPOINT")
	}

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

	return &config, nil
}
