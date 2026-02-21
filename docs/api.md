# API Reference

Base URL: `http://localhost:3000`

All REST endpoints are prefixed with `/api/v1`. The WebSocket endpoint is mounted at the root level.

---

## Authentication

Protected endpoints require a valid JWT in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

Tokens are obtained via the `/identity/login` endpoint. Access tokens expire in **15 minutes**; refresh tokens in **7 days**.

---

## Global Error Response Format

Every error response follows a consistent envelope:

```json
{
  "success": false,
  "error": {
    "code": "MACHINE_READABLE_CODE",
    "message": "human-readable description"
  }
}
```

| HTTP Status | When                                                           |
| ----------- | -------------------------------------------------------------- |
| `400`       | Invalid request body, bad parameters                           |
| `401`       | Missing or invalid JWT token                                   |
| `403`       | Valid token but insufficient permissions                       |
| `404`       | Resource not found                                             |
| `409`       | Conflict (e.g., email taken, trip already accepted)            |
| `422`       | Business rule violation (e.g., invalid trip status transition) |
| `426`       | WebSocket upgrade required                                     |
| `500`       | Unexpected internal error (details logged server-side)         |

---

## Identity Module

### Register

```
POST /api/v1/identity/register
```

**Request Body:**

```json
{
  "email": "rider@example.com",
  "password": "securePassword123",
  "first_name": "Jane",
  "last_name": "Doe",
  "role": "RIDER"
}
```

| Field        | Type   | Required | Notes               |
| ------------ | ------ | -------- | ------------------- |
| `email`      | string | yes      | Must be unique      |
| `password`   | string | yes      | Hashed with bcrypt  |
| `first_name` | string | yes      |                     |
| `last_name`  | string | yes      |                     |
| `role`       | string | yes      | `RIDER` or `DRIVER` |

**Response:** `201 Created`

```json
{
  "id": "uuid",
  "email": "rider@example.com",
  "first_name": "Jane",
  "last_name": "Doe",
  "role": "RIDER",
  "status": "ACTIVE",
  "created_at": "2026-02-21T10:00:00Z",
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci..."
}
```

**Error Codes:** `EMAIL_ALREADY_TAKEN` (409), `INVALID_ROLE` (400)

---

### Login

```
POST /api/v1/identity/login
```

**Request Body:**

```json
{
  "email": "rider@example.com",
  "password": "securePassword123"
}
```

**Response:** `200 OK` — same shape as Register response.

**Error Codes:** `INVALID_CREDENTIALS` (401), `ACCOUNT_SUSPENDED` (403)

---

## Location Module

> All location endpoints require authentication.

### Update Driver Location

```
POST /api/v1/location/update
```

**Request Body:**

```json
{
  "latitude": 41.0082,
  "longitude": 28.9784,
  "trip_id": "optional-trip-uuid"
}
```

| Field       | Type          | Required | Notes                                                               |
| ----------- | ------------- | -------- | ------------------------------------------------------------------- |
| `latitude`  | float64       | yes      | -90 to 90                                                           |
| `longitude` | float64       | yes      | -180 to 180                                                         |
| `trip_id`   | string (UUID) | no       | When set, triggers real-time broadcast to riders tracking this trip |

> `driver_id` is extracted from the JWT — not accepted from the request body.

**Response:** `200 OK`

```json
{
  "message": "location updated successfully"
}
```

**Error Codes:** `INVALID_COORDINATES` (400)

---

### Find Nearby Drivers

```
GET /api/v1/location/nearby?latitude=41.0082&longitude=28.9784&radius_km=5
```

| Query Param | Type    | Required | Notes       |
| ----------- | ------- | -------- | ----------- |
| `latitude`  | float64 | yes      | -90 to 90   |
| `longitude` | float64 | yes      | -180 to 180 |
| `radius_km` | float64 | yes      | Must be > 0 |

**Response:** `200 OK`

```json
{
  "drivers": ["uuid-1", "uuid-2", "uuid-3"]
}
```

**Error Codes:** `INVALID_COORDINATES` (400), `INVALID_RADIUS` (400)

---

## Trip Module

> All trip endpoints require authentication.

### Request a Trip

```
POST /api/v1/trip/request
```

**Request Body:**

```json
{
  "pickup_lat": 41.0082,
  "pickup_lon": 28.9784,
  "dropoff_lat": 41.0136,
  "dropoff_lon": 28.955
}
```

> `rider_id` is extracted from the JWT.

**Response:** `201 Created`

```json
{
  "id": "trip-uuid",
  "rider_id": "rider-uuid",
  "driver_id": null,
  "pickup_lat": 41.0082,
  "pickup_lon": 28.9784,
  "dropoff_lat": 41.0136,
  "dropoff_lon": 28.955,
  "status": "REQUESTED",
  "fare": null,
  "created_at": "2026-02-21T10:00:00Z",
  "updated_at": "2026-02-21T10:00:00Z"
}
```

**Error Codes:** `INVALID_COORDINATES` (400), `INVALID_RIDER_ID` (400)

---

### Accept a Trip (Driver)

```
PUT /api/v1/trip/:id/accept
```

> `driver_id` is extracted from the JWT. Uses **optimistic locking** — only succeeds if the trip's current status is `REQUESTED`.

**Response:** `200 OK` — `TripResponse` with `status: "ACCEPTED"` and the driver assigned.

**Error Codes:** `TRIP_NOT_FOUND` (404), `TRIP_ALREADY_ACCEPTED` (409), `INVALID_TRIP_ID` (400)

---

### Complete a Trip

```
PUT /api/v1/trip/:id/complete
```

**Response:** `200 OK` — `TripResponse` with `status: "COMPLETED"`.

**Error Codes:** `TRIP_NOT_FOUND` (404), `INVALID_TRIP_STATUS` (422), `INVALID_TRIP_ID` (400)

---

## Matching Module

> Requires authentication.

### Match Drivers to a Trip

```
POST /api/v1/matching/:trip_id/match
```

**Request Body:**

```json
{
  "lat": 41.0082,
  "lon": 28.9784
}
```

Finds all drivers within a 5 km radius and fans out notifications. Uses the `LocationPort` adapter to query the location module without direct Redis coupling.

**Response:** `200 OK`

```json
{
  "message": "drivers notified successfully"
}
```

**Error Codes:** `NO_DRIVERS_AVAILABLE` (404), `INVALID_TRIP_ID` (400), `INVALID_REQUEST_BODY` (400)

---

## Tracking Module (WebSocket)

### Track a Trip in Real Time

```
GET ws://localhost:3000/ws/trip/:trip_id
```

> This is a WebSocket endpoint — not a REST call. Connect with any WS client.

**Connection flow:**

1. Client opens a WebSocket connection to `/ws/trip/{trip_id}`.
2. Server registers the connection to the internal Hub for that trip.
3. When the driver sends GPS updates with the matching `trip_id`, the server pushes location frames to all connected clients.
4. Connection is push-only (server → client). Incoming messages from the client are discarded.

**Server Push Payload:**

```json
{
  "type": "LOCATION_UPDATE",
  "trip_id": "trip-uuid",
  "lat": 41.0085,
  "lon": 28.979
}
```

**Disconnection:** Close the WebSocket connection normally. The server automatically unregisters the client from the Hub.

> Non-WebSocket requests to `/ws/*` receive `426 Upgrade Required`.
