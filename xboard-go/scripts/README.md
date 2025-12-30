# Migration Scripts

## Quick Start - One Command Migration

### If Xboard is in Docker (MariaDB or MySQL)

```bash
cd /home/kexi/Next-Board/xboard-go/scripts && \
XBOARD_DOCKER=your_container_name \
XBOARD_DB=xboard \
NEXTBOARD_DB=xboard_go \
./migrate.sh
```

**Replace `your_container_name` with your actual Docker container name**. Find it with:
```bash
docker ps | grep xboard
```

**Note**: The script automatically detects MariaDB in Docker and uses `mariadb-dump` and `mariadb` commands instead of `mysqldump` and `mysql`.

### If Xboard is Local (Not Docker)

```bash
cd /home/kexi/Next-Board/xboard-go/scripts && \
XBOARD_DB=xboard \
NEXTBOARD_DB=xboard_go \
./migrate.sh
```

## What This Does

The script automatically:

1. **Dumps** Xboard database (from Docker or local)
2. **Converts** data to Next-Board format:
   - Users (preserves passwords, roles, telegram links)
   - Plans (quota, reset periods)
   - Server Groups → Labels
   - Nodes (all protocol types)
   - Current usage data
3. **Imports** into Next-Board database
4. **Verifies** migration success
5. **Creates backups** of both databases

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `XBOARD_DOCKER` | (empty) | Docker container name if Xboard runs in Docker |
| `XBOARD_DB` | `xboard` | Source database name |
| `NEXTBOARD_DB` | `xboard_go` | Target database name |
| `MYSQL_USER` | `root` | MySQL username |
| `MYSQL_CMD` | `mysql` | MySQL client (`mysql` or `mycli`) |

## Examples

### Example 1: Xboard in Docker with custom database names

```bash
XBOARD_DOCKER=xboard_php \
XBOARD_DB=v2board \
NEXTBOARD_DB=nextboard \
./migrate.sh
```

### Example 2: Both databases local, use mycli

```bash
MYSQL_CMD=mycli \
XBOARD_DB=xboard \
NEXTBOARD_DB=xboard_go \
./migrate.sh
```

### Example 3: Xboard in Docker, Next-Board local

```bash
XBOARD_DOCKER=xboard-app \
./migrate.sh
```

## What Gets Migrated

✅ **Migrated:**
- Users (email, passwords, roles, plan assignments)
- Plans (quota, reset periods)
- Server Groups → Labels (with automatic mapping)
- Nodes (all protocol types: vmess, vless, trojan, shadowsocks, hysteria)
- Current usage data (approximate billable traffic)
- User UUIDs
- Telegram chat IDs

❌ **Not Migrated:**
- Historical traffic statistics
- Orders and payments
- Tickets and knowledge base
- User tokens (users need to re-login)
- Commission logs

## After Migration

1. **Test user login:**
   ```bash
   cd /home/kexi/Next-Board/xboard-go
   make run

   # In another terminal:
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"user@example.com","password":"password"}'
   ```

2. **Configure CORS** in `xboard-go/config.json`:
   ```json
   {
     "server": {
       "cors_origins": ["https://your-frontend.pages.dev"]
     }
   }
   ```

3. **Update Web UI** backend URL in Cloudflare Pages:
   ```
   VITE_API_BASE_URL = http://YOUR_BACKEND_IP:8080
   ```

4. **Update nodes** to report to Next-Board:
   - Change node config to point to new backend
   - Update `NODE_SERVER_TOKEN` in node configs

## Troubleshooting

### Check Docker container name
```bash
docker ps
docker ps -a | grep xboard
```

### Check database names
```bash
# In Docker:
docker exec your_container_name mysql -uroot -p -e "SHOW DATABASES;"

# Local:
mysql -uroot -p -e "SHOW DATABASES;"
```

### View backups
```bash
ls -lh ./backups/
```

### Rollback migration
```bash
mysql -uroot -p xboard_go < ./backups/xboard_go_backup_YYYYMMDD_HHMMSS.sql
```

### Manual migration (if script fails)
```bash
# Use the generated SQL file:
mysql -uroot -p < ./backups/migration_YYYYMMDD_HHMMSS.sql
```

## Verification Queries

After migration, verify data in Next-Board:

```sql
-- Connect to Next-Board database
mysql -uroot -p xboard_go

-- Check counts
SELECT 'Users' as Item, COUNT(*) FROM users
UNION ALL SELECT 'Plans', COUNT(*) FROM plans
UNION ALL SELECT 'Labels', COUNT(*) FROM labels
UNION ALL SELECT 'Nodes', COUNT(*) FROM nodes
UNION ALL SELECT 'Usage Periods', COUNT(*) FROM usage_periods;

-- Check sample user
SELECT id, email, role, plan_id, banned FROM users LIMIT 5;

-- Check nodes have labels
SELECT n.name, l.name as label
FROM nodes n
JOIN node_labels nl ON nl.node_id = n.id
JOIN labels l ON l.id = nl.label_id
LIMIT 10;

-- Check plans have labels
SELECT p.name as plan, l.name as label
FROM plans p
JOIN plan_labels pl ON pl.plan_id = p.id
JOIN labels l ON l.id = pl.label_id;
```

## File Locations

- **Migration Script**: `/home/kexi/Next-Board/xboard-go/scripts/migrate.sh`
- **Backups**: `/home/kexi/Next-Board/xboard-go/scripts/backups/`
- **Generated SQL**: `/home/kexi/Next-Board/xboard-go/scripts/backups/migration_*.sql`
- **Full Guide**: `/home/kexi/Next-Board/MIGRATION_GUIDE.md`

## Support

For detailed information, see:
- Full Migration Guide: `/home/kexi/Next-Board/MIGRATION_GUIDE.md`
- API Documentation: `/home/kexi/Next-Board/xboard-go/API.md`
- Next-Board README: `/home/kexi/Next-Board/xboard-go/README.md`
