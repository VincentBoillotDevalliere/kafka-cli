package cmd

import (
	"context"

	"github.com/fatih/color"
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

		// Create admin client using utility function
		client, adminClient, err := cfg.NewAdminClient()
		if err != nil {
			return err
		}
		defer client.Close()

		// List topics using admin client
		topicsMetadata, err := adminClient.ListTopics(context.Background())
		if err != nil {
			return err
		}

		color.Blue("Topics:")
		for topicName := range topicsMetadata {
			color.Yellow(" - %s", topicName)
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
