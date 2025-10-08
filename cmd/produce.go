/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "os"

    "github.com/fatih/color"
    kafkaGo "github.com/segmentio/kafka-go"
    "github.com/spf13/cobra"

    "github.com/VincentBoillotDevalliere/kafka-cli/kafka"
)

type MessageEnvelope struct {
    Topic  string                 `json:"topic"`
    Object map[string]interface{} `json:"object"`
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
        if message == "" && inputFile == "" {
            return fmt.Errorf("either --message or --input must be provided")
        }
        if inputFile != "" {
            messages, err := HandleFileInput(inputFile)
            if err != nil {
                return fmt.Errorf("failed to handle file input: %v", err)
            }
            for _, msg := range messages {
                color.Cyan("Producing message to topic: %s", msg.Topic)
                jsonMsg, err := json.Marshal(msg.Object)
                if err != nil {
                    return fmt.Errorf("failed to marshal message object: %v", err)
                }
                err = ProduceMessage(msg.Topic, string(jsonMsg))
                if err != nil {
                    return fmt.Errorf("failed to produce message: %v", err)
                }
                color.Green("Produced message to topic %s: %s", msg.Topic, string(jsonMsg))
            }
        } else {
            if len(args) < 1 {
                return fmt.Errorf("topic argument is required when using --message")
            }
            topic := args[0]
            color.Cyan("Producing message to topic: %s", topic)
            err := ProduceMessage(topic, message)
            if err != nil {
                return fmt.Errorf("failed to produce message: %v", err)
            }
            color.Green("Produced message to topic %s: %s", topic, message)
        }
        return nil
    },
}

func init() {
    rootCmd.AddCommand(produceCmd)
    produceCmd.Flags().StringVarP(&message, "message", "m", "", "Message to send")
    produceCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Optional input file containing messages (one per line)")
}

func ProduceMessage(topic, jsonInput string) error {
    cfg := kafka.LoadConfig()
    w := kafkaGo.NewWriter(kafkaGo.WriterConfig{
        Brokers:  cfg.Brokers,
        Topic:    topic,
        Balancer: &kafkaGo.Hash{},
    })

    err := w.WriteMessages(context.Background(),
        kafkaGo.Message{
            Value: []byte(jsonInput),
        },
    )
    if err != nil {
        return fmt.Errorf("failed to write message: %w", err)
    }
    return nil
}

func HandleFileInput(inputFile string) ([]MessageEnvelope, error) {
    color.Blue("📂 Reading messages from file: %s", inputFile)
    var inputs []MessageEnvelope
    file, err := os.Open(inputFile)
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
    defer func() { _ = file.Close() }()

    data, err := io.ReadAll(file)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
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

    color.Blue("🧾 Found %d messages in file", len(envelopes))

    for i, obj := range envelopes {
        bytes, err := json.Marshal(obj)
        if err != nil {
            return nil, fmt.Errorf("failed to encode JSON object #%d: %w", i+1, err)
        }
        inputs = append(inputs, MessageEnvelope{
            Topic:  obj.Topic,
            Object: obj.Object,
        })
        color.Yellow("  - Parsed message #%d for topic '%s': %s", i+1, obj.Topic, string(bytes))
    }

    return inputs, nil
}

