#!/bin/bash

# Complete Migration Script: Xboard → Next-Board
# This script automates the full migration process
#
# Usage:
#   ./migrate.sh
#
# Or with custom settings:
#   XBOARD_DB=my_xboard NEXTBOARD_DB=my_nextboard ./migrate.sh
#   XBOARD_DOCKER=container_name ./migrate.sh

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration (can be overridden by environment variables)
XBOARD_DB=${XBOARD_DB:-xboard}
NEXTBOARD_DB=${NEXTBOARD_DB:-xboard_go}
XBOARD_DOCKER=${XBOARD_DOCKER:-}  # Docker container name (if Xboard is in Docker)
MYSQL_USER=${MYSQL_USER:-root}
MYSQL_CMD=${MYSQL_CMD:-mysql}  # Can be 'mysql' or 'mycli'
BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                                                        ║${NC}"
echo -e "${BLUE}║         Xboard → Next-Board Migration Tool            ║${NC}"
echo -e "${BLUE}║                                                        ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Function to print section headers
print_header() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Function to execute MySQL command (handles Docker)
exec_mysql() {
    local db=$1
    shift
    local sql="$@"

    if [ -n "$XBOARD_DOCKER" ] && [ "$db" = "$XBOARD_DB" ]; then
        # Execute in Docker container
        docker exec -i "$XBOARD_DOCKER" mysql -u"$MYSQL_USER" -p"$MYSQL_PASS" "$db" -e "$sql" 2>/dev/null
    else
        # Execute locally
        $MYSQL_CMD -u"$MYSQL_USER" -p "$db" -e "$sql" 2>/dev/null
    fi
}

# Function to check MySQL connection
check_mysql() {
    local db=$1
    if [ -n "$XBOARD_DOCKER" ] && [ "$db" = "$XBOARD_DB" ]; then
        # Check Docker container
        if docker exec -i "$XBOARD_DOCKER" mysql -u"$MYSQL_USER" -p"$MYSQL_PASS" -e "USE $db;" 2>/dev/null; then
            echo -e "${GREEN}✓${NC} Connected to database: $db (in Docker: $XBOARD_DOCKER)"
            return 0
        else
            echo -e "${RED}✗${NC} Cannot connect to database: $db (in Docker: $XBOARD_DOCKER)"
            return 1
        fi
    else
        # Check local database
        if $MYSQL_CMD -u"$MYSQL_USER" -p -e "USE $db;" 2>/dev/null; then
            echo -e "${GREEN}✓${NC} Connected to database: $db"
            return 0
        else
            echo -e "${RED}✗${NC} Cannot connect to database: $db"
            return 1
        fi
    fi
}

# Function to get table count
get_table_count() {
    local db=$1
    local table=$2

    if [ -n "$XBOARD_DOCKER" ] && [ "$db" = "$XBOARD_DB" ]; then
        # Get count from Docker
        docker exec -i "$XBOARD_DOCKER" mysql -u"$MYSQL_USER" -p"$MYSQL_PASS" -N -e "SELECT COUNT(*) FROM $db.$table;" 2>/dev/null || echo "0"
    else
        # Get count from local
        $MYSQL_CMD -u"$MYSQL_USER" -p -N -e "SELECT COUNT(*) FROM $db.$table;" 2>/dev/null || echo "0"
    fi
}

# ============================================================================
# STEP 0: Pre-flight Checks
# ============================================================================
print_header "Pre-flight Checks"

echo -n "Checking MySQL credentials... "
if ! mysql -u"$MYSQL_USER" -p -e "SELECT 1;" >/dev/null 2>&1; then
    echo -e "${RED}Failed${NC}"
    echo "Please ensure MySQL credentials are correct"
    exit 1
fi
echo -e "${GREEN}OK${NC}"

echo ""
echo "Configuration:"
echo "  Source DB:      $XBOARD_DB"
if [ -n "$XBOARD_DOCKER" ]; then
    echo "  Source:         Docker container: $XBOARD_DOCKER"
fi
echo "  Target DB:      $NEXTBOARD_DB"
echo "  MySQL User:     $MYSQL_USER"
echo "  MySQL Command:  $MYSQL_CMD"
echo "  Backup Dir:     $BACKUP_DIR"
echo ""

