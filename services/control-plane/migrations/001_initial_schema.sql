-- ===================================
-- Antisky Platform — Initial Database Schema
-- ===================================
-- PostgreSQL 15+

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ===================================
-- Users & Authentication
-- ===================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    avatar_url TEXT,
    password_hash TEXT,
    mfa_secret TEXT,
    mfa_enabled BOOLEAN DEFAULT FALSE,
    email_verified BOOLEAN DEFAULT FALSE,
    provider TEXT DEFAULT 'email', -- email, github, google, gitlab
    provider_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_provider ON users(provider, provider_id);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    key_prefix TEXT NOT NULL,     -- first 8 chars for identification
    key_hash TEXT NOT NULL,       -- bcrypt hash of full key
    scopes TEXT[] DEFAULT '{}',
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_prefix ON api_keys(key_prefix);

-- ===================================
-- Organizations & Teams
-- ===================================
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    avatar_url TEXT,
    plan TEXT NOT NULL DEFAULT 'free',    -- free, hobby, pro, enterprise
    billing_email TEXT,
    stripe_customer_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orgs_slug ON organizations(slug);

CREATE TABLE org_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'developer', -- owner, admin, developer, viewer
    invited_by UUID REFERENCES users(id),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, user_id)
);

CREATE INDEX idx_org_members_org ON org_members(org_id);
CREATE INDEX idx_org_members_user ON org_members(user_id);

-- ===================================
-- Projects
-- ===================================
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    description TEXT,
    repo_url TEXT,
    repo_provider TEXT,           -- github, gitlab, bitbucket
    repo_id TEXT,                 -- provider-specific repo ID
    repo_branch TEXT DEFAULT 'main',
    framework TEXT,               -- nextjs, express, django, laravel, etc.
    runtime TEXT NOT NULL DEFAULT 'nodejs', -- nodejs, go, python, php, ruby, static, docker
    build_command TEXT,
    start_command TEXT,
    output_dir TEXT,
    root_dir TEXT DEFAULT '/',
    install_command TEXT,
    node_version TEXT,
    auto_deploy BOOLEAN DEFAULT TRUE,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, slug)
);

CREATE INDEX idx_projects_org ON projects(org_id);
CREATE INDEX idx_projects_repo ON projects(repo_provider, repo_id);

CREATE TABLE project_domains (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    domain TEXT UNIQUE NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    verified BOOLEAN DEFAULT FALSE,
    verification_token TEXT,
    ssl_status TEXT DEFAULT 'pending', -- pending, active, failed
    ssl_cert_arn TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_domains_project ON project_domains(project_id);
CREATE INDEX idx_domains_domain ON project_domains(domain);

CREATE TABLE project_env_vars (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    encrypted_value TEXT NOT NULL,
    target TEXT[] DEFAULT '{production,preview,development}', -- which environments
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, key)
);

CREATE INDEX idx_env_vars_project ON project_env_vars(project_id);

-- ===================================
-- Deployments
-- ===================================
CREATE TYPE deployment_status AS ENUM (
    'queued', 'building', 'deploying', 'ready', 'failed', 'cancelled', 'rolled_back'
);

CREATE TYPE deployment_type AS ENUM (
    'production', 'preview', 'rollback'
);

CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    triggered_by UUID REFERENCES users(id),
    ref TEXT,                       -- branch or tag
    commit_sha TEXT,
    commit_message TEXT,
    commit_author TEXT,
    status deployment_status NOT NULL DEFAULT 'queued',
    type deployment_type NOT NULL DEFAULT 'production',
    url TEXT,                       -- deployment URL
    preview_url TEXT,
    build_log_url TEXT,
    build_duration_ms INTEGER,
    meta JSONB DEFAULT '{}',        -- source (cli, dashboard, webhook), etc.
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deployments_project ON deployments(project_id);
CREATE INDEX idx_deployments_status ON deployments(status);
CREATE INDEX idx_deployments_created ON deployments(created_at DESC);

CREATE TABLE deployment_functions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    deployment_id UUID NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    path TEXT NOT NULL,             -- /api/hello
    runtime TEXT NOT NULL,          -- nodejs, python, go
    handler TEXT NOT NULL,
    memory_mb INTEGER DEFAULT 128,
    timeout_seconds INTEGER DEFAULT 10,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deploy_funcs_deployment ON deployment_functions(deployment_id);

-- ===================================
-- Usage & Billing
-- ===================================
CREATE TABLE usage_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    type TEXT NOT NULL,              -- build_minutes, bandwidth_gb, function_invocations, compute_hours
    quantity NUMERIC NOT NULL,
    unit TEXT NOT NULL,              -- minutes, bytes, invocations, seconds
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_usage_org ON usage_records(org_id, recorded_at);
CREATE INDEX idx_usage_type ON usage_records(type, recorded_at);

-- ===================================
-- Audit Logs
-- ===================================
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_id UUID REFERENCES users(id),
    action TEXT NOT NULL,            -- project.created, deployment.triggered, member.invited, etc.
    resource_type TEXT NOT NULL,     -- project, deployment, member, env_var, domain
    resource_id UUID,
    metadata JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_org ON audit_logs(org_id, created_at DESC);
CREATE INDEX idx_audit_actor ON audit_logs(actor_id);
CREATE INDEX idx_audit_action ON audit_logs(action);

-- ===================================
-- Triggers for updated_at
-- ===================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_orgs_updated_at BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_domains_updated_at BEFORE UPDATE ON project_domains
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_env_vars_updated_at BEFORE UPDATE ON project_env_vars
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
