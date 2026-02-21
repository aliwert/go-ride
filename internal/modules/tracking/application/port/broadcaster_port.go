package port

import (
	"context"

	"github.com/google/uuid"
)

// decouples the location module from the WebSocket transport.
// the location use case calls this after persisting a GPS ping so the rider
// sees the driver move in real time without the location module knowing about WS.
type BroadcasterPort interface {
	BroadcastLocation(ctx context.Context, tripID uuid.UUID, lat, lon float64) error
}
