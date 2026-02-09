# Kafka Integration Guide

## Overview

This document provides detailed information about Kafka integration in the IDP platform for event-driven service-to-service communication.

## Why Kafka?

Kafka provides:
- **Asynchronous Communication**: Decouple services for better scalability
- **Event Streaming**: Real-time data pipelines and event processing
- **Reliability**: Persistent, replicated message storage
- **Scalability**: Handle millions of messages per second
- **Replay Capability**: Reprocess events from any point in time

## Use Cases

### 1. Event-Driven Architecture
```
Service A → Kafka Topic → Service B, C, D (subscribers)
```
- User registration events
- Order processing events
- Notification triggers
- Data synchronization

### 2. Change Data Capture (CDC)
```
Database → Kafka → Analytics/Search/Cache
```
- Real-time data replication
- Search index updates
- Cache invalidation

### 3. Log Aggregation
```
Multiple Services → Kafka → Log Processing → Storage
```
- Centralized logging
- Audit trails
- Analytics

## Kafka in the IDP

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Kafka Cluster                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Broker 1   │  │   Broker 2   │  │   Broker 3   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                              │
│  Topics:                                                     │
│  - user-events (partitions: 3, replication: 3)             │
│  - order-events (partitions: 5, replication: 3)            │
│  - notification-events (partitions: 3, replication: 3)     │
└─────────────────────────────────────────────────────────────┘
         ▲                    │                    ▼
         │                    │                    │
    ┌─────────┐          ┌─────────┐         ┌─────────┐
    │Service A│          │Service B│         │Service C│
    │Producer │          │Consumer │         │Consumer │
    └─────────┘          └─────────┘         └─────────┘
```

### Kafka Operator (Strimzi)

The platform uses **Strimzi** to manage Kafka clusters on Kubernetes:

- Automated Kafka cluster deployment
- Topic management via CRDs
- User and ACL management
- Monitoring and metrics
- Rolling updates with zero downtime

### Crossplane Integration

Services can request Kafka topics via Crossplane:

```yaml
apiVersion: platform.example.com/v1alpha1
kind: GoService
metadata:
  name: user-service
spec:
  serviceName: user-service
  kafka:
    enabled: true
    topics:
      - name: user-events
        partitions: 3
        replicationFactor: 3
        config:
          retention.ms: "604800000"  # 7 days
      - name: user-notifications
        partitions: 5
        replicationFactor: 3
```

This automatically creates:
- Kafka topics with specified configuration
- Kafka user with appropriate ACLs
- Kubernetes secrets with connection details

## Go Service Integration

### 1. Kafka Client Setup

**Dependencies** (`go.mod`):
```go
require (
    github.com/confluentinc/confluent-kafka-go/v2/kafka v2.3.0
    // OR
    github.com/IBM/sarama v1.42.1
)
```

### 2. Producer Implementation

**File**: `internal/pkg/kafka/producer.go`

```go
package kafka

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/confluentinc/confluent-kafka-go/v2/kafka"
    "github.com/rs/zerolog/log"
)

type Producer struct {
    producer *kafka.Producer
}

func NewProducer(brokers string) (*Producer, error) {
    p, err := kafka.NewProducer(&kafka.ConfigMap{
        "bootstrap.servers": brokers,
        "client.id":         "go-service",
        "acks":              "all",
        "retries":           3,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create producer: %w", err)
    }
    
    return &Producer{producer: p}, nil
}

func (p *Producer) Publish(ctx context.Context, topic string, key string, value interface{}) error {
    data, err := json.Marshal(value)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }
    
    err = p.producer.Produce(&kafka.Message{
        TopicPartition: kafka.TopicPartition{
            Topic:     &topic,
            Partition: kafka.PartitionAny,
        },
        Key:   []byte(key),
        Value: data,
    }, nil)
    
    if err != nil {
        return fmt.Errorf("failed to produce message: %w", err)
    }
    
    return nil
}

func (p *Producer) Close() {
    p.producer.Flush(15 * 1000)
    p.producer.Close()
}
```

### 3. Consumer Implementation

**File**: `internal/pkg/kafka/consumer.go`

```go
package kafka

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/confluentinc/confluent-kafka-go/v2/kafka"
    "github.com/rs/zerolog/log"
)

type Consumer struct {
    consumer *kafka.Consumer
}

type MessageHandler func(ctx context.Context, key, value []byte) error

