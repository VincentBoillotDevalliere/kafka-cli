// Package kafka provides Kafka configuration and client utilities with AWS MSK IAM support
package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/aws"
)

// Config holds the Kafka configuration
type Config struct {
	Brokers    []string
	UseAWSIAM  bool
	AWSRegion  string
	TLSEnabled bool
	awsConfig  *awssdk.Config
	tlsConfig  *tls.Config
}

// LoadConfig is a convenience function that creates a new Kafka configuration
// and panics if there's an error (for backward compatibility)
func LoadConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load Kafka configuration: %v", err))
	}
	return cfg
}

// NewConfig creates a new Kafka configuration from environment variables
func NewConfig() (*Config, error) {
	cfg := &Config{
		TLSEnabled: true, // Default to true for security
	}

	// Parse broker addresses
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "localhost:9092" // Default for local development
		cfg.TLSEnabled = false
	}
	cfg.Brokers = strings.Split(brokersEnv, ",")

	// Allow env override for TLS enablement
	if tlsEnabled, ok := lookupEnvBool("KAFKA_TLS_ENABLED"); ok {
		cfg.TLSEnabled = tlsEnabled
	}

	// Check if AWS IAM is enabled - auto-detect MSK or explicit setting
	if useIAM, ok := lookupEnvBool("KAFKA_USE_AWS_IAM"); ok {
		cfg.UseAWSIAM = useIAM
	}

	// Auto-detect AWS MSK based on broker URLs
	if !cfg.UseAWSIAM {
		for _, broker := range cfg.Brokers {
			if strings.Contains(broker, ".kafka.") && strings.Contains(broker, ".amazonaws.com") {
				cfg.UseAWSIAM = true
				break
			}
		}
	}

	if cfg.UseAWSIAM {
		// Get AWS region
		cfg.AWSRegion = os.Getenv("AWS_REGION")
		if cfg.AWSRegion == "" {
			return nil, fmt.Errorf("AWS_REGION environment variable is required when using AWS MSK")
		}

		// Load AWS configuration
		awsCfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(cfg.AWSRegion))
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
		cfg.awsConfig = &awsCfg
	}

	if cfg.TLSEnabled {
		tlsCfg, err := buildTLSConfigFromEnv()
		if err != nil {
			return nil, err
		}
		cfg.tlsConfig = tlsCfg
	}

	return cfg, nil
}

func buildTLSConfigFromEnv() (*tls.Config, error) {
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if insecureSkipVerify, ok := lookupEnvBool("KAFKA_TLS_INSECURE_SKIP_VERIFY"); ok {
		tlsCfg.InsecureSkipVerify = insecureSkipVerify
	}

	if caFile := strings.TrimSpace(os.Getenv("KAFKA_TLS_CA_FILE")); caFile != "" {
		data, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read KAFKA_TLS_CA_FILE %q: %w", caFile, err)
		}
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(data); !ok {
			return nil, fmt.Errorf("failed to parse CA certificates in %q", caFile)
		}
		tlsCfg.RootCAs = pool
	}

	return tlsCfg, nil
}

func lookupEnvBool(key string) (bool, bool) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return false, false
	}

	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		return false, false
	}
}

// CreateProducer creates a new Kafka producer with the configuration
func (c *Config) CreateProducer(opts ...ProducerOption) (*kgo.Client, error) {
	options := c.getBaseOptions()

	// Apply producer-specific options
	for _, opt := range opts {
		opt(&options)
	}

	// Add producer-specific configurations
	options = append(options,
		kgo.RequiredAcks(kgo.AllISRAcks()), // Wait for all replicas
		kgo.ProducerBatchMaxBytes(1000000), // 1MB batches
		kgo.ProducerBatchCompression(kgo.GzipCompression()),
		kgo.ProducerLinger(100*time.Millisecond), // Batch for up to 100ms
	)

	client, err := kgo.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return client, nil
}

