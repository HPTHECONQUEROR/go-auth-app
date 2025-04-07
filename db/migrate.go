package db

import (
	"context"
	"fmt"
	"log"
)

func RunMigrations() {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	messagesTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		sender_id INTEGER NOT NULL REFERENCES users(id),
		receiver_id INTEGER NOT NULL REFERENCES users(id),
		content TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	conversationsTable := `
	CREATE TABLE IF NOT EXISTS conversations (
		id SERIAL PRIMARY KEY,
		user1_id INTEGER NOT NULL REFERENCES users(id),
		user2_id INTEGER NOT NULL REFERENCES users(id),
		last_message TEXT DEFAULT '',
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT unique_conversation UNIQUE (user1_id, user2_id)
	);
	`

	snmpMetricsTable := `
	CREATE TABLE IF NOT EXISTS snmp_metrics (
		id SERIAL PRIMARY KEY,
		device_id VARCHAR(100) NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		device_name VARCHAR(255),
		device_description TEXT,
		device_uptime VARCHAR(255),
		cpu_load INTEGER,
		memory_total BIGINT,
		in_octets BIGINT,
		out_octets BIGINT,
		in_errors INTEGER,
		out_errors INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	// Create index on timestamp for faster queries
	metricsIndex := `
	CREATE INDEX IF NOT EXISTS idx_snmp_metrics_timestamp ON snmp_metrics (timestamp);
	`

	// Execute migrations
	migrations := []string{usersTable, messagesTable, conversationsTable, snmpMetricsTable, metricsIndex}

	for _, migration := range migrations {
		_, err := DB.Exec(context.Background(), migration)
		if err != nil {
			log.Fatal("Migration failed: ", err)
		}
	}

	fmt.Println("Database migrations completed successfully!")
}