package database

import (
	md "agent/models"
	"database/sql"
	"fmt"

	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // To register the driver.
)

type DBRepo interface {
	InitDB() error
	SaveLog(metrics md.Metrics) error
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

func (sv *Service) SaveLog(metrics md.Metrics) error {
	var agent_id int64
	var err error

	if !sv.checkAgent(metrics.Host) {
		new_agent := md.Agent{
			AgentName: metrics.Host,
			AgentIP:   "0.0.0.0",
		}
		agent_id, err = sv.regNewAgent(new_agent)
		if err != nil {
			return err
		}
	} else {
		agent_id, err = sv.getAgentId(metrics.Host)
		if err != nil {
			return err
		}
	}

	_, err = sv.db.Exec("insert into logs (agent_id, cpu_usage, mem_total, mem_usage, space_total, space_usage, sent_at) values($1, $2, $3, $4, $5, $6, $7)",
		agent_id, metrics.Cpu, metrics.MemTotal, metrics.MemUsage, metrics.SpaceTotal, metrics.SpaceUsage, metrics.SentAt)
	if err != nil {
		return err
	}
	return nil
}

func (sv *Service) checkAgent(agent_name string) bool {
	var agent md.Agent
	row := sv.db.QueryRow("select * from agents where agent_name = $1", agent_name)
	if err := row.Scan(&agent.ID, &agent.AgentName, &agent.AgentIP, &agent.LastSeen); err != nil {
		return false
	}
	return true
}
func (sv *Service) regNewAgent(agent md.Agent) (int64, error) {
	var id int64
	err := sv.db.QueryRow("insert into agents (agent_name, agent_ip, last_seen) values($1,$2, NOW()) RETURNING id",
		agent.AgentName, agent.AgentIP).Scan(&id)
	return id, err
}
func (sv *Service) getAgentId(agent_name string) (int64, error) {
	var id int64
	err := sv.db.QueryRow("select id from agents where agent_name = $1", agent_name).Scan(&id)
	return id, err
}
