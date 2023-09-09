package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan Message)
)

// Message represents a chat message
type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	r := gin.Default()

	r.GET("/ws", func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer ws.Close()

		clients[ws] = true

		for {
			var msg Message
			err := ws.ReadJSON(&msg)
			if err != nil {
				delete(clients, ws)
				break
			}
			broadcast <- msg
		}
	})

	go handleMessages()

	r.Run(":8080")
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
	}
}
