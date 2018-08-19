package database

import (
	"database/sql"

	// Driver for sqlite3
	_ "github.com/mattn/go-sqlite3"

	"github.com/hakasec/japanbot-go/bot/config"
)

// DBConnection is an extension of sql.DB
type DBConnection struct {
	*sql.DB

	config *config.DBConfiguration
}

// OpenFromConfig creates a DBConnection from a given DBConfiguration
func OpenFromConfig(config *config.DBConfiguration) (*DBConnection, error) {
	db, err := sql.Open(config.DriverName, config.ConnString)
	if err != nil {
		return nil, err
	}

	return &DBConnection{DB: db, config: config}, nil
}
