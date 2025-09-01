package models

import (
	"math/rand"
	"time"
)

// EventType defines the type of analytic event, such as a page view or login.
type EventType string

// EventSubject defines the NATS subject where events are published.
type EventSubject string

const (
	// PageView represents a user viewing a page.
	PageView EventType = "page_view"
	// UserLogin represents a user logging into the system.
	UserLogin EventType = "user_login"

	// AnalyticsSub is the NATS subject used for all analytics events.
	AnalyticsSub EventSubject = "events.analytics"
)

// AnalyticsEvent represents a single analytics event.
// This is the data structure that is marchalled to JSON and published to NATS.
type AnalyticsEvent struct {
	// EventType indicates the kind of event that occurred.
	EventType EventType `json:"event_type"`
	// UserID is the unique identifier for the user who triggered the event.
	UserID int64 `json:"user_id"`
	// CreatedAt is the UTC time at which the event was generated.
	CreatedAt time.Time `json:"created_at"`
	// EventData contains additional, unstructured data specific to the event type.
	// For a PageView, this might include the URL.
	EventData map[string]any `json:"event_data"`
}

// GenerateEvent creates a new AnalyticsEvent with randomized data.
func GenerateEvent() AnalyticsEvent {
	return AnalyticsEvent{
		EventType: randomType(),
		UserID:    randomID(),
		CreatedAt: time.Now().UTC(),
		EventData: randomDataMap(),
	}
}

func randomID() int64 {
	return rand.Int63n(300)
}
func randomType() EventType {
	if rand.Intn(2) == 0 {
		return PageView
	}
	return UserLogin
}
func randomDataMap() map[string]any {
	s1, s2 := randomData()
	return map[string]any{s1: s2}
}
func randomData() (string, string) {
	return "page", "/dashboard"
}
