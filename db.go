package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var Pool *sql.DB

// Подключение к БД
func ConnectDB(url string) {
	var err error
	Pool, err = sql.Open("postgres", url)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	if err = Pool.Ping(); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	fmt.Println("Connected to PostgreSQL")
}

// Закрытие БД
func Close() {
	Pool.Close()
	fmt.Println("Database closed")
}
