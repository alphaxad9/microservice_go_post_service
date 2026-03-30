package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	external_services "github.com/alphaxad9/my-go-backend/post_service/external/services"
	"github.com/alphaxad9/my-go-backend/post_service/internal/config"
	"github.com/alphaxad9/my-go-backend/post_service/internal/db"
	router "github.com/alphaxad9/my-go-backend/post_service/internal/http"
	postapi "github.com/alphaxad9/my-go-backend/post_service/src/posts/api/controllers"
	postservices "github.com/alphaxad9/my-go-backend/post_service/src/posts/application/posts/services"
	outboxrepos "github.com/alphaxad9/my-go-backend/post_service/src/posts/infra/repositories/outbox"
	postrepos "github.com/alphaxad9/my-go-backend/post_service/src/posts/infra/repositories/posts"
	events "github.com/alphaxad9/my-go-backend/post_service/src/posts/messaging/events"
	posts "github.com/alphaxad9/my-go-backend/post_service/src/posts/messaging/events/posts"
)

func main() {
	cfg := config.Load()

	// Validate required config at startup
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Set Gin mode from config
	gin.SetMode(cfg.GinMode)

	// Initialize PostgreSQL database
	err := db.InitializeDB(
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresDB,
	)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	if err := db.CreateAllTables(); err != nil {
		log.Fatalf("Failed to create database tables: %v", err)
	}

	// === Repositories ===
	postCmdRepo := postrepos.NewPostCommandRepository(db.DB)
	postQueryRepo := postrepos.NewPostQueryRepository(db.DB)
	outboxRepo := outboxrepos.NewOutboxRepository(db.DB)

	// === Application Services ===
	postCmdService := postservices.NewPostCommandService(postCmdRepo, outboxRepo)
	postQueryService := postservices.NewPostQueryService(postQueryRepo)

	// === External HTTP Client for User Service ===
	httpClient := external_services.NewHTTPClient(cfg.InternalAPIKey)
	userClient := external_services.NewUserAPIClient(httpClient, cfg.AuthServiceURL)

	// === Event Bus & Handlers ===
	eventBus := events.NewInMemoryEventBus()
	postEventHandler := posts.NewPostEventHandler(eventBus)
	postEventHandler.RegisterSubscriptions()
	log.Println("Post event handlers registered successfully")

	// === Controllers (API Layer) ===
	// Pass eventBus to command controller if needed for publishing
	postCmdController := postapi.NewPostCommandController(postCmdService) // ← ensure your controller accepts eventBus
	postQueryController := postapi.NewPostQueryController(postQueryService, userClient)

	// === Router ===
	rtr := router.NewRouter(postCmdController, postQueryController)
	engine := router.SetupRouter(cfg, rtr)

	// Graceful shutdown
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: engine,
	}

	go func() {
		log.Printf("Post service starting on port %s in %s mode...", cfg.Port, cfg.GinMode)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
