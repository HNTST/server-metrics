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
	GetLastLog() ([]md.Metrics, error)
	GetAgents() ([]md.Agent, error)
	GetDailyMetrics() ([]md.DailyMetrics, error)
	GenerateDailyReport() error
	CleanUpOldLogs(days int) error
	GetAgentNameByID(id int) string
}

type Service struct {
	db *sql.DB
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Printf("%s не найден, использую дефолт: %s", key, fallback)
	return fallback
}

func (sv *Service) InitDB() error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "192.168.31.234"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "diplom"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_NAME", "diplom"),
		getEnv("SSLMODE", "disable"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	sv.db = db
	err = db.Ping()
	if err != nil {
		return err
	}
	if _, err = db.Exec(Agents_table); err != nil {
		return err
	}
	if _, err = db.Exec(Logs_table); err != nil {
		return err
	}
	if _, err = db.Exec(Conts_table); err != nil {
		return err
	}
	if _, err = db.Exec(Daily_metrics); err != nil {
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
	err = tx.QueryRow(`insert into logs (agent_id, cpu_usage, mem_total, mem_usage, space_total, space_usage, sent_at)
		values($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
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

func (sv *Service) GetLastLog() ([]md.Metrics, error) {
	var (
		result []md.Metrics
	)

	rowsLogs, err := sv.db.Query(`SELECT DISTINCT ON (agent_id)
    id, agent_id, cpu_usage, mem_total, mem_usage, space_total, space_usage, sent_at
	FROM logs
	ORDER BY agent_id, id DESC;
	`)
	if err != nil {
		return result, err
	}
	defer rowsLogs.Close()

	for rowsLogs.Next() {
		var l md.Metrics
		err = rowsLogs.Scan(&l.ID, &l.AgentID, &l.Cpu, &l.MemTotal, &l.MemUsage, &l.SpaceTotal, &l.SpaceUsage, &l.SentAt)
		if err != nil {
			log.Printf("Ошибка получения лога: %v", err)
			continue
		}
		l.Host = sv.GetAgentNameByID(l.AgentID)

		rowsContLogs, err := sv.db.Query("select container_id, name, image, state from containers where log_id = $1", l.ID)
		if err != nil {
			log.Printf("Ошибка сканирования контейнера: %v", err)
			continue
		}

		for rowsContLogs.Next() {
			var c md.ContainerInfo
			err = rowsContLogs.Scan(&c.ContainerID, &c.Name, &c.Image, &c.State)
			if err != nil {
				log.Printf("Ошибка сканирования контейнера: %v", err)
				continue
			}
			l.Containers = append(l.Containers, c)
		}
		rowsContLogs.Close()
		result = append(result, l)
	}
	for _, m := range result {
		log.Printf("Агент: %s, Контейнеров в структуре: %d", m.Host, len(m.Containers))
	}
	return result, nil
}
func (sv *Service) GetAgentNameByID(id int) string {
	var ret_id string
	_ = sv.db.QueryRow("select agent_name from agents where id = $1", id).Scan(&ret_id)
	return ret_id
}

func (sv *Service) GetAgents() ([]md.Agent, error) {
	var (
		result []md.Agent
	)
	rows, err := sv.db.Query("select id, agent_name, agent_ip, last_seen from agents")
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var a md.Agent
		err = rows.Scan(&a.ID, &a.AgentName, &a.AgentIP, &a.LastSeen)
		if err != nil {
			log.Printf("Ошибка сканирования контейнера: %v", err)
			continue
		}
		result = append(result, a)
	}
	return result, nil
}

func (sv *Service) GetDailyMetrics() ([]md.DailyMetrics, error) {
	var reports []md.DailyMetrics

	rows, err := sv.db.Query(`
        SELECT agent_id, avg_cpu, max_cpu, avg_mem, max_mem, avg_disk, report_date
        FROM daily_metrics
        WHERE report_date = (SELECT MAX(report_date) FROM daily_metrics)
        ORDER BY agent_id ASC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d md.DailyMetrics
		err := rows.Scan(
			&d.AgentID, &d.AvgCPU, &d.MaxCPU,
			&d.AvgMEM, &d.MaxMEM, &d.AvgSpace, &d.ReportDate,
		)
		if err != nil {
			log.Printf("Ошибка сканирования отчета: %v", err)
			continue
		}
		reports = append(reports, d)
	}

	return reports, nil
}

func (sv *Service) GenerateDailyReport() error {
	query := `
    INSERT INTO daily_metrics (agent_id, avg_cpu, max_cpu, avg_mem, max_mem, avg_disk, report_date)
    SELECT
        agent_id,
        AVG(cpu_usage),
        MAX(cpu_usage),
        AVG(mem_usage),
        MAX(mem_usage),
        MAX(space_usage),
        (CURRENT_DATE - INTERVAL '1 day') -- Отчет за вчера
    FROM logs
    WHERE sent_at >= CURRENT_DATE - INTERVAL '1 day'
      AND sent_at < CURRENT_DATE
    GROUP BY agent_id
    ON CONFLICT (agent_id, report_date)
    DO UPDATE SET
        avg_cpu = EXCLUDED.avg_cpu,
        max_cpu = EXCLUDED.max_cpu,
        avg_mem = EXCLUDED.avg_mem,
        max_mem = EXCLUDED.max_mem,
        avg_disk = EXCLUDED.avg_disk;
    `

	_, err := sv.db.Exec(query)
	if err != nil {
		log.Printf("Ошибка при генерации ежедневного отчета: %v", err)
		return err
	}

	log.Println("Ежедневный отчет успешно сформирован.")
	return nil
}

func (sv *Service) CleanUpOldLogs(days int) error {
	query := `
    DELETE FROM logs
    WHERE sent_at < NOW() - ($1 * INTERVAL '1 day');
    `

	_, err := sv.db.Exec(query, days)
	if err != nil {
		log.Printf("Ошибка при очистке логов: %v", err)
		return err
	}

	log.Printf("Старые логи (старше %d дней) успешно удалены", days)
	return nil
}
