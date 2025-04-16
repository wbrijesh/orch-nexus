-- Schema created by: wbrijesh
-- Date: 2025-04-16 17:19:14 UTC
-- Note: Amazon Keyspaces does not support DEFAULT values

-- Create keyspace with SingleRegionStrategy (required for AWS Keyspaces)
CREATE KEYSPACE IF NOT EXISTS orchestrator
WITH REPLICATION = {'class': 'SingleRegionStrategy'}
AND DURABLE_WRITES = true;

---------------------------
-- AUTHENTICATION TABLES --
---------------------------

-- Users table - Core user information
CREATE TABLE IF NOT EXISTS orchestrator.users (
    user_id text,
    email text,
    display_name text,
    full_name text,
    avatar_url text,
    created_at timestamp,
    last_login timestamp,
    provider text, -- 'google' or 'github'
    account_status text, -- 'active', 'suspended', 'deleted'
    is_admin boolean,
    credits_balance bigint, -- Credits = subscription price in dollars * 1000
    is_beta_tester boolean,
    beta_expires_at timestamp,
    PRIMARY KEY (user_id)
);

-- OAuth accounts - Main authentication table
CREATE TABLE IF NOT EXISTS orchestrator.oauth_accounts (
    provider text, -- 'google', 'github'
    provider_user_id text, -- ID from the provider
    user_id text,
    email text,
    created_at timestamp,
    last_login timestamp,
    PRIMARY KEY ((provider, provider_user_id))
);

-- User lookup by email
CREATE TABLE IF NOT EXISTS orchestrator.users_by_email (
    email text,
    user_id text,
    provider text, -- 'google' or 'github'
    PRIMARY KEY (email)
);

--------------------
-- BETA USER MGMT --
--------------------

-- Beta invitation codes
CREATE TABLE IF NOT EXISTS orchestrator.beta_codes (
    code text,
    created_by_user_id text,
    created_at timestamp,
    is_used boolean,
    used_by_user_id text,
    used_at timestamp,
    free_credits bigint, -- Amount in credits
    beta_period_days int, -- Days beta access lasts after redemption
    is_active boolean,
    notes text,
    PRIMARY KEY (code)
);

-------------------------
-- TASK RELATED TABLES --
-------------------------

-- Tasks by user - the primary way tasks will be accessed
CREATE TABLE IF NOT EXISTS orchestrator.tasks (
    user_id text,
    task_id text,
    title text,
    description text,
    status text, -- 'queued', 'running', 'completed', 'failed', 'cancelled'
    created_at timestamp,
    started_at timestamp,
    completed_at timestamp,
    error_message text,
    total_cost bigint, -- Total cost in credits
    PRIMARY KEY ((user_id), created_at, task_id)
) WITH CLUSTERING ORDER BY (created_at DESC, task_id ASC);

-- Task steps - Optimized for retrieving steps for a specific task
CREATE TABLE IF NOT EXISTS orchestrator.task_steps (
    task_id text,
    step_id uuid,
    sequence_number int,
    step_type text, -- 'navigation', 'click', 'input', 'read', 'screenshot', 'wait', 'custom'
    status text, -- 'planned', 'executing', 'succeeded', 'failed'
    error_message text,
    timestamp timestamp,
    tokens_used int, -- LLM token usage
    step_cost bigint, -- Cost in credits
    PRIMARY KEY ((task_id), sequence_number)
) WITH CLUSTERING ORDER BY (sequence_number ASC);

-- Transaction history - Tracks credit usage
CREATE TABLE IF NOT EXISTS orchestrator.transactions (
    user_id text,
    transaction_id uuid,
    timestamp timestamp,
    amount bigint, -- Positive for credits added, negative for credits used
    balance_after bigint,
    description text,
    related_task_id text,
    related_step_id uuid,
    PRIMARY KEY ((user_id), timestamp, transaction_id)
) WITH CLUSTERING ORDER BY (timestamp DESC, transaction_id ASC);
