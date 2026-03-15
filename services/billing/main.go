package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.Println("╔══════════════════════════════════════╗")
	log.Println("║    Antisky Billing Service v2.0      ║")
	log.Println("║    Powered by Razorpay               ║")
	log.Println("╚══════════════════════════════════════╝")

	port := getEnv("BILLING_PORT", "8087")
	dbURL := getEnv("DATABASE_URL", "postgres://antisky:antisky_dev_password@localhost:5432/antisky?sslmode=disable")
	razorpayKeyID := getEnv("RAZORPAY_KEY_ID", "")
	razorpayKeySecret := getEnv("RAZORPAY_KEY_SECRET", "")

	if razorpayKeyID != "" {
		log.Println("✓ Razorpay configured")
	} else {
		log.Println("⚠ Razorpay keys not set — billing disabled")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()
	log.Println("✓ Connected to PostgreSQL")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"healthy","service":"billing","provider":"razorpay"}`))
	})

	// Plans
	r.Get("/api/v1/billing/plans", func(w http.ResponseWriter, r *http.Request) {
		rows, err := pool.Query(r.Context(), `SELECT id, name, slug, price_cents, features, limits, is_active FROM billing_plans WHERE is_active = true ORDER BY price_cents`)
		if err != nil {
			jsonError(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var plans []map[string]interface{}
		for rows.Next() {
			var id, name, slug string
			var priceCents int
			var features, limits, isActive interface{}
			rows.Scan(&id, &name, &slug, &priceCents, &features, &limits, &isActive)
			plans = append(plans, map[string]interface{}{
				"id": id, "name": name, "slug": slug,
				"price_cents": priceCents, "features": features, "limits": limits,
			})
		}
		jsonResponse(w, 200, map[string]interface{}{"plans": plans})
	})

	// Create Razorpay order
	r.Post("/api/v1/billing/checkout", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			OrgID    string `json:"org_id"`
			PlanSlug string `json:"plan_slug"`
			Email    string `json:"email"`
			Amount   int    `json:"amount"` // amount in paise (INR smallest unit)
		}
		json.NewDecoder(r.Body).Decode(&req)

		if razorpayKeyID == "" {
			jsonError(w, "Razorpay not configured", 500)
			return
		}

		// Get plan from DB
		var planName string
		var priceCents int
		err := pool.QueryRow(r.Context(), `SELECT name, price_cents FROM billing_plans WHERE slug = $1`, req.PlanSlug).Scan(&planName, &priceCents)
		if err != nil {
			// Use the provided amount if plan not in DB
			if req.Amount > 0 {
				priceCents = req.Amount
				planName = req.PlanSlug
			} else {
				jsonError(w, "plan not found", 400)
				return
			}
		}

		// Create Razorpay order via HTTP API
		orderPayload := map[string]interface{}{
			"amount":   priceCents, // Razorpay expects amount in paise
			"currency": "INR",
			"receipt":  fmt.Sprintf("order_%s_%s", req.OrgID, req.PlanSlug),
			"notes": map[string]string{
				"org_id":    req.OrgID,
				"plan_slug": req.PlanSlug,
				"email":     req.Email,
			},
		}

		orderJSON, _ := json.Marshal(orderPayload)
		httpReq, _ := http.NewRequest("POST", "https://api.razorpay.com/v1/orders", bytes.NewBuffer(orderJSON))
		httpReq.SetBasicAuth(razorpayKeyID, razorpayKeySecret)
		httpReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			jsonError(w, "razorpay error: "+err.Error(), 500)
			return
		}
		defer resp.Body.Close()

		var orderResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&orderResp)

		if resp.StatusCode != 200 {
			jsonError(w, fmt.Sprintf("razorpay order creation failed: %v", orderResp), resp.StatusCode)
			return
		}

		jsonResponse(w, 200, map[string]interface{}{
			"order_id": orderResp["id"],
			"amount":   orderResp["amount"],
			"currency": orderResp["currency"],
			"key_id":   razorpayKeyID,
		})
	})

	// Verify Razorpay payment
	r.Post("/api/v1/billing/verify", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			OrderID   string `json:"razorpay_order_id"`
			PaymentID string `json:"razorpay_payment_id"`
			Signature string `json:"razorpay_signature"`
			OrgID     string `json:"org_id"`
			PlanSlug  string `json:"plan_slug"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		// Verify payment by fetching it from Razorpay
		httpReq, _ := http.NewRequest("GET", "https://api.razorpay.com/v1/payments/"+req.PaymentID, nil)
		httpReq.SetBasicAuth(razorpayKeyID, razorpayKeySecret)

		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			jsonError(w, "payment verification failed", 500)
			return
		}
		defer resp.Body.Close()

		var payment map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&payment)

		status, _ := payment["status"].(string)
		if status != "captured" && status != "authorized" {
			jsonError(w, "payment not successful: "+status, 400)
			return
		}

		// Record payment in DB
		pool.Exec(r.Context(),
			`INSERT INTO billing_payments (org_id, razorpay_payment_id, razorpay_order_id, amount, currency, status)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (razorpay_payment_id) DO UPDATE SET status = $6`,
			req.OrgID, req.PaymentID, req.OrderID, payment["amount"], payment["currency"], status,
		)

		// Update subscription
		pool.Exec(r.Context(),
			`INSERT INTO billing_subscriptions (org_id, plan_id, razorpay_payment_id, status, current_period_start, current_period_end)
			 SELECT $1, id, $3, 'active', NOW(), NOW() + INTERVAL '30 days'
			 FROM billing_plans WHERE slug = $2
			 ON CONFLICT (org_id) DO UPDATE SET plan_id = EXCLUDED.plan_id, status = 'active', razorpay_payment_id = $3, current_period_start = NOW(), current_period_end = NOW() + INTERVAL '30 days', updated_at = NOW()`,
			req.OrgID, req.PlanSlug, req.PaymentID,
		)

		log.Printf("💰 Payment verified: %s for org %s", req.PaymentID, req.OrgID)
		jsonResponse(w, 200, map[string]string{"status": "payment verified", "payment_id": req.PaymentID})
	})

	// Get subscription info
	r.Get("/api/v1/billing/subscription/{orgID}", func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "orgID")
		var sub struct {
			PlanName    string `json:"plan_name"`
			Status      string `json:"status"`
			PriceCents  int    `json:"price_cents"`
			PeriodStart string `json:"period_start"`
			PeriodEnd   string `json:"period_end"`
		}

		err := pool.QueryRow(r.Context(),
			`SELECT bp.name, bs.status, bp.price_cents, bs.current_period_start, bs.current_period_end
			 FROM billing_subscriptions bs JOIN billing_plans bp ON bp.id = bs.plan_id
			 WHERE bs.org_id = $1 AND bs.status = 'active' LIMIT 1`, orgID,
		).Scan(&sub.PlanName, &sub.Status, &sub.PriceCents, &sub.PeriodStart, &sub.PeriodEnd)

		if err != nil {
			jsonResponse(w, 200, map[string]string{"plan": "free", "status": "active"})
			return
		}
		jsonResponse(w, 200, sub)
	})

	// Usage tracking
	r.Post("/api/v1/billing/usage", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			OrgID     string  `json:"org_id"`
			ProjectID string  `json:"project_id"`
			Type      string  `json:"type"`
			Quantity  float64 `json:"quantity"`
			Unit      string  `json:"unit"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		_, err := pool.Exec(r.Context(),
			`INSERT INTO usage_records (org_id, project_id, type, quantity, unit) VALUES ($1, $2, $3, $4, $5)`,
			req.OrgID, req.ProjectID, req.Type, req.Quantity, req.Unit,
		)
		if err != nil {
			jsonError(w, err.Error(), 500)
			return
		}
		jsonResponse(w, 201, map[string]string{"status": "recorded"})
	})

	// Usage summary
	r.Get("/api/v1/billing/usage/{orgID}", func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "orgID")
		rows, err := pool.Query(r.Context(),
			`SELECT type, SUM(quantity) as total, unit
			 FROM usage_records WHERE org_id = $1 AND recorded_at > NOW() - INTERVAL '30 days'
			 GROUP BY type, unit`, orgID,
		)
		if err != nil {
			jsonError(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		usage := make(map[string]interface{})
		for rows.Next() {
			var typ, unit string
			var total float64
			rows.Scan(&typ, &total, &unit)
			usage[typ] = map[string]interface{}{"total": total, "unit": unit}
		}
		jsonResponse(w, 200, map[string]interface{}{"usage": usage, "period": "30d"})
	})

	// Razorpay webhook
	r.Post("/api/v1/billing/webhook", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		var event map[string]interface{}
		if err := json.Unmarshal(body, &event); err != nil {
			jsonError(w, "invalid webhook payload", 400)
			return
		}

		eventType, _ := event["event"].(string)
		log.Printf("📧 Razorpay webhook: %s", eventType)

		switch eventType {
		case "payment.captured":
			log.Println("💰 Payment captured via webhook")
		case "payment.failed":
			log.Println("⚠ Payment failed via webhook")
		case "subscription.activated":
			log.Println("✓ Subscription activated")
		case "subscription.cancelled":
			log.Println("✗ Subscription cancelled")
		}

		w.WriteHeader(200)
	})

	// Cancel subscription
	r.Post("/api/v1/billing/cancel/{orgID}", func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "orgID")

		pool.Exec(r.Context(),
			`UPDATE billing_subscriptions SET status = 'cancelled', cancel_at = NOW() WHERE org_id = $1 AND status = 'active'`, orgID)

		jsonResponse(w, 200, map[string]string{"status": "subscription cancelled"})
	})

	server := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		log.Printf("💳 Billing service (Razorpay) listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	server.Shutdown(context.Background())
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error":"%s"}`, message)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
