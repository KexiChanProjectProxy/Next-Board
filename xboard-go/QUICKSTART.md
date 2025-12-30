# Quick Start Guide

Get Xboard Go running in 5 minutes!

## Option 1: Docker Compose (Recommended)

### 1. Clone and Configure

```bash
cd xboard-go
cp .env.example .env
```

Edit `.env` and set at minimum:
```bash
JWT_SECRET=your-random-secret-here
NODE_SERVER_TOKEN=your-node-token-here
```

### 2. Start Everything

```bash
docker-compose up -d
```

This starts:
- MariaDB on port 3306
- Prometheus on port 9090
- Grafana on port 3000
- Xboard Go on port 8080

### 3. Run Migrations

```bash
docker-compose exec xboard-go migrate -path /app/migrations \
  -database "mysql://xboard:xboard_password@tcp(mariadb:3306)/xboard_go" up
```

Or if you have `golang-migrate` installed locally:
```bash
make migrate-up
```

### 4. Create Admin User

```bash
# Connect to the database
docker-compose exec mariadb mysql -uxboard -pxboard_password xboard_go

# Run SQL
INSERT INTO users (email, password_hash, role) VALUES
('admin@example.com', '$2a$10$YourBcryptHashHere', 'admin');
```

Or use the API once the server is running:
```bash
# This requires an existing admin to call, so bootstrap the first one via SQL
```

### 5. Access the Services

- **Web UI**: http://localhost:8080
- **Login**: http://localhost:8080/login
- **Dashboard**: http://localhost:8080/dashboard
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **API**: http://localhost:8080/api/v1

## Option 2: Local Development

### Prerequisites

- Go 1.24+
- MariaDB 11.2+
- (Optional) golang-migrate

### 1. Install Dependencies

```bash
cd xboard-go
go mod download
```

### 2. Start MariaDB

```bash
# Using Docker
docker run -d \
  --name xboard-mariadb \
  -e MYSQL_ROOT_PASSWORD=rootpassword \
  -e MYSQL_DATABASE=xboard_go \
  -e MYSQL_USER=xboard \
  -e MYSQL_PASSWORD=xboard_password \
  -p 3306:3306 \
  mariadb:11.2

# Or use your existing MariaDB instance
```

### 3. Configure

```bash
cp .env.example .env
```

Edit `.env`:
```bash
DB_HOST=localhost
DB_PORT=3306
DB_USER=xboard
DB_PASSWORD=xboard_password
DB_NAME=xboard_go
JWT_SECRET=your-secret-here
NODE_SERVER_TOKEN=your-node-token-here
```

### 4. Run Migrations

```bash
# Install golang-migrate if needed
go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
make migrate-up
```

### 5. Run the Server

```bash
make run
# Or
go run ./cmd/server
```

## First Steps

### 1. Create Labels

```bash
curl -X POST http://localhost:8080/api/v1/admin/labels \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Premium",
    "description": "Premium tier nodes"
  }'
```

### 2. Create a Plan

```bash
curl -X POST http://localhost:8080/api/v1/admin/plans \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Basic Plan",
    "quota_bytes": 107374182400,
    "reset_period": "monthly",
    "base_multiplier": 1.0,
    "label_ids": [1]
  }'
```

### 3. Create a Node

```bash
curl -X POST http://localhost:8080/api/v1/admin/nodes \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "US Node 1",
    "node_type": "vmess",
    "host": "us1.example.com",
    "port": 443,
    "node_multiplier": 1.0,
    "label_ids": [1]
  }'
```

### 4. Create a User

```bash
curl -X POST http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "role": "user",
    "plan_id": 1
  }'
```

### 5. Test Node Connection

From your node, configure it to connect to:
- Server URL: `http://localhost:8080`
- Token: `your-node-token-here`
- Node ID: `1`

The node should successfully fetch config and user list.

## Verifying Installation

### Check Health

```bash
curl http://localhost:8080/health
# Should return: {"status":"ok"}
```

### Check Metrics

```bash
curl http://localhost:8080/metrics
# Should return Prometheus metrics
```

### Test Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

Should return:
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "abc123...",
  "token_type": "Bearer"
}
```

## Common Issues

### Database Connection Failed

**Error**: `Failed to connect to database`

**Solution**:
- Verify MariaDB is running: `docker ps` or `systemctl status mariadb`
- Check credentials in `.env` or `config.json`
- Test connection: `mysql -h localhost -u xboard -p xboard_go`

### Migrations Failed

**Error**: `Dirty database version`

**Solution**:
```bash
migrate -path migrations \
  -database "mysql://xboard:xboard_password@tcp(localhost:3306)/xboard_go" \
  force VERSION
```

### Port Already in Use

**Error**: `bind: address already in use`

**Solution**:
- Change `SERVER_PORT` in `.env`
- Or stop the conflicting service: `lsof -ti:8080 | xargs kill`

### Token Authentication Failed

**Error**: `Invalid token` when node connects

**Solution**:
- Verify `NODE_SERVER_TOKEN` matches in both server config and node config
- Check node is using correct query parameter name: `token`

## Next Steps

1. **Set up Telegram bot**: Add `TELEGRAM_TOKEN` to `.env` and restart
2. **Configure Prometheus**: Import dashboards for monitoring
3. **Set up SSL**: Use nginx or Caddy as reverse proxy
4. **Configure backups**: Set up automated MariaDB backups
5. **Review logs**: Check application logs for any warnings

## Getting Help

- Read the full [README.md](README.md)
- Check [Xboard compatibility docs](docs/xboard-compat.md)
- Review API examples in README
- Check application logs: `docker-compose logs xboard-go`

## Security Checklist

Before going to production:

- [ ] Change `JWT_SECRET` to a strong random value (32+ characters)
- [ ] Change `NODE_SERVER_TOKEN` to a strong random value
- [ ] Set `SERVER_MODE=release` in production
- [ ] Use strong database passwords
- [ ] Set up SSL/TLS (never expose HTTP in production)
- [ ] Configure firewall rules
- [ ] Enable database backups
- [ ] Review and restrict CORS settings
- [ ] Set up log rotation
- [ ] Configure rate limiting (via reverse proxy)

## Useful Commands

```bash
# View logs
docker-compose logs -f xboard-go

# Restart service
docker-compose restart xboard-go

# Access database
docker-compose exec mariadb mysql -uxboard -pxboard_password xboard_go

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Build binary
make build

# Run tests
make test

# Stop all services
docker-compose down

# Clean volumes
docker-compose down -v
```

Enjoy using Xboard Go!