read -p "Continue with these settings? [y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Migration cancelled"
    exit 1
fi

# ============================================================================
# STEP 1: Backup Databases
# ============================================================================
print_header "Step 1: Backup Databases"

mkdir -p "$BACKUP_DIR"

echo "Backing up $XBOARD_DB..."
if [ -n "$XBOARD_DOCKER" ]; then
    # Dump from Docker container
    echo "  (Dumping from Docker container: $XBOARD_DOCKER)"
    read -sp "Enter MySQL password for Xboard: " MYSQL_PASS
    echo ""
    docker exec "$XBOARD_DOCKER" mysqldump -u"$MYSQL_USER" -p"$MYSQL_PASS" "$XBOARD_DB" > "$BACKUP_DIR/${XBOARD_DB}_backup_${TIMESTAMP}.sql"
else
    # Dump from local
    mysqldump -u"$MYSQL_USER" -p "$XBOARD_DB" > "$BACKUP_DIR/${XBOARD_DB}_backup_${TIMESTAMP}.sql"
fi

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓${NC} Xboard backup saved: $BACKUP_DIR/${XBOARD_DB}_backup_${TIMESTAMP}.sql"
else
    echo -e "${RED}✗${NC} Backup failed!"
    exit 1
fi

echo "Backing up $NEXTBOARD_DB (if exists)..."
if mysql -u"$MYSQL_USER" -p -e "USE $NEXTBOARD_DB;" 2>/dev/null; then
    mysqldump -u"$MYSQL_USER" -p "$NEXTBOARD_DB" > "$BACKUP_DIR/${NEXTBOARD_DB}_backup_${TIMESTAMP}.sql"
    echo -e "${GREEN}✓${NC} Next-Board backup saved: $BACKUP_DIR/${NEXTBOARD_DB}_backup_${TIMESTAMP}.sql"
else
    echo -e "${YELLOW}⚠${NC} Next-Board database doesn't exist yet (this is OK)"
fi

# ============================================================================
# STEP 2: Verify Source Data
# ============================================================================
print_header "Step 2: Verify Source Data"

XBOARD_USERS=$(get_table_count "$XBOARD_DB" "v2_user")
XBOARD_PLANS=$(get_table_count "$XBOARD_DB" "v2_plan")
XBOARD_GROUPS=$(get_table_count "$XBOARD_DB" "v2_server_group")

echo "Source database statistics:"
echo "  Users:          $XBOARD_USERS"
echo "  Plans:          $XBOARD_PLANS"
echo "  Server Groups:  $XBOARD_GROUPS"
echo ""

if [ "$XBOARD_USERS" -eq 0 ]; then
    echo -e "${YELLOW}⚠${NC} Warning: No users found in source database"
    read -p "Continue anyway? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# ============================================================================
# STEP 3: Prepare Migration SQL
# ============================================================================
print_header "Step 3: Prepare Migration SQL"

MIGRATION_SQL="$BACKUP_DIR/migration_${TIMESTAMP}.sql"

cat > "$MIGRATION_SQL" << 'EOSQL'
-- Auto-generated migration script
-- Xboard → Next-Board

USE __NEXTBOARD_DB__;

SET FOREIGN_KEY_CHECKS = 0;
SET NAMES utf8mb4;
SET CHARACTER_SET_CLIENT = utf8mb4;

-- ============================================================================
-- STEP 1: Migrate Labels (from server groups)
-- ============================================================================

INSERT INTO labels (name, description, created_at, updated_at)
SELECT
    name,
    CONCAT('Migrated from Xboard group ID: ', id),
    FROM_UNIXTIME(created_at),
    FROM_UNIXTIME(updated_at)
FROM __XBOARD_DB__.v2_server_group
ON DUPLICATE KEY UPDATE name=name;

-- Temporary mapping table
CREATE TEMPORARY TABLE IF NOT EXISTS temp_group_label_map (
    old_group_id INT,
    new_label_id BIGINT UNSIGNED,
    PRIMARY KEY (old_group_id)
);

