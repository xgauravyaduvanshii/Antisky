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

	"github.com/antisky/services/auth/internal/handlers"
	"github.com/antisky/services/auth/internal/middleware"
	"github.com/antisky/services/auth/internal/store"
)

func main() {
	// Load config from environment
	port := getEnv("PORT", "8081")
	dbURL := getEnv("DATABASE_URL", "postgres://antisky:antisky_dev_password@localhost:5432/antisky?sslmode=disable")
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	jwtSecret := getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production")
	jwtExpiry := getEnvDuration("JWT_EXPIRY", 15*time.Minute)
	refreshExpiry := getEnvDuration("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour)

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

	// Connect to Redis (optional — falls back to DB-only sessions)
	var rdb *redis.Client
	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("⚠ Redis URL parse error: %v (continuing without cache)", err)
	} else {
		rdb = redis.NewClient(redisOpts)
		if err := rdb.Ping(ctx).Err(); err != nil {
			log.Printf("⚠ Redis connection failed: %v (continuing without cache)", err)
			rdb.Close()
			rdb = nil
		} else {
			defer rdb.Close()
			log.Println("✓ Connected to Redis")
		}
	}

	// Initialize stores and handlers
	userStore := store.NewUserStore(pool)
	sessionStore := store.NewSessionStore(pool, rdb)
	apiKeyStore := store.NewAPIKeyStore(pool)

	jwtManager := middleware.NewJWTManager(jwtSecret, jwtExpiry, refreshExpiry)
	authHandler := handlers.NewAuthHandler(userStore, sessionStore, apiKeyStore, jwtManager)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyStore, jwtManager)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://*.antisky.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"healthy","service":"auth"}`)
	})

	// Auth routes (public)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.RefreshToken)

		// OAuth callbacks
		r.Get("/github/callback", authHandler.GitHubCallback)
	})

	// Protected routes
	r.Route("/auth", func(r chi.Router) {
		r.Use(middleware.JWTAuth(jwtManager))
		r.Post("/logout", authHandler.Logout)
		r.Get("/me", authHandler.GetCurrentUser)
		r.Put("/me", authHandler.UpdateProfile)
	})

	// API Key management (protected)
	r.Route("/api-keys", func(r chi.Router) {
		r.Use(middleware.JWTAuth(jwtManager))
		r.Get("/", apiKeyHandler.List)
		r.Post("/", apiKeyHandler.Create)
		r.Delete("/{id}", apiKeyHandler.Revoke)
	})

	// Start server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("🚀 Auth service listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down auth service...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Auth service stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fallback
		}
		return d
	}
	return fallback
}
