// my-go-backend/post_service/src/posts/messaging/events/kafka/consumer.go
package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"my-go-backend/post_service/internal/config"
	events "my-go-backend/post_service/src/posts/messaging/events"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type EventHandler func(ctx context.Context, eventType string, payload json.RawMessage) error

type KafkaConsumer struct {
	consumer   *kafka.Consumer
	topic      string
	eventBus   events.EventBus
	logger     *slog.Logger
	handlers   map[string]EventHandler
	mu         sync.RWMutex
	cancelFunc context.CancelFunc
}

func NewKafkaConsumer(
	cfg config.KafkaConfig,
	eventBus events.EventBus,
) (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  cfg.Brokers,
		"group.id":           cfg.GroupID,
		"auto.offset.reset":  cfg.AutoOffsetReset,
		"enable.auto.commit": false,
		"isolation.level":    "read_uncommitted",
		"security.protocol":  cfg.SecurityProtocol,
	})
	if err != nil {
		return nil, err
	}

	if err := c.SubscribeTopics([]string{cfg.Topic}, nil); err != nil {
		return nil, err
	}

	logger := slog.With("component", "kafka_consumer")
	logger.Info("Subscribed to Kafka topic", "topic", cfg.Topic, "group_id", cfg.GroupID)

	return &KafkaConsumer{
		consumer: c,
		topic:    cfg.Topic,
		eventBus: eventBus,
		logger:   logger,
		handlers: make(map[string]EventHandler),
	}, nil
}

func (kc *KafkaConsumer) RegisterHandler(eventType string, handler EventHandler) {
	kc.mu.Lock()
	defer kc.mu.Unlock()
	kc.handlers[eventType] = handler
	kc.logger.Debug("Registered handler for event type", "event_type", eventType)
}

func (kc *KafkaConsumer) Start(ctx context.Context) {
	ctx, kc.cancelFunc = context.WithCancel(ctx)
	kc.logger.Info("Starting Kafka consumer loop...")

	for {
		select {
		case <-ctx.Done():
			kc.logger.Info("Shutting down Kafka consumer")
			kc.consumer.Close()
			return
		default:
			ev := kc.consumer.Poll(100) // 100ms timeout
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				kc.processMessage(ctx, e)
			case kafka.Error:
				kc.logger.Error("Kafka error", "error", e.Error())
				if e.Code() == kafka.ErrAllBrokersDown {
					time.Sleep(time.Second)
				}
			default:
				kc.logger.Debug("Ignored Kafka event", "type", e)
			}
		}
	}
}

func (kc *KafkaConsumer) processMessage(ctx context.Context, msg *kafka.Message) {
	// 1. Try to get event_type from headers
	eventType := ""
	for _, h := range msg.Headers {
		if h.Key == "event_type" && h.Value != nil {
			eventType = string(h.Value)
			break
		}
	}

	// 2. If not in headers, parse from message body
	if eventType == "" {
		var temp struct {
			EventType string `json:"event_type"`
		}
		if err := json.Unmarshal(msg.Value, &temp); err == nil {
			eventType = temp.EventType
		}
	}

	if eventType == "" {
		kc.logger.Warn("Missing event_type in headers and body", "offset", msg.TopicPartition.Offset)
		kc.commitMessage(msg)
		return
	}

	messageID := "unknown"
	if msg.Key != nil {
		messageID = string(msg.Key)
	}

	// Pass RAW msg.Value to handler (it will unmarshal fully)
	payload := json.RawMessage(msg.Value)

	// Dispatch
	kc.mu.RLock()
	handler, exists := kc.handlers[eventType]
	kc.mu.RUnlock()

	if !exists {
		kc.logger.Info("No handler for event type", "event_type", eventType)
		kc.commitMessage(msg)
		return
	}

	if err := handler(ctx, eventType, payload); err != nil {
		kc.logger.Error("Handler failed", "event_type", eventType, "error", err)
		return // don't commit on error
	}

	kc.logger.Info("Successfully processed event", "event_type", eventType, "message_id", messageID)
	kc.commitMessage(msg)
}

func (kc *KafkaConsumer) commitMessage(msg *kafka.Message) {
	_, err := kc.consumer.CommitMessage(msg)
	if err != nil {
		kc.logger.Error("Failed to commit offset", "error", err)
	}
}

func (kc *KafkaConsumer) Close() {
	if kc.cancelFunc != nil {
		kc.cancelFunc()
	}
	kc.consumer.Close()
}
