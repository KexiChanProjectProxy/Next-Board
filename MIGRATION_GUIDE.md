# Migrating from Xboard (PHP) to Next-Board (Go)

This guide will help you migrate your existing Xboard instance data to Next-Board.

## Overview

**What gets migrated:**
- ✅ Users (email, passwords, roles, plan assignments, telegram links)
- ✅ Plans (quota, reset periods)
- ✅ Server Groups → Labels
- ✅ Nodes (all protocol types)
- ✅ Current usage data (approximate)

**What does NOT migrate:**
- ❌ Historical traffic statistics
- ❌ Orders and payments
- ❌ Tickets and knowledge base
- ❌ User tokens (users need to re-login)

## Prerequisites

### 1. Backup Everything

```bash
# Backup Xboard database
mysqldump -u root -p xboard > xboard_backup_$(date +%Y%m%d).sql

# Backup Next-Board database (if you have data)
mysqldump -u root -p xboard_go > xboard_go_backup_$(date +%Y%m%d).sql
```

### 2. Verify Database Access

Make sure you can connect to both databases:

```bash
# Test Xboard connection
mysql -u root -p xboard -e "SELECT COUNT(*) FROM v2_user;"

# Test Next-Board connection
mysql -u root -p xboard_go -e "SELECT COUNT(*) FROM users;"
```

### 3. Set Up Next-Board

```bash
cd /home/kexi/Next-Board/xboard-go

# Install dependencies
go mod download

# Run migrations (creates fresh schema)
make migrate-up

# Verify schema is created
mysql -u root -p xboard_go -e "SHOW TABLES;"
```

## Migration Steps

### Method 1: Automated SQL Script (Recommended)

This method uses the provided SQL migration script.

#### Step 1: Review the Migration Script

```bash
cat /home/kexi/Next-Board/xboard-go/scripts/migrate_from_xboard.sql
```

**Important**: Review these sections:
- Database names (adjust if yours are different)
- Node migration query (depends on your Xboard version)
- Group ID to Label mapping

#### Step 2: Customize for Your Setup

If your databases have different names, edit the script:

```bash
# Edit the script
nano /home/kexi/Next-Board/xboard-go/scripts/migrate_from_xboard.sql

# Change these lines:
# FROM xboard.v2_user     →  FROM your_xboard_db.v2_user
# USE xboard_go           →  USE your_nextboard_db
```

#### Step 3: Run the Migration

```bash
# Dry run (check for errors without committing)
mysql -u root -p --dry-run < /home/kexi/Next-Board/xboard-go/scripts/migrate_from_xboard.sql

# Actual migration
mysql -u root -p < /home/kexi/Next-Board/xboard-go/scripts/migrate_from_xboard.sql
```

#### Step 4: Verify Migration

```bash
mysql -u root -p xboard_go

# Check counts
SELECT 'Users' as Table_Name, COUNT(*) as Count FROM users
UNION ALL
SELECT 'Plans', COUNT(*) FROM plans
UNION ALL
SELECT 'Labels', COUNT(*) FROM labels
UNION ALL
SELECT 'Nodes', COUNT(*) FROM nodes;

# Check a sample user
SELECT * FROM users LIMIT 3;

# Check nodes have labels
SELECT n.name, l.name as label
FROM nodes n
JOIN node_labels nl ON nl.node_id = n.id
JOIN labels l ON l.id = nl.label_id
LIMIT 10;

# Exit
exit
```

### Method 2: Manual Migration (Step by Step)

If the automated script doesn't work for your setup, follow these manual steps:

#### 1. Migrate Server Groups to Labels

```sql
USE xboard_go;

INSERT INTO labels (name, description, created_at, updated_at)
SELECT
    name,
    CONCAT('Group from Xboard: ', id),
    FROM_UNIXTIME(created_at),
    FROM_UNIXTIME(updated_at)
FROM xboard.v2_server_group;
```

#### 2. Migrate Plans

