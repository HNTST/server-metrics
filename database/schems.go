package database

// 1. Таблица агентов
var Agents_table = `
CREATE TABLE if not exists agents (
    id SERIAL PRIMARY KEY,
    agent_name VARCHAR(255) UNIQUE NOT NULL,
    agent_ip VARCHAR(50),
    last_seen TIMESTAMP WITH TIME ZONE -- Обновляем при каждом получении лога
);
`

// 2. Таблица логов (связана с агентом)
var Logs_table = `
CREATE TABLE if not exists logs (
    id SERIAL PRIMARY KEY,
    agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
    cpu_usage DOUBLE PRECISION,
    mem_total BIGINT,
    mem_usage BIGINT,
    space_total BIGINT,
    space_usage BIGINT,
    sent_at TIMESTAMP WITH TIME ZONE NOT NULL
);
`

// 3. Таблица контейнеров (связана с конкретным логом)
// Один лог содержит много контейнеров (One-to-Many)
var Conts_table = `
CREATE TABLE if not exists containers (
    id SERIAL PRIMARY KEY,
    log_id INTEGER REFERENCES logs(id) ON DELETE CASCADE,
    container_id VARCHAR(255),
    name VARCHAR(255),
    image VARCHAR(255),
    state VARCHAR(50)
);
`
