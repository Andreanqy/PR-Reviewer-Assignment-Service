package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func CreateSampleDB() {
	// Параметры подключения к PostgreSQL
	host := "localhost"
	port := 5432
	user := "postgres"
	password := "password"
	dbname := "mydb" // PullRequestsDataBase

	// Подключение к postgres для создания базы данных
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", host, port, user, password)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer db.Close()

	// Создание новой базе данных
	_, err = db.Exec("CREATE DATABASE " + dbname)
	if err != nil {
		log.Printf("База данных уже существует или ошибка: %v", err)
	} else {
		log.Println("База данных создана успешно")
	}

	// Подключение к новой базе данных
	db.Close()
	psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Ошибка подключения к новой базе: %v", err)
	}
	defer db.Close()

	// Создание таблиц
	createTables := []string{
		`CREATE TABLE IF NOT EXISTS Users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50),
			team_name VARCHAR(100),
			is_active BOOLEAN DEFAULT TRUE
		);`,
		`CREATE TABLE IF NOT EXISTS Teams (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50)
		);`,
		`CREATE TABLE IF NOT EXISTS PullRequests (
			id SERIAL PRIMARY KEY,
			title VARCHAR(100),
			author_id INT REFERENCES Users(id),
			status VARCHAR(20) CHECK (status IN ('OPEN', 'MERGED'))
		);`,
	}

	for _, query := range createTables {
		_, err := db.Exec(query)
		if err != nil {
			log.Fatalf("Ошибка создания таблицы: %v", err)
		}
	}
	log.Println("Таблицы созданы успешно")

	// Вставка примеров
	_, err = db.Exec(`
		INSERT INTO Users (name, team_name, is_active) VALUES
		('Alice', 'Backend', TRUE),
		('Bob', 'Frontend', TRUE),
		('Charlie', 'DevOps', TRUE)
		ON CONFLICT DO NOTHING;
	`)
	if err != nil {
		log.Fatalf("Ошибка вставки пользователей: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO Teams (title) VALUES
		('Backend'),
		('Frontend'),
		('DevOps')
		ON CONFLICT DO NOTHING;
	`)
	if err != nil {
		log.Fatalf("Ошибка вставки команд: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO PullRequests (name, author_id, status) VALUES
		('Add login feature', 1, 'OPEN'),
		('Fix UI bug', 2, 'OPEN'),
		('Update CI/CD pipeline', 3, 'OPEN')
		ON CONFLICT DO NOTHING;
	`)
	if err != nil {
		log.Fatalf("Ошибка вставки pull requests: %v", err)
	}

	log.Println("Примеры данных добавлены успешно")
}

func main() {
	host := "localhost"
	user := "postgres"
	pass := "password"
	port := 5432
	dbName := "mydb"

	// Настройки подключения к PostgreSQL
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", host, port, user, pass)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot ping database: %v", err)
	}

	// Создание базы mydb
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		log.Fatalf("Failed to drop database: %v", err)
	}
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	fmt.Println("Database created successfully!")

	// Подключение к новой базе mydb
	db.Close()
	db, err = sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable", user, pass, port, dbName))
	if err != nil {
		log.Fatalf("Failed to connect to new database: %v", err)
	}
	defer db.Close()

	// Создание таблиц
	createTables := []string{
		`CREATE TABLE IF NOT EXISTS Users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50),
			team VARCHAR(100),
			is_active BOOLEAN DEFAULT TRUE
		);`,
		`CREATE TABLE IF NOT EXISTS Teams (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50)
		);`,
		`CREATE TABLE IF NOT EXISTS PullRequests (
			id SERIAL PRIMARY KEY,
			title VARCHAR(100),
			author_id INT REFERENCES Users(id),
			status VARCHAR(20) CHECK (status IN ('OPEN', 'MERGED'))
		);`,
	}

	for _, query := range createTables {
		_, err := db.Exec(query)
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
	}
	fmt.Println("Tables created successfully!")

	// Добавляем 3 пользователей
	users := []struct {
		name      string
		team      string
		is_active bool
	}{
		{"Alice", "Frontend", true},
		{"Bob", "Frontend", true},
		{"Charlie", "Backend", true},
	}

	for _, u := range users {
		_, err := db.Exec("INSERT INTO users (name, team, is_active) VALUES ($1, $2, $3)", u.name, u.team, u.is_active)
		if err != nil {
			log.Fatalf("Failed to insert user: %v", err)
		}
	}
	fmt.Println("Users inserted successfully!")

	// Добавляем 2 команды
	teams := []string{"Frontend", "Backend"}
	for _, t := range teams {
		_, err := db.Exec("INSERT INTO teams (name) VALUES ($1)", t)
		if err != nil {
			log.Fatalf("Failed to insert team: %v", err)
		}
	}
	fmt.Println("Teams inserted successfully!")

	// Добавляем 2 Pull Request
	prs := []struct {
		title    string
		authorID int
		status   string
	}{
		{"Add login feature", 1, "OPEN"},
		{"Fix navbar bug", 2, "OPEN"},
	}

	for _, pr := range prs {
		_, err := db.Exec("INSERT INTO PullRequests (title, author_id, status) VALUES ($1, $2, $3)", pr.title, pr.authorID, pr.status)
		if err != nil {
			log.Fatalf("Failed to insert PR: %v", err)
		}
	}
	fmt.Println("Pull Requests inserted successfully!")
}
