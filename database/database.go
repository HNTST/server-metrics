package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // To register the driver.
)

type DBRepo interface {
	InitDB() error
}

type Service struct {
	db *sql.DB
}

func envLoad(path string) map[string]string {
	result := make(map[string]string)
	err := godotenv.Load(path)
	if err != nil {
		log.Println("Файл .env не найден, использую дефолтные настройки")
	}

	log.Println("Чтение DB_HOST")
	db_host := os.Getenv("DB_HOST")
	if db_host == "" {
		db_host = "localhost"
		log.Println("DB_HOST не найден, взят: ", db_host)
	}
	result["DB_HOST"] = db_host

	// Чтение DB_PORT
	log.Println("Чтение DB_PORT")
	db_port := os.Getenv("DB_PORT")
	if db_port == "" {
		db_port = "5432"
		log.Println("DB_PORT не найден, взят: ", db_port)
	}
	result["DB_PORT"] = db_port

	// Чтение DB_USER
	log.Println("Чтение DB_USER")
	db_user := os.Getenv("DB_USER")
	if db_user == "" {
		db_user = "diplom"
		log.Println("DB_USER не найден, взят: ", db_user)
	}
	result["DB_USER"] = db_user

	// Чтение DB_PASSWORD
	log.Println("Чтение DB_PASSWORD")
	db_password := os.Getenv("DB_PASSWORD")
	if db_password == "" {
		db_password = "password"
		log.Println("DB_PASSWORD не найден, взят: ", db_password)
	}
	result["DB_PASSWORD"] = db_password

	// Чтение DB_NAME
	log.Println("Чтение DB_NAME")
	db_name := os.Getenv("DB_NAME")
	if db_name == "" {
		db_name = "diplom"
		log.Println("DB_NAME не найден, взят: ", db_name)
	}
	result["DB_NAME"] = db_name

	// Чтение SSLMODE
	log.Println("Чтение SSLMODE")
	ssl_mode := os.Getenv("SSLMODE")
	if ssl_mode == "" {
		ssl_mode = "disable"
		log.Println("SSLMODE не найден, взят: ", ssl_mode)
	}
	result["SSLMODE"] = ssl_mode

	return result
}

func (sv *Service) InitDB() error {
	dsn := make(map[string]string)
	log.Println("Загрузка .env")
	dsn = envLoad("cmd/.env")
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dsn["DB_HOST"],
		dsn["DB_PORT"],
		dsn["DB_USER"],
		dsn["DB_PASSWORD"],
		dsn["DB_NAME"],
		dsn["SSLMODE"],
	)

	// 2. Открываем соединение
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	sv.db = db

	// 3. Проверяем соединение (sql.Open не создает само подключение сразу)
	err = db.Ping()
	if err != nil {
		return err
	}

	// создание таблиц
	if _, err = db.Exec(Agents_table); err != nil {
		return err
	}
	if _, err = db.Exec(Logs_table); err != nil {
		return err
	}
	if _, err = db.Exec(Conts_table); err != nil {
		return err
	}

	fmt.Println("Успешное подключение к базе данных!")
	return nil
}
