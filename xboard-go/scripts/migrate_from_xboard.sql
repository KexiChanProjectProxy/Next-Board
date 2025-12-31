-- Migration Script: Xboard (PHP) to Next-Board (Go)
--
-- Prerequisites:
-- 1. Both databases should be accessible (source: xboard, target: xboard)
-- 2. Run Next-Board migrations first (make migrate-up)
-- 3. Backup both databases before running this script
--
-- Usage:
--   mysql -u root -p < migrate_from_xboard.sql
--
-- This script assumes:
-- - Source DB: xboard (Xboard PHP database)
-- - Target DB: xboard (Next-Board Go database)

USE xboard;

-- ====================
-- STEP 1: Migrate Labels (from server groups)
-- ====================
-- Convert Xboard's server groups to Next-Board labels

INSERT INTO labels (name, description, created_at, updated_at)
SELECT
    name,
    CONCAT('Migrated from server group ID: ', id),
    FROM_UNIXTIME(created_at),
    FROM_UNIXTIME(updated_at)
FROM xboard.v2_server_group
ON DUPLICATE KEY UPDATE name=name; -- Skip duplicates

-- Store mapping for later use
CREATE TEMPORARY TABLE IF NOT EXISTS group_to_label_map (
    old_group_id INT,
    new_label_id BIGINT UNSIGNED,
    PRIMARY KEY (old_group_id)
);

INSERT INTO group_to_label_map (old_group_id, new_label_id)
SELECT
    sg.id as old_group_id,
    l.id as new_label_id
FROM xboard.v2_server_group sg
JOIN labels l ON l.name = sg.name;

-- ====================
-- STEP 2: Migrate Plans
-- ====================
-- Convert Xboard plans to Next-Board plans with proper reset periods

INSERT INTO plans (name, quota_bytes, reset_period, base_multiplier, created_at, updated_at)
SELECT
    p.name,
    p.transfer_enable as quota_bytes,
    CASE
        WHEN p.reset_traffic_method = 0 THEN 'monthly'  -- Every 1st of month
        WHEN p.reset_traffic_method = 1 THEN 'monthly'  -- Monthly reset
        WHEN p.reset_traffic_method = 2 THEN 'none'     -- No reset
        WHEN p.reset_traffic_method = 3 THEN 'yearly'   -- Every Jan 1st
        WHEN p.reset_traffic_method = 4 THEN 'yearly'   -- Yearly reset
        ELSE 'monthly'
    END as reset_period,
    1.0 as base_multiplier,
    FROM_UNIXTIME(p.created_at),
    FROM_UNIXTIME(p.updated_at)
FROM xboard.v2_plan p
ON DUPLICATE KEY UPDATE name=name;

-- Create plan-label associations
INSERT INTO plan_labels (plan_id, label_id, created_at)
SELECT
    np.id as plan_id,
    gtl.new_label_id as label_id,
    NOW()
FROM xboard.v2_plan p
JOIN plans np ON np.name = p.name
JOIN group_to_label_map gtl ON gtl.old_group_id = p.group_id
ON DUPLICATE KEY UPDATE plan_id=plan_id;

-- ====================
-- STEP 3: Migrate Users
-- ====================
-- Convert Xboard users to Next-Board users
-- Note: Password hashes should be compatible if both use bcrypt
-- Includes: balance, commission, tokens, and login tracking

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
    u.email,
    u.password as password_hash,
    CASE
        WHEN u.is_admin = 1 THEN 'admin'
        ELSE 'user'
    END as role,
    np.id as plan_id,
    u.telegram_id as telegram_chat_id,
    CASE
        WHEN u.telegram_id IS NOT NULL THEN FROM_UNIXTIME(u.created_at)
        ELSE NULL
    END as telegram_linked_at,
    u.banned,
    u.balance,
    u.discount,
    u.commission_type,
    u.commission_rate,
    u.commission_balance,
    u.token,
    CASE
        WHEN u.last_login_at > 0 THEN FROM_UNIXTIME(u.last_login_at)
        ELSE NULL
    END as last_login_at,
    CASE
        WHEN u.last_login_ip > 0 THEN INET_NTOA(u.last_login_ip)
        ELSE NULL
    END as last_login_ip,
    u.remarks,
    FROM_UNIXTIME(u.created_at),
    FROM_UNIXTIME(u.updated_at)
FROM xboard.v2_user u
LEFT JOIN plans np ON np.id = u.plan_id
ON DUPLICATE KEY UPDATE email=email;

-- Migrate user UUIDs
INSERT INTO user_uuids (user_id, uuid, created_at)
SELECT
    nu.id as user_id,
    ou.uuid,
    NOW()
FROM xboard.v2_user ou
JOIN users nu ON nu.email = ou.email
ON DUPLICATE KEY UPDATE uuid=uuid;

