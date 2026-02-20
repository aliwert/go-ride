CREATE TABLE IF NOT EXISTS trips (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rider_id    UUID           NOT NULL REFERENCES users(id),
    driver_id   UUID           REFERENCES users(id),
    pickup_lat  DECIMAL(9, 6)  NOT NULL,
    pickup_lon  DECIMAL(9, 6)  NOT NULL,
    dropoff_lat DECIMAL(9, 6)  NOT NULL,
    dropoff_lon DECIMAL(9, 6)  NOT NULL,
    status      VARCHAR(20)    NOT NULL CHECK (status IN ('REQUESTED', 'ACCEPTED', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED')) DEFAULT 'REQUESTED',
    fare        DECIMAL(10, 2),
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- reuses the set_updated_at() trigger function created in migration 000001
CREATE TRIGGER trigger_trips_updated_at
    BEFORE UPDATE ON trips
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