INSERT INTO temp_group_label_map (old_group_id, new_label_id)
SELECT sg.id, l.id
FROM __XBOARD_DB__.v2_server_group sg
JOIN labels l ON l.name = sg.name;

-- ============================================================================
-- STEP 2: Migrate Plans
-- ============================================================================

INSERT INTO plans (name, quota_bytes, reset_period, base_multiplier, created_at, updated_at)
SELECT
    p.name,
    p.transfer_enable as quota_bytes,
    CASE
        WHEN p.reset_traffic_method = 0 THEN 'monthly'
        WHEN p.reset_traffic_method = 1 THEN 'monthly'
        WHEN p.reset_traffic_method = 2 THEN 'none'
        WHEN p.reset_traffic_method = 3 THEN 'yearly'
        WHEN p.reset_traffic_method = 4 THEN 'yearly'
        ELSE 'monthly'
    END as reset_period,
    1.0 as base_multiplier,
    FROM_UNIXTIME(p.created_at),
    FROM_UNIXTIME(p.updated_at)
FROM __XBOARD_DB__.v2_plan p
ON DUPLICATE KEY UPDATE name=name;

-- Link plans to labels
INSERT INTO plan_labels (plan_id, label_id, created_at)
SELECT DISTINCT
    np.id as plan_id,
    tm.new_label_id as label_id,
    NOW()
FROM __XBOARD_DB__.v2_plan p
JOIN plans np ON np.name = p.name
JOIN temp_group_label_map tm ON tm.old_group_id = p.group_id
ON DUPLICATE KEY UPDATE plan_id=plan_id;

-- ============================================================================
-- STEP 3: Migrate Users
-- ============================================================================

INSERT INTO users (
    email,
    password_hash,
    role,
    plan_id,
    telegram_chat_id,
    telegram_linked_at,
    banned,
    created_at,
    updated_at
)
SELECT
    u.email,
    u.password as password_hash,
    CASE WHEN u.is_admin = 1 THEN 'admin' ELSE 'user' END as role,
    np.id as plan_id,
    u.telegram_id as telegram_chat_id,
    CASE WHEN u.telegram_id IS NOT NULL THEN FROM_UNIXTIME(u.created_at) ELSE NULL END as telegram_linked_at,
    u.banned,
    FROM_UNIXTIME(u.created_at),
    FROM_UNIXTIME(u.updated_at)
FROM __XBOARD_DB__.v2_user u
LEFT JOIN plans np ON np.id = u.plan_id
ON DUPLICATE KEY UPDATE email=email;

-- Migrate UUIDs
INSERT INTO user_uuids (user_id, uuid, created_at)
SELECT nu.id, ou.uuid, NOW()
FROM __XBOARD_DB__.v2_user ou
JOIN users nu ON nu.email = ou.email
ON DUPLICATE KEY UPDATE uuid=uuid;

-- ============================================================================
-- STEP 4: Migrate Nodes
-- ============================================================================

-- Check if v2_server table exists (newer Xboard versions)
SET @table_exists = (
    SELECT COUNT(*)
    FROM information_schema.tables
    WHERE table_schema = '__XBOARD_DB__'
    AND table_name = 'v2_server'
);

-- Migrate from unified v2_server table (if exists)
SET @migrate_servers = IF(@table_exists > 0,
    'INSERT INTO nodes (name, node_type, host, port, protocol_config, node_multiplier, status, created_at, updated_at)
    SELECT
        name,
        type as node_type,
        host,
        port,
        CASE
            WHEN type = \"vmess\" THEN JSON_OBJECT(\"network\", network, \"tls\", tls)
            WHEN type = \"vless\" THEN JSON_OBJECT(\"network\", network, \"tls\", tls, \"flow\", flow)
            WHEN type = \"trojan\" THEN JSON_OBJECT(\"server_name\", server_name)
            WHEN type = \"shadowsocks\" THEN JSON_OBJECT(\"cipher\", cipher)
            WHEN type = \"hysteria\" THEN JSON_OBJECT(\"up_mbps\", up_mbps, \"down_mbps\", down_mbps)
            ELSE JSON_OBJECT()
        END as protocol_config,
        CAST(rate as DECIMAL(10, 4)) as node_multiplier,
        CASE WHEN `show` = 1 THEN \"active\" ELSE \"inactive\" END as status,
        FROM_UNIXTIME(created_at),
        FROM_UNIXTIME(updated_at)
    FROM __XBOARD_DB__.v2_server
    ON DUPLICATE KEY UPDATE name=name',
    'SELECT 1'  -- No-op if table doesn't exist
);