func NewConsumer(brokers, groupID string, topics []string) (*Consumer, error) {
    c, err := kafka.NewConsumer(&kafka.ConfigMap{
        "bootstrap.servers": brokers,
        "group.id":          groupID,
        "auto.offset.reset": "earliest",
        "enable.auto.commit": false,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create consumer: %w", err)
    }
    
    err = c.SubscribeTopics(topics, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to subscribe to topics: %w", err)
    }
    
    return &Consumer{consumer: c}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            msg, err := c.consumer.ReadMessage(-1)
            if err != nil {
                log.Error().Err(err).Msg("Error reading message")
                continue
            }
            
            err = handler(ctx, msg.Key, msg.Value)
            if err != nil {
                log.Error().Err(err).Msg("Error handling message")
                // Send to DLQ or retry
                continue
            }
            
            // Commit offset after successful processing
            _, err = c.consumer.CommitMessage(msg)
            if err != nil {
                log.Error().Err(err).Msg("Error committing offset")
            }
        }
    }
}

func (c *Consumer) Close() error {
    return c.consumer.Close()
}
```

### 4. Event Publishing

**File**: `internal/messaging/producer/user_events.go`

```go
package producer

import (
    "context"
    "fmt"
    
    "github.com/yourorg/yourservice/internal/model/events"
    "github.com/yourorg/yourservice/internal/pkg/kafka"
)

type UserEventProducer struct {
    producer *kafka.Producer
    topic    string
}

func NewUserEventProducer(producer *kafka.Producer) *UserEventProducer {
    return &UserEventProducer{
        producer: producer,
        topic:    "user-events",
    }
}

func (p *UserEventProducer) PublishUserCreated(ctx context.Context, event *events.UserCreatedEvent) error {
    return p.producer.Publish(ctx, p.topic, event.UserID, event)
}

func (p *UserEventProducer) PublishUserUpdated(ctx context.Context, event *events.UserUpdatedEvent) error {
    return p.producer.Publish(ctx, p.topic, event.UserID, event)
}

func (p *UserEventProducer) PublishUserDeleted(ctx context.Context, event *events.UserDeletedEvent) error {
    return p.producer.Publish(ctx, p.topic, event.UserID, event)
}
```

### 5. Event Consumption

**File**: `internal/messaging/consumer/user_events.go`

```go
package consumer

import (
    "context"
    "encoding/json"
    
    "github.com/rs/zerolog/log"
    "github.com/yourorg/yourservice/internal/model/events"
    "github.com/yourorg/yourservice/internal/service"
)

type UserEventConsumer struct {
    userService *service.UserService
}

func NewUserEventConsumer(userService *service.UserService) *UserEventConsumer {
    return &UserEventConsumer{
        userService: userService,
    }
}

func (c *UserEventConsumer) HandleMessage(ctx context.Context, key, value []byte) error {
    var event events.UserEvent
    if err := json.Unmarshal(value, &event); err != nil {
        return err
    }
    
    switch event.Type {
    case "user.created":
        var userCreated events.UserCreatedEvent
        if err := json.Unmarshal(value, &userCreated); err != nil {
            return err
        }
        return c.handleUserCreated(ctx, &userCreated)
        
    case "user.updated":
        var userUpdated events.UserUpdatedEvent
        if err := json.Unmarshal(value, &userUpdated); err != nil {
            return err
        }
        return c.handleUserUpdated(ctx, &userUpdated)
        
    default:
        log.Warn().Str("event_type", event.Type).Msg("Unknown event type")
        return nil
    }
}

func (c *UserEventConsumer) handleUserCreated(ctx context.Context, event *events.UserCreatedEvent) error {
    log.Info().Str("user_id", event.UserID).Msg("Processing user created event")
    // Business logic here
    return nil
}

func (c *UserEventConsumer) handleUserUpdated(ctx context.Context, event *events.UserUpdatedEvent) error {
    log.Info().Str("user_id", event.UserID).Msg("Processing user updated event")
    // Business logic here
    return nil
}
```

### 6. Event Models

**File**: `internal/model/events/user_events.go`

```go
package events

import "time"

type UserEvent struct {
    Type      string    `json:"type"`
    Timestamp time.Time `json:"timestamp"`
}

type UserCreatedEvent struct {
    UserEvent
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Name   string `json:"name"`
}

type UserUpdatedEvent struct {
    UserEvent
    UserID string `json:"user_id"`
    Email  string `json:"email,omitempty"`
    Name   string `json:"name,omitempty"`
}

