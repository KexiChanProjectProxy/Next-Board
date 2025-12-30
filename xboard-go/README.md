# Xboard Go

A high-performance Go-based proxy management system that is fully wire-compatible with Xboard's node protocols. Built with Gin, GORM, and designed for scalability.

## Features

- **Xboard Node Protocol Compatible**: Supports all Xboard node protocols (UniProxy V1/V2, legacy formats)
- **Multi-Label Node Management**: Flexible node organization with label-based access control
- **Traffic Multipliers**: Node-level and plan-level multipliers for fine-grained traffic accounting
- **Prometheus Integration**: Built-in metrics exposition and historical data queries
- **Telegram Bot**: Real-time usage notifications and threshold alerts
- **Plan-Based User Management**: Flexible quota management with auto-reset periods
- **RESTful API**: Clean JSON API for user and admin operations
- **Web UI**: Minimal but functional web interface for dashboards

## Architecture

```
xboard-go/
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection
│   ├── handler/         # HTTP handlers (controllers)
│   ├── jobs/            # Background jobs
│   ├── metrics/         # Prometheus metrics
│   ├── middleware/      # Gin middleware (auth, node auth)
│   ├── models/          # Data models
│   ├── repository/      # Data access layer
│   ├── service/         # Business logic
│   └── telegram/        # Telegram bot
├── migrations/          # Database migrations
├── web/
│   └── templates/       # HTML templates
├── config.json          # Configuration file
└── docker-compose.yml   # Docker Compose setup
```

## Quick Start

### Prerequisites

- Go 1.24+
- MariaDB 11.2+
- (Optional) Docker & Docker Compose

### Local Development

1. **Clone and setup**

```bash
cd xboard-go
cp .env.example .env
# Edit .env with your configuration
```

2. **Start dependencies**

```bash
docker-compose up -d mariadb prometheus
```

3. **Run migrations**

```bash
# Install golang-migrate if not already installed
# brew install golang-migrate (macOS)
# Or download from https://github.com/golang-migrate/migrate

make migrate-up
```

4. **Run the application**

```bash
make run
# Or
go run ./cmd/server
```

The application will start on `http://localhost:8080`

### Docker Deployment

```bash
docker-compose up -d
```

This starts:
- MariaDB (port 3306)
- Prometheus (port 9090)
- Grafana (port 3000)
- Xboard Go (port 8080)

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | 8080 |
| `SERVER_MODE` | Gin mode (debug/release) | debug |
| `DB_HOST` | Database host | localhost |
| `DB_PORT` | Database port | 3306 |
| `DB_USER` | Database user | xboard |
| `DB_PASSWORD` | Database password | xboard_password |
| `DB_NAME` | Database name | xboard_go |
| `JWT_SECRET` | JWT signing secret | (required) |
| `NODE_SERVER_TOKEN` | Node authentication token | (required) |
| `PROMETHEUS_URL` | Prometheus server URL | http://localhost:9090 |
| `TELEGRAM_TOKEN` | Telegram bot token | (optional) |

### Configuration File

Alternatively, edit `config.json`:

```json
{
  "server": {
    "port": "8080",
    "mode": "debug"
  },
  "database": {
    "host": "localhost",
    "port": "3306",
    "user": "xboard",
    "password": "xboard_password",
    "dbname": "xboard_go"
  },
  "auth": {
    "jwt_secret": "your-secret-here",
    "access_token_duration": "15m",
    "refresh_token_duration": "168h"
  },
  "node": {
    "server_token": "your-node-token-here",
    "pull_interval": 60,
    "push_interval": 60
  }
}
```

## Xboard Node Compatibility

### Supported Endpoints

Xboard Go implements the following Xboard-compatible endpoints:

