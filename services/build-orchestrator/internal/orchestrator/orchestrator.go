package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// BuildEvent represents a build job from the control plane
type BuildEvent struct {
	DeploymentID string  `json:"deployment_id"`
	ProjectID    string  `json:"project_id"`
	RepoURL      *string `json:"repo_url"`
	Ref          string  `json:"ref"`
	Runtime      string  `json:"runtime"`
	Framework    *string `json:"framework"`
	BuildCommand *string `json:"build_command"`
	StartCommand *string `json:"start_command"`
	OutputDir    *string `json:"output_dir"`
	RootDir      string  `json:"root_dir"`
}

// BuildJob tracks an active build
type BuildJob struct {
	ID           string    `json:"id"`
	DeploymentID string    `json:"deployment_id"`
	ProjectID    string    `json:"project_id"`
	Status       string    `json:"status"` // queued, building, completed, failed
	StartedAt    time.Time `json:"started_at"`
	WorkerID     string    `json:"worker_id,omitempty"`
}

// Orchestrator manages build lifecycle
type Orchestrator struct {
	pool            *pgxpool.Pool
	rdb             *redis.Client
	controlPlaneURL string
	maxConcurrent   int
	activeBuilds    map[string]*BuildJob
}

func New(pool *pgxpool.Pool, rdb *redis.Client, controlPlaneURL string) *Orchestrator {
	return &Orchestrator{
		pool:            pool,
		rdb:             rdb,
		controlPlaneURL: controlPlaneURL,
		maxConcurrent:   5,
		activeBuilds:    make(map[string]*BuildJob),
	}
}

// ListenForBuilds subscribes to the build queue via Redis pub/sub
func (o *Orchestrator) ListenForBuilds(ctx context.Context) {
	sub := o.rdb.Subscribe(ctx, "builds:queue")
	defer sub.Close()

	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}

			var event BuildEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("Failed to parse build event: %v", err)
				continue
			}

			log.Printf("📦 Received build event: deployment=%s project=%s runtime=%s",
				event.DeploymentID, event.ProjectID, event.Runtime)

			go o.processBuild(ctx, &event)
		}
	}
}

// processBuild handles a single build
func (o *Orchestrator) processBuild(ctx context.Context, event *BuildEvent) {
	deploymentID := event.DeploymentID
	jobID := uuid.New().String()

	// Track active build
	job := &BuildJob{
		ID:           jobID,
		DeploymentID: deploymentID,
		ProjectID:    event.ProjectID,
		Status:       "building",
		StartedAt:    time.Now(),
	}
	o.activeBuilds[deploymentID] = job

	// Update deployment status to 'building'
	o.updateDeploymentStatus(ctx, deploymentID, "building", nil)
	o.appendLog(ctx, deploymentID, "🔨 Build started...")
	o.appendLog(ctx, deploymentID, fmt.Sprintf("   Runtime: %s", event.Runtime))
	if event.Framework != nil {
		o.appendLog(ctx, deploymentID, fmt.Sprintf("   Framework: %s", *event.Framework))
	}

	// Step 1: Detect build strategy
	strategy := o.detectBuildStrategy(event)
	o.appendLog(ctx, deploymentID, fmt.Sprintf("   Strategy: %s", strategy))

	// Step 2: Determine build commands
	buildCmd, startCmd := o.resolveBuildCommands(event)
	o.appendLog(ctx, deploymentID, fmt.Sprintf("   Build command: %s", buildCmd))
	o.appendLog(ctx, deploymentID, fmt.Sprintf("   Start command: %s", startCmd))

	// Step 3: In production, this would spawn an ECS Fargate task
	// For now, simulate the build process
	o.appendLog(ctx, deploymentID, "📥 Cloning repository...")
	time.Sleep(1 * time.Second)

	o.appendLog(ctx, deploymentID, "📦 Installing dependencies...")
	time.Sleep(1 * time.Second)

	o.appendLog(ctx, deploymentID, "🏗️  Building project...")
	time.Sleep(2 * time.Second)

	o.appendLog(ctx, deploymentID, "📤 Uploading artifacts...")
	time.Sleep(500 * time.Millisecond)

	// Step 4: Generate deployment URL
	previewURL := fmt.Sprintf("https://%s--%s.antisky.app", event.ProjectID[:8], deploymentID[:8])
	o.appendLog(ctx, deploymentID, fmt.Sprintf("✅ Build complete! URL: %s", previewURL))

	// Step 5: Update deployment status to 'ready'
	o.updateDeploymentStatus(ctx, deploymentID, "ready", nil)

	// Update deployment URL
	o.setDeploymentURL(ctx, deploymentID, previewURL)

	// Cleanup
	delete(o.activeBuilds, deploymentID)

	duration := time.Since(job.StartedAt)
	o.appendLog(ctx, deploymentID, fmt.Sprintf("⏱️  Total build time: %s", duration.Round(time.Millisecond)))
	log.Printf("✅ Build completed: deployment=%s duration=%s", deploymentID, duration.Round(time.Millisecond))
}

