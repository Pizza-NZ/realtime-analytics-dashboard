package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"pizza-nz/realtime-analytics-dashboard/pkg/models"

	"github.com/jackc/pgx/v5"
)

type DatabaseRepository interface {
	InsertData(ctx context.Context, event models.AnalyticsEvent) error
	GetTimeBucket(ctx context.Context) ([]models.TimeBucketStats, error)
	Close(ctx context.Context) error
}

type TimeSeriesDatabase struct {
	db *pgx.Conn
}

func NewTimeSeriesDatabase(conn *pgx.Conn) *TimeSeriesDatabase {
	return &TimeSeriesDatabase{
		db: conn,
	}
}

func ConnectToTimeSeriesDB(ctx context.Context) (*pgx.Conn, error) {
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

func (ts *TimeSeriesDatabase) InsertData(ctx context.Context, event models.AnalyticsEvent) error {
	tx, err := ts.db.Begin(ctx)
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

func (ts *TimeSeriesDatabase) GetTimeBucket(ctx context.Context) ([]models.TimeBucketStats, error) {
	sql := `SELECT time_bucket('1 second', created_at) AS "Time", COUNT(*) AS "Count"
		FROM analytics_events GROUP BY time_bucket('1 second', created_at);`

	var results []models.TimeBucketStats
	rows, err := ts.db.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var result models.TimeBucketStats
		if err := rows.Scan(&result.Time, &result.Count); err != nil {
			log.Printf("Failed to Scan data into time bucket stats: %v", err)
			continue
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (ts *TimeSeriesDatabase) Close(ctx context.Context) error {
	return ts.db.Close(ctx)
}
