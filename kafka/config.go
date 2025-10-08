package kafka

import (
	"log"
	"strings"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Brokers []string
}

func LoadConfig() Config {
	// Load .env into environment
	_ = godotenv.Load()

	// Viper will look at env variables
	viper.AutomaticEnv()

	// Set defaults - use localhost for local development, can be overridden by env vars
	viper.SetDefault("KAFKA_BROKERS", "localhost:9092")

	brokers := viper.GetString("KAFKA_BROKERS")
	if brokers == "" {
		log.Fatal(color.RedString("KAFKA_BROKERS environment variable is required"))
	}

	color.Blue("Connecting to Kafka brokers: %s", brokers)

	return Config{
		Brokers: strings.Split(brokers, ","),
	}
}
