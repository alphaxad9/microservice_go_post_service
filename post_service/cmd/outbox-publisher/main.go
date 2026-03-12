// github.com/alphaxad9/my-go-backend/post_service/cmd/outbox-publisher/main.go
package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alphaxad9/my-go-backend/post_service/internal/config"
	"github.com/alphaxad9/my-go-backend/post_service/internal/db"
	outboxrepos "github.com/alphaxad9/my-go-backend/post_service/src/posts/infra/repositories/outbox"
	kafkamessaging "github.com/alphaxad9/my-go-backend/post_service/src/posts/messaging/events/kafka"
)

func main() {
	// Load configs
	appCfg := config.Load()
	kafkaCfg := config.LoadKafkaConfig()

	// Initialize DB
	if err := db.InitializeDB(
		appCfg.PostgresHost,
		appCfg.PostgresPort,
		appCfg.PostgresUser,
		appCfg.PostgresPassword,
		appCfg.PostgresDB,
	); err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}
	defer db.Close()

	// Create Kafka producer
	producer, err := kafkamessaging.NewKafkaProducer(kafkaCfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Create outbox repo
	outboxRepo := outboxrepos.NewOutboxRepository(db.DB)

	// Create publisher
	publisher := kafkamessaging.NewOutboxPublisher(
		outboxRepo,
		producer,
		3*time.Second, // poll every 3 seconds
		20,            // batch size
		kafkaCfg.Retries,
	)

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		slog.Info("Shutdown signal received")
		cancel()
	}()

	slog.Info("Starting Kafka outbox publisher")
	publisher.Start(ctx)
	slog.Info("Kafka outbox publisher stopped")
}

// # From project root (github.com/alphaxad9/my-go-backend/post_service/)
// go mod tidy

// # Build the publisher binary
// go build -o bin/outbox-publisher ./post_service/cmd/outbox-publisher

// # Run it (make sure .env is loaded or set via export)
// ./bin/outbox-publisher
