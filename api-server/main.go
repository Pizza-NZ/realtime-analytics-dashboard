package apiserver

import (
	"context"
	"log"
	"net/http"
	"os"
	"pizza-nz/realtime-analytics-dashboard/pkg/storage"

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
		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}
		defer conn.Close()

		for {
			// Messages here
		}
	})
	router.GET("/stats", func(ctx *gin.Context) {
		data, err := db.GetTimeBucket(ctx.Request.Context())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, data)
	})

	router.Run()
}