type UserDeletedEvent struct {
    UserEvent
    UserID string `json:"user_id"`
}
```

## Configuration

### Environment Variables

```bash
# Kafka configuration
APP_KAFKA_BROKERS=kafka-cluster-kafka-bootstrap.kafka:9092
APP_KAFKA_GROUP_ID=user-service-group
APP_KAFKA_TOPICS=user-events,notification-events
APP_KAFKA_SASL_ENABLED=true
APP_KAFKA_SASL_USERNAME=user-service
APP_KAFKA_SASL_PASSWORD=secret
```

### Config File (`config.yaml`)

```yaml
kafka:
  brokers: localhost:9092
  groupId: user-service-group
  topics:
    - user-events
    - notification-events
  sasl:
    enabled: false
    username: ""
    password: ""
  consumer:
    autoOffsetReset: earliest
    enableAutoCommit: false
  producer:
    acks: all
    retries: 3
```

## Local Development

### Docker Compose

**File**: `local-dev/docker-compose.yaml`

```yaml
version: '3.8'

services:
  zookeeper:
    image: confluentinc/cp-zookeeper:7.5.0
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"

  kafka:
    image: confluentinc/cp-kafka:7.5.0
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"

  kafka-ui:
    image: provectuslabs/kafka-ui:latest
    depends_on:
      - kafka
    ports:
      - "8090:8080"
    environment:
      KAFKA_CLUSTERS_0_NAME: local
      KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS: kafka:9092

  postgres:
    image: postgres:15-alpine
    # ... (existing config)

  redis:
    image: redis:7-alpine
    # ... (existing config)
```

### Testing Kafka Locally

```bash
# Start all services
make dev

# Produce a test message
kafka-console-producer --broker-list localhost:9092 --topic user-events
> {"type":"user.created","user_id":"123","email":"test@example.com"}

# Consume messages
kafka-console-consumer --bootstrap-server localhost:9092 --topic user-events --from-beginning

# View Kafka UI
open http://localhost:8090
```

## Monitoring

### Prometheus Metrics

The Kafka client exports metrics:

```
kafka_producer_messages_total
kafka_producer_errors_total
kafka_consumer_messages_total
kafka_consumer_lag
kafka_consumer_errors_total
```

### Grafana Dashboards

Pre-built dashboards for:
- Kafka cluster health
- Topic throughput
- Consumer lag
- Producer/consumer errors

## Best Practices

### 1. Event Design
- Use clear event names: `user.created`, `order.placed`
- Include event version for schema evolution
- Add correlation IDs for tracing
- Include timestamps

### 2. Error Handling
- Implement retry logic with exponential backoff
- Use dead letter queues (DLQ) for failed messages
- Log all errors with context

### 3. Idempotency
- Design consumers to be idempotent
- Use unique message IDs
- Track processed messages

### 4. Performance
- Batch messages when possible
- Use appropriate partition keys
- Monitor consumer lag
- Tune consumer group size

### 5. Security
- Enable SASL/SSL in production
- Use ACLs to restrict topic access
- Rotate credentials regularly

## Common Patterns

### 1. Transactional Outbox
```go
// Save to database and publish event atomically
tx, _ := db.Begin()
userRepo.Create(tx, user)
outboxRepo.Create(tx, event)
tx.Commit()

// Background worker publishes from outbox
```

### 2. Saga Pattern
```go
// Orchestrate distributed transactions
orderService.CreateOrder() → Kafka
paymentService.ProcessPayment() → Kafka
inventoryService.ReserveItems() → Kafka
```

### 3. CQRS
```go
// Command side
userService.CreateUser() → Kafka → user.created

// Query side
userEventConsumer → Update read model
```

## Troubleshooting

### Consumer Lag
```bash
# Check consumer lag
kafka-consumer-groups --bootstrap-server localhost:9092 --describe --group user-service-group
```

### Topic Issues
```bash
# List topics
kafka-topics --bootstrap-server localhost:9092 --list

# Describe topic
kafka-topics --bootstrap-server localhost:9092 --describe --topic user-events
```

### Connection Issues
- Verify broker addresses
- Check network connectivity
- Validate SASL credentials
- Review firewall rules

## Resources

- [Strimzi Documentation](https://strimzi.io/docs/)
- [Confluent Kafka Go Client](https://docs.confluent.io/kafka-clients/go/current/overview.html)
- [Kafka Best Practices](https://kafka.apache.org/documentation/#bestpractices)
