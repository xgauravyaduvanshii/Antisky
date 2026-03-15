-- ===================================
-- Antisky Platform Upgrade — Distributed Infrastructure
-- ===================================
-- Migration 002: Servers, Admin, Terminal, Billing, OAuth

-- ===================================
-- Server Nodes (Distributed Fleet)
-- ===================================
CREATE TYPE server_status AS ENUM (
    'provisioning', 'online', 'offline', 'maintenance', 'error', 'decommissioned'
);

CREATE TABLE servers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    hostname TEXT UNIQUE NOT NULL,
    ip_address INET NOT NULL,
    port INTEGER DEFAULT 8090,
    region TEXT NOT NULL DEFAULT 'us-east-1',
    zone TEXT,
    server_type TEXT DEFAULT 'worker',       -- worker, builder, edge, database
    status server_status NOT NULL DEFAULT 'provisioning',
    auth_token_hash TEXT NOT NULL,           -- bcrypt hash of server auth token
    server_key TEXT UNIQUE NOT NULL,         -- unique server identifier
    os_info TEXT,
    docker_version TEXT,
    cpu_cores INTEGER,
    ram_mb INTEGER,
    disk_gb INTEGER,
    labels JSONB DEFAULT '{}',
    config JSONB DEFAULT '{}',
    last_heartbeat_at TIMESTAMPTZ,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_servers_status ON servers(status);
CREATE INDEX idx_servers_region ON servers(region);
CREATE INDEX idx_servers_key ON servers(server_key);
CREATE INDEX idx_servers_heartbeat ON servers(last_heartbeat_at);

-- ===================================
-- Server Metrics (Time-Series)
-- ===================================
CREATE TABLE server_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    cpu_percent REAL NOT NULL,
    ram_used_mb INTEGER NOT NULL,
    ram_total_mb INTEGER NOT NULL,
    disk_used_gb REAL NOT NULL,
    disk_total_gb REAL NOT NULL,
    network_rx_bytes BIGINT DEFAULT 0,
    network_tx_bytes BIGINT DEFAULT 0,
    active_containers INTEGER DEFAULT 0,
    load_average REAL[] DEFAULT '{}',
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_metrics_server ON server_metrics(server_id, recorded_at DESC);
CREATE INDEX idx_metrics_time ON server_metrics(recorded_at);

-- Partition by time for performance (optional, documented for production)
-- In production, use pg_partman or TimescaleDB

-- ===================================
-- Server Commands (Remote Execution)
-- ===================================
CREATE TYPE command_status AS ENUM (
    'pending', 'running', 'completed', 'failed', 'timeout'
);

CREATE TABLE server_commands (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    issued_by UUID NOT NULL REFERENCES users(id),
    command TEXT NOT NULL,
    args JSONB DEFAULT '{}',
    status command_status NOT NULL DEFAULT 'pending',
    output TEXT,
    error TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    timeout_seconds INTEGER DEFAULT 300,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_commands_server ON server_commands(server_id, created_at DESC);
CREATE INDEX idx_commands_status ON server_commands(status);

-- ===================================
-- Admin Users
-- ===================================
ALTER TABLE users ADD COLUMN IF NOT EXISTS role TEXT DEFAULT 'user'; -- user, admin, super_admin
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_banned BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS banned_reason TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS login_count INTEGER DEFAULT 0;

CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_banned ON users(is_banned);

-- ===================================
-- Admin Impersonation Logs
-- ===================================
CREATE TABLE admin_impersonation_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID NOT NULL REFERENCES users(id),
    target_user_id UUID NOT NULL REFERENCES users(id),
    reason TEXT,
    ip_address INET,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ
);

CREATE INDEX idx_impersonation_admin ON admin_impersonation_logs(admin_id);
CREATE INDEX idx_impersonation_target ON admin_impersonation_logs(target_user_id);

-- ===================================
-- Terminal Sessions
-- ===================================
CREATE TABLE terminal_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    server_id UUID NOT NULL REFERENCES servers(id),
    status TEXT DEFAULT 'active',            -- active, closed
    ip_address INET,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    commands_executed INTEGER DEFAULT 0,
    bytes_transferred BIGINT DEFAULT 0
);

