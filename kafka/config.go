package kafka

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/aws"
)

type Config struct {
	Brokers   []string
	UseAWSIAM bool
	AWSRegion string
}

func LoadConfig() Config {
	// Load .env into environment
	_ = godotenv.Load()

	// Viper will look at env variables
	viper.AutomaticEnv()

	// Set defaults - use localhost for local development, can be overridden by env vars
	viper.SetDefault("KAFKA_BROKERS", "localhost:9092")
	viper.SetDefault("KAFKA_USE_AWS_IAM", "false")
	viper.SetDefault("AWS_REGION", "us-east-1")

	brokers := viper.GetString("KAFKA_BROKERS")
	if brokers == "" {
		log.Fatal(color.RedString("KAFKA_BROKERS environment variable is required"))
	}

	useAWSIAM := viper.GetBool("KAFKA_USE_AWS_IAM")
	awsRegion := viper.GetString("AWS_REGION")

	if useAWSIAM {
		color.Blue("Connecting to AWS MSK brokers: %s (region: %s) with IAM authentication", brokers, awsRegion)
	} else {
		color.Blue("Connecting to Kafka brokers: %s", brokers)
	}

	return Config{
		Brokers:   strings.Split(brokers, ","),
		UseAWSIAM: useAWSIAM,
		AWSRegion: awsRegion,
	}
}

// getCommonOptions returns common kgo options including AWS IAM authentication if enabled
func (c *Config) getCommonOptions() ([]kgo.Opt, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(c.Brokers...),
	}

	if c.UseAWSIAM {
		// Load AWS configuration
		awsConfig, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, err
		}

		// Get AWS credentials
		creds, err := awsConfig.Credentials.Retrieve(context.Background())
		if err != nil {
			return nil, err
		}

		// Create AWS MSK IAM SASL mechanism
		saslMech := aws.Auth{
			AccessKey:    creds.AccessKeyID,
			SecretKey:    creds.SecretAccessKey,
			SessionToken: creds.SessionToken,
			UserAgent:    "kafka-cli-franz-go",
		}

		opts = append(opts, kgo.SASL(saslMech.AsManagedStreamingIAMMechanism()))
	}

	return opts, nil
}

// NewConsumerClient creates a new franz-go client optimized for consuming messages
func (c *Config) NewConsumerClient(groupID, topic string) (*kgo.Client, error) {
	commonOpts, err := c.getCommonOptions()
	if err != nil {
		return nil, err
	}

	opts := append(commonOpts,
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topic),
		kgo.FetchMinBytes(1),                   // Start fetching immediately when any data is available
		kgo.FetchMaxBytes(1024*1024),           // Allow up to 1MB per fetch (much more reasonable)
		kgo.FetchMaxWait(100*time.Millisecond), // Don't wait more than 100ms for more data
		kgo.SessionTimeout(10*time.Second),     // Faster session timeout
		kgo.HeartbeatInterval(3*time.Second),   // More frequent heartbeats
	)

	return kgo.NewClient(opts...)
}

// NewProducerClient creates a new franz-go client optimized for producing messages
func (c *Config) NewProducerClient(defaultTopic string) (*kgo.Client, error) {
	commonOpts, err := c.getCommonOptions()
	if err != nil {
		return nil, err
	}

	opts := commonOpts
	if defaultTopic != "" {
		opts = append(opts, kgo.DefaultProduceTopic(defaultTopic))
	}

	return kgo.NewClient(opts...)
}

// NewAdminClient creates a new franz-go admin client for metadata operations
func (c *Config) NewAdminClient() (*kgo.Client, *kadm.Client, error) {
	commonOpts, err := c.getCommonOptions()
	if err != nil {
		return nil, nil, err
	}

	client, err := kgo.NewClient(commonOpts...)
	if err != nil {
		return nil, nil, err
	}

	adminClient := kadm.NewClient(client)
	return client, adminClient, nil
}

// NewPartitionConsumerClient creates a franz-go client for consuming specific partitions with offset control
func (c *Config) NewPartitionConsumerClient(topic string, partition int32, offset int64) (*kgo.Client, error) {
	commonOpts, err := c.getCommonOptions()
	if err != nil {
		return nil, err
	}

	opts := append(commonOpts,
		kgo.ConsumePartitions(map[string]map[int32]kgo.Offset{
			topic: {partition: kgo.NewOffset().At(offset)},
		}),
		kgo.FetchMinBytes(1),
		kgo.FetchMaxBytes(1024*1024),
	)

	return kgo.NewClient(opts...)
}
