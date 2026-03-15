package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/antisky/services/server-manager/internal/handlers"
	"github.com/antisky/services/server-manager/internal/store"
)

func main() {
	log.Println("╔══════════════════════════════════════╗")
	log.Println("║   Antisky Server Manager v1.0        ║")
	log.Println("╚══════════════════════════════════════╝")

	port := getEnv("SERVER_MANAGER_PORT", "8083")
	dbURL := getEnv("DATABASE_URL", "postgres://antisky:antisky_dev_password@localhost:5432/antisky?sslmode=disable")
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	clusterSecret := getEnv("CLUSTER_SECRET", "antisky-cluster-secret-2026")

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()
	log.Println("✓ Connected to PostgreSQL")

	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("⚠ Redis connection warning: %v (continuing without Redis caching)", err)
	} else {
		log.Println("✓ Connected to Redis")
	}

	serverStore := store.NewServerStore(pool)
	h := handlers.NewServerHandler(serverStore, rdb, clusterSecret)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(corsMiddleware)

	// Health
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"healthy","service":"server-manager"}`))
	})

	// Server registration (from builder nodes)
	r.Post("/api/v1/servers/register", h.RegisterServer)
	r.Post("/api/v1/servers/heartbeat", h.ReceiveHeartbeat)

	// Admin routes (require admin auth)
	r.Route("/api/v1/admin/servers", func(r chi.Router) {
		r.Get("/", h.ListServers)
		r.Get("/{serverID}", h.GetServer)
		r.Delete("/{serverID}", h.DecommissionServer)
		r.Post("/{serverID}/command", h.SendCommand)
		r.Get("/{serverID}/metrics", h.GetServerMetrics)
		r.Get("/{serverID}/commands", h.GetServerCommands)
		r.Post("/{serverID}/drain", h.DrainServer)
	})

	r.Post("/api/v1/admin/cache/flush", h.FlushCache)

	// Admin user management
	r.Route("/api/v1/admin/users", func(r chi.Router) {
		r.Get("/", h.ListUsers)
		r.Get("/{userID}", h.GetUser)
		r.Delete("/{userID}", h.DeleteUser)
		r.Post("/{userID}/ban", h.BanUser)
		r.Post("/{userID}/unban", h.UnbanUser)
		r.Post("/{userID}/impersonate", h.ImpersonateUser)
		r.Put("/{userID}/role", h.SetUserRole)
	})

	// Admin platform stats
	r.Get("/api/v1/admin/stats", h.GetPlatformStats)
	r.Get("/api/v1/admin/cluster", h.GetClusterConfig)
	r.Put("/api/v1/admin/cluster/{key}", h.UpdateClusterConfig)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("🚀 Server Manager listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Start stale server detector
	go detectStaleServers(ctx, serverStore)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server manager...")
	server.Shutdown(context.Background())
}

func detectStaleServers(ctx context.Context, s *store.ServerStore) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, _ := s.MarkStaleServers(ctx, 2*time.Minute)
			if count > 0 {
				log.Printf("⚠ Marked %d servers as offline (no heartbeat)", count)
			}
		}
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Server-Key, X-Cluster-Secret")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