```sql
INSERT INTO plans (name, quota_bytes, reset_period, base_multiplier, created_at, updated_at)
SELECT
    name,
    transfer_enable,
    'monthly' as reset_period,
    1.0 as base_multiplier,
    FROM_UNIXTIME(created_at),
    FROM_UNIXTIME(updated_at)
FROM xboard.v2_plan;
```

#### 3. Migrate Users

```sql
INSERT INTO users (email, password_hash, role, plan_id, telegram_chat_id, banned, created_at, updated_at)
SELECT
    email,
    password,
    IF(is_admin = 1, 'admin', 'user'),
    plan_id,
    telegram_id,
    banned,
    FROM_UNIXTIME(created_at),
    FROM_UNIXTIME(updated_at)
FROM xboard.v2_user;
```

#### 4. Migrate User UUIDs

```sql
INSERT INTO user_uuids (user_id, uuid, created_at)
SELECT
    nu.id,
    ou.uuid,
    NOW()
FROM xboard.v2_user ou
JOIN users nu ON nu.email = ou.email;
```

#### 5. Migrate Nodes

This depends on your Xboard version. For newer Xboard with unified `v2_server` table:

```sql
INSERT INTO nodes (name, node_type, host, port, node_multiplier, status, created_at, updated_at)
SELECT
    name,
    type,
    host,
    port,
    CAST(rate as DECIMAL(10, 4)),
    IF(`show` = 1, 'active', 'inactive'),
    FROM_UNIXTIME(created_at),
    FROM_UNIXTIME(updated_at)
FROM xboard.v2_server;
```

For older Xboard with separate protocol tables, you'll need to migrate each:

```sql
-- VMess nodes
INSERT INTO nodes (name, node_type, host, port, protocol_config, node_multiplier, status, created_at, updated_at)
SELECT
    name,
    'vmess',
    host,
    port,
    JSON_OBJECT('network', network, 'tls', tls),
    CAST(rate as DECIMAL(10, 4)),
    IF(`show` = 1, 'active', 'inactive'),
    FROM_UNIXTIME(created_at),
    FROM_UNIXTIME(updated_at)
FROM xboard.v2_server_vmess;

-- Repeat for v2_server_vless, v2_server_trojan, v2_server_shadowsocks, etc.
```

## Post-Migration Tasks

### 1. Create Initial Admin User

If you didn't have an admin in Xboard, create one:

```bash
cd /home/kexi/Next-Board/xboard-go

# Create admin user manually
mysql -u root -p xboard_go

INSERT INTO users (email, password_hash, role, created_at, updated_at)
VALUES (
    'admin@example.com',
    '$2a$10$YourBcryptHashHere',  -- Generate with: htpasswd -bnBC 10 "" yourpassword
    'admin',
    NOW(),
    NOW()
);
```

Or use bcrypt online generator: https://bcrypt.online/

### 2. Configure Label Multipliers

Set up traffic multipliers for labels:

```sql
USE xboard_go;

-- Example: Premium label has 2x multiplier
INSERT INTO plan_label_multipliers (plan_id, label_id, multiplier, created_at, updated_at)
SELECT
    p.id,
    l.id,
    2.0,
    NOW(),
    NOW()
FROM plans p
CROSS JOIN labels l
WHERE l.name = 'Premium';
```

### 3. Initialize Usage Periods

Create current billing periods for all users:

```sql
INSERT INTO usage_periods (user_id, plan_id, period_start, period_end, is_current, created_at, updated_at)
SELECT
    u.id,
    u.plan_id,
    DATE_FORMAT(NOW(), '%Y-%m-01 00:00:00'),
    DATE_FORMAT(DATE_ADD(NOW(), INTERVAL 1 MONTH), '%Y-%m-01 00:00:00'),
    TRUE,
    NOW(),
    NOW()
FROM users u
WHERE u.plan_id IS NOT NULL;
```

### 4. Test User Login

