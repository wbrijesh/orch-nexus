package database

import (
	"fmt"
	"log"
)

func (s *service) InitSchema() error {
	log.Println("Initializing Amazon Keyspaces schema...")

	// Create keyspace
	if err := s.Session.Query(`
        CREATE KEYSPACE IF NOT EXISTS orchestrator
        WITH REPLICATION = {'class': 'SingleRegionStrategy'}
        AND DURABLE_WRITES = true;
    `).Exec(); err != nil {
		return fmt.Errorf("failed to create keyspace: %v", err)
	}

	// Create users table
	if err := s.Session.Query(`
        CREATE TABLE IF NOT EXISTS orchestrator.users (
            user_id text,
            email text,
            display_name text,
            full_name text,
            avatar_url text,
            created_at timestamp,
            last_login timestamp,
            provider text,
            account_status text,
            is_admin boolean,
            credits_balance bigint,
            is_beta_tester boolean,
            beta_expires_at timestamp,
            PRIMARY KEY (user_id)
        );
    `).Exec(); err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	// Create oauth_accounts table - Updated to match new schema
	if err := s.Session.Query(`
        CREATE TABLE IF NOT EXISTS orchestrator.oauth_accounts (
            provider text,
            provider_user_id text,
            user_id text,
            email text,
            created_at timestamp,
            PRIMARY KEY ((provider, provider_user_id))
        );
    `).Exec(); err != nil {
		return fmt.Errorf("failed to create oauth_accounts table: %v", err)
	}

	// Create users_by_email table - Updated to match new schema
	if err := s.Session.Query(`
        CREATE TABLE IF NOT EXISTS orchestrator.users_by_email (
            email text,
            user_id text,
            provider text,
            PRIMARY KEY (email)
        );
    `).Exec(); err != nil {
		return fmt.Errorf("failed to create users_by_email table: %v", err)
	}

	// Create beta_codes table
	if err := s.Session.Query(`
        CREATE TABLE IF NOT EXISTS orchestrator.beta_codes (
            code text,
            created_by_user_id text,
            created_at timestamp,
            is_used boolean,
            used_by_user_id text,
            used_at timestamp,
            free_credits bigint,
            beta_period_days int,
            is_active boolean,
            notes text,
            PRIMARY KEY (code)
        );
    `).Exec(); err != nil {
		return fmt.Errorf("failed to create beta_codes table: %v", err)
	}

	// Create tasks table
	if err := s.Session.Query(`
        CREATE TABLE IF NOT EXISTS orchestrator.tasks (
            user_id text,
            task_id text,
            title text,
            description text,
            status text,
            created_at timestamp,
            started_at timestamp,
            completed_at timestamp,
            error_message text,
            total_cost bigint,
            PRIMARY KEY ((user_id), created_at, task_id)
        ) WITH CLUSTERING ORDER BY (created_at DESC, task_id ASC);
    `).Exec(); err != nil {
		return fmt.Errorf("failed to create tasks table: %v", err)
	}

	// Create task_steps table
	if err := s.Session.Query(`
        CREATE TABLE IF NOT EXISTS orchestrator.task_steps (
            task_id text,
            step_id uuid,
            sequence_number int,
            step_type text,
            status text,
            error_message text,
            timestamp timestamp,
            tokens_used int,
            step_cost bigint,
            PRIMARY KEY ((task_id), sequence_number)
        ) WITH CLUSTERING ORDER BY (sequence_number ASC);
    `).Exec(); err != nil {
		return fmt.Errorf("failed to create task_steps table: %v", err)
	}

	// Create transactions table
	if err := s.Session.Query(`
        CREATE TABLE IF NOT EXISTS orchestrator.transactions (
            user_id text,
            transaction_id uuid,
            timestamp timestamp,
            amount bigint,
            balance_after bigint,
            description text,
            related_task_id text,
            related_step_id uuid,
            PRIMARY KEY ((user_id), timestamp, transaction_id)
        ) WITH CLUSTERING ORDER BY (timestamp DESC, transaction_id ASC);
    `).Exec(); err != nil {
		return fmt.Errorf("failed to create transactions table: %v", err)
	}

	log.Println("Amazon Keyspaces schema initialized successfully")
	return nil
}
