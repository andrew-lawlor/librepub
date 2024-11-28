package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/andrew-lawlor/librepub/config"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func init() {
	// Check if DB exists.
	_, err := os.Stat(config.DbPath)
	if err == nil {
		fmt.Println("DB Exists, exiting.")
		setDB()
		return
	}
	fmt.Println("DB does not exist, creating.")
	// Set up connection.
	setDB()
	// Execute schema and insert starting data.
	err = initDB()
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("DB Created successfully.")

}

func setDB() {
	var err error
	db, err = sql.Open("sqlite3", config.DbPath)
	if err != nil {
		log.Fatal(err.Error())
	}
	// Engage WAL mode for MAX PERFORMANCE
	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		log.Fatalf("Failed to set WAL mode: %v", err)
	}
}

func GetDB() *sql.DB {
	return db
}

// Should I load this from a text file?
func initDB() error {
	schema := `CREATE TABLE IF NOT EXISTS "config" (
	"id"	INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
	"name"	TEXT NOT NULL UNIQUE,
	"value"	TEXT NOT NULL,
	"type"	TEXT NOT NULL DEFAULT 'boolean'
);
	CREATE TABLE IF NOT EXISTS users (
	user_name	TEXT NOT NULL UNIQUE,
	user_id	INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
	password	TEXT NOT NULL,
	display_name	TEXT NOT NULL,
	created	TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS "logs" (
	"id"	INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
	"level"	TEXT NOT NULL DEFAULT 'Info',
	"created"	TEXT NOT NULL,
	"message"	TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS "questions" (
	"id"	INTEGER NOT NULL UNIQUE,
	"question"	TEXT NOT NULL,
	"answer"	TEXT NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE IF NOT EXISTS "books" (
	"id"	INTEGER NOT NULL UNIQUE,
	"title"	TEXT NOT NULL,
	"ereader_id"	TEXT NOT NULL,
	"user_id"	TEXT NOT NULL,
	FOREIGN KEY("user_id") REFERENCES "users"("user_id"),
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE IF NOT EXISTS "annotations" (
	"id"	INTEGER NOT NULL UNIQUE,
	"text"	TEXT,
	"note"	TEXT,
	"type"	TEXT,
	"created" TEXT NOT NULL,
	"ereader_id"	TEXT NOT NULL,
	"book_id"	INTEGER NOT NULL,
	"user_id"	TEXT NOT NULL,
	FOREIGN KEY("user_id") REFERENCES "users"("user_id"),
	FOREIGN KEY("book_id") REFERENCES "books"("id"),
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE IF NOT EXISTS "vocab" (
	"id"	INTEGER NOT NULL UNIQUE,
	"word"	TEXT NOT NULL,
	"definition"	TEXT,
	"created" TEXT NOT NULL,
	"user_id"	TEXT NOT NULL,
	FOREIGN KEY("user_id") REFERENCES "users"("user_id"),
	PRIMARY KEY("id")
);
`
	_, err := db.Exec(schema)
	if err != nil {
		return err
	}
	statement, err := db.Prepare("INSERT INTO config (name, value, type) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = statement.Exec("allow_registration", "1", "boolean")
	if err != nil {
		return err
	}
	// Init capcha q + a's, which are stored in text file.
	content, err := os.ReadFile("./config/qa.txt")
	if err != nil {
		log.Fatal(err.Error())
	}
	var lines = strings.Split(string(content), "\n")
	for _, line := range lines {
		var subLine = strings.Split(line, "|")
		statement, err := db.Prepare("INSERT INTO questions (question, answer) VALUES (?, ?)")
		if err != nil {
			return err
		}
		_, err = statement.Exec(subLine[0], subLine[1])
		if err != nil {
			return err
		}
	}
	return nil
}
