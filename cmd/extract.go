package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/VincentBoillotDevalliere/kafka-cli/kafka"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	topic   string
	fromStr string
	toStr   string
	output  string
)

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract messages from a Kafka topic to a file",
	Long: `Extract messages from a specified Kafka topic within an optional time range and save them to a file.
You can specify the time range using --from and --to flags in RFC3339 format.
The output file can be specified with the --output flag. If not provided, it defaults to 'extracted_messages.json'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		defaultWindows := 15 // 15 minutes
		if topic == "" {
			return fmt.Errorf("Topic input in mandatory")
		}
		if fromStr == "" || toStr == "" {
			color.HiYellow("fromStr or toStr undefined, backup to default values: %d minutes", defaultWindows)
			fromStr = time.Now().Add(-time.Duration(defaultWindows) * time.Minute).Format(time.RFC3339)
			toStr = time.Now().Format(time.RFC3339)
		}
		color.Yellow("🕒 Extract window: from %s to %s", fromStr, toStr)

		// Parse time with flexible timezone support
		from, err := parseTimeWithTimezone(fromStr)
		if err != nil {
			return fmt.Errorf("invalid --from: %v", err)
		}
		to, err := parseTimeWithTimezone(toStr)
		if err != nil {
			return fmt.Errorf("invalid --to: %v", err)
		}

		color.Cyan("🌍 Using timezone: %s", from.Location())

		cfg := kafka.LoadConfig()
		ctx := context.Background()

		// 1️⃣ Create admin client using utility function
		client, adminClient, err := cfg.NewAdminClient()
		if err != nil {
			return fmt.Errorf("failed to create kafka client: %w", err)
		}
		defer client.Close()

		// 2️⃣ Get partition metadata and basic offset information
		topicDetails, err := adminClient.ListTopics(ctx, topic)
		if err != nil {
			return fmt.Errorf("failed to get topic details: %w", err)
		}

		topicInfo, exists := topicDetails[topic]
		if !exists {
			return fmt.Errorf("topic %s does not exist", topic)
		}

		if len(topicInfo.Partitions) == 0 {
			return fmt.Errorf("topic %s has no partitions", topic)
		}

		// Get earliest and latest offsets for partition 0 using timestamp lookups
		earliestOffsets, err := adminClient.ListOffsetsAfterMilli(ctx, 0, topic) // 0 = earliest
		if err != nil {
			return fmt.Errorf("failed to get earliest offsets: %w", err)
		}

		latestOffsets, err := adminClient.ListOffsetsAfterMilli(ctx, -1, topic) // -1 = latest
		if err != nil {
			return fmt.Errorf("failed to get latest offsets: %w", err)
		}

		var firstOffset, lastOffset int64
		if offsets, exists := earliestOffsets[topic]; exists {
			if partOffset, partExists := offsets[0]; partExists {
				firstOffset = partOffset.Offset
			}
		}

		if offsets, exists := latestOffsets[topic]; exists {
			if partOffset, partExists := offsets[0]; partExists {
				lastOffset = partOffset.Offset
			}
		}

		color.Cyan("📊 Topic %s partition 0: %d (first) → %d (last)", topic, firstOffset, lastOffset)

		if firstOffset >= lastOffset {
			color.Yellow("⚠️  No messages found in topic %s", topic)
			return fmt.Errorf("no messages found in topic")
		}

		// 3️⃣ Try time-based offset lookup with proper error handling
		color.Blue("🕐 Finding start offset for time: %s", fromStr)
		startOffsetResults, err := adminClient.ListOffsetsAfterMilli(ctx, from.UnixMilli(), topic)
		var startOffset int64 = firstOffset
		if err == nil && startOffsetResults != nil {
			if offsets, exists := startOffsetResults[topic]; exists {
				if partOffset, partExists := offsets[0]; partExists {
					startOffset = partOffset.Offset
				}
			}
		} else {
			color.Yellow("⚠️  Could not find offset for start time, using first offset")
		}

		color.Blue("🕐 Finding end offset for time: %s", toStr)
		endOffsetResults, err := adminClient.ListOffsetsAfterMilli(ctx, to.UnixMilli(), topic)
		var endOffset int64 = lastOffset
		if err == nil && endOffsetResults != nil {
			if offsets, exists := endOffsetResults[topic]; exists {
				if partOffset, partExists := offsets[0]; partExists {
					endOffset = partOffset.Offset
				}
			}
		} else {
			color.Yellow("⚠️  Could not find offset for end time, using last offset")
		}

		color.Cyan("📊 Time-based offset range: %d → %d", startOffset, endOffset)

		if startOffset >= endOffset {
			color.Yellow("⚠️  No messages in the specified time range")
			return fmt.Errorf("no messages found in time range %s to %s", fromStr, toStr)
		}

		messageCount := endOffset - startOffset
		color.Green("🎯 Will read approximately %d messages", messageCount)

		// 3️⃣ Create partition consumer using utility function
		consumerClient, err := cfg.NewPartitionConsumerClient(topic, 0, startOffset)
		if err != nil {
			return fmt.Errorf("failed to create consumer client: %w", err)
		}
		defer consumerClient.Close()

		color.Blue("✅ Consumer positioned at offset %d", startOffset)

		// 4️⃣ Read messages until endOffset (no timestamp filtering needed)
		var messages []MessageEnvelope
		readCount := 0
		expectedMessages := endOffset - startOffset

		color.Cyan("🔍 Reading %d messages from offset %d to %d...", expectedMessages, startOffset, endOffset)

		// Set a reasonable timeout based on expected message count
		timeoutDuration := 10*time.Second + time.Duration(expectedMessages)*time.Millisecond
		readCtx, cancel := context.WithTimeout(ctx, timeoutDuration)
		defer cancel()

		reachedEnd := false
		for !reachedEnd {
			select {
			case <-readCtx.Done():
				color.Yellow("📝 Finished reading: timeout reached")
				reachedEnd = true
				continue
			default:
			}

			fetches := consumerClient.PollFetches(readCtx)
			if errs := fetches.Errors(); len(errs) > 0 {
				for _, err := range errs {
					if err.Err.Error() != "context deadline exceeded" {
						color.Red("fetch error: %v", err)
					}
				}
				continue
			}

			if fetches.Empty() {
				continue
			}

			// Process records
			fetches.EachPartition(func(p kgo.FetchTopicPartition) {
				for _, record := range p.Records {
					readCount++
					if readCount%1000 == 0 || readCount <= 10 {
						color.Blue("📊 Read message %d/%d at offset %d", readCount, expectedMessages, record.Offset)
					}

					if record.Offset >= endOffset {
						color.Blue("🛑 Reached end offset %d", endOffset)
						reachedEnd = true
						return
					}

					// Convert headers
					headers := make(map[string]string)
					for _, h := range record.Headers {
						headers[h.Key] = string(h.Value)
					}

					// Parse message body
					var body map[string]interface{}
					if jsonErr := json.Unmarshal(record.Value, &body); jsonErr != nil {
						body = map[string]interface{}{"raw": string(record.Value)}
					}

					messages = append(messages, MessageEnvelope{
						Topic:   topic,
						Headers: headers,
						Message: body,
					})
				}
			})
		}

		color.Blue("📊 Total messages extracted: %d", len(messages))
		// write to JSON file
		file, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()

		enc := json.NewEncoder(file)
		enc.SetIndent("", "  ")
		if err := enc.Encode(messages); err != nil {
			return fmt.Errorf("failed to encode messages: %w", err)
		}

		color.Green("✅ Extracted %d messages → %s", len(messages), output)
		return nil
	},
}

// parseTimeWithTimezone parses time strings with flexible timezone support
func parseTimeWithTimezone(timeStr string) (time.Time, error) {
	// List of supported time formats, in order of preference
	formats := []string{
		time.RFC3339,          // "2006-01-02T15:04:05Z07:00" (with timezone)
		time.RFC3339Nano,      // "2006-01-02T15:04:05.999999999Z07:00" (with nanoseconds)
		"2006-01-02T15:04:05", // "2006-01-02T15:04:05" (no timezone - will use local)
		"2006-01-02 15:04:05", // "2006-01-02 15:04:05" (space separator, no timezone)
		"2006-01-02T15:04",    // "2006-01-02T15:04" (no seconds)
		"2006-01-02 15:04",    // "2006-01-02 15:04" (space separator, no seconds)
	}

	// First, try parsing with timezone info
	for _, format := range formats[:2] {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	// If no timezone specified, try parsing without timezone and assume local time
	for _, format := range formats[2:] {
		if t, err := time.ParseInLocation(format, timeStr, time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time %q. Supported formats: RFC3339 (2006-01-02T15:04:05Z07:00), or without timezone (2006-01-02T15:04:05) which will use local timezone", timeStr)
}

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmd.Flags().StringVarP(&topic, "topic", "", "", "topic")
	extractCmd.Flags().StringVarP(&fromStr, "from", "", "", "Start time (RFC3339 format with timezone or local time without timezone)")
	extractCmd.Flags().StringVarP(&toStr, "to", "", "", "End time (RFC3339 format with timezone or local time without timezone)")
	extractCmd.Flags().StringVarP(&output, "output", "o", "", "Optional output file")
}
