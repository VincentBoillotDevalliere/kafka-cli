/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
    "context"
    "time"

    "github.com/fatih/color"
    kafkaGo "github.com/segmentio/kafka-go"
    "github.com/spf13/cobra"

	"github.com/VincentBoillotDevalliere/kafka-cli/kafka"
)

// consumeCmd represents the consume command
var consumeCmd = &cobra.Command{
	Use:   "consume",
	Short: "Consume messages from a Kafka topic",
	Run: func(cmd *cobra.Command, args []string) {
		topic := args[0]
        if topic == "" {
            color.Red("Topic is required")
        }
        color.Cyan("Consuming messages from topic: %s", topic)

		readWithReader(topic, "consumer-through-kafka 1")
	},
}

// Read from the topic using kafka.Reader
// Readers can use consumer groups (but are not required to)
func readWithReader(topic, groupID string) {
	cfg := kafka.LoadConfig()
	r := kafkaGo.NewReader(kafkaGo.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  groupID,
		Topic:    topic,
		MaxBytes: 100, // per message
		// more options are available
	})

	// Create a deadline
	readDeadline, cancel := context.WithDeadline(context.Background(),
		time.Now().Add(60*time.Second))
	defer cancel()
	for {
		msg, err := r.ReadMessage(readDeadline)
		if err != nil {
			break
		}
        color.Yellow("message at topic/partition/offset %v/%v/%v: %s = %s",
            msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
	}

    if err := r.Close(); err != nil {
        color.Red("failed to close reader: %v", err)
    }
}

func init() {
	rootCmd.AddCommand(consumeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// consumeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// consumeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
