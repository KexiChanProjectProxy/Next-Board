-- Remove balance, commission, and token fields from users table

ALTER TABLE users
    DROP INDEX idx_last_login,
    DROP INDEX idx_token,
    DROP COLUMN remarks,
    DROP COLUMN last_login_ip,
    DROP COLUMN last_login_at,
    DROP COLUMN token,
    DROP COLUMN commission_balance,
    DROP COLUMN commission_rate,
    DROP COLUMN commission_type,
    DROP COLUMN discount,
    DROP COLUMN balance;
