# Real-Time Analytics Dashboard

This project is a real-time analytics dashboard that demonstrates a simple yet powerful architecture for collecting, processing, and visualizing data in real time.

## Architecture

The project is composed of the following services:

*   **`producer`**: A Go service that generates mock analytics events and publishes them to a NATS topic.
*   **`consumer`**: A Go service that subscribes to the NATS topic, receives the analytics events, and stores them in a TimescaleDB database.
*   **`api-server`**: A Go service that provides a WebSocket endpoint for real-time data streaming and a REST API for querying historical analytics data.
*   **NATS**: A lightweight, high-performance messaging system that acts as the backbone for our data pipeline.
*   **TimescaleDB**: A time-series database built on top of PostgreSQL, used for storing and querying our analytics data.
*   **Grafana**: An open-source platform for monitoring and observability, which can be used to visualize the analytics data.

## Getting Started

To get the project up and running, you will need to have Docker and Docker Compose installed.

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/your-username/realtime-analytics-dashboard.git
    cd realtime-analytics-dashboard
    ```

2.  **Create a `.env` file:**

    Create a `.env` file in the root of the project and add the following environment variables:

    ```
    DB_USER=user
    DB_PASS=password
    DB_NAME=analytics
    GRAF_USER=admin
    GRAF_PASS=admin
    ```

3.  **Run the services:**

    ```bash
    docker-compose up -d
    ```

    This will start all the services in the background.

4.  **Run the Go applications:**

    In separate terminal windows, run the following commands to start the Go applications:

    ```bash
    go run producer/main.go
    go run consumer/main.go
    go run api-server/main.go
    ```

## Services

### `producer`

The `producer` service generates mock analytics events and publishes them to the `analytics` NATS topic. Each event has the following structure:

```json
{
  "event_type": "page_view",
  "user_id": 123,
  "created_at": "2025-10-10T10:00:00Z",
  "event_data": {
    "path": "/home"
  }
}
```

### `consumer`

The `consumer` service subscribes to the `analytics` NATS topic, receives the analytics events, and stores them in the `analytics_events` table in the TimescaleDB database.

### `api-server`

The `api-server` service provides the following endpoints:

*   **`GET /ping`**: A simple health check endpoint that returns `pong`.
*   **`GET /ws`**: A WebSocket endpoint that streams analytics data in real time to connected clients.
*   **`GET /stats`**: A REST endpoint that returns aggregated analytics data from the database.

## Database Schema

The `analytics_events` table has the following schema:

| Column     | Type        | Description                               |
| ---------- | ----------- | ----------------------------------------- |
| event_type | `varchar`   | The type of the event (e.g., `page_view`). |
| user_id    | `bigint`    | The ID of the user who triggered the event. |
| created_at | `timestamptz` | The timestamp of when the event occurred. |
| event_data | `jsonb`     | Additional data associated with the event. |
| event_id   | `UUID`      | A unique identifier for the event.        |

The table is a hypertable partitioned by the `created_at` column, which is an optimization for time-series data in TimescaleDB.
