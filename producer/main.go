package producer

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/nats-io/nats.go"
)

type EventType string
type EventSubject string

const (
	PageView  EventType = "page_view"
	UserLogin EventType = "user_login"

	AnalyticsSub EventSubject = "events.analytics"
)

type AnalyticsEvent struct {
	EventType EventType      `json:"event_type"`
	UserID    int64          `json:"user_id"`
	Timestamp time.Time      `json:"timestamp"`
	EventData map[string]any `json:"event_data"`
}

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("failed to connect to nats gateway: %v", err)
		panic(1)
	}
	defer nc.Close()

	for {
		event := makeEvent()

		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("error marshalling event: %v", err)
			continue
		}

		err = nc.Publish(string(AnalyticsSub), eventJSON)
		if err != nil {
			log.Printf("error publishing event: %v", err)
		} else {
			log.Printf("published event: %v", err)
		}

		time.Sleep(1 * time.Second)
	}
}

func makeEvent() AnalyticsEvent {
	return AnalyticsEvent{
		EventType: randomType(),
		UserID:    randomID(),
		Timestamp: time.Now(),
		EventData: randomDataMap(),
	}
}

func randomID() int64 {
	return rand.Int63n(300)
}
func randomType() EventType {
	if 1+rand.Intn(2)%2 == 0 {
		return PageView
	} else {
		return UserLogin
	}
}
func randomDataMap() map[string]any {
	s1, s2 := randomData()
	return map[string]any{s1: s2}
}
func randomData() (string, string) {
	return "page", "/dashboard"
}
