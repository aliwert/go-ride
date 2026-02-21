package port

import (
	"context"

	"github.com/google/uuid"
)

// abstracts the push notification delivery mechanism.
// swappable between a mock logger, FCM, APNs, or a Kafka-backed fanout.
type NotificationPort interface {
	NotifyDriver(ctx context.Context, driverID uuid.UUID, tripID uuid.UUID) error
}
