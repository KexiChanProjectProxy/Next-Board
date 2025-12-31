#!/bin/bash

# Simple User Migration: Xboard → Next-Board
# Focus: Users can re-login with their existing credentials
#
# Usage: XBOARD_DOCKER=container_name ./migrate_users_simple.sh

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

XBOARD_DOCKER=${XBOARD_DOCKER:-}
XBOARD_DB=${XBOARD_DB:-xboard}
NEXTBOARD_DB=${NEXTBOARD_DB:-xboard}
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
TEMP_DIR="./temp_migration_${TIMESTAMP}"

echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     Simple User Migration: Xboard → Next-Board        ║${NC}"
echo -e "${BLUE}║     Focus: Users can re-login immediately             ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

if [ -z "$XBOARD_DOCKER" ]; then
    echo -e "${RED}Error: XBOARD_DOCKER environment variable not set${NC}"
    echo "Usage: XBOARD_DOCKER=container_name ./migrate_users_simple.sh"
    exit 1
fi

echo "Configuration:"
echo "  Source: Docker container $XBOARD_DOCKER"
echo "  Source DB: $XBOARD_DB"
echo "  Target DB: $NEXTBOARD_DB (local)"
echo ""

read -p "Continue? [y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 0
fi

mkdir -p "$TEMP_DIR"

echo ""
echo -e "${BLUE}Step 1: Export users from Xboard (Docker)${NC}"
docker exec "$XBOARD_DOCKER" mariadb -uroot "$XBOARD_DB" -e "
SELECT
    id,
    email,
    password,
    CASE WHEN is_admin = 1 THEN 'admin' ELSE 'user' END as role,
    plan_id,
    telegram_id,
    banned,
    balance,
    discount,
    commission_type,
    commission_rate,
    commission_balance,
    token,
    last_login_at,
    last_login_ip,
    remarks,
    created_at,
    updated_at,
    uuid
FROM v2_user
" > "$TEMP_DIR/users.tsv"

USER_COUNT=$(tail -n +2 "$TEMP_DIR/users.tsv" | wc -l)
echo -e "${GREEN}✓${NC} Exported $USER_COUNT users"

echo ""
echo -e "${BLUE}Step 2: Import users to Next-Board (Local)${NC}"

# Create import SQL
cat > "$TEMP_DIR/import_users.sql" << 'EOF'
-- Import users from Xboard to Next-Board

SET FOREIGN_KEY_CHECKS = 0;

-- Load data into temporary table first
CREATE TEMPORARY TABLE temp_xboard_users (
    old_id INT,
    email VARCHAR(255),
    password_hash VARCHAR(255),
    role VARCHAR(10),
    plan_id INT,
    telegram_id BIGINT,
    banned BOOLEAN,
    balance INT,
    discount INT,
    commission_type TINYINT,
    commission_rate INT,
    commission_balance INT,
    token VARCHAR(32),
    last_login_at INT,
    last_login_ip INT,
    remarks TEXT,
    created_at INT,
    updated_at INT,
    uuid VARCHAR(36)
);

-- Load from TSV file
LOAD DATA LOCAL INFILE 'TEMP_DIR_PLACEHOLDER/users.tsv'
INTO TABLE temp_xboard_users
IGNORE 1 LINES
(old_id, email, password_hash, role, @plan_id, @telegram_id, banned,
 balance, @discount, commission_type, @commission_rate, commission_balance,
 @token, @last_login_at, @last_login_ip, @remarks, created_at, updated_at, uuid)
SET
    plan_id = NULLIF(@plan_id, 'NULL'),
    telegram_id = NULLIF(@telegram_id, 'NULL'),
    discount = NULLIF(@discount, 'NULL'),
    commission_rate = NULLIF(@commission_rate, 'NULL'),
    token = NULLIF(@token, 'NULL'),
    last_login_at = NULLIF(@last_login_at, 'NULL'),
    last_login_ip = NULLIF(@last_login_ip, 'NULL'),
    remarks = NULLIF(@remarks, 'NULL');

-- Insert into users table
INSERT INTO users (
    email,
    password_hash,
    role,
    plan_id,
    telegram_chat_id,
    telegram_linked_at,
    banned,
    balance,
    discount,
    commission_type,
    commission_rate,
    commission_balance,
    token,
    last_login_at,
    last_login_ip,
    remarks,
    created_at,
    updated_at
)
SELECT
    email,
    password_hash,
    role,
    plan_id,
    telegram_id,
    CASE WHEN telegram_id IS NOT NULL THEN FROM_UNIXTIME(created_at) ELSE NULL END,
    banned,
    balance,
    discount,
    commission_type,
    commission_rate,
    commission_balance,
    token,
    CASE WHEN last_login_at > 0 THEN FROM_UNIXTIME(last_login_at) ELSE NULL END,
    CASE WHEN last_login_ip > 0 THEN INET_NTOA(last_login_ip) ELSE NULL END,
    remarks,
    FROM_UNIXTIME(created_at),
    FROM_UNIXTIME(updated_at)
FROM temp_xboard_users
ON DUPLICATE KEY UPDATE email=email;

-- Insert UUIDs
INSERT INTO user_uuids (user_id, uuid, created_at)
SELECT u.id, t.uuid, NOW()
FROM temp_xboard_users t
JOIN users u ON u.email = t.email
WHERE t.uuid IS NOT NULL
ON DUPLICATE KEY UPDATE uuid=uuid;

-- Show summary
SELECT 'Migration Summary:' as '';
SELECT CONCAT('Users migrated: ', COUNT(*)) as '' FROM users;
SELECT CONCAT('UUIDs migrated: ', COUNT(*)) as '' FROM user_uuids;

SET FOREIGN_KEY_CHECKS = 1;
EOF

# Replace placeholder with actual temp dir
sed -i "s|TEMP_DIR_PLACEHOLDER|$PWD/$TEMP_DIR|g" "$TEMP_DIR/import_users.sql"

# Execute import
mysql -uroot --local-infile=1 "$NEXTBOARD_DB" < "$TEMP_DIR/import_users.sql"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓${NC} Users imported successfully"
else
    echo -e "${RED}✗${NC} Import failed"
    exit 1
fi

echo ""
echo -e "${BLUE}Step 3: Verify migration${NC}"
mysql -uroot "$NEXTBOARD_DB" -e "
SELECT
    COUNT(*) as total_users,
    SUM(CASE WHEN role = 'admin' THEN 1 ELSE 0 END) as admins,
    SUM(CASE WHEN token IS NOT NULL THEN 1 ELSE 0 END) as users_with_tokens,
    SUM(balance) as total_balance,
    SUM(commission_balance) as total_commission
FROM users;
"

echo ""
echo -e "${GREEN}════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Migration Complete!${NC}"
echo -e "${GREEN}════════════════════════════════════════════════════════${NC}"
echo ""
echo "✅ Users can now login with their existing credentials"
echo "✅ Tokens preserved (no re-authentication needed)"
echo "✅ Balances and commissions migrated"
echo ""
echo "Cleanup: rm -rf $TEMP_DIR"
echo ""
