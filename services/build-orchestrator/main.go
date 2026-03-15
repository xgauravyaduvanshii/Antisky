package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/antisky/services/build-orchestrator/internal/orchestrator"
)

func main() {
	log.Println("╔══════════════════════════════════════╗")
	log.Println("║   Antisky Build Orchestrator v0.1    ║")
	log.Println("╚══════════════════════════════════════╝")

	dbURL := getEnv("DATABASE_URL", "postgres://antisky:antisky_dev_password@localhost:5432/antisky?sslmode=disable")
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	controlPlaneURL := getEnv("CONTROL_PLANE_URL", "http://localhost:8080")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to PostgreSQL
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()
	log.Println("✓ Connected to PostgreSQL")

	// Connect to Redis
	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()
	log.Println("✓ Connected to Redis")

	// Create orchestrator
	orch := orchestrator.New(pool, rdb, controlPlaneURL)

	// Start listening for build events
	go orch.ListenForBuilds(ctx)

	// Start periodic health check for running builds
	go orch.MonitorBuilds(ctx)

	log.Println("🔨 Build orchestrator is running, waiting for build events...")

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down build orchestrator...")
	cancel()
	time.Sleep(2 * time.Second)
	log.Println("Build orchestrator stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// BuildEvent is the event published to Redis when a deploy is triggered
type BuildEvent struct {
	DeploymentID string `json:"deployment_id"`
	ProjectID    string `json:"project_id"`
	RepoURL      string `json:"repo_url"`
	Ref          string `json:"ref"`
	Runtime      string `json:"runtime"`
	Framework    string `json:"framework"`
	BuildCommand string `json:"build_command"`
	StartCommand string `json:"start_command"`
	OutputDir    string `json:"output_dir"`
	RootDir      string `json:"root_dir"`
}

func parseBuildEvent(data string) (*BuildEvent, error) {
	var event BuildEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return nil, err
	}
	return &event, nil
}
