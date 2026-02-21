package ws

import (
	"log"

	"github.com/gofiber/contrib/websocket"

	platformws "github.com/aliwert/go-ride/internal/platform/websocket"
)

// upgrades the HTTP connection to WebSocket, registers the rider
// to the hub for the given trip, and holds the connection open.
// only push from server → client; incoming messages are discarded.
func TrackTrip(hub *platformws.Hub) func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		tripID := c.Params("trip_id")
		if tripID == "" {
			log.Println("WARN: ws connection rejected — missing trip_id")
			c.Close()
			return
		}

		hub.Register(tripID, c)
		defer func() {
			hub.Unregister(tripID, c)
			c.Close()
		}()

		log.Printf("INFO: ws client connected for trip %s", tripID)

		// keep-alive read loop, discard incoming frames but need the loop
		// so the connection stays open and we detect client disconnects
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				break
			}
		}

		log.Printf("INFO: ws client disconnected from trip %s", tripID)
	}
}
