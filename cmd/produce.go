package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/VincentBoillotDevalliere/kafka-cli/kafka"
)

type MessageEnvelope struct {
	Topic   string                 `json:"Topic"`
	Headers map[string]string      `json:"Headers,omitempty"`
	Message map[string]interface{} `json:"Message"`
}

var (
	message   string
	inputFile string
)

// produceCmd represents the produce command
var produceCmd = &cobra.Command{
	Use:   "produce",
	Short: "Produce a message to a Kafka topic",
	Long: `Produce messages directly or from a JSON file.
If a JSON file is provided with -i, each element should contain a "topic" and an "object" field:
[
  { "topic": "topicA", "object": {...} },
  { "topic": "topicB", "object": {...} }
]`,
	RunE: func(cmd *cobra.Command, args []string) error {
		color.Cyan("🚀 Kafka Producer")

		if message == "" && inputFile == "" {
			return fmt.Errorf("either --message or --input must be provided")
		}
		if inputFile != "" {
			messages, err := HandleFileInput(inputFile)
			if err != nil {
				return fmt.Errorf("failed to handle file input: %v", err)
			}
			for i, msg := range messages {
				jsonMsg, err := json.Marshal(msg.Message)
				if err != nil {
					return fmt.Errorf("failed to marshal message object: %v", err)
				}
				err = ProduceMessage(msg.Topic, string(jsonMsg), msg.Headers)
				if err != nil {
					return fmt.Errorf("failed to produce message: %v", err)
				}
				color.Green("✅ Produced message #%d to topic '%s'", i+1, msg.Topic)
			}
			color.Magenta("🎉 Done!")
		} else {
			if len(args) < 1 {
				return fmt.Errorf("topic argument is required when using --message")
			}
			topic := args[0]
			err := ProduceMessage(topic, message, nil)
			if err != nil {
				return fmt.Errorf("failed to produce message: %v", err)
			}
			color.Green("✅ Produced message to topic '%s'", topic)
			color.Magenta("🎉 Done!")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(produceCmd)
	produceCmd.Flags().StringVarP(&message, "message", "m", "", "Message to send")
	produceCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Optional input file containing messages (one per line)")
}

func ProduceMessage(topic, jsonInput string, headers map[string]string) error {
	cfg := kafka.LoadConfig()
	
	// Create franz-go client with producer configuration
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.DefaultProduceTopic(topic), // Set default topic for convenience
	)
	if err != nil {
		return fmt.Errorf("failed to create kafka client: %w", err)
	}
	defer client.Close()

	// Convert headers to franz-go format
	var franzHeaders []kgo.RecordHeader
	for k, v := range headers {
		franzHeaders = append(franzHeaders, kgo.RecordHeader{Key: k, Value: []byte(v)})
	}

	// Create and send the record
	record := &kgo.Record{
		Topic:   topic,
		Value:   []byte(jsonInput),
		Headers: franzHeaders,
	}

	// Produce the message synchronously
	ctx := context.Background()
	results := client.ProduceSync(ctx, record)
	
	// Check for errors
	for _, result := range results {
		if result.Err != nil {
			return fmt.Errorf("failed to produce message: %w", result.Err)
		}
	}
	
	return nil
}

func HandleFileInput(inputFile string) ([]MessageEnvelope, error) {
	color.Blue("📂 Reading file: %s", inputFile)
	var inputs []MessageEnvelope
	data, err := readFile(inputFile)
	if err != nil {
		return nil, err
	}

	var envelopes []MessageEnvelope
	if err := json.Unmarshal(data, &envelopes); err != nil {
		// Handle non array JSON (single object)
		var singleMessage MessageEnvelope
		if err := json.Unmarshal(data, &singleMessage); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
		envelopes = append(envelopes, singleMessage)
	}

	color.Blue("🧾 Found %d messages", len(envelopes))

	for _, obj := range envelopes {
		inputs = append(inputs, MessageEnvelope{
			Headers: obj.Headers,
			Topic:   obj.Topic,
			Message: obj.Message,
		})
	}

	return inputs, nil
}

func readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return data, nil
}