#### V1 API (UniProxy)
- `GET /api/v1/server/UniProxy/config` - Node configuration
- `GET /api/v1/server/UniProxy/user` - User list
- `POST /api/v1/server/UniProxy/push` - Traffic report
- `POST /api/v1/server/UniProxy/alive` - Online users
- `GET /api/v1/server/UniProxy/alivelist` - Device limits
- `POST /api/v1/server/UniProxy/status` - Node status

#### V2 API
- `GET /api/v2/server/config`
- `GET /api/v2/server/user`
- `POST /api/v2/server/push`
- `POST /api/v2/server/alive`
- `GET /api/v2/server/alivelist`
- `POST /api/v2/server/status`

### Authentication

Nodes authenticate using query parameters:
- `token` - Server token (must match `node.server_token` in config)
- `node_id` - Node ID
- `node_type` - Protocol type (vmess, vless, trojan, shadowsocks, etc.)

### Traffic Reporting

Traffic reports use **DELTA format** (incremental):

```json
POST /api/v1/server/UniProxy/push?token=xxx&node_id=1&node_type=vmess

[
  [1, [1000000, 2000000]],
  [2, [500000, 1500000]]
]
```

Format: `[[user_id, [upload_bytes, download_bytes]], ...]`

### ETag Support

Config and user endpoints support ETags for efficient caching:

```bash
GET /api/v1/server/UniProxy/user
If-None-Match: "abc123"

# Returns 304 Not Modified if unchanged
```

## API Documentation

### Authentication

#### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

Response:
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "abc123...",
  "token_type": "Bearer"
}
```

#### Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"abc123..."}'
```

### User Endpoints

All user endpoints require `Authorization: Bearer <token>` header.

#### Get Current User
```bash
curl http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer <token>"
```

#### Get Plan
```bash
curl http://localhost:8080/api/v1/me/plan \
  -H "Authorization: Bearer <token>"
```

#### Get Allowed Nodes
```bash
curl http://localhost:8080/api/v1/me/nodes \
  -H "Authorization: Bearer <token>"
```

#### Get Current Usage
```bash
curl http://localhost:8080/api/v1/me/usage \
  -H "Authorization: Bearer <token>"
```

Response:
```json
{
  "usage": {
    "real_bytes_up": 1000000,
    "real_bytes_down": 2000000,
    "billable_bytes_up": 1500000,
    "billable_bytes_down": 3000000,
    "period_start": "2025-01-01T00:00:00Z",
    "period_end": "2025-02-01T00:00:00Z"
  }
}
```

### Admin Endpoints

All admin endpoints require admin role.

#### Create User
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

#### Create Node
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

#### Create Plan
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

#### Create Label
```bash
curl -X POST http://localhost:8080/api/v1/admin/labels \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Premium",
    "description": "Premium tier nodes"
  }'
```

## Traffic Accounting

### Multiplier Calculation

The final billable traffic is calculated using:

```
billable_bytes = real_bytes × node_multiplier × plan_base_multiplier × Π(label_multipliers)
```

Where:
- `node_multiplier` - Per-node multiplier (default: 1.0)
- `plan_base_multiplier` - Plan-wide multiplier (default: 1.0)
- `label_multipliers` - Product of all matching label multipliers

### Example

Node: `node_multiplier = 1.5`, labels: [Premium, US]
Plan: `base_multiplier = 1.0`, label_multipliers: {Premium: 2.0}

```
billable = real × 1.5 × 1.0 × 2.0 = real × 3.0
```

### Label Matching Semantics

A plan's label list defines which nodes are **allowed**. A node is accessible if it has **at least one** label that matches any label in the plan.

```
Plan labels: [Premium, Standard]
Node A labels: [Premium] -> ✓ Allowed
Node B labels: [Standard, US] -> ✓ Allowed
Node C labels: [Free] -> ✗ Not allowed
```

## Prometheus Metrics

### Service Metrics

Available at `/metrics`:

