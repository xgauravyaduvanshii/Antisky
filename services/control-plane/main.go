package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/antisky/services/control-plane/internal/handlers"
	"github.com/antisky/services/control-plane/internal/middleware"
	"github.com/antisky/services/control-plane/internal/store"
)

func main() {
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://antisky:antisky_dev_password@localhost:5432/antisky?sslmode=disable")
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	jwtSecret := getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production")

	ctx := context.Background()

	// Connect to PostgreSQL
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping PostgreSQL: %v", err)
	}
	log.Println("✓ Connected to PostgreSQL")

	// Connect to Redis
	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}
	log.Println("✓ Connected to Redis")

	// Initialize stores
	projectStore := store.NewProjectStore(pool)
	deployStore := store.NewDeploymentStore(pool)
	orgStore := store.NewOrgStore(pool)
	envVarStore := store.NewEnvVarStore(pool)

	// Initialize handlers
	projectHandler := handlers.NewProjectHandler(projectStore, orgStore)
	deployHandler := handlers.NewDeploymentHandler(deployStore, projectStore, rdb)
	orgHandler := handlers.NewOrgHandler(orgStore)
	envVarHandler := handlers.NewEnvVarHandler(envVarStore, projectStore)
	webhookHandler := handlers.NewWebhookHandler(deployStore, projectStore)

	// JWT middleware
	jwtAuth := middleware.NewJWTAuth(jwtSecret)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://*.antisky.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"healthy","service":"control-plane"}`)
	})

	// Webhook routes (no auth — verified by signature)
	r.Route("/api/v1/webhooks", func(r chi.Router) {
		r.Post("/github", webhookHandler.GitHubWebhook)
		r.Post("/gitlab", webhookHandler.GitLabWebhook)
	})

	// API v1 routes (authenticated)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(jwtAuth.Middleware)

		// Organizations
		r.Route("/orgs", func(r chi.Router) {
			r.Get("/", orgHandler.List)
			r.Post("/", orgHandler.Create)
			r.Route("/{orgId}", func(r chi.Router) {
				r.Get("/", orgHandler.Get)
				r.Put("/", orgHandler.Update)
				r.Delete("/", orgHandler.Delete)

				// Members
				r.Get("/members", orgHandler.ListMembers)
				r.Post("/members", orgHandler.InviteMember)
				r.Delete("/members/{userId}", orgHandler.RemoveMember)
			})
		})

		// Projects
		r.Route("/projects", func(r chi.Router) {
			r.Get("/", projectHandler.List)
			r.Post("/", projectHandler.Create)
			r.Route("/{projectId}", func(r chi.Router) {
				r.Get("/", projectHandler.Get)
				r.Put("/", projectHandler.Update)
				r.Delete("/", projectHandler.Delete)

				// Deployments
				r.Post("/deploy", deployHandler.TriggerDeploy)
				r.Get("/deployments", deployHandler.List)
				r.Route("/deployments/{deployId}", func(r chi.Router) {
					r.Get("/", deployHandler.Get)
					r.Post("/cancel", deployHandler.Cancel)
					r.Post("/rollback", deployHandler.Rollback)
					r.Get("/logs", deployHandler.GetLogs)
				})

				// Environment variables
				r.Get("/env", envVarHandler.List)
				r.Post("/env", envVarHandler.Set)
				r.Delete("/env/{key}", envVarHandler.Delete)

				// Domains
				r.Get("/domains", projectHandler.ListDomains)
				r.Post("/domains", projectHandler.AddDomain)
				r.Delete("/domains/{domainId}", projectHandler.RemoveDomain)
			})
		})
	})

	// Start server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("🚀 Control Plane API listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down control plane...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Control plane stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
