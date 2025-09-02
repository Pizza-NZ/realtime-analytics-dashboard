// Package consumer contains a standalone program that subscribes to analytic
// events and logs msgs from the NATS topic.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"pizza-nz/realtime-analytics-dashboard/pkg/models"

	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("failed to connect to nats gateway: %v", err)
	}
	defer nc.Close()

	handler := func(m *nats.Msg) {
		var event models.AnalyticsEvent
		json.Unmarshal(m.Data, &event)
		log.Printf("Received a message: %s\n", string(m.Data))
	}
	_, err = nc.Subscribe(string(models.AnalyticsSub), handler)
	if err != nil {
		os.Exit(1)
	}

	// block forever
	select {}
}

func connectDB() (*pgx.Conn, error) {
	dsn, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err = conn.Ping(context.Background()); err != nil {
		conn.Close(context.Background())
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Successfully connected to the database.")
	return conn, nil
}
