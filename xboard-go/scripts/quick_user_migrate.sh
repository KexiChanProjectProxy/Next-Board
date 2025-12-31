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

# Step 1: Dump v2_user from Docker
echo "[1/3] Dumping users from Xboard Docker..."
docker exec "$XBOARD_DOCKER" mariadb-dump -uroot "$XBOARD_DB" v2_user > /tmp/v2_user.sql

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
ON DUPLICATE KEY UPDATE email=email;

-- Migrate UUIDs
INSERT INTO user_uuids (user_id, uuid, created_at)
SELECT u.id, v.uuid, NOW()
FROM v2_user v
JOIN users u ON u.email = v.email
WHERE v.uuid IS NOT NULL AND v.uuid != ''
ON DUPLICATE KEY UPDATE uuid=uuid;

SET FOREIGN_KEY_CHECKS = 1;

-- Summary
SELECT '✓ Migration Complete' as Status;
SELECT COUNT(*) as 'Total Users' FROM users;
SELECT COUNT(*) as 'With Tokens' FROM users WHERE token IS NOT NULL;
SELECT COUNT(*) as 'Admins' FROM users WHERE role = 'admin';
SELECT SUM(balance)/100 as 'Total Balance (currency units)' FROM users;
EOF

# Step 3: Execute migration
echo "[3/3] Migrating users..."
mysql -uroot "$NEXTBOARD_DB" < /tmp/do_user_migration.sql

echo ""
echo "✅ Done! Users can now login with their existing credentials."
echo ""
echo "Cleanup: rm /tmp/v2_user.sql /tmp/do_user_migration.sql"
