// Package producer contains a standalone program that generates mock analytic
// events and publishes them to a NATS topic.
//
// This tool is designed to be used as a data source for the consumer and
// dashboard services in the real-time-analytics project. To run it, ensure a
// NATS server is accessible and execute `go run main.go`.
package main

import (
	"encoding/json"
	"log"
	"pizza-nz/realtime-analytics-dashboard/pkg/models"
	"time"

	"github.com/nats-io/nats.go"
)

// main is the entry point for the producer service.
// It connects to the default NATS server and enters an infinite loop,
// publishing a new, randomly generated AnalyticsEvent every second.
func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("failed to connect to nats gateway: %v", err)
	}
	defer nc.Close()

	for {
		event := models.GenerateEvent()

		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("error marshalling event: %v", err)
			continue
		}

		err = nc.Publish(string(models.AnalyticsSub), eventJSON)
		if err != nil {
			log.Printf("error publishing event: %v", err)
		} else {
			log.Printf("published event: %v", err)
		}

		time.Sleep(1 * time.Second)
	}
}
