// github.com/alphaxad9/my-go-backend/post_service/cmd/kafka-consumer/main.go
package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alphaxad9/my-go-backend/post_service/internal/config"
	events "github.com/alphaxad9/my-go-backend/post_service/src/posts/messaging/events"
	"github.com/alphaxad9/my-go-backend/post_service/src/posts/messaging/events/kafka"
	handlers "github.com/alphaxad9/my-go-backend/post_service/src/posts/messaging/events/posts"
)

func main() {
	// Load environment variables from .env file FIRST
	config.Load()

	kafkaCfg := config.LoadKafkaConfig()

	// Create in-memory event bus
	eventBus := events.NewInMemoryEventBus()

	// CRITICAL: Register your PostEventHandler
	postHandler := handlers.NewPostEventHandler(eventBus)
	postHandler.RegisterSubscriptions() // ← THIS WAS MISSING!

	// Create and start consumer
	consumer, err := kafka.NewKafkaConsumer(kafkaCfg, eventBus)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	// Register dispatchers
	dispatcher := kafka.NewDispatcher(eventBus)
	dispatcher.RegisterHandlers(consumer)

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		slog.Info("Shutdown signal received")
		consumer.Close()
	}()

	slog.Info("Starting Kafka consumer...")
	consumer.Start(ctx)
	slog.Info("Kafka consumer stopped")
}

// # Build consumer
// go build -o bin/kafka-consumer ./post_service/cmd/kafka-consumer

// # Run it
// ./bin/kafka-consumer
