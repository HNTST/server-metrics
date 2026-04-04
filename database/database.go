package database

import (
	md "agent/models"
	"database/sql"
	"fmt"

	"log"
	"os"

	_ "github.com/lib/pq" // To register the driver.
)

type DBRepo interface {
	InitDB() error
	SaveLog(metrics md.Metrics) error
}

type Service struct {
	db *sql.DB
}

// getEnv — берет значение из ОС или возвращает default
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Printf("%s не найден, использую дефолт: %s", key, fallback)
	return fallback
}

func (sv *Service) InitDB() error {

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "diplom"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_NAME", "diplom"),
		getEnv("SSLMODE", "disable"),
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
	var (
		agent_id, log_id int64
		err              error
	)

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

	tx, err := sv.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, _ = tx.Exec("UPDATE agents SET last_seen = NOW() WHERE id = $1", agent_id)
	// ПОЛУЧИТЬ ID ВСТАВЛЕННОГО ЛОГА
	err = tx.QueryRow("insert into logs (agent_id, cpu_usage, mem_total, mem_usage, space_total, space_usage, sent_at) values($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		agent_id, metrics.Cpu, metrics.MemTotal, metrics.MemUsage, metrics.SpaceTotal, metrics.SpaceUsage, metrics.SentAt).Scan(&log_id)
	if err != nil {
		return err
	}
	err = sv.saveDockerLog(tx, log_id, metrics)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (sv *Service) saveDockerLog(tx *sql.Tx, log_id int64, metrics md.Metrics) error {
	var (
		err error
	)
	if len(metrics.Containers) == 0 {
		return nil
	}
	for _, c := range metrics.Containers {
		_, err = tx.Exec(
			"insert into containers (log_id, container_id, name, image, state) values($1, $2, $3, $4, $5)",
			log_id, c.ContainerID, c.Name, c.Image, c.State,
		)
		if err != nil {
			return err
		}
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
