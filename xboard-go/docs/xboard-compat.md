# Xboard Node Protocol Compatibility

This document details how Xboard Go achieves wire-level compatibility with Xboard's node protocols.

## Overview

Xboard Go implements the exact same node-facing APIs as Xboard, ensuring that existing nodes can connect without any modifications. All request/response formats, authentication mechanisms, and behavioral semantics are preserved.

## Protocol Implementation

### Authentication

**Xboard Implementation** (`app/Http/Middleware/Server.php`):
- Token-based authentication via query parameter `token`
- Required parameters: `token`, `node_id`, `node_type`
- Validates against `admin_setting('server_token')`

**Xboard Go Implementation** (`internal/middleware/node_auth.go`):
- Identical token validation
- Same required parameters
- Token configured via `node.server_token` in config
- Returns same error messages

### Endpoint Compatibility Matrix

| Endpoint | Xboard Path | Xboard Go Path | Status |
|----------|-------------|----------------|--------|
| Config | `/api/v1/server/UniProxy/config` | ✓ Same | ✅ Compatible |
| Users | `/api/v1/server/UniProxy/user` | ✓ Same | ✅ Compatible |
| Traffic | `/api/v1/server/UniProxy/push` | ✓ Same | ✅ Compatible |
| Online | `/api/v1/server/UniProxy/alive` | ✓ Same | ✅ Compatible |
| Device Limit | `/api/v1/server/UniProxy/alivelist` | ✓ Same | ✅ Compatible |
| Status | `/api/v1/server/UniProxy/status` | ✓ Same | ✅ Compatible |

### Traffic Report Format

**Critical Detail**: Xboard uses **DELTA** (incremental) traffic reporting, not cumulative.

**Xboard Implementation** (`app/Http/Controllers/Server/UniProxyController.php:push()`):
```php
User::where('id', $uid)
    ->incrementEach([
        'u' => $upload * $server_rate,
        'd' => $download * $server_rate,
    ], ['t' => time()]);
```

**Xboard Go Implementation** (`internal/service/accounting_service.go`):
```go
// Increment usage (DELTA format)
usageRepo.IncrementUsage(
    userID, nodeID,
    report.Upload,      // New traffic since last report
    report.Download,    // New traffic since last report
    billableUp,
    billableDown,
)
```

### Request/Response Examples

#### 1. Get Config

**Request** (same in both):
```
GET /api/v1/server/UniProxy/config?token=xxx&node_id=1&node_type=vmess
If-None-Match: "abc123"
```

**Response** (Xboard):
```json
{
  "protocol": "vmess",
  "listen_ip": "0.0.0.0",
  "server_port": 443,
  "base_config": {
    "push_interval": 60,
    "pull_interval": 60
  }
}
```

**Response** (Xboard Go):
```json
{
  "protocol": "vmess",
  "listen_ip": "0.0.0.0",
  "server_port": 443,
  "base_config": {
    "push_interval": 60,
    "pull_interval": 60
  }
}
```
✅ Identical

#### 2. Get Users

**Request** (same in both):
```
GET /api/v1/server/UniProxy/user?token=xxx&node_id=1&node_type=vmess
```

**Response** (Xboard):
```json
{
  "users": [
    {
      "id": 1,
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "speed_limit": 0,
      "device_limit": 0
    }
  ]
}
```

**Response** (Xboard Go):
```json
{
  "users": [
    {
      "id": 1,
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "speed_limit": 0,
      "device_limit": 0
    }
  ]
}
```
✅ Identical

#### 3. Push Traffic

**Request** (Xboard):
```json
POST /api/v1/server/UniProxy/push?token=xxx&node_id=1

[
  [1, [1024000, 2048000]],
  [2, [512000, 1024000]]
]
```

**Request** (Xboard Go):
```json
POST /api/v1/server/UniProxy/push?token=xxx&node_id=1

[
  [1, [1024000, 2048000]],
  [2, [512000, 1024000]]
]
```
✅ Identical format

**Response** (both):
```json
{
  "code": 0,
  "message": "ok",
  "data": true
}
```
✅ Identical

#### 4. Push Online Users

**Request** (both):
```json
POST /api/v1/server/UniProxy/alive?token=xxx&node_id=1

{
  "1": ["192.168.1.1_node1", "192.168.1.2_node1"],
  "2": ["192.168.1.3_node2"]
}
```

**Response** (both):
```json
{
  "data": true
}
```
✅ Identical

#### 5. Get Device Limits

**Request** (both):
```
GET /api/v1/server/UniProxy/alivelist?token=xxx&node_id=1
```

**Response** (both):
```json
{
  "alive": {
    "1": 2,
    "5": 3
  }
}
```
✅ Identical

#### 6. Push Status

**Request** (both):
```json
POST /api/v1/server/UniProxy/status?token=xxx&node_id=1

{
  "cpu": 45.5,
  "mem": {
    "total": 8589934592,
    "used": 4294967296
  },
  "swap": {
    "total": 2147483648,
    "used": 536870912
  },
  "disk": {
    "total": 107374182400,
    "used": 53687091200
  }
}
```

