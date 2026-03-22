package services

import (
	"log"
	"os"
	"testing"

	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// TestMain acts as the global setup/teardown wrapper for all unit tests in the services package.
// It executes ONCE before the suite starts.
func TestMain(m *testing.M) {
	// 1. Setup an isolated In-Memory SQLite Database specifically for unit testing
	// This ensures zero data loss for the real application.
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to testing database: %v", err)
	}

	// 2. Override the global config.DB with our clean test DB
	config.DB = db
	config.RedisClient = nil // Force Redis fallback paths explicitly for local tests

	// 3. Migrate all schemas required by the test scenarios automatically
	err = db.AutoMigrate(
		&models.User{},
		&models.Medicine{},
		&models.Prescription{},
		&models.Order{},
	)
	if err != nil {
		log.Fatalf("Test DB Migration failed: %v", err)
	}

	log.Println("✅ In-Memory Test DB successfully initialized and migrated.")

	// 4. Run the suite
	code := m.Run()

	// Exit code
	os.Exit(code)
}
