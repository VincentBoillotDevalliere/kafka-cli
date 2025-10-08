/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
    "github.com/fatih/color"
    kafkaGo "github.com/segmentio/kafka-go"
    "github.com/spf13/cobra"

	"github.com/VincentBoillotDevalliere/kafka-cli/kafka"
)

// topicCmd represents the topic command
var topicCmd = &cobra.Command{
	Use:   "topic",
	Short: "Manage Kafka topics",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all topics",
	RunE: func(cmd *cobra.Command, args []string) error {
        color.Cyan("Listing all topics")
		cfg := kafka.LoadConfig()
		conn, err := kafkaGo.Dial("tcp", cfg.Brokers[0])
		if err != nil {
			return err
		}
		defer func() { _ = conn.Close() }()

		partitions, err := conn.ReadPartitions()
		if err != nil {
			return err
		}
		topicsMap := make(map[string]struct{})
		for _, p := range partitions {
			topicsMap[p.Topic] = struct{}{}
		}
        color.Blue("Topics:")
		for topic := range topicsMap {
            color.Yellow(" - %s", topic)
        }
        return nil
    },
}

func init() {
	rootCmd.AddCommand(topicCmd)
	topicCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// topicCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// topicCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
