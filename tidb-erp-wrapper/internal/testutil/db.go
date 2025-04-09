package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/fowlerlee/tidb/tidb-erp-wrapper/config"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/internal/db"

	_ "github.com/go-sql-driver/mysql"
)

// TestDB represents a test database instance
type TestDB struct {
	DB     *db.DBHandler
	dbName string
}

// NewTestDB creates a new test database with a unique name
func NewTestDB() (*TestDB, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}

	// Create a unique test database name
	dbName := fmt.Sprintf("test_db_%d", time.Now().UnixNano())
	cfg.Database.DBName = dbName

	// Connect to MySQL server without database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
	)
	tempDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database server: %v", err)
	}
	defer tempDB.Close()

	// Create test database
	_, err = tempDB.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		return nil, fmt.Errorf("error creating test database: %v", err)
	}

	// Initialize schema
	schema, err := os.ReadFile("../internal/db/schema.sql")
	if err != nil {
		return nil, fmt.Errorf("error reading schema file: %v", err)
	}

	// Connect to the new database
	dbHandler, err := db.NewDBHandler(cfg)
	if err != nil {
		return nil, fmt.Errorf("error connecting to test database: %v", err)
	}

	// Execute schema
	_, err = dbHandler.DB().Exec(string(schema))
	if err != nil {
		return nil, fmt.Errorf("error executing schema: %v", err)
	}

	return &TestDB{
		DB:     dbHandler,
		dbName: dbName,
	}, nil
}

// Cleanup removes the test database
func (tdb *TestDB) Cleanup() error {
	if err := tdb.DB.Close(); err != nil {
		return fmt.Errorf("error closing database connection: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	// Connect to MySQL server without database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
	)
	tempDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("error connecting to database server: %v", err)
	}
	defer tempDB.Close()

	// Drop test database
	_, err = tempDB.Exec("DROP DATABASE IF EXISTS " + tdb.dbName)
	if err != nil {
		return fmt.Errorf("error dropping test database: %v", err)
	}

	return nil
}
