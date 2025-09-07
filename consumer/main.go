// Package consumer contains a standalone program that subscribes to analytic
// events and logs msgs from the NATS topic.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"pizza-nz/realtime-analytics-dashboard/pkg/models"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go"
)

func main() {
	// --- NATS Connection ---
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("failed to connect to nats gateway: %v", err)
	}
	defer nc.Close()

	// --- Signal Handling ---
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// --- Database Connection ---
	conn, err := connectDB(ctx)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// --- NATS Subscription Handler
	handler := func(m *nats.Msg) {
		var event models.AnalyticsEvent
		err = json.Unmarshal(m.Data, &event)
		if err != nil {
			log.Printf("Error unmarshaling message data: %v", err)
			m.Term()
			return
		}

		log.Printf("Received message for user %d: event type %s", event.UserID, event.EventType)

		insertCtx, cancelInsert := context.WithTimeout(ctx, 5*time.Second)
		defer cancelInsert()

		if err := insertData(conn, insertCtx, event); err != nil {
			log.Printf("Error inserting data: %v", err)
			m.Nak()
		} else {
			m.Ack()
		}
	}

	sub, err := nc.Subscribe(string(models.AnalyticsSub), handler)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic %s: %v", models.AnalyticsSub, err)
	}
	log.Println("Consumer successfully started. Waiting for messages...")

	// --- Wait for Shutdown Signal ---
	// Block here until the context created by signal.NotifyContext is canceled.
	<-ctx.Done()

	// --- Graceful Shutdown Sequence ---
	log.Println("Shutdown signal received. Starting graceful shutdown...")

	// 1. Drain NATS subscription.
	// This stops receiving new messages and waits for processing messages to finish.
	log.Println("Draining NATS subscription...")
	if err := sub.Drain(); err != nil {
		log.Printf("Error draining NATS subscription: %v", err)
	}

	// 2. Close NATS connection.
	nc.Close()
	log.Println("NATS connection closed.")

	// 3. Close Database Connection.
	log.Println("Closing database connection...")
	// Use context.Background() for cleanup tasks to ensure they run even if the main context timed out.
	if err := conn.Close(context.Background()); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}

	log.Println("Shutdown complete.")
}

func connectDB(ctx context.Context) (*pgx.Conn, error) {
	dsn, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err = conn.Ping(ctx); err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Successfully connected to the database.")
	return conn, nil
}

func insertData(conn *pgx.Conn, ctx context.Context, event models.AnalyticsEvent) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	sql := `INSERT INTO analytics_events (event_type, user_id, created_at, event_data) 
VALUES ($1, $2, $3, $4);`

	if _, err := tx.Exec(ctx, sql, event.EventType, event.UserID, event.CreatedAt, event.EventData); err != nil {
		return fmt.Errorf("failed to execute insert: %w", err)
	}

	return tx.Commit(ctx)
}
