package dbConnection

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func GetTestDbConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")

	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		_ = db.Close()
		fmt.Println("db closed")
		log.Fatalf("Unable to ping the database: %v\n", err)
	}

	fmt.Println("Successfully connected to the test database!")

	return db, err
}
