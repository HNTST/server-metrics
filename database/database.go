package database

import (
    "database/sql"
    "log"

    _ "github.com/lib/pq" // To register the driver.
)

type DBRepo interface {
	InitDB() error
}

type Service struct {
	db   *sql.DB
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

	// DB_HOST уже обработан в вашем примере

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
        log.Println("DB_PASSWORD не найден, взят: [скрыто]")
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

func (sv *Service) InitDB() error{


	cfg := pq.Config{
    Host: "localhost",
    Port: 5432,
    User: "pqgo",
}
// Or: create a new Config from the defaults, environment, and DSN.
// cfg, err := pq.NewConfig("host=postgres dbname=pqgo")
// if err != nil {
//     log.Fatal(err)
// }

c, err := pq.NewConnectorConfig(cfg)
if err != nil {
    log.Fatal(err)
}

// Create connection pool.
db := sql.OpenDB(c)
defer db.Close()

// Make sure it works.
err = db.Ping()
if err != nil {
    log.Fatal(err)
}
	db, err :=
}
