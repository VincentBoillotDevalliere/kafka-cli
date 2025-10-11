# Kafka CLI Configuration Examples

## Local Kafka Configuration (default)
```bash
export KAFKA_BROKERS="localhost:9092"
export KAFKA_USE_AWS_IAM="false"
```

## AWS MSK with IAM Authentication
```bash
# Your MSK cluster bootstrap servers
export KAFKA_BROKERS="b-1.your-cluster.abc123.c1.kafka.us-east-1.amazonaws.com:9098,b-2.your-cluster.abc123.c1.kafka.us-east-1.amazonaws.com:9098"

# Enable AWS IAM authentication for MSK
export KAFKA_USE_AWS_IAM="true"

# AWS region where your MSK cluster is located
export AWS_REGION="us-east-1"

# AWS credentials (can be set via various methods)
# Option 1: Environment variables
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"

# Option 2: AWS profile
export AWS_PROFILE="your-profile"

# Option 3: Use IAM role (if running on EC2)
# No additional env vars needed - will use instance role automatically
```

## .env File Example
Create a `.env` file in your project directory:

```env
KAFKA_BROKERS=b-1.your-cluster.abc123.c1.kafka.us-east-1.amazonaws.com:9098,b-2.your-cluster.abc123.c1.kafka.us-east-1.amazonaws.com:9098
KAFKA_USE_AWS_IAM=true
AWS_REGION=us-east-1
AWS_PROFILE=your-profile
```

## Usage Examples

### Local Kafka
```bash
./kafka-cli topic list
./kafka-cli produce my-topic -m '{"message": "hello world"}'
./kafka-cli consume my-topic
```

### AWS MSK with IAM
```bash
# Set environment variables
export KAFKA_USE_AWS_IAM=true
export KAFKA_BROKERS="your-msk-brokers:9098"
export AWS_REGION="us-east-1"

# Use the CLI normally
./kafka-cli topic list
./kafka-cli produce my-topic -m '{"message": "hello from aws msk"}'
./kafka-cli consume my-topic
./kafka-cli extract --topic my-topic --output messages.json
```

## Required IAM Permissions

Your AWS IAM user/role needs the following permissions for MSK:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "kafka-cluster:Connect",
                "kafka-cluster:AlterCluster",
                "kafka-cluster:DescribeCluster"
            ],
            "Resource": "arn:aws:kafka:us-east-1:123456789012:cluster/your-cluster-name/*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "kafka-cluster:*Topic*",
                "kafka-cluster:WriteData",
                "kafka-cluster:ReadData"
            ],
            "Resource": "arn:aws:kafka:us-east-1:123456789012:topic/your-cluster-name/*/*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "kafka-cluster:AlterGroup",
                "kafka-cluster:DescribeGroup"
            ],
            "Resource": "arn:aws:kafka:us-east-1:123456789012:group/your-cluster-name/*/*"
        }
    ]
}
```