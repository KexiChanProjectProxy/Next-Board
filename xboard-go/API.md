# Xboard Go API Documentation

Complete API reference for Xboard Go. All timestamps are in RFC3339 format.

## Table of Contents

- [Authentication](#authentication)
- [User Endpoints](#user-endpoints)
- [Admin Endpoints](#admin-endpoints)
- [Node Protocol Endpoints](#node-protocol-endpoints)
- [Error Handling](#error-handling)

---

## Base URL

```
http://localhost:8080
```

For production, use your domain with HTTPS.

---

## Authentication

### Login

Authenticate with email and password to receive JWT tokens.

**Endpoint:** `POST /api/v1/auth/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Validation:**
- `email`: Required, must be valid email format
- `password`: Required, minimum 6 characters

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "b8f3a7c2d1e4f9a8b7c6d5e4f3a2b1c0",
  "token_type": "Bearer"
}
```

**Error Response:** `401 Unauthorized`
```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "invalid email or password"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

---

### Refresh Token

Get a new access token using a refresh token.

**Endpoint:** `POST /api/v1/auth/refresh`

**Request Body:**
```json
{
  "refresh_token": "b8f3a7c2d1e4f9a8b7c6d5e4f3a2b1c0"
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer"
}
```

**Error Response:** `401 Unauthorized`
```json
{
  "error": {
    "code": "INVALID_REFRESH_TOKEN",
    "message": "invalid or expired refresh token"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "b8f3a7c2d1e4f9a8b7c6d5e4f3a2b1c0"
  }'
```

---

## User Endpoints

All user endpoints require authentication. Include the access token in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

### Get Current User

Get the authenticated user's profile.

**Endpoint:** `GET /api/v1/me`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:** `200 OK`
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "role": "user",
    "plan_id": 2,
    "telegram_chat_id": null,
    "telegram_linked_at": null,
    "created_at": "2025-01-15T10:30:00Z"
  }
}
```

**Example:**
```bash
curl http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer <access_token>"
```

---

### Get User Plan

Get the authenticated user's plan details.

**Endpoint:** `GET /api/v1/me/plan`

**Response:** `200 OK`
```json
{
  "plan": {
    "id": 2,
    "name": "Premium Plan",
    "quota_bytes": 107374182400,
    "reset_period": "monthly",
    "base_multiplier": 1.0,
    "created_at": "2025-01-01T00:00:00Z",
    "labels": [
      {
        "id": 1,
        "name": "Premium",
        "description": "Premium tier nodes",
        "multiplier": 2.0
      },
      {
        "id": 2,
        "name": "US",
        "description": "US nodes",
        "multiplier": 1.0
      }
    ]
  }
}
```

**Response (no plan):** `200 OK`
```json
{
  "plan": null
}
```

**Example:**
```bash
curl http://localhost:8080/api/v1/me/plan \
  -H "Authorization: Bearer <access_token>"
```

---

### Get Allowed Nodes

Get nodes accessible to the authenticated user based on their plan.

**Endpoint:** `GET /api/v1/me/nodes`

**Response:** `200 OK`
```json
{
  "nodes": [
    {
      "id": 1,
      "name": "US West 1",
      "node_type": "vmess",
      "host": "us-west-1.example.com",
      "port": 443,
      "node_multiplier": 1.5,
      "status": "active",
      "labels": [
        {
          "id": 1,
          "name": "Premium",
          "multiplier": 2.0
        },
        {
          "id": 2,
          "name": "US",
          "multiplier": 1.0
        }
      ]
    }
  ]
}
```

**Notes:**
- Only returns nodes with at least one label matching the user's plan
- Empty array if user has no plan or no matching nodes

**Example:**
```bash
curl http://localhost:8080/api/v1/me/nodes \
  -H "Authorization: Bearer <access_token>"
```

---

### Get Current Usage

Get the authenticated user's current billing period usage.

**Endpoint:** `GET /api/v1/me/usage`

**Response:** `200 OK`
```json
{
  "usage": {
    "real_bytes_up": 10737418240,
    "real_bytes_down": 21474836480,
    "billable_bytes_up": 32212254720,
    "billable_bytes_down": 64424509440,
    "period_start": "2025-01-01T00:00:00Z",
    "period_end": "2025-02-01T00:00:00Z"
  }
}
```

**Fields:**
- `real_bytes_up`: Actual upload bytes (before multipliers)
- `real_bytes_down`: Actual download bytes (before multipliers)
- `billable_bytes_up`: Billed upload bytes (after multipliers)
- `billable_bytes_down`: Billed download bytes (after multipliers)
- `period_start`: Current billing period start
- `period_end`: Current billing period end

**Example:**
```bash
curl http://localhost:8080/api/v1/me/usage \
  -H "Authorization: Bearer <access_token>"
```

---

### Get Usage History

Get historical usage data from Prometheus.

**Endpoint:** `GET /api/v1/me/usage/history`

**Query Parameters:**
- `range` (optional): Time range (default: 30d)

**Response:** `200 OK`
```json
{
  "message": "Prometheus integration not yet implemented",
  "note": "This endpoint will query Prometheus for historical usage data",
  "params": {
    "start": "2024-12-15T10:30:00Z",
    "end": "2025-01-15T10:30:00Z"
  }
}
```

**Note:** This endpoint is planned for future implementation with Prometheus integration.

**Example:**
```bash
curl http://localhost:8080/api/v1/me/usage/history?range=30d \
  -H "Authorization: Bearer <access_token>"
```

---

### Generate Telegram Link Token

Generate a one-time token to link Telegram account.

**Endpoint:** `POST /api/v1/me/telegram/link`

**Response:** `200 OK`
```json
{
  "link_token": "a1b2c3d4e5f6g7h8i9j0",
  "expires_in": 300,
  "instructions": "Send this token to the bot using /link <token>"
}
```

**Usage:**
1. Call this endpoint to get a token
2. Send `/link <token>` to the Telegram bot
3. Account will be linked

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/me/telegram/link \
  -H "Authorization: Bearer <access_token>"
```

---

## Admin Endpoints

All admin endpoints require authentication with an admin role account.

**Authentication:**
```
Authorization: Bearer <admin_access_token>
```

### User Management

#### Create User

Create a new user account.

**Endpoint:** `POST /api/v1/admin/users`

**Request Body:**
```json
{
  "email": "newuser@example.com",
  "password": "password123",
  "role": "user",
  "plan_id": 1
}
```

**Validation:**
- `email`: Required, valid email format
- `password`: Required, minimum 6 characters
- `role`: Required, must be "admin" or "user"
- `plan_id`: Optional, integer

**Response:** `201 Created`
```json
{
  "user": {
    "id": 5,
    "email": "newuser@example.com",
    "role": "user",
    "plan_id": 1,
    "banned": false,
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

**Error Response:** `400 Bad Request`
```json
{
  "error": {
    "code": "USER_CREATION_FAILED",
    "message": "email already exists"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "password123",
    "role": "user",
    "plan_id": 1
  }'
```

---

#### List Users

Get paginated list of all users.

**Endpoint:** `GET /api/v1/admin/users`

**Query Parameters:**
- `page` (optional): Page number, default 1
- `limit` (optional): Items per page, default 20

**Response:** `200 OK`
```json
{
  "users": [
    {
      "id": 1,
      "email": "user1@example.com",
      "role": "user",
      "plan_id": 1,
      "banned": false,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "email": "admin@example.com",
      "role": "admin",
      "plan_id": null,
      "banned": false,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 42,
    "page": 1,
    "limit": 20,
    "pages": 3
  }
}
```

**Example:**
```bash
curl "http://localhost:8080/api/v1/admin/users?page=1&limit=20" \
  -H "Authorization: Bearer <admin_token>"
```

---

#### Get User

Get a specific user by ID.

**Endpoint:** `GET /api/v1/admin/users/:id`

**Path Parameters:**
- `id`: User ID (integer)

**Response:** `200 OK`
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "role": "user",
    "plan_id": 1,
    "banned": false,
    "telegram_chat_id": 123456789,
    "telegram_linked_at": "2025-01-10T15:20:00Z",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

**Error Response:** `404 Not Found`
```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User not found"
  }
}
```

**Example:**
```bash
curl http://localhost:8080/api/v1/admin/users/1 \
  -H "Authorization: Bearer <admin_token>"
```

---

#### Update User

Update user information.

**Endpoint:** `PUT /api/v1/admin/users/:id`

**Path Parameters:**
- `id`: User ID (integer)

**Request Body:** (all fields optional)
```json
{
  "email": "newemail@example.com",
  "plan_id": 2,
  "banned": true
}
```

**Response:** `200 OK`
```json
{
  "user": {
    "id": 1,
    "email": "newemail@example.com",
    "role": "user",
    "plan_id": 2,
    "banned": true,
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-15T10:35:00Z"
  }
}
```

**Example:**
```bash
curl -X PUT http://localhost:8080/api/v1/admin/users/1 \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "banned": true
  }'
```

---

#### Delete User

Delete a user account.

**Endpoint:** `DELETE /api/v1/admin/users/:id`

**Path Parameters:**
- `id`: User ID (integer)

**Response:** `200 OK`
```json
{
  "message": "User deleted successfully"
}
```

**Error Response:** `500 Internal Server Error`
```json
{
  "error": {
    "code": "DELETE_FAILED",
    "message": "failed to delete user"
  }
}
```

**Example:**
```bash
curl -X DELETE http://localhost:8080/api/v1/admin/users/5 \
  -H "Authorization: Bearer <admin_token>"
```

---

### Node Management

#### Create Node

Create a new proxy node.

**Endpoint:** `POST /api/v1/admin/nodes`

**Request Body:**
```json
{
  "name": "US West 1",
  "node_type": "vmess",
  "host": "us-west-1.example.com",
  "port": 443,
  "protocol_config": "{\"network\":\"tcp\",\"tls\":1}",
  "node_multiplier": 1.5,
  "label_ids": [1, 2]
}
```

**Field Details:**
- `name`: Required, node display name
- `node_type`: Required, protocol type (vmess, vless, trojan, shadowsocks, etc.)
- `host`: Required, node hostname or IP
- `port`: Required, node port
- `protocol_config`: Optional, JSON string with protocol-specific config
- `node_multiplier`: Optional, traffic multiplier (default: 1.0)
- `label_ids`: Optional, array of label IDs to assign

**Response:** `201 Created`
```json
{
  "node": {
    "id": 10,
    "name": "US West 1",
    "node_type": "vmess",
    "host": "us-west-1.example.com",
    "port": 443,
    "protocol_config": "{\"network\":\"tcp\",\"tls\":1}",
    "node_multiplier": 1.5,
    "status": "active",
    "created_at": "2025-01-15T10:40:00Z",
    "updated_at": "2025-01-15T10:40:00Z"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/admin/nodes \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "US West 1",
    "node_type": "vmess",
    "host": "us-west-1.example.com",
    "port": 443,
    "node_multiplier": 1.5,
    "label_ids": [1, 2]
  }'
```

---

#### List Nodes

Get paginated list of all nodes.

**Endpoint:** `GET /api/v1/admin/nodes`

**Query Parameters:**
- `page` (optional): Page number, default 1
- `limit` (optional): Items per page, default 20

**Response:** `200 OK`
```json
{
  "nodes": [
    {
      "id": 1,
      "name": "US West 1",
      "node_type": "vmess",
      "host": "us-west-1.example.com",
      "port": 443,
      "node_multiplier": 1.5,
      "status": "active",
      "last_seen_at": "2025-01-15T10:35:00Z",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-15T10:35:00Z"
    }
  ],
  "pagination": {
    "total": 15,
    "page": 1,
    "limit": 20,
    "pages": 1
  }
}
```

**Example:**
```bash
curl "http://localhost:8080/api/v1/admin/nodes?page=1&limit=20" \
  -H "Authorization: Bearer <admin_token>"
```

---

### Plan Management

#### Create Plan

Create a new subscription plan.

**Endpoint:** `POST /api/v1/admin/plans`

**Request Body:**
```json
{
  "name": "Premium Plan",
  "quota_bytes": 107374182400,
  "reset_period": "monthly",
  "base_multiplier": 1.0,
  "label_ids": [1, 2, 3]
}
```

**Field Details:**
- `name`: Required, plan name
- `quota_bytes`: Required, total quota in bytes
- `reset_period`: Required, one of: "none", "daily", "weekly", "monthly", "yearly"
- `base_multiplier`: Optional, base traffic multiplier (default: 1.0)
- `label_ids`: Optional, array of label IDs for node access

**Response:** `201 Created`
```json
{
  "plan": {
    "id": 5,
    "name": "Premium Plan",
    "quota_bytes": 107374182400,
    "reset_period": "monthly",
    "base_multiplier": 1.0,
    "created_at": "2025-01-15T10:45:00Z",
    "updated_at": "2025-01-15T10:45:00Z"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/admin/plans \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Premium Plan",
    "quota_bytes": 107374182400,
    "reset_period": "monthly",
    "base_multiplier": 1.0,
    "label_ids": [1, 2, 3]
  }'
```

---

#### List Plans

Get paginated list of all plans.

**Endpoint:** `GET /api/v1/admin/plans`

**Query Parameters:**
- `page` (optional): Page number, default 1
- `limit` (optional): Items per page, default 20

**Response:** `200 OK`
```json
{
  "plans": [
    {
      "id": 1,
      "name": "Free Plan",
      "quota_bytes": 1073741824,
      "reset_period": "monthly",
      "base_multiplier": 1.0,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "name": "Premium Plan",
      "quota_bytes": 107374182400,
      "reset_period": "monthly",
      "base_multiplier": 1.0,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 5,
    "page": 1,
    "limit": 20,
    "pages": 1
  }
}
```

**Example:**
```bash
curl "http://localhost:8080/api/v1/admin/plans?page=1&limit=20" \
  -H "Authorization: Bearer <admin_token>"
```

---

### Label Management

#### Create Label

Create a new label for organizing nodes and plans.

**Endpoint:** `POST /api/v1/admin/labels`

**Request Body:**
```json
{
  "name": "Premium",
  "description": "Premium tier nodes"
}
```

**Response:** `201 Created`
```json
{
  "label": {
    "id": 8,
    "name": "Premium",
    "description": "Premium tier nodes",
    "multiplier": 1.0,
    "created_at": "2025-01-15T10:50:00Z",
    "updated_at": "2025-01-15T10:50:00Z"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/admin/labels \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Premium",
    "description": "Premium tier nodes"
  }'
```

---

#### List Labels

Get paginated list of all labels.

**Endpoint:** `GET /api/v1/admin/labels`

**Query Parameters:**
- `page` (optional): Page number, default 1
- `limit` (optional): Items per page, default 20

**Response:** `200 OK`
```json
{
  "labels": [
    {
      "id": 1,
      "name": "Premium",
      "description": "Premium tier nodes",
      "multiplier": 2.0,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "name": "US",
      "description": "United States nodes",
      "multiplier": 1.0,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 10,
    "page": 1,
    "limit": 20,
    "pages": 1
  }
}
```

**Example:**
```bash
curl "http://localhost:8080/api/v1/admin/labels?page=1&limit=20" \
  -H "Authorization: Bearer <admin_token>"
```

---

## Node Protocol Endpoints

These endpoints are used by Xboard-compatible proxy nodes to communicate with the server. They implement the UniProxy protocol.

### Authentication

Node endpoints use query parameter authentication:

```
?token=<server_token>&node_id=<node_id>&node_type=<protocol>
```

**Parameters:**
- `token`: Server token from config (`node.server_token`)
- `node_id`: Node database ID
- `node_type`: Protocol type (vmess, vless, trojan, shadowsocks, etc.)

### Endpoints

Both V1 and V2 APIs are supported with identical behavior:

**V1 API:**
- `/api/v1/server/UniProxy/config`
- `/api/v1/server/UniProxy/user`
- `/api/v1/server/UniProxy/push`
- `/api/v1/server/UniProxy/alive`
- `/api/v1/server/UniProxy/alivelist`
- `/api/v1/server/UniProxy/status`

**V2 API:**
- `/api/v2/server/config`
- `/api/v2/server/user`
- `/api/v2/server/push`
- `/api/v2/server/alive`
- `/api/v2/server/alivelist`
- `/api/v2/server/status`

---

### Get Node Configuration

Get node configuration for initialization.

**Endpoint:** `GET /api/v1/server/UniProxy/config`

**Query Parameters:**
- `token`: Server token
- `node_id`: Node ID
- `node_type`: Protocol type

**Response:** `200 OK`
```json
{
  "protocol": "vmess",
  "listen_ip": "0.0.0.0",
  "server_port": 443,
  "base_config": {
    "push_interval": 60,
    "pull_interval": 60
  },
  "network": "tcp",
  "tls": 1
}
```

**ETag Support:**
- Response includes `ETag` header
- Send `If-None-Match` header with previous ETag
- Returns `304 Not Modified` if unchanged

**Example:**
```bash
curl "http://localhost:8080/api/v1/server/UniProxy/config?token=your-token&node_id=1&node_type=vmess"
```

**With ETag:**
```bash
curl "http://localhost:8080/api/v1/server/UniProxy/config?token=your-token&node_id=1&node_type=vmess" \
  -H "If-None-Match: \"abc123def456\""
```

---

### Get User List

Get list of users allowed on this node.

**Endpoint:** `GET /api/v1/server/UniProxy/user`

**Query Parameters:**
- `token`: Server token
- `node_id`: Node ID
- `node_type`: Protocol type

**Response:** `200 OK`
```json
{
  "users": [
    {
      "id": 1,
      "uuid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "speed_limit": 0,
      "device_limit": 0
    },
    {
      "id": 2,
      "uuid": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
      "speed_limit": 0,
      "device_limit": 0
    }
  ]
}
```

**Field Details:**
- `id`: User database ID
- `uuid`: User's protocol UUID
- `speed_limit`: Speed limit in bytes/sec (0 = unlimited)
- `device_limit`: Max concurrent devices (0 = unlimited)

**Filtering:**
- Only returns users with active plans
- Only users with at least one matching label
- Excludes banned users
- Excludes users who exceeded quota

**ETag Support:**
- Response includes `ETag` header
- Send `If-None-Match` header to check for changes
- Returns `304 Not Modified` if user list unchanged

**Example:**
```bash
curl "http://localhost:8080/api/v1/server/UniProxy/user?token=your-token&node_id=1&node_type=vmess"
```

---

### Push Traffic Data

Report user traffic to server (DELTA format).

**Endpoint:** `POST /api/v1/server/UniProxy/push`

**Query Parameters:**
- `token`: Server token
- `node_id`: Node ID
- `node_type`: Protocol type

**Request Body (Array Format):**
```json
[
  [1, [10485760, 20971520]],
  [2, [5242880, 15728640]]
]
```

**Format:** `[[user_id, [upload_bytes, download_bytes]], ...]`

**Request Body (Object Format):**
```json
{
  "1": [10485760, 20971520],
  "2": [5242880, 15728640]
}
```

**Format:** `{"user_id": [upload_bytes, download_bytes], ...}`

**Important:** Traffic values are **DELTA** (incremental) not absolute.

**Response:** `200 OK`
```json
{
  "code": 0,
  "message": "ok",
  "data": true
}
```

**Traffic Processing:**
1. Raw traffic recorded
2. Multipliers applied: `billable = raw × node_multiplier × plan_base_multiplier × Π(label_multipliers)`
3. Billable traffic added to user's current usage

**Example:**
```bash
curl -X POST "http://localhost:8080/api/v1/server/UniProxy/push?token=your-token&node_id=1&node_type=vmess" \
  -H "Content-Type: application/json" \
  -d '[
    [1, [10485760, 20971520]],
    [2, [5242880, 15728640]]
  ]'
```

---

### Push Online Users

Report currently connected users and their IPs.

**Endpoint:** `POST /api/v1/server/UniProxy/alive`

**Query Parameters:**
- `token`: Server token
- `node_id`: Node ID
- `node_type`: Protocol type

**Request Body:**
```json
{
  "1": ["192.168.1.100_node1", "192.168.1.101_node1"],
  "2": ["10.0.0.50_node1"]
}
```

**Format:** `{"user_id": ["ip_identifier", ...], ...}`

**Response:** `200 OK`
```json
{
  "data": true
}
```

**Example:**
```bash
curl -X POST "http://localhost:8080/api/v1/server/UniProxy/alive?token=your-token&node_id=1&node_type=vmess" \
  -H "Content-Type: application/json" \
  -d '{
    "1": ["192.168.1.100_node1"],
    "2": ["10.0.0.50_node1"]
  }'
```

---

### Get Device Limit List

Get online device counts for enforcing device limits.

**Endpoint:** `GET /api/v1/server/UniProxy/alivelist`

**Query Parameters:**
- `token`: Server token
- `node_id`: Node ID
- `node_type`: Protocol type

**Response:** `200 OK`
```json
{
  "alive": {
    "1": 2,
    "2": 1,
    "5": 3
  }
}
```

**Format:** `{"user_id": device_count, ...}`

**Usage:**
- Node calls this endpoint to get current device counts
- Node enforces device limits by checking against user's `device_limit`
- Used to prevent account sharing

**Example:**
```bash
curl "http://localhost:8080/api/v1/server/UniProxy/alivelist?token=your-token&node_id=1&node_type=vmess"
```

---

### Push Node Status

Report node system status (CPU, memory, disk).

**Endpoint:** `POST /api/v1/server/UniProxy/status`

**Query Parameters:**
- `token`: Server token
- `node_id`: Node ID
- `node_type`: Protocol type

**Request Body:**
```json
{
  "cpu": 45.5,
  "mem": {
    "total": 16777216000,
    "used": 8388608000
  },
  "swap": {
    "total": 4294967296,
    "used": 1073741824
  },
  "disk": {
    "total": 107374182400,
    "used": 53687091200
  }
}
```

**Field Details:**
- `cpu`: CPU usage percentage (0-100)
- `mem.total`: Total memory in bytes
- `mem.used`: Used memory in bytes
- `swap.total`: Total swap in bytes
- `swap.used`: Used swap in bytes
- `disk.total`: Total disk in bytes
- `disk.used`: Used disk in bytes

**Response:** `200 OK`
```json
{
  "data": true,
  "code": 0,
  "message": "success"
}
```

**Side Effects:**
- Updates node's `last_seen_at` timestamp
- Status data can be cached for monitoring dashboards

**Example:**
```bash
curl -X POST "http://localhost:8080/api/v1/server/UniProxy/status?token=your-token&node_id=1&node_type=vmess" \
  -H "Content-Type: application/json" \
  -d '{
    "cpu": 45.5,
    "mem": {
      "total": 16777216000,
      "used": 8388608000
    },
    "swap": {
      "total": 4294967296,
      "used": 1073741824
    },
    "disk": {
      "total": 107374182400,
      "used": 53687091200
    }
  }'
```

---

## Error Handling

### Error Response Format

All errors follow this structure:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message"
  }
}
```

### Common Error Codes

#### Authentication Errors (401 Unauthorized)

- `INVALID_CREDENTIALS`: Wrong email or password
- `INVALID_REFRESH_TOKEN`: Refresh token is invalid or expired
- `UNAUTHORIZED`: Missing or invalid access token

#### Validation Errors (400 Bad Request)

- `INVALID_REQUEST`: Request body validation failed
- `INVALID_ID`: Invalid ID format in URL parameter

#### Not Found Errors (404 Not Found)

- `USER_NOT_FOUND`: User does not exist
- `PLAN_NOT_FOUND`: Plan does not exist
- `NODE_NOT_FOUND`: Node does not exist

#### Forbidden Errors (403 Forbidden)

- `FORBIDDEN`: User lacks required permissions (not admin)

#### Server Errors (500 Internal Server Error)

- `INTERNAL_ERROR`: Generic server error
- `UPDATE_FAILED`: Database update failed
- `DELETE_FAILED`: Database delete failed
- `USER_CREATION_FAILED`: User creation failed
- `NODE_CREATION_FAILED`: Node creation failed
- `PLAN_CREATION_FAILED`: Plan creation failed
- `LABEL_CREATION_FAILED`: Label creation failed

### HTTP Status Codes

| Code | Meaning | When Used |
|------|---------|-----------|
| 200 | OK | Successful GET, PUT, DELETE request |
| 201 | Created | Successful POST creating new resource |
| 304 | Not Modified | ETag match, resource unchanged |
| 400 | Bad Request | Invalid request body or parameters |
| 401 | Unauthorized | Authentication failed or missing |
| 403 | Forbidden | Authenticated but lacks permissions |
| 404 | Not Found | Requested resource doesn't exist |
| 422 | Unprocessable Entity | Node protocol validation errors |
| 500 | Internal Server Error | Server-side error |

---

## Rate Limiting

Currently not implemented. Consider implementing rate limiting for production deployments:

- Login endpoint: 5 requests per minute per IP
- User endpoints: 60 requests per minute per user
- Admin endpoints: 120 requests per minute per admin
- Node endpoints: No limit (trusted)

---

## Pagination

List endpoints support pagination:

**Query Parameters:**
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)

**Response Format:**
```json
{
  "data": [...],
  "pagination": {
    "total": 150,
    "page": 2,
    "limit": 20,
    "pages": 8
  }
}
```

---

## Traffic Multiplier Calculation

Traffic multipliers are applied in layers:

```
billable_bytes = real_bytes × node_multiplier × plan_base_multiplier × Π(label_multipliers)
```

**Example:**

Given:
- Node: `node_multiplier = 1.5`, labels: [Premium, US]
- Plan: `base_multiplier = 1.0`, labels: [Premium (2.0×), US (1.0×)]
- Real traffic: 1 GB upload, 2 GB download

Calculation:
```
billable_upload = 1GB × 1.5 × 1.0 × 2.0 × 1.0 = 3 GB
billable_download = 2GB × 1.5 × 1.0 × 2.0 × 1.0 = 6 GB
```

**Label Matching:**
- Only labels present on **both** node and plan are multiplied
- If node has `[Premium, US]` but plan only has `[Premium]`, only Premium multiplier applies

---

## Timestamps

All timestamps use RFC3339 format with UTC timezone:

```
2025-01-15T10:30:00Z
```

Convert to local time in your client application as needed.

---

## CORS

Server accepts requests from any origin with these allowed headers:

- `Origin`
- `Content-Type`
- `Authorization`
- `X-Forwarded-For`
- `X-Real-IP`

All methods are allowed: GET, POST, PUT, DELETE, OPTIONS

For production, restrict `AllowOrigins` to your specific domain.

---

## Monitoring

### Health Check

**Endpoint:** `GET /health`

**Response:** `200 OK`
```json
{
  "status": "ok"
}
```

### Prometheus Metrics

**Endpoint:** `GET /metrics`

Returns Prometheus metrics in text format:

```
# HELP http_request_duration_seconds HTTP request duration
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",path="/api/v1/me",le="0.1"} 150

# HELP active_nodes Number of active nodes
# TYPE active_nodes gauge
active_nodes 5

# HELP traffic_reports_total Total traffic reports received
# TYPE traffic_reports_total counter
traffic_reports_total{node_id="1"} 1523
```

See [Prometheus Metrics](#prometheus-metrics) section in README for more details.

---

## Best Practices

### Authentication

1. **Store tokens securely**
   - Never expose tokens in URLs
   - Use HTTPS in production
   - Store refresh token securely (httpOnly cookie or secure storage)

2. **Token refresh**
   - Refresh access tokens before expiry
   - Handle 401 errors by refreshing and retrying

3. **Admin operations**
   - Audit all admin actions
   - Use separate admin accounts
   - Implement 2FA for admin accounts (future)

### Traffic Reporting

1. **Delta format**
   - Always send incremental traffic, not absolute
   - Track last reported values on node side

2. **Batching**
   - Batch traffic reports (every 60 seconds)
   - Don't send empty reports

3. **Error handling**
   - Retry failed reports with exponential backoff
   - Cache reports locally if server is unreachable

### Performance

1. **ETag usage**
   - Always send `If-None-Match` for config and user endpoints
   - Cache responses and only update on 200 OK

2. **Pagination**
   - Use appropriate page sizes
   - Don't fetch all records at once

3. **Connection pooling**
   - Reuse HTTP connections
   - Implement connection pooling in node clients

---

## Support

For issues, feature requests, or questions:

- GitHub Issues: https://github.com/KexiChanProjectProxy/Next-Board/issues
- Documentation: https://github.com/KexiChanProjectProxy/Next-Board

---

**Last Updated:** 2025-01-15
**API Version:** 1.0
**Xboard Compatibility:** UniProxy V1/V2