// CreateConsumer creates a new Kafka consumer with the configuration
func (c *Config) CreateConsumer(consumerGroup string, topics []string, opts ...ConsumerOption) (*kgo.Client, error) {
	if consumerGroup == "" {
		return nil, fmt.Errorf("consumer group is required")
	}
	if len(topics) == 0 {
		return nil, fmt.Errorf("at least one topic is required")
	}

	options := c.getBaseOptions()

	// Apply consumer-specific options
	for _, opt := range opts {
		opt(&options)
	}

	// Add consumer-specific configurations
	options = append(options,
		kgo.ConsumerGroup(consumerGroup),
		kgo.ConsumeTopics(topics...),
		kgo.FetchMinBytes(1),                   // Fetch at least 1 byte
		kgo.FetchMaxBytes(52428800),            // 50MB max fetch
		kgo.FetchMaxWait(500*time.Millisecond), // Wait up to 500ms for data
		kgo.SessionTimeout(30*time.Second),
		kgo.HeartbeatInterval(3*time.Second),
		kgo.RebalanceTimeout(30*time.Second),
		kgo.AutoCommitInterval(1*time.Second),
	)

	client, err := kgo.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return client, nil
}

// getBaseOptions returns the base kgo options for both producers and consumers
func (c *Config) getBaseOptions() []kgo.Opt {
	options := []kgo.Opt{
		kgo.SeedBrokers(c.Brokers...),
		kgo.RequestTimeoutOverhead(10 * time.Second),
		kgo.ConnIdleTimeout(30 * time.Second),
	}

	// Add TLS configuration if enabled
	if c.TLSEnabled && c.tlsConfig != nil {
		options = append(options, kgo.DialTLSConfig(c.tlsConfig.Clone()))
	}

	// Add AWS IAM SASL configuration if enabled
	if c.UseAWSIAM && c.awsConfig != nil {
		saslMech := aws.ManagedStreamingIAM(func(ctx context.Context) (aws.Auth, error) {
			creds, err := c.awsConfig.Credentials.Retrieve(ctx)
			if err != nil {
				return aws.Auth{}, fmt.Errorf("failed to retrieve AWS credentials: %w", err)
			}

			return aws.Auth{
				AccessKey:    creds.AccessKeyID,
				SecretKey:    creds.SecretAccessKey,
				SessionToken: creds.SessionToken,
				UserAgent:    "kafka-cli",
			}, nil
		})

		options = append(options, kgo.SASL(saslMech))
	}

	return options
}

// ProducerOption is a function type for configuring producer options
type ProducerOption func(*[]kgo.Opt)

// WithProducerBatchSize sets the producer batch size
func WithProducerBatchSize(size int32) ProducerOption {
	return func(opts *[]kgo.Opt) {
		*opts = append(*opts, kgo.ProducerBatchMaxBytes(size))
	}
}

// WithProducerLinger sets the producer linger time
func WithProducerLinger(linger time.Duration) ProducerOption {
	return func(opts *[]kgo.Opt) {
		*opts = append(*opts, kgo.ProducerLinger(linger))
	}
}

// WithProducerCompression sets the producer compression type
func WithProducerCompression(compression kgo.CompressionCodec) ProducerOption {
	return func(opts *[]kgo.Opt) {
		*opts = append(*opts, kgo.ProducerBatchCompression(compression))
	}
}

// WithRequiredAcks sets the required acknowledgments for the producer
func WithRequiredAcks(acks kgo.Acks) ProducerOption {
	return func(opts *[]kgo.Opt) {
		*opts = append(*opts, kgo.RequiredAcks(acks))
	}
}

// ConsumerOption is a function type for configuring consumer options
type ConsumerOption func(*[]kgo.Opt)

// WithConsumerOffset sets the initial offset for the consumer
func WithConsumerOffset(offset kgo.Offset) ConsumerOption {
	return func(opts *[]kgo.Opt) {
		*opts = append(*opts, kgo.ConsumeResetOffset(offset))
	}
}

// WithAutoCommit enables or disables auto-commit for the consumer
func WithAutoCommit(enabled bool) ConsumerOption {
	return func(opts *[]kgo.Opt) {
		if !enabled {
			*opts = append(*opts, kgo.DisableAutoCommit())
		}
	}
}

// WithSessionTimeout sets the session timeout for the consumer
func WithSessionTimeout(timeout time.Duration) ConsumerOption {
	return func(opts *[]kgo.Opt) {
		*opts = append(*opts, kgo.SessionTimeout(timeout))
	}
}

