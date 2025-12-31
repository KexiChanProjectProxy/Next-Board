#!/bin/bash

# Ultra-Simple User Migration
# Only migrates users so they can re-login immediately
#
# Usage: XBOARD_DOCKER=container_name ./migrate_users_only.sh

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

XBOARD_DOCKER=${XBOARD_DOCKER:-}
XBOARD_DB=${XBOARD_DB:-xboard}
NEXTBOARD_DB=${NEXTBOARD_DB:-xboard}

echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║          Quick User Migration Tool                    ║${NC}"
echo -e "${BLUE}║          Users can re-login immediately               ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

if [ -z "$XBOARD_DOCKER" ]; then
    echo -e "${RED}Error: Set XBOARD_DOCKER=container_name${NC}"
    exit 1
fi

echo "Migrating users from Docker ($XBOARD_DOCKER) to local MariaDB..."
echo ""

# Dump ONLY v2_user table from Docker
echo -e "${BLUE}[1/3]${NC} Dumping users from Xboard..."
docker exec "$XBOARD_DOCKER" mariadb-dump -uroot --no-create-info --complete-insert "$XBOARD_DB" v2_user > /tmp/xboard_users_dump.sql

# Transform v2_user INSERT to users table INSERT
echo -e "${BLUE}[2/3]${NC} Converting format..."
cat > /tmp/migrate_users.sql << 'EOSQL'
SET FOREIGN_KEY_CHECKS = 0;

-- Insert users
INSERT INTO users (
    email, password_hash, role, plan_id, telegram_chat_id, telegram_linked_at,
    banned, balance, discount, commission_type, commission_rate, commission_balance,
    token, last_login_at, last_login_ip, remarks, created_at, updated_at
)
SELECT
    email,
    password as password_hash,
    CASE WHEN is_admin = 1 THEN 'admin' ELSE 'user' END as role,
    plan_id,
    telegram_id as telegram_chat_id,
    CASE WHEN telegram_id IS NOT NULL THEN FROM_UNIXTIME(created_at) ELSE NULL END as telegram_linked_at,
    banned,
    COALESCE(balance, 0) as balance,
    discount,
    COALESCE(commission_type, 0) as commission_type,
    commission_rate,
    COALESCE(commission_balance, 0) as commission_balance,
    token,
    CASE WHEN last_login_at > 0 THEN FROM_UNIXTIME(last_login_at) ELSE NULL END as last_login_at,
    CASE WHEN last_login_ip > 0 THEN INET_NTOA(last_login_ip) ELSE NULL END as last_login_ip,
    remarks,
    FROM_UNIXTIME(created_at) as created_at,
    FROM_UNIXTIME(updated_at) as updated_at
FROM (
EOSQL

# Extract just the VALUES part from the dump and create a subquery
grep "^INSERT INTO" /tmp/xboard_users_dump.sql | \
    sed "s/INSERT INTO \`v2_user\` VALUES //" | \
    sed 's/;$//' | \
    awk '{
        # Parse the INSERT VALUES and create SELECT statements
        print "SELECT * FROM (SELECT"
        gsub(/\),\(/, " UNION ALL SELECT ")
        gsub(/^\(/, "")
        gsub(/\)$/, "")
        # Add column aliases
        split($0, values, " UNION ALL SELECT ")
        for (i in values) {
            if (i == 1) {
                print values[i] " as row_data"
            } else {
                print "UNION ALL SELECT " values[i]
            }
        }
        print ") as source_data"
    }' >> /tmp/migrate_users.sql

# Finish the SQL
cat >> /tmp/migrate_users.sql << 'EOSQL'
) as source_users
ON DUPLICATE KEY UPDATE email=email;

-- Migrate UUIDs
INSERT INTO user_uuids (user_id, uuid, created_at)
SELECT u.id, src.uuid, NOW()
FROM (
EOSQL

# Add UUID extraction
grep "^INSERT INTO" /tmp/xboard_users_dump.sql | \
    sed "s/INSERT INTO \`v2_user\` //" | \
    awk -F, '{print "SELECT " $1 " as old_id, " $16 " as uuid"}' >> /tmp/migrate_users.sql

cat >> /tmp/migrate_users.sql << 'EOSQL'
) as src
JOIN users u ON u.id = src.old_id
WHERE src.uuid IS NOT NULL
ON DUPLICATE KEY UPDATE uuid=uuid;

SET FOREIGN_KEY_CHECKS = 1;

SELECT CONCAT('✓ Migrated ', COUNT(*), ' users') as Result FROM users;
EOSQL

# Execute migration
echo -e "${BLUE}[3/3]${NC} Importing to Next-Board..."
mysql -uroot "$NEXTBOARD_DB" < /tmp/migrate_users.sql 2>&1 | grep -E "Result|Error" || true

# Verify
MIGRATED=$(mysql -uroot -N "$NEXTBOARD_DB" -e "SELECT COUNT(*) FROM users;")
echo ""
echo -e "${GREEN}✓ Migration complete: $MIGRATED users${NC}"
echo ""
echo "Cleanup: rm /tmp/xboard_users_dump.sql /tmp/migrate_users.sql"
