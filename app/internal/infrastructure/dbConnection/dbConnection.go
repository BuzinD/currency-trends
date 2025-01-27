package dbConnection

import (
	"cur/internal/config/dbConfig"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func GetDbConnection() (*sql.DB, error) {

	config, err := dbConfig.GetDbConfig()

	if err != nil {
		log.Fatal(err)
	}

	// Format the connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DbName)

	// Open a connection to the database
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalf("Unable to connect to the database: %v\n", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Unable to ping the database: %v\n", err)
		defer db.Close()
	}

	fmt.Println("Successfully connected to the database!")

	return db, err
}
