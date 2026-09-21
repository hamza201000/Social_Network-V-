package database

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/mattn/go-sqlite3"
)

func ConfigDb() {

	db, err := sql.Open("sqlite3", "storage/app.db")
	if err != nil {
		log.Fatalf("Error opening database file: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to SQLite3 database: %v", err)
	}

	fmt.Println("Successfully connected to app.db using sqlite3 driver!")

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}
}
