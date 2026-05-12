package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	connStr string
	conn    *sql.DB
}

func NewDB(destination string) (*DB, error) {
	db := &DB{conn: nil, connStr: destination}

	err := db.Init()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// "./test.db"
func (db *DB) Init() error {
	var err error
	db.conn, err = sql.Open("sqlite3", db.connStr)
	if err != nil {
		return err
	}

	sqlStmt := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        name TEXT
    );
    `
	_, err = db.conn.Exec(sqlStmt)
	if err != nil {
		return err
	}

	log.Println("Table 'users' created successfully")
	return nil
}

func (db *DB) Close() error {
	if err := db.conn.Close(); err != nil {
		return err
	}
	return nil
}
