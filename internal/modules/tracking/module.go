package tracking

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	"github.com/aliwert/go-ride/internal/modules/tracking/application/port"
	"github.com/aliwert/go-ride/internal/modules/tracking/infrastructure/adapter"
	trackingws "github.com/aliwert/go-ride/internal/modules/tracking/presentation/ws"
	platformws "github.com/aliwert/go-ride/internal/platform/websocket"
)

// wires the tracking domain stack and returns the BroadcasterPort
// so the location module can push updates without knowing about WebSockets
func InitModule(app *fiber.App, hub *platformws.Hub) port.BroadcasterPort {
	broadcaster := adapter.NewHubBroadcaster(hub)

	// websocket upgrade check must run before the handler to reject non-WS requests
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/trip/:trip_id", websocket.New(trackingws.TrackTrip(hub)))

	return broadcaster
}
