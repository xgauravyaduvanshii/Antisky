package models

import (
	"time"

	"github.com/google/uuid"
)

// --- Organizations ---

type Organization struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Slug             string    `json:"slug"`
	AvatarURL        *string   `json:"avatar_url,omitempty"`
	Plan             string    `json:"plan"`
	BillingEmail     *string   `json:"billing_email,omitempty"`
	StripeCustomerID *string   `json:"-"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type OrgMember struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	InvitedBy *uuid.UUID `json:"invited_by,omitempty"`
	JoinedAt  time.Time `json:"joined_at"`
	// Joined fields
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

type CreateOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type InviteMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// --- Projects ---

type Project struct {
	ID             uuid.UUID `json:"id"`
	OrgID          uuid.UUID `json:"org_id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    *string   `json:"description,omitempty"`
	RepoURL        *string   `json:"repo_url,omitempty"`
	RepoProvider   *string   `json:"repo_provider,omitempty"`
	RepoID         *string   `json:"repo_id,omitempty"`
	RepoBranch     string    `json:"repo_branch"`
	Framework      *string   `json:"framework,omitempty"`
	Runtime        string    `json:"runtime"`
	BuildCommand   *string   `json:"build_command,omitempty"`
	StartCommand   *string   `json:"start_command,omitempty"`
	OutputDir      *string   `json:"output_dir,omitempty"`
	RootDir        string    `json:"root_dir"`
	InstallCommand *string   `json:"install_command,omitempty"`
	NodeVersion    *string   `json:"node_version,omitempty"`
	AutoDeploy     bool      `json:"auto_deploy"`
	Config         map[string]interface{} `json:"config,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	// Computed fields
	LatestDeployment *Deployment `json:"latest_deployment,omitempty"`
	DomainCount      int         `json:"domain_count,omitempty"`
}

type CreateProjectRequest struct {
	OrgID          uuid.UUID `json:"org_id"`
	Name           string    `json:"name"`
	RepoURL        *string   `json:"repo_url,omitempty"`
	RepoProvider   *string   `json:"repo_provider,omitempty"`
	Framework      *string   `json:"framework,omitempty"`
	Runtime        string    `json:"runtime"`
	BuildCommand   *string   `json:"build_command,omitempty"`
	StartCommand   *string   `json:"start_command,omitempty"`
	OutputDir      *string   `json:"output_dir,omitempty"`
	RootDir        string    `json:"root_dir"`
	RepoBranch     string    `json:"repo_branch"`
}

type UpdateProjectRequest struct {
	Name           *string   `json:"name,omitempty"`
	Description    *string   `json:"description,omitempty"`
	BuildCommand   *string   `json:"build_command,omitempty"`
	StartCommand   *string   `json:"start_command,omitempty"`
	OutputDir      *string   `json:"output_dir,omitempty"`
	Framework      *string   `json:"framework,omitempty"`
	Runtime        *string   `json:"runtime,omitempty"`
	AutoDeploy     *bool     `json:"auto_deploy,omitempty"`
	RepoBranch     *string   `json:"repo_branch,omitempty"`
}

// --- Domains ---

type ProjectDomain struct {
	ID                uuid.UUID `json:"id"`
	ProjectID         uuid.UUID `json:"project_id"`
	Domain            string    `json:"domain"`
	IsPrimary         bool      `json:"is_primary"`
	Verified          bool      `json:"verified"`
	VerificationToken *string   `json:"verification_token,omitempty"`
	SSLStatus         string    `json:"ssl_status"`
	SSLCertARN        *string   `json:"-"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type AddDomainRequest struct {
	Domain string `json:"domain"`
}

// --- Deployments ---

type Deployment struct {
	ID              uuid.UUID  `json:"id"`
	ProjectID       uuid.UUID  `json:"project_id"`
	TriggeredBy     *uuid.UUID `json:"triggered_by,omitempty"`
	Ref             *string    `json:"ref,omitempty"`
	CommitSHA       *string    `json:"commit_sha,omitempty"`
	CommitMessage   *string    `json:"commit_message,omitempty"`
	CommitAuthor    *string    `json:"commit_author,omitempty"`
	Status          string     `json:"status"`
	Type            string     `json:"type"`
	URL             *string    `json:"url,omitempty"`
	PreviewURL      *string    `json:"preview_url,omitempty"`
	BuildLogURL     *string    `json:"build_log_url,omitempty"`
	BuildDurationMs *int       `json:"build_duration_ms,omitempty"`
	Meta            map[string]interface{} `json:"meta,omitempty"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	// Computed
	TriggeredByName string `json:"triggered_by_name,omitempty"`
}

type TriggerDeployRequest struct {
	Ref     string `json:"ref"`
	Source  string `json:"source"` // cli, dashboard, webhook, ide
}

// --- Environment Variables ---

type EnvVar struct {
	ID             uuid.UUID `json:"id"`
	ProjectID      uuid.UUID `json:"project_id"`
	Key            string    `json:"key"`
	EncryptedValue string    `json:"-"`
	Value          string    `json:"value,omitempty"` // only populated on read
	Target         []string  `json:"target"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type SetEnvVarRequest struct {
	Key    string   `json:"key"`
	Value  string   `json:"value"`
	Target []string `json:"target,omitempty"`
}

// --- Common ---

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	TotalCount int         `json:"total_count"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
}