CREATE INDEX idx_terminal_user ON terminal_sessions(user_id);
CREATE INDEX idx_terminal_server ON terminal_sessions(server_id);
CREATE INDEX idx_terminal_status ON terminal_sessions(status);

-- ===================================
-- OAuth Connections
-- ===================================
CREATE TABLE oauth_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,                   -- github, google, gitlab, bitbucket
    provider_user_id TEXT NOT NULL,
    access_token_encrypted TEXT NOT NULL,
    refresh_token_encrypted TEXT,
    token_expires_at TIMESTAMPTZ,
    scopes TEXT[] DEFAULT '{}',
    provider_username TEXT,
    provider_avatar TEXT,
    connected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, provider)
);

CREATE INDEX idx_oauth_user ON oauth_connections(user_id);
CREATE INDEX idx_oauth_provider ON oauth_connections(provider, provider_user_id);

-- ===================================
-- Billing & Subscriptions
-- ===================================
CREATE TABLE billing_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,                       -- Free, Hobby, Pro, Enterprise
    slug TEXT UNIQUE NOT NULL,
    price_cents INTEGER NOT NULL DEFAULT 0,
    stripe_price_id TEXT,
    features JSONB DEFAULT '{}',
    limits JSONB DEFAULT '{}',               -- build_minutes, bandwidth_gb, etc.
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed billing plans
INSERT INTO billing_plans (name, slug, price_cents, features, limits) VALUES
    ('Free', 'free', 0,
     '{"deployments": true, "custom_domains": false, "team_members": 1}',
     '{"build_minutes": 100, "bandwidth_gb": 10, "projects": 3}'),
    ('Hobby', 'hobby', 500,
     '{"deployments": true, "custom_domains": true, "team_members": 3, "preview_deploys": true}',
     '{"build_minutes": 500, "bandwidth_gb": 100, "projects": 10}'),
    ('Pro', 'pro', 2000,
     '{"deployments": true, "custom_domains": true, "team_members": 10, "preview_deploys": true, "analytics": true, "priority_support": true}',
     '{"build_minutes": 2000, "bandwidth_gb": 1000, "projects": 50}'),
    ('Enterprise', 'enterprise', 0,
     '{"deployments": true, "custom_domains": true, "team_members": -1, "preview_deploys": true, "analytics": true, "sla": true, "dedicated_support": true}',
     '{"build_minutes": -1, "bandwidth_gb": -1, "projects": -1}');

CREATE TABLE billing_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES billing_plans(id),
    stripe_subscription_id TEXT,
    stripe_customer_id TEXT,
    status TEXT NOT NULL DEFAULT 'active',    -- active, cancelled, past_due, trialing
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    cancel_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subs_org ON billing_subscriptions(org_id);
CREATE INDEX idx_subs_status ON billing_subscriptions(status);

CREATE TABLE billing_invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    stripe_invoice_id TEXT,
    amount_cents INTEGER NOT NULL,
    currency TEXT DEFAULT 'usd',
    status TEXT NOT NULL DEFAULT 'pending',   -- pending, paid, failed, void
    period_start TIMESTAMPTZ,
    period_end TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    invoice_pdf_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoices_org ON billing_invoices(org_id);

-- ===================================
-- Cluster Configuration
-- ===================================
CREATE TABLE cluster_config (
    key TEXT PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO cluster_config (key, value, description) VALUES
    ('platform.name', '"Antisky"', 'Platform name'),
    ('platform.version', '"1.0.0"', 'Platform version'),
    ('cluster.max_servers', '100', 'Maximum servers in cluster'),
    ('build.max_concurrent', '10', 'Max concurrent builds'),
    ('build.timeout_minutes', '30', 'Build timeout in minutes'),
    ('deployment.auto_ssl', 'true', 'Auto provision SSL for custom domains');

-- ===================================
-- Triggers
-- ===================================
CREATE TRIGGER trg_servers_updated_at BEFORE UPDATE ON servers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_oauth_updated_at BEFORE UPDATE ON oauth_connections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_subs_updated_at BEFORE UPDATE ON billing_subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
