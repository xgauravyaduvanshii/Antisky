package main

import (
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
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/webhook"
)

func main() {
	log.Println("╔══════════════════════════════════════╗")
	log.Println("║    Antisky Billing Service v1.0      ║")
	log.Println("╚══════════════════════════════════════╝")

	port := getEnv("BILLING_PORT", "8087")
	dbURL := getEnv("DATABASE_URL", "postgres://antisky:antisky_dev_password@localhost:5432/antisky?sslmode=disable")
	stripeKey := getEnv("STRIPE_SECRET_KEY", "")
	webhookSecret := getEnv("STRIPE_WEBHOOK_SECRET", "")

	if stripeKey != "" {
		stripe.Key = stripeKey
		log.Println("✓ Stripe configured")
	} else {
		log.Println("⚠ Stripe key not set — billing disabled")
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
		w.Write([]byte(`{"status":"healthy","service":"billing"}`))
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

	// Create Stripe checkout session
	r.Post("/api/v1/billing/checkout", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			OrgID   string `json:"org_id"`
			PlanSlug string `json:"plan_slug"`
			Email   string `json:"email"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		// Get plan price ID
		var stripePriceID string
		pool.QueryRow(r.Context(), `SELECT stripe_price_id FROM billing_plans WHERE slug = $1`, req.PlanSlug).Scan(&stripePriceID)

		if stripePriceID == "" {
			jsonError(w, "plan not found or no Stripe price configured", 400)
			return
		}

		// Create or get Stripe customer
		var stripeCustomerID string
		pool.QueryRow(r.Context(), `SELECT stripe_customer_id FROM organizations WHERE id = $1`, req.OrgID).Scan(&stripeCustomerID)

		if stripeCustomerID == "" {
			cust, err := customer.New(&stripe.CustomerParams{
				Email: stripe.String(req.Email),
				Params: stripe.Params{Metadata: map[string]string{"org_id": req.OrgID}},
			})
			if err != nil {
				jsonError(w, "stripe error: "+err.Error(), 500)
				return
			}
			stripeCustomerID = cust.ID
			pool.Exec(r.Context(), `UPDATE organizations SET stripe_customer_id = $1 WHERE id = $2`, stripeCustomerID, req.OrgID)
		}

		// Create checkout session
		params := &stripe.CheckoutSessionParams{
			Customer: stripe.String(stripeCustomerID),
			Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{Price: stripe.String(stripePriceID), Quantity: stripe.Int64(1)},
			},
			SuccessURL: stripe.String(getEnv("PLATFORM_URL", "http://localhost:3000") + "/billing/success"),
			CancelURL:  stripe.String(getEnv("PLATFORM_URL", "http://localhost:3000") + "/billing/cancel"),
		}

		s, err := session.New(params)
		if err != nil {
			jsonError(w, "checkout error: "+err.Error(), 500)
			return
		}

		jsonResponse(w, 200, map[string]string{"checkout_url": s.URL, "session_id": s.ID})
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

	// Stripe webhook
	r.Post("/api/v1/billing/webhook", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		sig := r.Header.Get("Stripe-Signature")

		event, err := webhook.ConstructEvent(body, sig, webhookSecret)
		if err != nil {
			jsonError(w, "webhook verification failed", 400)
			return
		}

		log.Printf("📧 Stripe webhook: %s", event.Type)

		switch event.Type {
		case "customer.subscription.created", "customer.subscription.updated":
			var sub stripe.Subscription
			json.Unmarshal(event.Data.Raw, &sub)
			handleSubscriptionChange(r.Context(), pool, &sub)
		case "customer.subscription.deleted":
			var sub stripe.Subscription
			json.Unmarshal(event.Data.Raw, &sub)
			handleSubscriptionCancelled(r.Context(), pool, &sub)
		case "invoice.paid":
			log.Println("💰 Invoice paid")
		case "invoice.payment_failed":
			log.Println("⚠ Payment failed")
		}

		w.WriteHeader(200)
	})

	// Cancel subscription
	r.Post("/api/v1/billing/cancel/{orgID}", func(w http.ResponseWriter, r *http.Request) {
		orgID := chi.URLParam(r, "orgID")
		var stripeSubID string
		pool.QueryRow(r.Context(),
			`SELECT stripe_subscription_id FROM billing_subscriptions WHERE org_id = $1 AND status = 'active'`, orgID,
		).Scan(&stripeSubID)

		if stripeSubID == "" {
			jsonError(w, "no active subscription", 400)
			return
		}

		_, err := subscription.Cancel(stripeSubID, nil)
		if err != nil {
			jsonError(w, "cancel failed: "+err.Error(), 500)
			return
		}

		pool.Exec(r.Context(),
			`UPDATE billing_subscriptions SET status = 'cancelled', cancel_at = NOW() WHERE org_id = $1 AND status = 'active'`, orgID)

		jsonResponse(w, 200, map[string]string{"status": "subscription cancelled"})
	})

	server := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		log.Printf("💳 Billing service listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	server.Shutdown(context.Background())
}

func handleSubscriptionChange(ctx context.Context, pool *pgxpool.Pool, sub *stripe.Subscription) {
	customerID := sub.Customer.ID
	var orgID string
	pool.QueryRow(ctx, `SELECT id FROM organizations WHERE stripe_customer_id = $1`, customerID).Scan(&orgID)
	if orgID == "" {
		return
	}

	pool.Exec(ctx,
		`INSERT INTO billing_subscriptions (org_id, plan_id, stripe_subscription_id, stripe_customer_id, status, current_period_start, current_period_end)
		 SELECT $1, id, $3, $4, $5, to_timestamp($6), to_timestamp($7)
		 FROM billing_plans WHERE stripe_price_id = $2
		 ON CONFLICT (org_id) DO UPDATE SET status = $5, current_period_start = to_timestamp($6), current_period_end = to_timestamp($7), updated_at = NOW()`,
		orgID, sub.Items.Data[0].Price.ID, sub.ID, customerID, string(sub.Status),
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd,
	)
}

func handleSubscriptionCancelled(ctx context.Context, pool *pgxpool.Pool, sub *stripe.Subscription) {
	pool.Exec(ctx,
		`UPDATE billing_subscriptions SET status = 'cancelled', updated_at = NOW() WHERE stripe_subscription_id = $1`,
		sub.ID,
	)
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