- `http_request_duration_seconds` - HTTP request duration histogram
- `active_nodes` - Number of active nodes
- `traffic_reports_total` - Total traffic reports received
- `telegram_notifications_total` - Total Telegram notifications sent
- `accounting_errors_total` - Total accounting errors
- `user_traffic_bytes_total` - User traffic counters
- `online_users_total` - Currently online users

### Example PromQL Queries

**Total traffic by user (last 24h)**
```promql
sum by (user_id) (
  increase(user_traffic_bytes_total{direction="up",type="billable"}[24h])
)
```

**Traffic report rate**
```promql
rate(traffic_reports_total[5m])
```

**Active nodes**
```promql
active_nodes
```

**HTTP request latency (p95)**
```promql
histogram_quantile(0.95,
  rate(http_request_duration_seconds_bucket[5m])
)
```

## Telegram Bot

### Setup

1. Create a bot via [@BotFather](https://t.me/botfather)
2. Set the token in config or environment: `TELEGRAM_TOKEN=<your_token>`
3. Restart the application

### Linking Account

1. In the web dashboard, click "Generate Link Token"
2. Send `/link <token>` to your bot
3. You'll receive notifications based on configured thresholds

### Notification Types

- **Threshold alerts**: 50%, 80%, 95% quota usage
- **Quota exceeded**: When user exceeds plan quota

### Example Notification

```
Usage Alert for user@example.com

Real Usage:
  Upload: 50.0 GiB
  Download: 100.0 GiB
  Total: 150.0 GiB

Billable Usage:
  Upload: 75.0 GiB
  Download: 150.0 GiB
  Total: 225.0 GiB

Quota: 250.0 GiB
Used: 90.0%
```

## Background Jobs

### Plan Reset Job

Runs every hour. Checks all current usage periods and resets if `period_end` has passed.

### Notification Job

Runs every 5 minutes. Checks user quotas against thresholds and sends Telegram notifications.

### Online User Cleanup

Runs every 10 minutes. Removes stale online user records.

## Development

### Running Tests

```bash
make test
```

### Building

```bash
make build
```

### Database Migrations

Create new migration:
```bash
migrate create -ext sql -dir migrations -seq <name>
```

Apply migrations:
```bash
make migrate-up
```

Rollback:
```bash
make migrate-down
```

## Production Deployment

### Checklist

- [ ] Change `JWT_SECRET` to a strong random value
- [ ] Change `NODE_SERVER_TOKEN` to a strong random value
- [ ] Set `SERVER_MODE=release`
- [ ] Configure proper database credentials
- [ ] Set up SSL/TLS (use reverse proxy like nginx)
- [ ] Configure firewall rules
- [ ] Set up log aggregation
- [ ] Configure Prometheus retention
- [ ] Set up Grafana dashboards
- [ ] Configure backup strategy for MariaDB

### Recommended Architecture

```
Internet
    ↓
[Nginx/Caddy with SSL]
    ↓
[Xboard Go :8080]
    ↓
[MariaDB :3306]
[Prometheus :9090]
[Grafana :3000]
```

## Troubleshooting

### Node Connection Issues

1. Verify `NODE_SERVER_TOKEN` matches on both sides
2. Check node can reach server (network, firewall)
3. Verify `node_id` exists in database
4. Check server logs for authentication errors

### Traffic Not Recording

1. Ensure user has an active plan
2. Verify user is not banned
3. Check that node has labels matching user's plan
4. Review server logs for accounting errors
5. Check Prometheus metrics: `accounting_errors_total`

### Database Migration Errors

1. Ensure database is running and accessible
2. Verify credentials in config
3. Check migration version: `migrate -path migrations -database "..." version`
4. Manually inspect database state if needed

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License

## Acknowledgments

- Built to be compatible with [Xboard](https://github.com/cedar2025/Xboard)
- Uses [Gin](https://github.com/gin-gonic/gin) web framework
- Database ORM: [GORM](https://gorm.io)
- Metrics: [Prometheus](https://prometheus.io)
