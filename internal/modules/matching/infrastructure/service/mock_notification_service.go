package service

import (
	"context"
	"log"

	"github.com/google/uuid"
)

// logs push notifications to stdout.
// replace with FCM / APNs / Kafka fanout when the real service is ready.
type MockNotificationService struct{}

func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{}
}

func (s *MockNotificationService) NotifyDriver(ctx context.Context, driverID uuid.UUID, tripID uuid.UUID) error {
	log.Printf("INFO: [PUSH NOTIFICATION] sending trip %s to driver %s", tripID, driverID)
	return nil
}
