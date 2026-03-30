package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database
	PostgresHost     string
	PostgresPort     int
	PostgresDB       string
	PostgresUser     string
	PostgresPassword string

	// Kafka
	KafkaBrokers          []string
	KafkaTopic            string
	KafkaGroupID          string
	KafkaSecurityProtocol string
	KafkaEnableAutoCommit bool
	KafkaAutoOffsetReset  string
	KafkaRetries          int
	KafkaPollTimeout      float64

	// Auth & Security
	JWTSecret        string
	AuthPublicKeyURL string
	AuthPublicKeyTTL int
	AuthServiceURL   string
	InternalAPIKey   string

	// Server
	FrontendURLs []string
	Port         string
	GinMode      string
}

func Load() *Config {
	loadEnvFile()

	return &Config{
		// Database
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnvAsInt("POSTGRES_PORT", 5432),
		PostgresDB:       getEnv("POSTGRES_DB", ""),
		PostgresUser:     getEnv("POSTGRES_USER", ""),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", ""),

		// Kafka
		KafkaBrokers:          getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		KafkaTopic:            getEnv("KAFKA_TOPIC", ""),
		KafkaGroupID:          getEnv("KAFKA_GROUP_ID", ""),
		KafkaSecurityProtocol: getEnv("KAFKA_SECURITY_PROTOCOL", "PLAINTEXT"),
		KafkaEnableAutoCommit: getEnvAsBool("KAFKA_ENABLE_AUTO_COMMIT", false),
		KafkaAutoOffsetReset:  getEnv("KAFKA_AUTO_OFFSET_RESET", "earliest"),
		KafkaRetries:          getEnvAsInt("KAFKA_RETRIES", 5),
		KafkaPollTimeout:      getEnvAsFloat("KAFKA_POLL_TIMEOUT", 1.0),

		// Auth & Security
		JWTSecret:        getEnv("JWT_SECRET", ""),
		AuthPublicKeyURL: getEnv("AUTH_PUBLIC_KEY_URL", ""),
		AuthPublicKeyTTL: getEnvAsInt("AUTH_PUBLIC_KEY_TTL", 300),
		AuthServiceURL:   getEnv("AUTH_SERVICE_URL", ""),
		InternalAPIKey:   getEnv("INTERNAL_API_KEY", ""),

		// Server
		FrontendURLs: getEnvAsSlice("FRONTEND_URL", []string{"http://localhost:3000"}),
		Port:         getEnv("PORT", "8080"),
		GinMode:      getEnv("GIN_MODE", "debug"),
	}
}

func loadEnvFile() {
	// Check if running in Kubernetes
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		log.Println("ℹ️ Running inside Kubernetes, skipping .env loading")
		return
	}

	// Try current working directory
	if err := godotenv.Load(); err == nil {
		log.Println("✅ .env loaded from working directory")
		return
	}

	// Try project root (walk up directories)
	dir, err := os.Getwd()
	if err != nil {
		log.Println("⚠️  Could not determine working directory")
		return
	}

	for i := 0; i < 5; i++ {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err == nil {
				log.Printf("✅ .env loaded from: %s", envPath)
				return
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	log.Println("ℹ️ No .env file found, using system environment variables")
}

// Helper: Get string env var with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if defaultValue != "" {
		log.Printf("⚠️  Environment variable %s not set, using default", key)
		return defaultValue
	}
	log.Printf("❌ Environment variable %s is required but not set", key)
	return ""
}

// Helper: Get int env var with default
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("⚠️  Invalid integer for %s: %s, using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

// Helper: Get float env var with default
func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Printf("⚠️  Invalid float for %s: %s, using default: %f", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

// Helper: Get bool env var with default
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Printf("⚠️  Invalid boolean for %s: %s, using default: %t", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

// Helper: Get slice env var (comma-separated)
func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	// Split by comma and trim spaces
	parts := strings.Split(valueStr, ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return parts
}

// Validate required config (call after Load())
func (c *Config) Validate() error {
	required := map[string]string{
		"POSTGRES_DB":         c.PostgresDB,
		"POSTGRES_USER":       c.PostgresUser,
		"POSTGRES_PASSWORD":   c.PostgresPassword,
		"JWT_SECRET":          c.JWTSecret,
		"KAFKA_TOPIC":         c.KafkaTopic,
		"KAFKA_GROUP_ID":      c.KafkaGroupID,
		"AUTH_PUBLIC_KEY_URL": c.AuthPublicKeyURL,
		"AUTH_SERVICE_URL":    c.AuthServiceURL,
		"INTERNAL_API_KEY":    c.InternalAPIKey,
	}

	missing := []string{}
	for key, value := range required {
		if value == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return &MissingConfigError{Missing: missing}
	}

	return nil
}

type MissingConfigError struct {
	Missing []string
}

func (e *MissingConfigError) Error() string {
	return "missing required environment variables: " + strings.Join(e.Missing, ", ")
}
