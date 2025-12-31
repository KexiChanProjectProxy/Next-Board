#!/bin/bash

# Quickest User Migration - Users can re-login immediately
# Usage: XBOARD_DOCKER=container_name ./quick_user_migrate.sh

set -e

XBOARD_DOCKER=${XBOARD_DOCKER:-}
XBOARD_DB=${XBOARD_DB:-xboard}
NEXTBOARD_DB=${NEXTBOARD_DB:-xboard}

if [ -z "$XBOARD_DOCKER" ]; then
    echo "Error: Set XBOARD_DOCKER=container_name"
    exit 1
fi

echo "🚀 Quick User Migration Starting..."
echo ""
echo "What will be migrated:"
echo "  ✓ Users (email, password) - can login immediately"
echo "  ✓ UUIDs - subscription URLs work immediately"
echo "  ✓ Tokens - no re-authentication needed"
echo "  ✓ Balances - financial data preserved"
echo ""

# Step 1: Dump v2_user from Docker
echo "[1/3] Dumping users from Xboard Docker..."
read -sp "Enter MySQL password for Docker container (or press Enter if none): " DOCKER_MYSQL_PASS
echo ""

if [ -n "$DOCKER_MYSQL_PASS" ]; then
    docker exec "$XBOARD_DOCKER" mariadb-dump -uroot -p"$DOCKER_MYSQL_PASS" "$XBOARD_DB" v2_user > /tmp/v2_user.sql
else
    docker exec "$XBOARD_DOCKER" mariadb-dump -uroot "$XBOARD_DB" v2_user > /tmp/v2_user.sql
fi

if [ $? -ne 0 ]; then
    echo "Error: Failed to dump users from Docker"
    exit 1
fi

USER_COUNT=$(grep "^INSERT INTO" /tmp/v2_user.sql | wc -l)
echo "✓ Dumped users from Docker"

# Step 2: Create migration SQL
echo "[2/3] Creating migration SQL..."
cat > /tmp/do_user_migration.sql << 'EOF'
SET FOREIGN_KEY_CHECKS = 0;

-- Create temporary table with Xboard structure
CREATE TEMPORARY TABLE v2_user (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(64),
    password VARCHAR(64),
    is_admin BOOLEAN DEFAULT FALSE,
    plan_id INT,
    telegram_id BIGINT,
    banned BOOLEAN DEFAULT FALSE,
    balance INT DEFAULT 0,
    discount INT,
    commission_type TINYINT DEFAULT 0,
    commission_rate INT,
    commission_balance INT DEFAULT 0,
    token VARCHAR(32),
    last_login_at INT,
    last_login_ip INT,
    remarks TEXT,
    created_at INT,
    updated_at INT,
    uuid VARCHAR(36)
);

-- Load Xboard dump into temporary table
SOURCE /tmp/v2_user.sql;

-- Migrate to Next-Board users table
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
    password,
    CASE WHEN is_admin = 1 THEN 'admin' ELSE 'user' END,
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
FROM v2_user
ON DUPLICATE KEY UPDATE
    password_hash=VALUES(password_hash),
    role=VALUES(role),
    plan_id=VALUES(plan_id),
    telegram_chat_id=VALUES(telegram_chat_id),
    telegram_linked_at=VALUES(telegram_linked_at),
    banned=VALUES(banned),
    balance=VALUES(balance),
    discount=VALUES(discount),
    commission_type=VALUES(commission_type),
    commission_rate=VALUES(commission_rate),
    commission_balance=VALUES(commission_balance),
    token=VALUES(token),
    last_login_at=VALUES(last_login_at),
    last_login_ip=VALUES(last_login_ip),
    remarks=VALUES(remarks),
    updated_at=VALUES(updated_at);

-- Migrate UUIDs
INSERT INTO user_uuids (user_id, uuid, created_at)
SELECT u.id, v.uuid, NOW()
FROM v2_user v
JOIN users u ON u.email = v.email
WHERE v.uuid IS NOT NULL AND v.uuid != ''
ON DUPLICATE KEY UPDATE uuid=VALUES(uuid);

SET FOREIGN_KEY_CHECKS = 1;

-- Summary
SELECT '✓ Migration Complete' as Status;
SELECT COUNT(*) as 'Total Users' FROM users;
SELECT COUNT(*) as 'UUIDs Migrated' FROM user_uuids;
SELECT COUNT(*) as 'With Tokens' FROM users WHERE token IS NOT NULL;
SELECT COUNT(*) as 'Admins' FROM users WHERE role = 'admin';
SELECT SUM(balance)/100 as 'Total Balance (currency units)' FROM users;
EOF

# Step 3: Execute migration
echo "[3/3] Migrating users..."
mysql -uroot "$NEXTBOARD_DB" < /tmp/do_user_migration.sql

echo ""
echo "✅ Migration Complete!"
echo ""
echo "What works now:"
echo "  ✓ Users can login with existing email/password"
echo "  ✓ Subscription URLs work immediately (UUIDs preserved)"
echo "  ✓ User tokens preserved (no re-authentication needed)"
echo "  ✓ Balances and commissions preserved"
echo ""
echo "Next steps:"
echo "  1. Update node remote URLs to point to Next-Board"
echo "  2. Users can use nodes immediately (no sub update needed)"
echo ""
echo "Cleanup: rm /tmp/v2_user.sql /tmp/do_user_migration.sql"
