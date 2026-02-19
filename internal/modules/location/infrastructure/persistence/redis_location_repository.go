package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// single sorted-set holding all driver geo positions
	geoKey = "drivers:locations"

	// per-driver heartbeat key — when this expires the driver is considered offline.
	// redis GEOADD has no per-member TTL, so we track liveness separately and
	// filter stale entries at read time.
	heartbeatPrefix = "drivers:heartbeat:"
	heartbeatTTL    = 5 * time.Minute
)

type RedisLocationRepository struct {
	client *redis.Client
}

func NewRedisLocationRepository(client *redis.Client) *RedisLocationRepository {
	return &RedisLocationRepository{client: client}
}

func (r *RedisLocationRepository) UpdateLocation(ctx context.Context, driverID uuid.UUID, lat, lon float64) error {
	pipe := r.client.Pipeline()

	pipe.GeoAdd(ctx, geoKey, &redis.GeoLocation{
		Name:      driverID.String(),
		Longitude: lon,
		Latitude:  lat,
	})

	// refresh the heartbeat so this driver is considered active
	pipe.Set(ctx, heartbeatPrefix+driverID.String(), "1", heartbeatTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis: update location: %w", err)
	}

	return nil
}

func (r *RedisLocationRepository) FindNearbyDrivers(ctx context.Context, lat, lon float64, radiusInKm float64) ([]uuid.UUID, error) {
	results, err := r.client.GeoSearch(ctx, geoKey, &redis.GeoSearchQuery{
		Longitude:  lon,
		Latitude:   lat,
		Radius:     radiusInKm,
		RadiusUnit: "km",
		Sort:       "ASC",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("redis: geo search: %w", err)
	}

	// filter out drivers whose heartbeat has expired — they stopped sending pings
	drivers := make([]uuid.UUID, 0, len(results))
	for _, member := range results {
		alive, err := r.client.Exists(ctx, heartbeatPrefix+member).Result()
		if err != nil {
			continue // skip on transient error, don't break the whole list
		}
		if alive == 0 {
			continue
		}

		id, err := uuid.Parse(member)
		if err != nil {
			continue
		}
		drivers = append(drivers, id)
	}

	return drivers, nil
}