// detectBuildStrategy determines how to build based on project config
func (o *Orchestrator) detectBuildStrategy(event *BuildEvent) string {
	runtime := strings.ToLower(event.Runtime)

	switch runtime {
	case "static":
		return "static-upload"
	case "docker":
		return "dockerfile"
	case "nodejs":
		if event.Framework != nil {
			framework := strings.ToLower(*event.Framework)
			switch framework {
			case "nextjs":
				return "nextjs-ssr"
			case "nuxt", "nuxtjs":
				return "nuxt-ssr"
			case "remix":
				return "remix-ssr"
			case "astro":
				return "astro-ssg"
			default:
				return "nodejs-container"
			}
		}
		return "nodejs-container"
	case "go":
		return "go-binary"
	case "python":
		if event.Framework != nil {
			framework := strings.ToLower(*event.Framework)
			switch framework {
			case "django":
				return "python-django"
			case "fastapi":
				return "python-fastapi"
			default:
				return "python-container"
			}
		}
		return "python-container"
	case "php":
		return "php-container"
	case "ruby":
		return "ruby-container"
	default:
		return "auto-detect"
	}
}

// resolveBuildCommands returns build and start commands for the runtime
func (o *Orchestrator) resolveBuildCommands(event *BuildEvent) (string, string) {
	// Use explicit commands if provided
	if event.BuildCommand != nil && *event.BuildCommand != "" {
		buildCmd := *event.BuildCommand
		startCmd := "npm start"
		if event.StartCommand != nil && *event.StartCommand != "" {
			startCmd = *event.StartCommand
		}
		return buildCmd, startCmd
	}

	// Default commands by runtime
	runtime := strings.ToLower(event.Runtime)
	switch runtime {
	case "nodejs":
		return "npm run build", "npm start"
	case "go":
		return "go build -o app ./cmd/app", "./app"
	case "python":
		return "pip install -r requirements.txt", "gunicorn app:app"
	case "php":
		return "composer install --no-dev", "php-fpm"
	case "ruby":
		return "bundle install", "bundle exec puma"
	case "static":
		return "npm run build", ""
	default:
		return "echo 'no build step'", "echo 'no start command'"
	}
}

// MonitorBuilds periodically checks health of running builds
func (o *Orchestrator) MonitorBuilds(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for deployID, job := range o.activeBuilds {
				// Timeout builds after 30 minutes
				if time.Since(job.StartedAt) > 30*time.Minute {
					log.Printf("⚠️  Build timed out: deployment=%s", deployID)
					errMsg := "Build timed out after 30 minutes"
					o.updateDeploymentStatus(ctx, deployID, "failed", &errMsg)
					o.appendLog(ctx, deployID, "❌ Build timed out")
					delete(o.activeBuilds, deployID)
				}
			}

			if len(o.activeBuilds) > 0 {
				log.Printf("📊 Active builds: %d", len(o.activeBuilds))
			}
		}
	}
}

// appendLog adds a log line to the deployment's log stream in Redis
func (o *Orchestrator) appendLog(ctx context.Context, deploymentID, message string) {
	timestamp := time.Now().Format("15:04:05.000")
	logLine := fmt.Sprintf("[%s] %s", timestamp, message)

	logKey := "deploy:logs:" + deploymentID
	o.rdb.RPush(ctx, logKey, logLine)
	o.rdb.Expire(ctx, logKey, 24*time.Hour) // Keep logs for 24h

	// Also publish to live stream channel for real-time tailing
	o.rdb.Publish(ctx, "deploy:logs:stream:"+deploymentID, logLine)
}

// updateDeploymentStatus calls the control plane to update deployment status
func (o *Orchestrator) updateDeploymentStatus(ctx context.Context, deploymentID, status string, errorMsg *string) {
	// Update directly in database
	query := `UPDATE deployments SET status = $1`
	args := []interface{}{status}

	if status == "building" {
		query += `, started_at = NOW()`
	}
	if status == "ready" || status == "failed" || status == "cancelled" {
		query += `, completed_at = NOW()`
		if status == "ready" {
			query += `, build_duration_ms = EXTRACT(EPOCH FROM (NOW() - started_at)) * 1000`
		}
	}
	if errorMsg != nil {
		query += fmt.Sprintf(`, error_message = $%d`, len(args)+1)
		args = append(args, *errorMsg)
	}

	query += fmt.Sprintf(` WHERE id = $%d`, len(args)+1)
	args = append(args, deploymentID)

	_, err := o.pool.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("Failed to update deployment status: %v", err)
	}
}

// setDeploymentURL updates the deployment URL
func (o *Orchestrator) setDeploymentURL(ctx context.Context, deploymentID, url string) {
	_, err := o.pool.Exec(ctx,
		`UPDATE deployments SET url = $1, preview_url = $1 WHERE id = $2`,
		url, deploymentID,
	)
	if err != nil {
		log.Printf("Failed to set deployment URL: %v", err)
	}
}

// In production, this would use the AWS SDK to create ECS tasks
func (o *Orchestrator) spawnFargateWorker(ctx context.Context, event *BuildEvent) (string, error) {
	// TODO: Implement ECS Fargate task creation
	// ecs.RunTask with:
	// - task definition for build worker
	// - environment variables (repo URL, ref, build commands, etc.)
	// - network configuration (private subnet, security group)
	// - log configuration (CloudWatch logs)
	_ = ctx
	_ = event
	return "simulated-task-id", nil
}

// Unused but kept for when we integrate with the control plane HTTP API
func (o *Orchestrator) callControlPlane(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	_ = ctx
	_ = body
	url := o.controlPlaneURL + path
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}