-- ====================
-- STEP 4: Migrate Nodes
-- ====================
-- Note: Xboard has separate tables for each protocol type
-- Next-Board uses a unified nodes table with node_type column

-- Temporary table to store all nodes from different server types
CREATE TEMPORARY TABLE IF NOT EXISTS temp_all_nodes (
    old_id INT,
    name VARCHAR(255),
    node_type VARCHAR(50),
    host VARCHAR(255),
    port INT,
    protocol_config JSON,
    node_multiplier DECIMAL(10, 4),
    status VARCHAR(20),
    group_ids TEXT,
    old_created_at INT,
    old_updated_at INT
);

-- Migrate from v2_server table (if exists, this is the unified table in newer Xboard)
INSERT INTO temp_all_nodes (old_id, name, node_type, host, port, protocol_config, node_multiplier, status, group_ids, old_created_at, old_updated_at)
SELECT
    id,
    name,
    type as node_type,
    host,
    port,
    CASE
        WHEN type = 'vmess' THEN JSON_OBJECT('network', network, 'network_settings', network_settings, 'tls', tls)
        WHEN type = 'vless' THEN JSON_OBJECT('network', network, 'network_settings', network_settings, 'tls', tls, 'flow', flow)
        WHEN type = 'trojan' THEN JSON_OBJECT('network', network, 'network_settings', network_settings)
        WHEN type = 'shadowsocks' THEN JSON_OBJECT('cipher', cipher, 'server_key', server_key)
        WHEN type = 'hysteria' THEN JSON_OBJECT('up_mbps', up_mbps, 'down_mbps', down_mbps)
        ELSE JSON_OBJECT()
    END as protocol_config,
    CAST(rate as DECIMAL(10, 4)) as node_multiplier,
    CASE
        WHEN `show` = 1 THEN 'active'
        ELSE 'inactive'
    END as status,
    group_id as group_ids,
    created_at,
    updated_at
FROM xboard.v2_server
WHERE 1=1;  -- Only if v2_server table exists

-- Insert nodes into Next-Board
INSERT INTO nodes (name, node_type, host, port, protocol_config, node_multiplier, status, created_at, updated_at)
SELECT
    name,
    node_type,
    host,
    port,
    protocol_config,
    node_multiplier,
    status,
    FROM_UNIXTIME(old_created_at),
    FROM_UNIXTIME(old_updated_at)
FROM temp_all_nodes
ON DUPLICATE KEY UPDATE name=name;

-- Create node-label associations
-- This requires parsing the group_ids which might be JSON array or comma-separated
INSERT INTO node_labels (node_id, label_id, created_at)
SELECT DISTINCT
    n.id as node_id,
    gtl.new_label_id as label_id,
    NOW()
FROM temp_all_nodes tan
JOIN nodes n ON n.name = tan.name
JOIN group_to_label_map gtl ON FIND_IN_SET(gtl.old_group_id, REPLACE(REPLACE(tan.group_ids, '[', ''), ']', '')) > 0
ON DUPLICATE KEY UPDATE node_id=node_id;

-- ====================
-- STEP 5: Migrate Current Usage Data (Optional)
-- ====================
-- Create current usage periods for all users with plans

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
    ou.u as billable_bytes_up,  -- Initial approximation
    ou.d as billable_bytes_down, -- Initial approximation
    TRUE as is_current,
    NOW(),
    NOW()
FROM xboard.v2_user ou
JOIN users nu ON nu.email = ou.email
WHERE nu.plan_id IS NOT NULL
ON DUPLICATE KEY UPDATE user_id=user_id;

-- ====================
-- Cleanup
-- ====================
DROP TEMPORARY TABLE IF EXISTS group_to_label_map;
DROP TEMPORARY TABLE IF EXISTS temp_all_nodes;

-- ====================
-- Post-Migration Verification
-- ====================
SELECT 'Migration Summary:' as Status;
SELECT 'Labels:', COUNT(*) FROM labels;
SELECT 'Plans:', COUNT(*) FROM plans;
SELECT 'Users:', COUNT(*) FROM users;
SELECT 'Nodes:', COUNT(*) FROM nodes;
SELECT 'Usage Periods:', COUNT(*) FROM usage_periods;

-- ====================
-- Important Notes
-- ====================
-- 1. Review the migrated data carefully
-- 2. Test user login (passwords should work if both use bcrypt)
-- 3. Verify node configurations in protocol_config JSON
-- 4. Set up label multipliers if needed:
--    UPDATE plan_label_multipliers SET multiplier = X WHERE label_id = Y;
-- 5. Users will need to re-authenticate (tokens won't migrate)
-- 6. Historical traffic stats are NOT migrated (only current period)
