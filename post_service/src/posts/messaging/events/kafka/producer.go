// github.com/alphaxad9/my-go-backend/post_service/src/posts/messaging/events/kafka/producer.go

package kafka

import (
	"context"
	"log/slog"

	"github.com/alphaxad9/my-go-backend/post_service/internal/config"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Producer interface {
	Produce(ctx context.Context, key, value []byte, headers map[string]string) error
	Close()
}

type kafkaProducer struct {
	producer *kafka.Producer
	topic    string
	logger   *slog.Logger
}

func NewKafkaProducer(cfg config.KafkaConfig) (Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":                     cfg.Brokers,
		"enable.idempotence":                    true,
		"acks":                                  "all",
		"retries":                               2147483647,
		"max.in.flight.requests.per.connection": 5,
		"security.protocol":                     cfg.SecurityProtocol,
	})
	if err != nil {
		return nil, err
	}

	logger := slog.With("component", "kafka_producer")
	logger.Info("Kafka producer created with idempotence enabled", "brokers", cfg.Brokers)

	return &kafkaProducer{
		producer: p,
		topic:    cfg.Topic,
		logger:   logger,
	}, nil
}

func (kp *kafkaProducer) Produce(ctx context.Context, key, value []byte, headers map[string]string) error {
	kafkaHeaders := make([]kafka.Header, 0, len(headers))
	for k, v := range headers {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: k, Value: []byte(v)})
	}

	deliveryChan := make(chan kafka.Event, 1)
	err := kp.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &kp.topic, Partition: kafka.PartitionAny},
		Key:            key,
		Value:          value,
		Headers:        kafkaHeaders,
	}, deliveryChan)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-deliveryChan:
		close(deliveryChan)
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			return m.TopicPartition.Error
		}
		kp.logger.Debug("Message delivered", "topic", *m.TopicPartition.Topic, "partition", m.TopicPartition.Partition)
		return nil
	}
}

func (kp *kafkaProducer) Close() {
	kp.producer.Close()
	kp.logger.Info("Kafka producer closed")
}