PREPARE stmt FROM @migrate_servers;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================================================
-- STEP 5: Create Node-Label Associations
-- ============================================================================

-- Link nodes to labels based on group_id
INSERT INTO node_labels (node_id, label_id, created_at)
SELECT DISTINCT
    n.id as node_id,
    tm.new_label_id as label_id,
    NOW()
FROM __XBOARD_DB__.v2_server s
JOIN nodes n ON n.name = s.name
JOIN temp_group_label_map tm ON FIND_IN_SET(tm.old_group_id, REPLACE(REPLACE(s.group_id, '[', ''), ']', '')) > 0
ON DUPLICATE KEY UPDATE node_id=node_id;

-- ============================================================================
-- STEP 6: Initialize Usage Periods
-- ============================================================================

INSERT INTO usage_periods (
    user_id,
    plan_id,
    period_start,
    period_end,
    real_bytes_up,
    real_bytes_down,
    billable_bytes_up,
    billable_bytes_down,
    is_current,
    created_at,
    updated_at
)
SELECT
    nu.id as user_id,
    nu.plan_id,
    DATE_FORMAT(NOW(), '%Y-%m-01 00:00:00') as period_start,
    DATE_FORMAT(DATE_ADD(NOW(), INTERVAL 1 MONTH), '%Y-%m-01 00:00:00') as period_end,
    ou.u as real_bytes_up,
    ou.d as real_bytes_down,
    ou.u as billable_bytes_up,
    ou.d as billable_bytes_down,
    TRUE as is_current,
    NOW(),
    NOW()
FROM __XBOARD_DB__.v2_user ou
JOIN users nu ON nu.email = ou.email
WHERE nu.plan_id IS NOT NULL
ON DUPLICATE KEY UPDATE user_id=user_id;

-- Cleanup
DROP TEMPORARY TABLE IF EXISTS temp_group_label_map;

SET FOREIGN_KEY_CHECKS = 1;

-- Summary
SELECT '============================================' as '';
SELECT 'Migration Summary' as '';
SELECT '============================================' as '';
SELECT CONCAT('Labels:         ', COUNT(*)) as '' FROM labels;
SELECT CONCAT('Plans:          ', COUNT(*)) as '' FROM plans;
SELECT CONCAT('Users:          ', COUNT(*)) as '' FROM users;
SELECT CONCAT('Nodes:          ', COUNT(*)) as '' FROM nodes;
SELECT CONCAT('Usage Periods:  ', COUNT(*)) as '' FROM usage_periods;
SELECT '============================================' as '';
EOSQL

# Replace database names
sed -i "s/__XBOARD_DB__/$XBOARD_DB/g" "$MIGRATION_SQL"
sed -i "s/__NEXTBOARD_DB__/$NEXTBOARD_DB/g" "$MIGRATION_SQL"

echo -e "${GREEN}✓${NC} Migration SQL prepared: $MIGRATION_SQL"

# ============================================================================
# STEP 4: Run Migration
# ============================================================================
print_header "Step 4: Run Migration"

echo "This will migrate data from $XBOARD_DB to $NEXTBOARD_DB"
echo ""
read -p "Start migration now? [y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Migration cancelled"
    echo "You can run it manually later with:"
    echo "  mysql -u$MYSQL_USER -p < $MIGRATION_SQL"
    exit 0
fi

echo "Running migration..."
echo ""

if mysql -u"$MYSQL_USER" -p < "$MIGRATION_SQL"; then
    echo ""
    echo -e "${GREEN}✓${NC} Migration completed successfully!"
