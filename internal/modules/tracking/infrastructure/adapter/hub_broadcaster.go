package adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	platformws "github.com/aliwert/go-ride/internal/platform/websocket"
)

type locationUpdate struct {
	Type   string  `json:"type"`
	TripID string  `json:"trip_id"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
}

// adapts the platform Hub to the tracking module's BroadcasterPort.
// serialisation happens here so the use case stays transport-agnostic.
type HubBroadcaster struct {
	hub *platformws.Hub
}

func NewHubBroadcaster(hub *platformws.Hub) *HubBroadcaster {
	return &HubBroadcaster{hub: hub}
}

func (b *HubBroadcaster) BroadcastLocation(_ context.Context, tripID uuid.UUID, lat, lon float64) error {
	payload, err := json.Marshal(locationUpdate{
		Type:   "LOCATION_UPDATE",
		TripID: tripID.String(),
		Lat:    lat,
		Lon:    lon,
	})
	if err != nil {
		return fmt.Errorf("hub broadcaster: marshal: %w", err)
	}

	b.hub.BroadcastToTrip(tripID.String(), payload)
	return nil
}
