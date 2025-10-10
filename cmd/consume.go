/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/twmb/franz-go/pkg/kgo"

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

// Read from the topic using franz-go client
// Clients can use consumer groups for distributed consumption
func readWithReader(topic, groupID string) {
	cfg := kafka.LoadConfig()

	// Create optimized consumer client
	client, err := cfg.NewConsumerClient(groupID, topic)
	if err != nil {
		color.Red("failed to create kafka client: %v", err)
		return
	}
	defer client.Close()

	// Create a deadline context
	readDeadline, cancel := context.WithDeadline(context.Background(),
		time.Now().Add(60*time.Second))
	defer cancel()

	// Poll for messages with shorter intervals for better responsiveness
	for {
		// Use a shorter context for each poll to make it more responsive
		pollCtx, pollCancel := context.WithTimeout(readDeadline, 2*time.Second)
		fetches := client.PollFetches(pollCtx)
		pollCancel()

		if errs := fetches.Errors(); len(errs) > 0 {
			// Only log non-timeout errors to reduce noise
			for _, err := range errs {
				if err.Err.Error() != "context deadline exceeded" {
					color.Red("fetch error: %v", err)
				}
			}
			// Don't continue immediately on error, check if context is done
			select {
			case <-readDeadline.Done():
				color.Blue("Consumer timeout reached")
				return
			default:
				continue
			}
		}

		if fetches.Empty() {
			// Check if we've reached the deadline
			select {
			case <-readDeadline.Done():
				color.Blue("Consumer timeout reached, no messages received")
				return
			default:
				// Continue polling - no messages right now but keep trying
				continue
			}
		}

		// Process all records
		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			for _, record := range p.Records {
				color.Yellow("message at topic/partition/offset %v/%v/%v: %s = %s",
					record.Topic, record.Partition, record.Offset, string(record.Key), string(record.Value))
			}
		})
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
