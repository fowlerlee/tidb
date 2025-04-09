package db

import (
	"database/sql"
	"fmt"
	"github.com/fowlerlee/tidb/tidb-erp-wrapper/config"

	_ "github.com/go-sql-driver/mysql"
)

type DBHandler struct {
	db *sql.DB
}

func NewDBHandler(cfg *config.Config) (*DBHandler, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error connecting to TiDB: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging TiDB: %v", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &DBHandler{db: db}, nil
}

func (h *DBHandler) Close() error {
	return h.db.Close()
}

func (h *DBHandler) DB() *sql.DB {
	return h.db
}
