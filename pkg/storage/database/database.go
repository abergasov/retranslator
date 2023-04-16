package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // justifying it
)

type DBConnect struct {
	db *sqlx.DB
}

func InitDBConnect(dbPath string) (DBConnector, error) {
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error connect to db: %w", err)
	}
	return &DBConnect{db}, err
}

func InitMemory() (DBConnector, error) {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("error connect to db: %w", err)
	}
	return &DBConnect{db}, err
}

func (d *DBConnect) Close() error {
	return d.db.Close()
}

func (d *DBConnect) Client() *sqlx.DB {
	return d.db
}
