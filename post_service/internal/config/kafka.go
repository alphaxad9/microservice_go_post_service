// github.com/alphaxad9/my-go-backend/post_service/internal/config/kafka.go
package config

import (
	"strconv"
	"strings"
)

type KafkaConfig struct {
	Brokers          string
	Topic            string
	GroupID          string // ← ADD THIS
	Retries          int
	EnableAutoCommit bool
	AutoOffsetReset  string
	SecurityProtocol string
	PollTimeout      float64
}

func LoadKafkaConfig() KafkaConfig {
	return KafkaConfig{
		Brokers:          getEnv("KAFKA_BROKERS", "localhost:9092"),
		Topic:            getEnv("KAFKA_TOPIC", "microservice_one.events"),
		GroupID:          getEnv("KAFKA_GROUP_ID", "post_service_outbox_publisher"), // ← USE IT
		Retries:          mustParseInt(getEnv("KAFKA_RETRIES", "5"), 5),
		EnableAutoCommit: parseBool(getEnv("KAFKA_ENABLE_AUTO_COMMIT", "false")),
		AutoOffsetReset:  getEnv("KAFKA_AUTO_OFFSET_RESET", "earliest"),
		SecurityProtocol: getEnv("KAFKA_SECURITY_PROTOCOL", "PLAINTEXT"),
		PollTimeout:      mustParseFloat(getEnv("KAFKA_POLL_TIMEOUT", "1.0"), 1.0),
	}
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes"
}

func mustParseInt(s string, fallback int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return fallback
}

func mustParseFloat(s string, fallback float64) float64 {
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	return fallback
}
