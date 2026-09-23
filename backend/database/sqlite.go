package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func ConfigDb()*sql.DB {

	db, err := sql.Open("sqlite3", "database/storage/app.db")
	if err != nil {
		log.Fatalf("Error opening database file: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to SQLite3 database: %v", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}
	fmt.Println("Successfully connected to app.db using sqlite3 driver!")
	m, err := migrate.New(
		"file://database/migrations",
		"sqlite3://database/storage/app.db",
	)
	if err != nil {
		log.Fatal("failed to initialize migrations:", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("failed to run migrations:", err)
	}

	log.Println("Database migrations completed")
	return  db
}
