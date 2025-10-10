package kafka

import (
	"log"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
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

// NewConsumerClient creates a new franz-go client optimized for consuming messages
func (c *Config) NewConsumerClient(groupID, topic string) (*kgo.Client, error) {
	return kgo.NewClient(
		kgo.SeedBrokers(c.Brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topic),
		kgo.FetchMinBytes(1),                   // Start fetching immediately when any data is available
		kgo.FetchMaxBytes(1024*1024),           // Allow up to 1MB per fetch (much more reasonable)
		kgo.FetchMaxWait(100*time.Millisecond), // Don't wait more than 100ms for more data
		kgo.SessionTimeout(10*time.Second),     // Faster session timeout
		kgo.HeartbeatInterval(3*time.Second),   // More frequent heartbeats
	)
}

// NewProducerClient creates a new franz-go client optimized for producing messages
func (c *Config) NewProducerClient(defaultTopic string) (*kgo.Client, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(c.Brokers...),
	}

	if defaultTopic != "" {
		opts = append(opts, kgo.DefaultProduceTopic(defaultTopic))
	}

	return kgo.NewClient(opts...)
}

// NewAdminClient creates a new franz-go admin client for metadata operations
func (c *Config) NewAdminClient() (*kgo.Client, *kadm.Client, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(c.Brokers...),
	)
	if err != nil {
		return nil, nil, err
	}

	adminClient := kadm.NewClient(client)
	return client, adminClient, nil
}

// NewPartitionConsumerClient creates a franz-go client for consuming specific partitions with offset control
func (c *Config) NewPartitionConsumerClient(topic string, partition int32, offset int64) (*kgo.Client, error) {
	return kgo.NewClient(
		kgo.SeedBrokers(c.Brokers...),
		kgo.ConsumePartitions(map[string]map[int32]kgo.Offset{
			topic: {partition: kgo.NewOffset().At(offset)},
		}),
		kgo.FetchMinBytes(1),
		kgo.FetchMaxBytes(1024*1024),
	)
}
