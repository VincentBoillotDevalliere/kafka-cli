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
- 📁 **Batch Operations** - Process multiple messages from JSON files
- 🏷️ **Topic Management** - List and manage Kafka topics
- 🎨 **Colored Output** - Beautiful, colored terminal output for better readability
- ⚙️ **Flexible Configuration** - Environment variable support with sensible defaults
- 🔄 **Consumer Groups** - Full support for Kafka consumer groups

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

## 📖 Usage

### Producing Messages

#### Single Message
```bash
kafka-cli produce my-topic --message '{"event": "user_signup", "userId": 123}'
```

#### From JSON File
Create a JSON file with message envelopes:

```json
[
  {
    "topic": "user-events",
    "object": {
      "event": "signup",
      "userId": 123,
      "email": "user@example.com"
    }
  },
  {
    "topic": "notifications",
    "object": {
      "type": "welcome_email",
      "userId": 123
    }
  }
]
```

Then produce all messages:

```bash
kafka-cli produce --input messages.json
```

### Consuming Messages

```bash
# Consume from a specific topic
kafka-cli consume user-events

# The consumer will automatically use a generated consumer group
```

### Topic Management

```bash
# List all topics
kafka-cli topic list
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

## 📄 License

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