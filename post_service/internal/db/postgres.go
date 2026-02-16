// Package db handles database connection and schema management.
package db

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var (
	DB   *sql.DB
	once sync.Once
)

// InitializeDB opens a connection pool to the PostgreSQL database.
// It assumes the database already exists.
// This function is idempotent and safe for concurrent calls.
func InitializeDB(host string, port int, user, password, dbname string) error {
	var initErr error
	once.Do(func() {
		connStr := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname,
		)

		DB, initErr = sql.Open("postgres", connStr)
		if initErr != nil {
			initErr = fmt.Errorf("failed to open database connection: %w", initErr)
			return
		}

		// Verify the connection
		if pingErr := DB.Ping(); pingErr != nil {
			DB.Close()
			DB = nil
			initErr = fmt.Errorf("failed to ping database: %w", pingErr)
			return
		}

		log.Println("Connected to PostgreSQL successfully")
	})
	return initErr
}

// Close closes the database connection pool.
func Close() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			log.Println("PostgreSQL connection closed")
		}
		DB = nil
	}
}

// CreateAllTables creates all required tables.
func CreateAllTables() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	if err := CreatePostsTable(); err != nil {
		return fmt.Errorf("failed to create posts table: %w", err)
	}

	if err := CreateOutboxTable(); err != nil {
		return fmt.Errorf("failed to create outbox table: %w", err)
	}

	log.Println("All database tables created successfully")
	return nil
}