**Response** (both):
```json
{
  "data": true,
  "code": 0,
  "message": "success"
}
```
✅ Identical

## Key Differences from Xboard

While the node protocol is identical, Xboard Go has architectural differences:

### 1. Label System

**Xboard**:
- Labels (groups) are permission-based
- Users belong to groups
- Nodes serve specific groups

**Xboard Go**:
- Labels are organizational tags
- Nodes have multiple labels
- Plans allow nodes with matching labels
- No permission group binding

### 2. Traffic Multipliers

**Xboard**:
- Single `rate` field per server
- Optional time-based rates

**Xboard Go**:
- Node-level multiplier
- Plan base multiplier
- Label-specific multipliers
- Multiplicative stacking

### 3. Usage Tracking

**Xboard**:
- Updates user's `u` and `d` fields directly
- Single cumulative counter

**Xboard Go**:
- Period-based tracking
- Separate real and billable counters
- Per-node usage breakdown
- Historical periods preserved

### 4. User Selection

**Xboard** (for node user list):
```php
$users = User::whereIn('group_id', $server->group_ids)
    ->where('banned', 0)
    ->where(function($query) {
        $query->where('expired_at', '>=', time())
            ->orWhere('expired_at', NULL);
    })
    ->whereRaw('u + d < transfer_enable')
    ->get();
```

**Xboard Go** (equivalent logic):
```go
// Users with plans that have labels matching node's labels
// Not banned
// Not exceeded quota (billable_bytes < quota_bytes)
```

## Testing Compatibility

### Manual Testing

1. **Start Xboard Go**:
```bash
go run ./cmd/server
```

2. **Test Config Endpoint**:
```bash
curl "http://localhost:8080/api/v1/server/UniProxy/config?token=your-token&node_id=1&node_type=vmess"
```

3. **Test User Endpoint**:
```bash
curl "http://localhost:8080/api/v1/server/UniProxy/user?token=your-token&node_id=1&node_type=vmess"
```

4. **Test Traffic Push**:
```bash
curl -X POST "http://localhost:8080/api/v1/server/UniProxy/push?token=your-token&node_id=1" \
  -H "Content-Type: application/json" \
  -d '[[1, [1000000, 2000000]]]'
```

### Integration Testing

Use an actual Xboard-compatible node:

1. Configure node to point to Xboard Go
2. Set correct server token
3. Start node
4. Verify connection in logs
5. Generate traffic
6. Check usage counters

## Migration from Xboard

### Data Migration

Xboard Go uses a different schema. To migrate from Xboard:

1. **Export Users**:
```sql
SELECT id, email, password, group_id, transfer_enable, u, d, expired_at, banned
FROM users;
```

2. **Map to Xboard Go**:
- Create plans based on `group_id`
- Convert `transfer_enable` to plan `quota_bytes`
- Import users with plan assignments
- Initialize usage periods with current `u`, `d` values

3. **Export Servers**:
```sql
SELECT id, name, type, host, port, rate, group_id
FROM servers;
```

4. **Map to Xboard Go**:
- Create nodes with `node_multiplier = rate`
- Create labels for each `group_id`
- Assign labels to nodes

### Cutover Process

1. **Parallel Run**: Run both Xboard and Xboard Go
2. **Migrate Nodes Gradually**: Update node configs one by one
3. **Monitor**: Check logs and metrics
4. **Verify**: Ensure traffic is recorded correctly
5. **Complete**: Once all nodes migrated, decommission Xboard

## Troubleshooting

### Node Can't Connect

**Symptom**: Node logs show authentication errors

**Xboard Error**:
```json
{
  "message": "Invalid token"
}
```

**Xboard Go Error** (identical):
```json
{
  "message": "Invalid token"
}
```

**Solution**: Verify `NODE_SERVER_TOKEN` in Xboard Go config matches node's token

### Users Not Appearing

**Symptom**: `/user` endpoint returns empty list

**Xboard Cause**: User's `group_id` not in server's `group_ids`

**Xboard Go Cause**: User's plan has no labels matching node's labels

**Solution**: Assign appropriate labels to node or adjust plan's label list

### Traffic Not Recording

**Symptom**: Traffic pushed but usage not increasing

**Xboard Go Debug**:
```bash
# Check logs
docker logs xboard-go-app

# Check metrics
curl http://localhost:8080/metrics | grep accounting_errors
```

**Common Causes**:
- User has no plan
- User is banned
- Node ID doesn't exist
- Traffic multiplier calculation error

## Summary

Xboard Go maintains 100% wire-protocol compatibility with Xboard for all node-facing endpoints. Existing nodes can connect without modification by simply:

1. Updating server URL
2. Ensuring correct server token
3. Verifying node ID exists

The internal architecture differs to provide enhanced features (labels, multipliers, historical tracking), but the node communication protocol is identical.
