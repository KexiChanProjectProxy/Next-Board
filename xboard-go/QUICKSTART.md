# Quick Start Guide

Get Xboard Go running in 5 minutes!

## Prerequisites

- Go 1.24+
- MariaDB 11.2+
- (Optional) Prometheus for metrics

## Installation Steps

### 1. Setup Database

Install MariaDB 11.2+ and create the database:

```bash
# Install MariaDB (Ubuntu/Debian)
sudo apt update
sudo apt install mariadb-server

# Or on macOS
brew install mariadb
brew services start mariadb

# Create database and user
mysql -u root -p
```

In the MySQL console:
```sql
CREATE DATABASE xboard_go;
CREATE USER 'xboard'@'localhost' IDENTIFIED BY 'xboard_password';
GRANT ALL PRIVILEGES ON xboard_go.* TO 'xboard'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### 2. Clone and Install Dependencies

```bash
cd xboard-go
go mod download
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

### 5. Create Initial Admin User

```bash
# Connect to database
mysql -u xboard -p xboard_go

# Create admin user (password: admin123)
INSERT INTO users (email, password_hash, role, banned, created_at, updated_at)
VALUES (
    'admin@example.com',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'admin',
    0,
    NOW(),
    NOW()
);
```

### 6. Run the Server

```bash
make run
# Or
go run ./cmd/server
# Or build and run
make build
./bin/server
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
- Verify MariaDB is running: `systemctl status mariadb` or `brew services list`
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
- Check application logs in the terminal where server is running

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
# Run the application
make run

# Build binary
make build

# Run tests
make test

# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Access database
mysql -u xboard -p xboard_go

# Check MariaDB status
systemctl status mariadb
# Or on macOS
brew services list

# Restart MariaDB
sudo systemctl restart mariadb
# Or on macOS
brew services restart mariadb
```

Enjoy using Xboard Go!