else
    echo ""
    echo -e "${RED}✗${NC} Migration failed!"
    echo ""
    echo "To rollback, run:"
    echo "  mysql -u$MYSQL_USER -p $NEXTBOARD_DB < $BACKUP_DIR/${NEXTBOARD_DB}_backup_${TIMESTAMP}.sql"
    exit 1
fi

# ============================================================================
# STEP 5: Verify Migration
# ============================================================================
print_header "Step 5: Verify Migration"

NEXTBOARD_USERS=$(get_table_count "$NEXTBOARD_DB" "users")
NEXTBOARD_PLANS=$(get_table_count "$NEXTBOARD_DB" "plans")
NEXTBOARD_LABELS=$(get_table_count "$NEXTBOARD_DB" "labels")
NEXTBOARD_NODES=$(get_table_count "$NEXTBOARD_DB" "nodes")

echo "Migration Results:"
echo ""
echo "  Source (Xboard)         →  Target (Next-Board)"
echo "  ───────────────────────────────────────────────"
echo "  Users:  $XBOARD_USERS → $NEXTBOARD_USERS"
echo "  Plans:  $XBOARD_PLANS → $NEXTBOARD_PLANS"
echo "  Groups: $XBOARD_GROUPS → Labels: $NEXTBOARD_LABELS"
echo "  Nodes:  (various)       →  $NEXTBOARD_NODES"
echo ""

# Check for discrepancies
ISSUES=0

if [ "$XBOARD_USERS" -ne "$NEXTBOARD_USERS" ]; then
    echo -e "${YELLOW}⚠${NC} User count mismatch (expected: $XBOARD_USERS, got: $NEXTBOARD_USERS)"
    ISSUES=$((ISSUES + 1))
fi

if [ "$XBOARD_PLANS" -ne "$NEXTBOARD_PLANS" ]; then
    echo -e "${YELLOW}⚠${NC} Plan count mismatch (expected: $XBOARD_PLANS, got: $NEXTBOARD_PLANS)"
    ISSUES=$((ISSUES + 1))
fi

if [ "$XBOARD_GROUPS" -ne "$NEXTBOARD_LABELS" ]; then
    echo -e "${YELLOW}⚠${NC} Label count mismatch (expected: $XBOARD_GROUPS, got: $NEXTBOARD_LABELS)"
    ISSUES=$((ISSUES + 1))
fi

if [ $ISSUES -eq 0 ]; then
    echo -e "${GREEN}✓${NC} All counts match!"
else
    echo ""
    echo -e "${YELLOW}Found $ISSUES potential issues. Please review manually.${NC}"
fi

# ============================================================================
# STEP 6: Summary
# ============================================================================
print_header "Migration Complete!"

echo ""
echo "Backups saved to:"
echo "  • $BACKUP_DIR/${XBOARD_DB}_backup_${TIMESTAMP}.sql"
echo "  • $BACKUP_DIR/${NEXTBOARD_DB}_backup_${TIMESTAMP}.sql"
echo ""
echo "Migration SQL:"
echo "  • $MIGRATION_SQL"
echo ""
echo -e "${GREEN}Next Steps:${NC}"
echo ""
echo "1. Test user login:"
echo "   cd /home/kexi/Next-Board/xboard-go"
echo "   make run"
echo "   curl -X POST http://localhost:8080/api/v1/auth/login \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"email\":\"user@example.com\",\"password\":\"password\"}'"
echo ""
echo "2. Update web UI backend URL in Cloudflare Pages:"
echo "   VITE_API_BASE_URL = http://YOUR_BACKEND_IP:8080"
echo ""
echo "3. Configure CORS in xboard-go/config.json:"
echo "   \"cors_origins\": [\"https://nextboard.yuanzhisheng.pages.dev\"]"
echo ""
echo "4. Review migrated data in database:"
echo "   mysql -u$MYSQL_USER -p $NEXTBOARD_DB"
echo ""
echo -e "${YELLOW}If you need to rollback:${NC}"
echo "  mysql -u$MYSQL_USER -p $NEXTBOARD_DB < $BACKUP_DIR/${NEXTBOARD_DB}_backup_${TIMESTAMP}.sql"
echo ""
echo -e "${GREEN}✨ Migration successful! Your Next-Board is ready to use.${NC}"
echo ""
