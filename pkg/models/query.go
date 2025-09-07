package models

import "time"

// TimeBucketStats represents a Time Bucket query result.
type TimeBucketStats struct {
	Time  time.Time
	Count int64
}
