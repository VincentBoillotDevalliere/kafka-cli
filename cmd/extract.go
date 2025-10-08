package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/VincentBoillotDevalliere/kafka-cli/kafka"
	"github.com/fatih/color"
	kafkaGo "github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"
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

		// 1️⃣ Dial the partition leader
		conn, err := kafkaGo.DialLeader(ctx, "tcp", cfg.Brokers[0], topic, 0)
		if err != nil {
			return fmt.Errorf("failed to dial leader: %w", err)
		}
		defer conn.Close()

		// 2️⃣ Get basic offset information first
		firstOffset, err := conn.ReadFirstOffset()
		if err != nil {
			return fmt.Errorf("failed to read first offset: %w", err)
		}

		lastOffset, err := conn.ReadLastOffset()
		if err != nil {
			return fmt.Errorf("failed to read last offset: %w", err)
		}

		color.Cyan("📊 Topic %s partition 0: %d (first) → %d (last)", topic, firstOffset, lastOffset)

		if firstOffset == lastOffset {
			color.Yellow("⚠️  No messages found in topic %s", topic)
			return fmt.Errorf("no messages found in topic")
		}

		// 3️⃣ Try time-based offset lookup with proper error handling
		color.Blue("🕐 Finding start offset for time: %s", fromStr)
		startOffset, err := conn.ReadOffset(from)
		if err != nil {
			color.Yellow("⚠️  Could not find offset for start time, using first offset")
			startOffset = firstOffset
		} else if startOffset == -1 {
			color.Yellow("⚠️  Start time is before all messages, using first offset")
			startOffset = firstOffset
		}

		color.Blue("🕐 Finding end offset for time: %s", toStr)
		endOffset, err := conn.ReadOffset(to)
		if err != nil {
			color.Yellow("⚠️  Could not find offset for end time, using last offset")
			endOffset = lastOffset
		} else if endOffset == -1 {
			color.Yellow("⚠️  End time is after all messages, using last offset")
			endOffset = lastOffset
		}

		color.Cyan("📊 Time-based offset range: %d → %d", startOffset, endOffset)

		if startOffset >= endOffset {
			color.Yellow("⚠️  No messages in the specified time range")
			return fmt.Errorf("no messages found in time range %s to %s", fromStr, toStr)
		}

		messageCount := endOffset - startOffset
		color.Green("🎯 Will read approximately %d messages", messageCount)

		// 3️⃣ Create reader and set to start offset
		reader := kafkaGo.NewReader(kafkaGo.ReaderConfig{
			Brokers:   cfg.Brokers,
			Topic:     topic,
			Partition: 0,
		})
		defer func() {
			if closeErr := reader.Close(); closeErr != nil {
				color.Red("Warning: failed to close reader: %v", closeErr)
			}
		}()

		// Set the reader to start from our calculated offset
		if setErr := reader.SetOffset(startOffset); setErr != nil {
			return fmt.Errorf("failed to set reader offset to %d: %w", startOffset, setErr)
		}

		color.Blue("✅ Reader positioned at offset %d", startOffset)

		// 4️⃣ Read messages until endOffset (no timestamp filtering needed)
		var messages []MessageEnvelope
		readCount := 0
		expectedMessages := endOffset - startOffset

		color.Cyan("🔍 Reading %d messages from offset %d to %d...", expectedMessages, startOffset, endOffset)

		// Set a reasonable timeout based on expected message count
		timeoutDuration := 10*time.Second + time.Duration(expectedMessages)*time.Millisecond
		readCtx, cancel := context.WithTimeout(ctx, timeoutDuration)
		defer cancel()

		for {
			m, readErr := reader.ReadMessage(readCtx)
			if readErr != nil {
				color.Yellow("📝 Finished reading: %v", readErr)
				break // probably EOF or timeout
			}

			readCount++
			if readCount%1000 == 0 || readCount <= 10 {
				color.Blue("📊 Read message %d/%d at offset %d", readCount, expectedMessages, m.Offset)
			}

			if m.Offset >= endOffset {
				color.Blue("🛑 Reached end offset %d", endOffset)
				break
			}

			// Convert headers
			headers := make(map[string]string)
			for _, h := range m.Headers {
				headers[h.Key] = string(h.Value)
			}

			// Parse message body
			var body map[string]interface{}
			if jsonErr := json.Unmarshal(m.Value, &body); jsonErr != nil {
				body = map[string]interface{}{"raw": string(m.Value)}
			}

			messages = append(messages, MessageEnvelope{
				Topic:   topic,
				Headers: headers,
				Message: body,
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
