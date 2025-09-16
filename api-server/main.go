package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"pizza-nz/realtime-analytics-dashboard/pkg/storage"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ConnectionManager is a simple manager to keep track of all active clients.
type ConnectionManager struct {
	connections map[*websocket.Conn]bool // The set of active connections.
	mutex       sync.Mutex               // The mutex to protect access to the map.
}

var manager = &ConnectionManager{
	connections: map[*websocket.Conn]bool{},
	mutex:       sync.Mutex{},
}

func main() {
	router := gin.Default()
	conn, err := storage.ConnectToTimeSeriesDB(context.Background())
	if err != nil {
		// There is a issue here
		os.Exit(1)
	}
	db := storage.NewTimeSeriesDatabase(conn)

	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	router.GET("/ws", func(ctx *gin.Context) {
		handleWebSocketConnection(ctx.Writer, ctx.Request, manager)
	})
	router.GET("/stats", func(ctx *gin.Context) {
		data, err := db.GetTimeBucket(ctx.Request.Context())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, data)
	})

	go manager.RunTicker(context.Background(), db)

	router.Run()
}

func handleWebSocketConnection(w http.ResponseWriter, r *http.Request, manager *ConnectionManager) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	manager.HandleNewClient(conn)
}

func (manager *ConnectionManager) HandleNewClient(conn *websocket.Conn) {
	manager.AddConnection(conn)
	defer func() {
		manager.RemoveConnection(conn)
		conn.Close()
	}()

	// This loop reads messages from the client. When ReadMessage returns an error,
	// it means the client disconnected, and the loop breaks.
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected close error: %v", err)
			}
			break
		}
		log.Printf("Received message type %d: %s", messageType, message)
	}
}

func (manager *ConnectionManager) AddConnection(conn *websocket.Conn) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	manager.connections[conn] = true
}

func (manager *ConnectionManager) RemoveConnection(conn *websocket.Conn) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	delete(manager.connections, conn)
}

func (manager *ConnectionManager) RunTicker(ctx context.Context, db storage.DatabaseRepository) (*time.Ticker, chan bool) {
	ticker := time.NewTicker(time.Second)

	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
				data, err := db.GetTimeBucket(ctx)
				if err != nil {
					// Handle
				}
				jsonMessage, _ := json.Marshal(data)
				manager.Broadcast(jsonMessage)
			}
		}
	}()

	return ticker, done
}

func (manager *ConnectionManager) Broadcast(message []byte) {
	// local list of connections
	list := make([]*websocket.Conn, 0, len(manager.connections))
	manager.mutex.Lock()
	for conn := range manager.connections {
		list = append(list, conn)
	}
	manager.mutex.Unlock()

	// iterate over copy
	for _, conn := range list {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			// Handle, maybe remove connection
			manager.RemoveConnection(conn)
		}
	}
}
