package websocket

import (
	"log"
	"time"

	"github.com/gofiber/websocket/v2"
)

const (
	// Time allowed to write a message to the peer
	notificationWriteWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	notificationPongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	notificationPingPeriod = (notificationPongWait * 9) / 10

	// Maximum message size allowed from peer
	notificationMaxMessageSize = 512 * 1024 // 512KB
)

// ReadPump pumps messages from the websocket connection to the hub
func (c *NotificationClient) ReadPump() {
	defer func() {
		c.Hub.UnregisterClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(notificationPongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(notificationPongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Notification WebSocket error: %v", err)
			}
			break
		}

		// For notification hub, we mostly just receive ping/pong
		// Most messages are server -> client
		log.Printf("📨 Received notification message from user %s: %s", c.UserID, string(message))

		// You can handle client messages here if needed
		// For now, notifications are mostly one-way (server -> client)
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *NotificationClient) WritePump() {
	ticker := time.NewTicker(notificationPingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(notificationWriteWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Error writing notification message: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(notificationWriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
