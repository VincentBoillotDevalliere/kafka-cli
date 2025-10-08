# Kafka CLI 🚀

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/VincentBoillotDevalliere/kafka-cli?style=flat)](https://github.com/VincentBoillotDevalliere/kafka-cli/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/VincentBoillotDevalliere/kafka-cli)](https://goreportcard.com/report/github.com/VincentBoillotDevalliere/kafka-cli)
[![Build Status](https://img.shields.io/github/actions/workflow/status/VincentBoillotDevalliere/kafka-cli/ci.yml?branch=main)](https://github.com/VincentBoillotDevalliere/kafka-cli/actions)

A powerful and user-friendly command-line tool for interacting with Apache Kafka. Built with Go and designed for developers who need a simple yet robust way to produce, consume, and manage Kafka topics from the terminal.

## ✨ Features

- 🎯 **Simple CLI Interface** - Intuitive commands for all Kafka operations
- 📤 **Message Production** - Send messages to Kafka topics with ease
- 📥 **Message Consumption** - Consume messages from topics with real-time output
- � **Message Extraction** - Extract messages by time range with timezone support
- �📁 **Batch Operations** - Process multiple messages from JSON files
- 🏷️ **Topic Management** - List and manage Kafka topics
- � **Timezone Support** - Work with multiple timezones for time-based operations
- �🎨 **Colored Output** - Beautiful, colored terminal output for better readability
- ⚙️ **Flexible Configuration** - Environment variable support with sensible defaults
- 🔄 **Consumer Groups** - Full support for Kafka consumer groups
- 🚀 **Efficient Processing** - Time-based offset lookup for optimal performance

## 🛠️ Installation

### Using Go Install

```bash
go install github.com/VincentBoillotDevalliere/kafka-cli@latest
```

### From Source

```bash
git clone https://github.com/VincentBoillotDevalliere/kafka-cli.git
cd kafka-cli
go build -o kafka-cli
```

## 🚀 Quick Start

### 1. Configure Kafka Connection

Set up your Kafka broker connection using environment variables:

```bash
export KAFKA_BROKERS="localhost:9092"
```

Or create a `.env` file in your project directory:

```env
KAFKA_BROKERS=localhost:9092,broker2:9092
```

### 2. List Available Topics

```bash
kafka-cli topic list
```

### 3. Produce a Message

```bash
kafka-cli produce my-topic --message '{"user": "john", "action": "login"}'
```

### 4. Consume Messages

```bash
kafka-cli consume my-topic
```

### 5. Extract Messages by Time

```bash
# Extract last hour using local timezone
kafka-cli extract --topic my-topic --from "2025-10-08 15:00:00" --to "2025-10-08 16:00:00" -o messages.json
```

## 📝 Command Examples

### Complete Workflow Example

```bash
# 1. Check available topics
kafka-cli topic list

# 2. Produce some test messages
echo '{"event": "user_login", "userId": 123, "timestamp": "2025-10-08T15:30:00Z"}' | \
kafka-cli produce user-events --message "$(cat)"

# 3. Produce from a batch file
cat > batch_messages.json << 'EOF'
[
  {
    "Topic": "user-events",
    "Headers": {"source": "web", "version": "1.0"},
    "Message": {"event": "signup", "userId": 456}
  },
  {
    "Topic": "user-events", 
    "Headers": {"source": "mobile", "version": "2.1"},
    "Message": {"event": "purchase", "userId": 456, "amount": 99.99}
  }
]
EOF

kafka-cli produce -i batch_messages.json

# 4. Consume messages (Ctrl+C to stop)
kafka-cli consume user-events

# 5. Extract messages from specific time window
kafka-cli extract --topic user-events \
  --from "2025-10-08T15:00:00+02:00" \
  --to "2025-10-08T16:00:00+02:00" \
  -o extracted_events.json

# 6. Check the extracted file
cat extracted_events.json | jq '.[].Message.event'
```

### Production Use Cases

```bash
# Monitor errors in production
kafka-cli consume error-logs --group monitoring-team

# Extract audit logs for compliance
kafka-cli extract --topic audit-logs \
  --from "2025-10-01T00:00:00Z" \
  --to "2025-10-31T23:59:59Z" \
  -o october_audit.json

# Batch process user events
kafka-cli produce -i daily_user_events.json

# Debug specific time window
kafka-cli extract --topic app-logs \
  --from "2025-10-08 14:30:00" \
  --to "2025-10-08 14:45:00" \
  -o incident_logs.json
```

## 📖 Usage

### 📤 Producing Messages

#### Single Message
```bash
# Produce a single message to a topic
kafka-cli produce my-topic --message '{"event": "user_signup", "userId": 123}'

# Alternative syntax
kafka-cli produce --message '{"event": "login", "userId": 456}' my-topic
```

#### Batch Production from JSON File
Create a JSON file with message envelopes:

```json
[
  {
    "Topic": "user-events",
    "Headers": {
      "source": "web-app",
      "correlation-id": "abc-123"
    },
    "Message": {
      "event": "signup",
      "userId": 123,
      "email": "user@example.com"
    }
  },
  {
    "Topic": "notifications",
    "Headers": {
      "priority": "high"
    },
    "Message": {
      "type": "welcome_email",
      "userId": 123
    }
  }
]
```

Then produce all messages:

```bash
# Produce from file
kafka-cli produce --input messages.json

# Alternative short form
kafka-cli produce -i messages.json
```

**Output Example:**
```
🚀 Kafka Producer
📂 Reading file: messages.json
🧾 Found 2 messages
✅ Produced message #1 to topic 'user-events'
✅ Produced message #2 to topic 'notifications'
🎉 Done!
```

### 📥 Consuming Messages

```bash
# Consume from a specific topic (starts from latest by default)
kafka-cli consume user-events

# Consume with specific consumer group
kafka-cli consume user-events --group my-consumer-group

# Consume from beginning of topic
kafka-cli consume user-events --from-beginning
```

### 📊 Extracting Messages (Time-based)

Extract messages from a topic within a specific time range and save them to a JSON file:

#### Basic Extraction
```bash
# Extract messages from the last 15 minutes (default)
kafka-cli extract --topic user-events -o extracted.json

# Extract with specific time range (UTC)
kafka-cli extract --topic user-events \
  --from "2025-10-08T10:00:00Z" \
  --to "2025-10-08T11:00:00Z" \
  -o morning-events.json
```

#### Timezone Support
```bash
# Using your local timezone (no timezone specified)
kafka-cli extract --topic user-events \
  --from "2025-10-08 15:00:00" \
  --to "2025-10-08 16:00:00" \
  -o local-time.json

# Using explicit timezone (CEST/UTC+2)
kafka-cli extract --topic user-events \
  --from "2025-10-08T15:00:00+02:00" \
  --to "2025-10-08T16:00:00+02:00" \
  -o cest-events.json

# Different timezones
kafka-cli extract --topic logs \
  --from "2025-10-08T09:00:00-04:00" \  # EDT (New York)
  --to "2025-10-08T10:00:00-04:00" \
  -o ny-morning.json
```

**Output Example:**
```
🕒 Extract window: from 2025-10-08T15:00:00+02:00 to 2025-10-08T16:00:00+02:00
🌍 Using timezone: Local
📊 Topic user-events partition 0: 0 (first) → 1543 (last)
🕐 Finding start offset for time: 2025-10-08T15:00:00+02:00
🕐 Finding end offset for time: 2025-10-08T16:00:00+02:00
📊 Time-based offset range: 1200 → 1350
🎯 Will read approximately 150 messages
✅ Reader positioned at offset 1200
🔍 Reading 150 messages from offset 1200 to 1350...
📊 Total messages extracted: 147
✅ Extracted 147 messages → cest-events.json
```

#### Supported Time Formats
```bash
# RFC3339 with timezone (recommended)
--from "2025-10-08T15:00:00+02:00"    # CEST
--from "2025-10-08T13:00:00Z"         # UTC  
--from "2025-10-08T09:00:00-04:00"    # EDT

# Local timezone (automatic)
--from "2025-10-08T15:00:00"          # Uses local timezone
--from "2025-10-08 15:00:00"          # Space separator
--from "2025-10-08T15:00"             # Without seconds
```

### 🏷️ Topic Management

```bash
# List all topics
kafka-cli topic list

# Get detailed topic information
kafka-cli topic describe my-topic
```

## ⚙️ Configuration

Kafka CLI uses environment variables for configuration. You can set these in your shell or use a `.env` file:

| Variable | Description | Default |
|----------|-------------|---------|
| `KAFKA_BROKERS` | Comma-separated list of Kafka broker addresses | `localhost:9092` |

### Example Configuration

```env
# .env file
KAFKA_BROKERS=broker1:9092,broker2:9092,broker3:9092
```

## 🏗️ Development

### Prerequisites

- Go 1.23 or higher
- Access to a Kafka cluster (local or remote)

### Building from Source

```bash
git clone https://github.com/VincentBoillotDevalliere/kafka-cli.git
cd kafka-cli
go mod download
go build -o kafka-cli
```

### Running Tests

```bash
go test ./...
```

### Linting

We use `golangci-lint` for code quality:

```bash
# Install linter
make install-linter

# Run linter
make lint

# Auto-fix issues
make lint-fix
```

## 📁 Project Structure

```
kafka-cli/
├── cmd/                    # CLI commands
│   ├── consume.go         # Message consumption logic
│   ├── produce.go         # Message production logic  
│   ├── root.go            # Root command and CLI setup
│   └── topic.go           # Topic management commands
├── kafka/                 # Kafka client configuration
│   └── config.go          # Configuration management
├── utils/                 # Utility functions
├── main.go               # Application entry point
├── go.mod                # Go module definition
└── Makefile              # Build and development tasks
```

## 🤝 Contributing

We welcome contributions! Here's how you can help:

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Commit your changes** (`git commit -m 'Add amazing feature'`)
4. **Push to the branch** (`git push origin feature/amazing-feature`)
5. **Open a Pull Request**

### Development Guidelines

- Write clear, commented code
- Include tests for new features
- Run `make lint` before submitting
- Update documentation as needed
- Follow Go best practices

## 📋 Roadmap

- [ ] **Topic Creation/Deletion** - Full topic lifecycle management
- [ ] **Schema Registry Support** - Avro/JSON Schema integration
- [ ] **Interactive Mode** - Real-time interactive CLI mode
- [ ] **Message Filtering** - Advanced filtering and search capabilities
- [ ] **Performance Metrics** - Built-in performance monitoring
- [ ] **Configuration Profiles** - Multiple environment configurations
- [ ] **Docker Support** - Containerized deployment options

## 🐛 Bug Reports & Feature Requests

Found a bug or have a feature request? Please create an issue on GitHub:

[**Create an Issue**](https://github.com/VincentBoillotDevalliere/kafka-cli/issues/new)

## � Troubleshooting

### Common Issues

#### Connection Issues
```bash
# Check if Kafka brokers are accessible
telnet localhost 9092

# Verify your broker configuration
echo $KAFKA_BROKERS
```

#### Time Zone Issues
```bash
# Check your system timezone
date

# Use explicit timezone in commands
kafka-cli extract --topic my-topic \
  --from "2025-10-08T15:00:00+02:00" \
  --to "2025-10-08T16:00:00+02:00"
```

#### No Messages Found
```bash
# Check if topic exists and has messages
kafka-cli topic list

# Try extracting all available messages first
kafka-cli extract --topic my-topic -o all_messages.json

# Check message timestamps
cat all_messages.json | jq '.[0]' # View first message
```

### Performance Tips

- Use time-based extraction instead of consuming all messages for large topics
- Specify appropriate time windows to avoid reading too many messages
- Use batch production for multiple messages
- Monitor consumer lag when consuming in real-time

## �📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - Powerful CLI framework for Go
- [Kafka-Go](https://github.com/segmentio/kafka-go) - High-performance Kafka client library
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Color](https://github.com/fatih/color) - Colorful terminal output

## 📊 Stats

![GitHub stars](https://img.shields.io/github/stars/VincentBoillotDevalliere/kafka-cli?style=social)
![GitHub forks](https://img.shields.io/github/forks/VincentBoillotDevalliere/kafka-cli?style=social)
![GitHub watchers](https://img.shields.io/github/watchers/VincentBoillotDevalliere/kafka-cli?style=social)

---

<div align="center">

**[⬆ Back to Top](#kafka-cli-)**

Made with ❤️ by [VincentBoillotDevalliere](https://github.com/VincentBoillotDevalliere)

</div>