```bash
# Start Next-Board
cd /home/kexi/Next-Board/xboard-go
make run

# In another terminal, test login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

### 5. Configure CORS for Web UI

Edit `xboard-go/config.json`:

```json
{
  "server": {
    "cors_origins": ["https://your-pages-domain.pages.dev"]
  }
}
```

### 6. Update Backend URL in Web UI

In Cloudflare Pages environment variables:

```
VITE_API_BASE_URL = http://YOUR_BACKEND_IP:8080
```

Or for production:

```
VITE_API_BASE_URL = https://api.yourdomain.com
```

## Troubleshooting

### Password Login Fails

**Issue**: Users can't login after migration

**Solution**: Xboard might use different password hashing. Check:

```sql
SELECT password FROM xboard.v2_user LIMIT 1;
```

If it doesn't start with `$2a$` or `$2y$`, you'll need to reset passwords:

```bash
# Reset all user passwords (requires users to reset via email)
# Or manually update passwords with bcrypt hashes
```

### Node Labels Not Showing

**Issue**: Nodes don't have labels assigned

**Solution**: Manually link nodes to labels:

```sql
-- Get label IDs
SELECT id, name FROM labels;

-- Get node IDs
SELECT id, name FROM nodes;

-- Create associations
INSERT INTO node_labels (node_id, label_id, created_at)
VALUES (1, 1, NOW()), (2, 1, NOW()), (3, 2, NOW());
```

### Usage Data Shows Zero

**Issue**: All users show 0 usage

**Solution**: Current traffic wasn't migrated properly. Users will start fresh. To preserve old stats:

```sql
UPDATE usage_periods up
JOIN users u ON u.id = up.user_id
JOIN xboard.v2_user ou ON ou.email = u.email
SET
    up.real_bytes_up = ou.u,
    up.real_bytes_down = ou.d,
    up.billable_bytes_up = ou.u,
    up.billable_bytes_down = ou.d;
```

### Foreign Key Errors

**Issue**: Migration fails with foreign key constraint errors

**Solution**: Disable foreign key checks temporarily:

```sql
SET FOREIGN_KEY_CHECKS = 0;
-- Run your migration
SET FOREIGN_KEY_CHECKS = 1;
```

## Verification Checklist

After migration, verify:

- [ ] User count matches: `SELECT COUNT(*) FROM users;`
- [ ] Admin users have role='admin'
- [ ] Plans exist and have labels assigned
- [ ] Nodes exist and have labels assigned
- [ ] Test user can login via web UI
- [ ] Dashboard shows plan and usage correctly
- [ ] Nodes page shows available nodes
- [ ] Admin panel shows users (for admin accounts)

## Going Live

### 1. Switch DNS/Traffic

Once verified, switch traffic from Xboard to Next-Board:

- Update reverse proxy to point to Next-Board
- Or change DNS records
- Keep Xboard running in read-only mode temporarily

### 2. Monitor Logs

```bash
# Check Next-Board logs
cd /home/kexi/Next-Board/xboard-go
./bin/server  # Watch for errors

# Check database queries
mysql -u root -p xboard_go -e "SHOW PROCESSLIST;"
```

### 3. Update Nodes

Update your proxy nodes to report to Next-Board:

- Change node config to point to new backend URL
- Update `NODE_SERVER_TOKEN` in node configs
- Verify nodes appear in admin panel

## Rollback Plan

If migration fails:

```bash
# Stop Next-Board
pkill server

# Restore old database
mysql -u root -p xboard_go < xboard_go_backup_YYYYMMDD.sql

# Re-run migrations
cd /home/kexi/Next-Board/xboard-go
make migrate-down
make migrate-up

# Try migration again
```

## Getting Help

If you encounter issues:

1. Check Next-Board logs: `./bin/server` output
2. Check MySQL error log: `/var/log/mysql/error.log`
3. Review API documentation: `/home/kexi/Next-Board/xboard-go/API.md`
4. Create GitHub issue: https://github.com/your-repo/issues

## Success!

Once migration is complete:

- Users can login with their existing credentials
- All nodes are accessible based on plan labels
- Traffic accounting starts fresh (or from migrated values)
- Admin panel works for user management
- Web UI connects successfully

**Your Next-Board is now live! 🚀**