// WithHeartbeatInterval sets the heartbeat interval for the consumer
func WithHeartbeatInterval(interval time.Duration) ConsumerOption {
	return func(opts *[]kgo.Opt) {
		*opts = append(*opts, kgo.HeartbeatInterval(interval))
	}
}

// WithFetchMaxBytes sets the maximum bytes to fetch per request
func WithFetchMaxBytes(maxBytes int32) ConsumerOption {
	return func(opts *[]kgo.Opt) {
		*opts = append(*opts, kgo.FetchMaxBytes(maxBytes))
	}
}

// WithFetchMaxWait sets the maximum time to wait for fetch requests
func WithFetchMaxWait(maxWait time.Duration) ConsumerOption {
	return func(opts *[]kgo.Opt) {
		*opts = append(*opts, kgo.FetchMaxWait(maxWait))
	}
}

// TestConnection tests the connection to Kafka brokers
func (c *Config) TestConnection(ctx context.Context) error {
	client, err := kgo.NewClient(c.getBaseOptions()...)
	if err != nil {
		return fmt.Errorf("failed to create test client: %w", err)
	}
	defer client.Close()

	// Try to ping brokers to test connection
	err = client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka brokers: %w", err)
	}

	return nil
}

// GetBrokers returns the configured broker addresses
func (c *Config) GetBrokers() []string {
	return c.Brokers
}

// IsAWSIAMEnabled returns true if AWS IAM authentication is enabled
func (c *Config) IsAWSIAMEnabled() bool {
	return c.UseAWSIAM
}

// GetAWSRegion returns the configured AWS region
func (c *Config) GetAWSRegion() string {
	return c.AWSRegion
}

// IsTLSEnabled returns true if TLS is enabled
func (c *Config) IsTLSEnabled() bool {
	return c.TLSEnabled
}

// NewConsumerClient creates a new consumer client (for backward compatibility)
func (c *Config) NewConsumerClient(groupID, topic string) (*kgo.Client, error) {
	return c.CreateConsumer(groupID, []string{topic})
}

// NewProducerClient creates a new producer client (for backward compatibility)
func (c *Config) NewProducerClient(topic string) (*kgo.Client, error) {
	return c.CreateProducer()
}

// AdminClient wraps kadm.Client to provide the expected admin operations
type AdminClient struct {
	*kadm.Client
}

// ListTopics returns topic metadata
func (ac *AdminClient) ListTopics(ctx context.Context, topics ...string) (map[string]kadm.TopicDetail, error) {
	if len(topics) == 0 {
		return ac.Client.ListTopics(ctx)
	}
	return ac.Client.ListTopics(ctx, topics...)
}

// ListOffsetsAfterMilli returns offsets for topics after a given timestamp
func (ac *AdminClient) ListOffsetsAfterMilli(ctx context.Context, millis int64, topics ...string) (kadm.ListedOffsets, error) {
	return ac.Client.ListOffsetsAfterMilli(ctx, millis, topics...)
}

// NewAdminClient creates a new admin client (for backward compatibility)
// Returns client, adminClient, error to match existing code expectations
func (c *Config) NewAdminClient() (*kgo.Client, *AdminClient, error) {
	// For admin operations, we just need a basic client
	options := c.getBaseOptions()
	client, err := kgo.NewClient(options...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kafka admin client: %w", err)
	}

	// Create kadm admin client
	adminClient := &AdminClient{
		Client: kadm.NewClient(client),
	}

	return client, adminClient, nil
}

// NewPartitionConsumerClient creates a partition consumer client (for backward compatibility)
func (c *Config) NewPartitionConsumerClient(topic string, partition int, offset int64) (*kgo.Client, error) {
	// For partition consumption, we create a client that consumes from specific partition
	options := c.getBaseOptions()

	// Add partition-specific consumption options
	options = append(options,
		kgo.ConsumePartitions(map[string]map[int32]kgo.Offset{
			topic: {int32(partition): kgo.NewOffset().At(offset)},
		}),
	)

	client, err := kgo.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka partition consumer client: %w", err)
	}
	return client, nil
}
