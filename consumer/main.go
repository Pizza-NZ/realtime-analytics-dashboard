// Package consumer contains a standalone program that subscribes to analytic
// events and logs msgs from the NATS topic.
package main

import (
	"context"
	"encoding/json"
	"log"
	"os/signal"
	"pizza-nz/realtime-analytics-dashboard/pkg/models"
	"pizza-nz/realtime-analytics-dashboard/pkg/storage"
	"syscall"
	"time"

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
	conn, err := storage.ConnectToTimeSeriesDB(ctx)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	db := storage.NewTimeSeriesDatabase(conn)

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

		if err := db.InsertData(insertCtx, event); err != nil {
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
